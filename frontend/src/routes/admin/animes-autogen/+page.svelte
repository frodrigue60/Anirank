<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { getAuthToken } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import Wand2 from "lucide-svelte/icons/wand-2";
  import Download from "lucide-svelte/icons/download";
  import Search from "lucide-svelte/icons/search";
  import CloudDownload from "lucide-svelte/icons/cloud-download";
  import X from "lucide-svelte/icons/x";
  import Check from "lucide-svelte/icons/check";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import Image from "lucide-svelte/icons/image";
  import SearchCode from "lucide-svelte/icons/search-code";

  // Anilist search form state
  let anilistQuery = $state("");

  // Batch Form State
  let batchYear = $state(new Date().getFullYear());
  let batchSeason = $state("");
  let batchFormat = $state("");
  let isGenerating = $state(false);
  let isHydrating = $state(false);

  // Hydration Form State
  let atYear = $state(new Date().getFullYear());
  let atSeason = $state("");
  let atLanguage = $state("ja");
  let progressMessage = $state("");

  // AnimeThemes Hydration State
  let animeThemesQuery = $state("");
  let animeThemesResults = $state<any[]>([]);
  let selectedAnimeThemesIDs = $state<Set<number>>(new Set());
  let isSearchingAnimeThemes = $state(false);
  let showResultsModal = $state(false);

  // Anilist Search State
  let anilistResults = $state<any[]>([]);
  let selectedAnilistIDs = $state<Set<number>>(new Set());
  let isSearchingAnilist = $state(false);
  let showAnilistResultsModal = $state(false);
  let isImportingAnilist = $state(false);

  import { untrack } from "svelte";

  // API Status State
  let apiStatus = $state<{
    anilist: { status: "loading" | "online" | "offline"; message?: string };
    animethemes: { status: "loading" | "online" | "offline"; message?: string };
  }>({
    anilist: { status: "loading" },
    animethemes: { status: "loading" },
  });

  let isCheckingStatus = false;

  async function fetchApiStatus() {
    if (isCheckingStatus) return;
    isCheckingStatus = true;

    apiStatus.anilist.status = "loading";
    apiStatus.animethemes.status = "loading";

    try {
      const resp = await api.get("/admin/system/api-status");
      const data = resp.data;
      apiStatus.anilist = data.anilist;
      apiStatus.animethemes = data.animethemes;
    } catch (err: any) {
      apiStatus.anilist = { status: "offline", message: "Connection Error" };
      apiStatus.animethemes = {
        status: "offline",
        message: "Connection Error",
      };
    } finally {
      isCheckingStatus = false;
    }
  }

  $effect(() => {
    untrack(() => {
      fetchApiStatus();
    });
  });

  async function handleAnimeThemesSearch(e: Event) {
    e.preventDefault();
    if (!animeThemesQuery.trim()) return;

    isSearchingAnimeThemes = true;
    try {
      const resp = await api.get(
        `/admin/animes/animethemes/search?q=${encodeURIComponent(animeThemesQuery.trim())}`,
      );
      animeThemesResults = resp.data.data || [];
      selectedAnimeThemesIDs = new Set();
      showResultsModal = true;
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to search AnimeThemes"),
        "error",
      );
    } finally {
      isSearchingAnimeThemes = false;
    }
  }

  function toggleAnimeSelection(id: number) {
    if (selectedAnimeThemesIDs.has(id)) {
      selectedAnimeThemesIDs.delete(id);
    } else {
      selectedAnimeThemesIDs.add(id);
    }
    // Trigger reactivity for Set in Svelte 5
    selectedAnimeThemesIDs = new Set(selectedAnimeThemesIDs);
  }

  async function handleImportSelected() {
    if (selectedAnimeThemesIDs.size === 0) return;

    showResultsModal = false;
    isHydrating = true;
    progressMessage = "Starting targeted hydration...";

    try {
      const response = await fetch(
        `${api.defaults.baseURL}/admin/animes/animethemes/hydrate`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            ...(getAuthToken() ? { "Authorization": `Bearer ${getAuthToken()}` } : {}),
            "X-CSRF-Token": typeof document !== 'undefined' && document.cookie.includes('csrf_token=') ? `; ${document.cookie}`.split(`; csrf_token=`)[1].split(';')[0] : '',
          },
          body: JSON.stringify({
            ids: Array.from(selectedAnimeThemesIDs),
            language: atLanguage,
          }),
        },
      );

      if (!response.ok) throw new Error("Import failed");

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const messages = chunk.split("\n\n");

          for (const msg of messages) {
            if (msg.startsWith("data: ")) {
              progressMessage = msg.replace("data: ", "");
            }
          }
        }
      }

      toastState.addToast(
        `Imported ${selectedAnimeThemesIDs.size} animes successfully`,
        "success",
      );
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Import failed"), "error");
    } finally {
      isHydrating = false;
      progressMessage = "";
    }
  }

  async function handleAnilistSearch(e: Event) {
    e.preventDefault();
    if (!anilistQuery.trim()) return;

    isSearchingAnilist = true;
    try {
      const resp = await api.get(
        `/admin/animes/anilist-search?q=${encodeURIComponent(anilistQuery.trim())}`,
      );
      anilistResults = resp.data.data || [];
      selectedAnilistIDs = new Set();
      showAnilistResultsModal = true;
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to search AniList"),
        "error",
      );
    } finally {
      isSearchingAnilist = false;
    }
  }

  function toggleAnilistSelection(id: number) {
    if (selectedAnilistIDs.has(id)) {
      selectedAnilistIDs.delete(id);
    } else {
      selectedAnilistIDs.add(id);
    }
    selectedAnilistIDs = new Set(selectedAnilistIDs);
  }

  async function handleAnilistImportSelected() {
    if (selectedAnilistIDs.size === 0) return;

    isImportingAnilist = true;
    try {
      const resp = await api.post("/admin/animes/batch-from-anilist", {
        anilist_ids: Array.from(selectedAnilistIDs),
      });

      const result = resp.data.data;
      if (result.imported > 0) {
        toastState.addToast(
          `Imported ${result.imported} animes successfully`,
          "success",
        );
      }
      if (result.failed > 0) {
        toastState.addToast(
          `Failed to import ${result.failed} animes`,
          "warning",
        );
      }
      showAnilistResultsModal = false;
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Import failed"), "error");
    } finally {
      isImportingAnilist = false;
    }
  }

  async function handleBatchGenerate(e: Event) {
    e.preventDefault();
    if (!batchYear || !batchSeason || !batchFormat) return;

    isGenerating = true;
    try {
      await api.post("/admin/animes/batch", {
        year: batchYear,
        season: batchSeason,
        format: batchFormat,
      });
      toastState.addToast("Batch import initiated successfully", "success");
      // Optional: goto("/admin/animes");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "Failed to batch fetch animes"),
        "error",
      );
    } finally {
      isGenerating = false;
    }
  }

  async function handleATHydration(e: Event) {
    e.preventDefault();
    if (!atYear || !atSeason) return;

    isHydrating = true;
    progressMessage = "Connecting to hydration stream...";
    try {
      const response = await fetch(
        `${api.defaults.baseURL}/admin/animes/hydrate`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            ...(getAuthToken() ? { "Authorization": `Bearer ${getAuthToken()}` } : {}),
            "X-CSRF-Token": typeof document !== 'undefined' && document.cookie.includes('csrf_token=') ? `; ${document.cookie}`.split(`; csrf_token=`)[1].split(';')[0] : '',
          },
          body: JSON.stringify({
            year: atYear,
            season: atSeason,
            language: atLanguage,
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const messages = chunk.split("\n\n");

          for (const msg of messages) {
            if (msg.startsWith("data: ")) {
              progressMessage = msg.replace("data: ", "");
            }
          }
        }
      }

      toastState.addToast(
        `Hydration for ${atSeason} ${atYear} completed`,
        "success",
      );
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Hydration failed"), "error");
    } finally {
      isHydrating = false;
      progressMessage = "";
    }
  }
</script>

<svelte:head>
  <title>Batch Import | Admin</title>
</svelte:head>

<div class="space-y-8">
  <div
    class="bg-surface-container border border-outline-variant rounded-3xl p-4 flex flex-wrap items-center gap-6 shadow-xl"
  >
    <div class="flex flex-col">
      <span class="text-[10px] font-bold uppercase text-on-surface-variant/40 tracking-wider"
        >External Services</span
      >
      <h2 class="text-on-surface font-bold">API Health Check</h2>
    </div>

    <div class="flex flex-wrap items-center gap-3">
      <!-- Anilist Status -->
      <div
        class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/40 border border-outline-variant"
      >
        <span class="text-xs font-bold text-on-surface-variant/70">AniList</span>
        {#if apiStatus.anilist.status === "loading"}
          <div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></div>
        {:else if apiStatus.anilist.status === "online"}
          <div
            class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"
          ></div>
        {:else}
          <div
            class="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)]"
          ></div>
        {/if}
        <span
          class="text-[10px] font-bold uppercase {apiStatus.anilist.status ===
          'online'
            ? 'text-emerald-500'
            : apiStatus.anilist.status === 'offline'
              ? 'text-rose-500'
              : 'text-amber-500'}"
        >
          {apiStatus.anilist.status}
        </span>
      </div>

      <!-- AnimeThemes Status -->
      <div
        class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/40 border border-outline-variant"
      >
        <span class="text-xs font-bold text-on-surface-variant/70">AnimeThemes</span>
        {#if apiStatus.animethemes.status === "loading"}
          <div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></div>
        {:else if apiStatus.animethemes.status === "online"}
          <div
            class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"
          ></div>
        {:else}
          <div
            class="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_8_px_rgba(244,63,94,0.5)]"
          ></div>
        {/if}
        <span
          class="text-[10px] font-bold uppercase {apiStatus.animethemes
            .status === 'online'
            ? 'text-emerald-500'
            : apiStatus.animethemes.status === 'offline'
              ? 'text-rose-500'
              : 'text-amber-500'}"
        >
          {apiStatus.animethemes.status}
        </span>
      </div>

      <button
        onclick={fetchApiStatus}
        class="p-2 hover:bg-surface-highest rounded-full transition-colors text-on-surface-variant/70 hover:text-on-surface"
        title="Check status now"
      >
        <RefreshCw size={14} class={apiStatus.anilist.status === 'loading' || apiStatus.animethemes.status === 'loading' ? 'animate-spin' : ''} />

      </button>
    </div>

    {#if apiStatus.anilist.status === "offline" || apiStatus.animethemes.status === "offline"}
      <div
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-rose-500/10 border border-rose-500/20 animate-in fade-in slide-in-from-left-2"
      >
        <AlertTriangle size={16} class="text-rose-500" />
        <p class="text-[11px] text-rose-200 font-medium">
          Partial outages detected. Generation may be incomplete.
        </p>
      </div>
    {/if}
  </div>
  <div class="grid grid-cols-1 xl:grid-cols-2 gap-8 items-start">
    <!-- Section: AniList Tools -->
    <section class="space-y-6">
      <div class="px-1">
        <h2 class="text-xl font-bold text-on-surface leading-tight">
          AniList Integration
        </h2>
        <p
          class="text-xs text-on-surface-variant/40 uppercase tracking-widest font-bold mt-1"
        >
          Metadata & Resource Discovery
        </p>
      </div>

      <!-- Batch Import (AniList) -->
      <div
        class="bg-surface-container border border-outline-variant rounded-3xl p-6 shadow-2xl relative overflow-hidden group"
      >
        <div
          class="absolute top-0 right-0 p-8 opacity-5 -mr-4 -mt-4 group-hover:scale-110 transition-transform"
        >
          <Wand2 size={96} class="opacity-5" />
        </div>

        <div class="mb-6 relative">
          <h3 class="text-lg font-bold text-on-surface flex items-center gap-2">
            Batch Seasonal Import
          </h3>
          <p class="text-sm text-on-surface-variant/70">
            Fetch series data for a specific year and season.
          </p>
        </div>

        <form onsubmit={handleBatchGenerate} class="space-y-6 relative">
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="space-y-2">
              <label
                for="batchYear"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Year</label
              >
              <select
                id="batchYear"
                bind:value={batchYear}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all shadow-inner"
              >
                <option value={0}>Year</option>
                {#each Array.from({ length: 77 }, (_, i) => new Date().getFullYear() + 1 - i) as year}
                  <option value={year}>{year}</option>
                {/each}
              </select>
            </div>

            <div class="space-y-2">
              <label
                for="batchSeason"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Season</label
              >
              <select
                id="batchSeason"
                bind:value={batchSeason}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all shadow-inner"
              >
                <option value="">Season</option>
                <option value="WINTER">Winter</option>
                <option value="SPRING">Spring</option>
                <option value="SUMMER">Summer</option>
                <option value="FALL">Fall</option>
              </select>
            </div>

            <div class="space-y-2">
              <label
                for="batchFormat"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Format</label
              >
              <select
                id="batchFormat"
                bind:value={batchFormat}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all shadow-inner"
              >
                <option value="">Format</option>
                <option value="TV">TV</option>
                <option value="TV_SHORT">Short</option>
                <option value="MOVIE">Movie</option>
                <option value="OVA">OVA</option>
                <option value="ONA">ONA</option>
                <option value="SPECIAL">Special</option>
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={isGenerating || apiStatus.anilist.status !== "online"}
            class="w-full bg-blue-600 hover:bg-primary-container/80 disabled:bg-gray-700 py-3 rounded-xl text-on-surface font-bold transition-all flex items-center justify-center gap-2 shadow-lg shadow-anirank-primary/20"
          >
            {#if isGenerating}
              <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24"
                ><circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle><path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path></svg
              >
              Processing Batch...
            {:else}
              <Download size={20} />
              Fetch from AniList
            {/if}
          </button>
        </form>
      </div>

      <!-- Manual Search (AniList) -->
      <div
        class="bg-surface-container border border-outline-variant rounded-3xl p-6 shadow-2xl"
      >
        <div class="mb-4">
          <h3 class="text-lg font-bold text-on-surface flex items-center gap-2">
            Focused AniList Search
          </h3>
          <p class="text-sm text-on-surface-variant/70">
            Import a specific title by searching.
          </p>
        </div>

        <form onsubmit={handleAnilistSearch} class="flex gap-3">
          <input
            type="text"
            required
            bind:value={anilistQuery}
            placeholder="Series title on AniList..."
            class="flex-1 bg-black/60 border border-outline-variant rounded-xl px-4 py-3 text-on-surface placeholder-white/20 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
          />
          <button
            type="submit"
            disabled={isSearchingAnilist ||
              apiStatus.anilist.status !== "online"}
            class="px-6 bg-blue-600 hover:bg-primary-container/80 disabled:bg-gray-700 rounded-xl text-on-surface font-bold transition-all flex items-center gap-2 shadow-lg shadow-anirank-primary/20"
          >
            {#if isSearchingAnilist}
              <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24"
                ><circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle><path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path></svg
              >
            {:else}
              <Search size={20} />
            {/if}
          </button>
        </form>
      </div>
    </section>

    <!-- Section: AnimeThemes Tools -->
    <section class="space-y-6">
      <div class="px-1">
        <h2 class="text-xl font-bold text-on-surface leading-tight">
          AnimeThemes Integration
        </h2>
        <p
          class="text-xs text-on-surface-variant/40 uppercase tracking-widest font-bold mt-1"
        >
          Music Hydration & Synchro
        </p>
      </div>

      <!-- Hydration (AnimeThemes) -->
      <div
        class="bg-surface-container border border-outline-variant rounded-3xl p-6 shadow-2xl relative overflow-hidden group"
      >
        <div
          class="absolute top-0 right-0 p-8 opacity-5 -mr-4 -mt-4 group-hover:scale-110 transition-transform"
        >
          <RefreshCw size={96} class="opacity-5 text-purple-400" />
        </div>

        <div class="mb-6 relative">
          <h3 class="text-lg font-bold text-on-surface">Seasonal Music Hydration</h3>
          <p class="text-sm text-on-surface-variant/70">
            Enrich existing animes with songs and variants.
          </p>
        </div>

        <form onsubmit={handleATHydration} class="space-y-6 relative">
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="space-y-2">
              <label
                for="atYear"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Year</label
              >
              <select
                id="atYear"
                bind:value={atYear}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-purple-500 transition-all shadow-inner"
              >
                <option value={0}>Year</option>
                {#each Array.from({ length: 77 }, (_, i) => new Date().getFullYear() + 1 - i) as year}
                  <option value={year}>{year}</option>
                {/each}
              </select>
            </div>

            <div class="space-y-2">
              <label
                for="atSeason"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Season</label
              >
              <select
                id="atSeason"
                bind:value={atSeason}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-purple-500 transition-all shadow-inner"
              >
                <option value="">Season</option>
                <option value="WINTER">Winter</option>
                <option value="SPRING">Spring</option>
                <option value="SUMMER">Summer</option>
                <option value="FALL">Fall</option>
              </select>
            </div>

            <div class="space-y-2">
              <label
                for="atLanguage"
                class="text-[10px] font-bold uppercase text-on-surface-variant/70 ml-1"
                >Language Filter</label
              >
              <select
                id="atLanguage"
                bind:value={atLanguage}
                required
                class="w-full bg-black/60 border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-purple-500 transition-all shadow-inner"
              >
                <option value="all">All (Multi-lang)</option>
                <option value="ja">Japanese Only (Original)</option>
                <option value="en">English Only</option>
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={isHydrating || apiStatus.animethemes.status !== "online" || apiStatus.anilist.status !== "online"}
            class="w-full bg-purple-600 hover:bg-purple-600/80 disabled:bg-gray-700 py-3 rounded-xl text-on-surface font-bold transition-all flex items-center justify-center gap-2"
          >
            {#if isHydrating}
              <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24"
                ><circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle><path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path></svg
              >
              Hydrating Stream...
            {:else}
              <RefreshCw size={20} />
              Start Hydration
            {/if}
          </button>
        </form>

        {#if isHydrating}
          <div
            class="mt-6 p-4 bg-purple-500/10 border border-purple-500/20 rounded-xl animate-in fade-in slide-in-from-top-2 relative"
          >
            <div class="flex items-center gap-3">
              <div class="shrink-0">
                <svg
                  class="animate-spin h-4 w-4 text-purple-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  ><circle
                    class="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    stroke-width="4"
                  ></circle><path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path></svg
                >
              </div>
              <p class="text-sm font-medium text-on-surface truncate">
                {progressMessage}
              </p>
            </div>
            <div
              class="mt-3 w-full bg-surface-highest rounded-full h-1 overflow-hidden"
            >
              <div class="bg-purple-500 h-full w-full animate-pulse"></div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Manual Search (AnimeThemes) -->
      <div
        class="bg-surface-container border border-outline-variant rounded-3xl p-6 shadow-2xl"
      >
        <div class="mb-4">
          <h3 class="text-lg font-bold text-on-surface flex items-center gap-2">
            Focused AnimeThemes Search
          </h3>
          <p class="text-sm text-on-surface-variant/70">
            Import a specific title by searching on AnimeThemes.
          </p>
        </div>

        <form onsubmit={handleAnimeThemesSearch} class="flex gap-3">
          <input
            type="text"
            bind:value={animeThemesQuery}
            placeholder="Series title on AnimeThemes..."
            class="flex-1 bg-black/60 border border-outline-variant rounded-xl px-4 py-3 text-on-surface placeholder-white/20 focus:outline-none focus:border-purple-500 transition-all font-medium"
          />
          <button
            type="submit"
            disabled={isSearchingAnimeThemes ||
              apiStatus.animethemes.status !== "online" ||
              apiStatus.anilist.status !== "online"}
            class="px-6 bg-purple-600 hover:bg-purple-600/80 disabled:bg-gray-700 rounded-xl text-on-surface font-bold transition-all flex items-center gap-2"
          >
            {#if isSearchingAnimeThemes}
              <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24"
                ><circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle><path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path></svg
              >
            {:else}
              <Search size={20} />
            {/if}
          </button>
        </form>
      </div>
    </section>
  </div>
</div>

{#if showResultsModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 animate-in fade-in"
  >
    <div
      class="bg-black border border-outline-variant rounded-3xl w-full max-w-4xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden"
    >
      <!-- Header -->
      <div
        class="p-6 border-b border-outline-variant flex justify-between items-center bg-surface-highest"
      >
        <div>
          <h3 class="text-xl font-bold text-on-surface flex items-center gap-2">
            <CloudDownload size={24} class="text-primary" />
            AnimeThemes Results
          </h3>
          <p class="text-sm text-on-surface-variant/70">
            Showing results for: <span class="text-on-surface font-medium italic"
              >"{animeThemesQuery}"</span
            >
          </p>
        </div>

        <div class="flex items-center gap-4">
          <div class="flex flex-col items-end">
            <label for="modalLanguage" class="text-[10px] font-bold uppercase text-on-surface-variant/40 tracking-wider">Hydration Filter</label>
            <select
              id="modalLanguage"
              bind:value={atLanguage}
              class="bg-surface-container border border-outline-variant rounded-lg py-1 px-3 text-xs text-on-surface focus:outline-none focus:border-primary transition-all"
            >
              <option value="all">All (Multi-lang)</option>
              <option value="ja">Japanese Only</option>
              <option value="en">English Only</option>
            </select>
          </div>
          
          <button
            onclick={() => (showResultsModal = false)}
            class="p-2 hover:bg-surface-highest rounded-full transition-colors text-on-surface-variant/70"
          >
            <X size={20} />
          </button>
        </div>
      </div>

      <!-- Results List -->
      <div class="flex-1 overflow-y-auto p-6">
        {#if animeThemesResults.length === 0}
          <div class="text-center py-12">
            <SearchCode size={48} class="text-gray-600 mb-2 mx-auto" />
            <p class="text-on-surface-variant/70">No results found on AnimeThemes.</p>
          </div>
        {:else}
          <div
            class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
          >
            {#each animeThemesResults as anime}
              <button
                onclick={() => toggleAnimeSelection(anime.id)}
                class="flex flex-col gap-3 p-3 rounded-2xl border transition-all text-left group/card relative {selectedAnimeThemesIDs.has(
                  anime.id,
                )
                  ? 'bg-primary/20 border-primary shadow-lg shadow-anirank-primary/10'
                  : 'bg-surface-highest border-outline-variant hover:border-outline-variant'}"
              >
                <div
                  class="relative w-full aspect-2/3 shrink-0 rounded-xl overflow-hidden bg-black/40 shadow-inner"
                >
                  {#if anime.images?.find((i: any) => i.facet === "Large Cover" || i.facet === "Small Cover")}
                    <img
                      src={anime.images.find(
                        (i: any) =>
                          i.facet === "Large Cover" ||
                          i.facet === "Small Cover",
                      )?.link}
                      alt={anime.name}
                      title={anime.name}
                      class="w-full h-full object-cover group-hover/card:scale-110 transition-transform duration-500"
                    />
                  {:else}
                    <div class="w-full h-full flex items-center justify-center">
                      <Image size={24} class="text-gray-600" />
                    </div>
                  {/if}

                  {#if selectedAnimeThemesIDs.has(anime.id)}
                    <div
                      class="absolute inset-0 bg-primary/40 flex items-center justify-center"
                    >
                      <CheckCircle2 size={32} class="text-on-surface font-bold" />
                    </div>
                  {/if}
                </div>

                <div class="flex-1 min-w-0 px-1">
                  <h4
                    class="font-bold text-on-surface truncate text-sm"
                    title={anime.name}
                  >
                    {anime.name}
                  </h4>
                  <div class="flex flex-col mt-1">
                    <p
                      class="text-[10px] text-on-surface-variant/70 uppercase tracking-widest font-bold"
                    >
                      {anime.season}
                      {anime.year}
                    </p>
                    <p class="text-[10px] text-on-surface-variant/40 font-medium mt-0.5">
                      {anime.media_format || "TV"}
                    </p>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div
        class="p-6 border-t border-outline-variant flex justify-between items-center bg-surface-highest"
      >
        <p class="text-sm text-on-surface-variant/70">
          <span class="text-on-surface font-bold">{selectedAnimeThemesIDs.size}</span
          > selected
        </p>
        <div class="flex gap-3">
          <button
            onclick={() => (showResultsModal = false)}
            class="px-6 py-2.5 rounded-xl text-on-surface font-medium hover:bg-surface-highest transition-colors"
          >
            Cancel
          </button>
          <button
            onclick={handleImportSelected}
            disabled={selectedAnimeThemesIDs.size === 0 || apiStatus.anilist.status !== "online"}
            class="px-8 py-2.5 bg-primary hover:bg-primary-container disabled:bg-gray-700 text-on-surface font-bold rounded-xl transition-all shadow-lg shadow-anirank-primary/20"
          >
            Import Selected
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showAnilistResultsModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 animate-in fade-in"
  >
    <div
      class="bg-black border border-outline-variant rounded-3xl w-full max-w-4xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden"
    >
      <!-- Header -->
      <div
        class="p-6 border-b border-outline-variant flex justify-between items-center bg-surface-highest"
      >
        <div>
          <h3 class="text-xl font-bold text-on-surface flex items-center gap-2">
            <Search size={24} class="text-primary" />
            AniList Results
          </h3>
          <p class="text-sm text-on-surface-variant/70">
            Showing results for: <span class="text-on-surface font-medium italic"
              >"{anilistQuery}"</span
            >
          </p>
        </div>
        <button
          onclick={() => (showAnilistResultsModal = false)}
          class="p-2 hover:bg-surface-highest rounded-full transition-colors text-on-surface-variant/70"
        >
          <X size={20} />
        </button>
      </div>

      <!-- Results List -->
      <div class="flex-1 overflow-y-auto p-6">
        {#if anilistResults.length === 0}
          <div class="text-center py-12">
            <SearchCode size={48} class="text-gray-600 mb-2 mx-auto" />
            <p class="text-on-surface-variant/70">No results found on AniList.</p>
          </div>
        {:else}
          <div
            class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
          >
            {#each anilistResults as anime}
              <button
                onclick={() => toggleAnilistSelection(anime.id)}
                class="flex flex-col gap-3 p-3 rounded-2xl border transition-all text-left group/card relative {selectedAnilistIDs.has(
                  anime.id,
                )
                  ? 'bg-primary/20 border-primary shadow-lg shadow-anirank-primary/10'
                  : 'bg-surface-highest border-outline-variant hover:border-outline-variant'}"
              >
                <div
                  class="relative w-full aspect-2/3 shrink-0 rounded-xl overflow-hidden bg-black/40 shadow-inner"
                >
                  {#if anime.coverImage?.large}
                    <img
                      src={anime.coverImage.large}
                      alt={anime.title.romaji}
                      title={anime.title.romaji}
                      class="w-full h-full object-cover group-hover/card:scale-110 transition-transform duration-500"
                    />
                  {:else}
                    <div class="w-full h-full flex items-center justify-center">
                      <Image size={24} class="text-gray-600" />
                    </div>
                  {/if}

                  {#if selectedAnilistIDs.has(anime.id)}
                    <div
                      class="absolute inset-0 bg-primary/40 flex items-center justify-center"
                    >
                      <CheckCircle2 size={32} class="text-on-surface font-bold" />
                    </div>
                  {/if}
                </div>

                <div class="flex-1 min-w-0 px-1">
                  <h4
                    class="font-bold text-on-surface truncate text-sm"
                    title={anime.title.romaji}
                  >
                    {anime.title.romaji || anime.title.english}
                  </h4>
                  <div class="flex flex-col mt-1">
                    <p
                      class="text-[10px] text-on-surface-variant/70 uppercase tracking-widest font-bold"
                    >
                      {anime.season || "N/A"}
                      {anime.seasonYear || ""}
                    </p>
                    <p class="text-[10px] text-on-surface-variant/40 font-medium mt-0.5">
                      {anime.format || "TV"}
                    </p>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div
        class="p-6 border-t border-outline-variant flex justify-between items-center bg-surface-highest"
      >
        <p class="text-sm text-on-surface-variant/70">
          <span class="text-on-surface font-bold">{selectedAnilistIDs.size}</span> selected
        </p>
        <div class="flex gap-3">
          <button
            onclick={() => (showAnilistResultsModal = false)}
            class="px-6 py-2.5 rounded-xl text-on-surface font-medium hover:bg-surface-highest transition-colors"
          >
            Cancel
          </button>
          <button
            onclick={handleAnilistImportSelected}
            disabled={selectedAnilistIDs.size === 0 || isImportingAnilist}
            class="px-8 py-2.5 bg-primary hover:bg-primary-container disabled:bg-gray-700 text-on-surface font-bold rounded-xl transition-all shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
          >
            {#if isImportingAnilist}
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              Importing...
            {:else}
              Import Selected
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
