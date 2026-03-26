<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { getAuthToken } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";

  // Anilist search form state
  let anilistQuery = $state("");

  // Batch Form State
  let batchYear = $state("");
  let batchSeason = $state("");
  let batchFormat = $state("");
  let isGenerating = $state(false);
  let isHydrating = $state(false);

  // Hydration Form State
  let atYear = $state(new Date().getFullYear().toString());
  let atSeason = $state("");
  let progressMessage = $state("");

  function handleAnilistSearch(e: Event) {
    e.preventDefault();
    if (!anilistQuery.trim()) return;
    goto(
      `/admin/animes/anilist-search?q=${encodeURIComponent(anilistQuery.trim())}`,
    );
  }

  async function handleBatchGenerate(e: Event) {
    e.preventDefault();
    if (!batchYear || !batchSeason || !batchFormat) return;

    isGenerating = true;
    try {
      await api.post("/admin/animes/batch", {
        year: parseInt(batchYear),
        season: batchSeason,
        format: batchFormat,
      });
      toastState.addToast("Batch import initiated successfully", "success");
      // Optional: goto("/admin/animes");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        `Failed to batch fetch animes: ${err.message || err}`,
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
      const response = await fetch(`${api.defaults.baseURL}/admin/animes/hydrate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${getAuthToken()}`, // Use standardized token getter
        },
        body: JSON.stringify({
          year: parseInt(atYear),
          season: atSeason,
        }),
      });

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
      toastState.addToast(`Hydration failed: ${err.message || err}`, "error");
    } finally {
      isHydrating = false;
      progressMessage = "";
    }
  }
</script>

<svelte:head>
  <title>Batch Import | Admin</title>
</svelte:head>

<!-- <div class="mb-8">
  <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
    AniList Autogen
  </h1>
  <p class="text-gray-400">
    Import animes in batch or search specifically on AniList to add them to the catalog.
  </p>
</div> -->

<!-- batch generate -->
<div class="mb-8">
  <form
    onsubmit={handleBatchGenerate}
    class="flex flex-wrap gap-4 items-center bg-white/5 p-6 rounded-2xl border border-white/10 shadow-xl"
  >
    <div class="w-full mb-2">
      <h2 class="text-lg font-semibold text-white flex items-center gap-2">
        <span class="material-symbols-outlined text-anirank-primary"
          >auto_fix_high</span
        >
        Batch Import by Season
      </h2>
      <p class="text-sm text-gray-400">
        Select parameters to fetch multiple animes at once.
      </p>
    </div>

    <div class="flex flex-wrap gap-4 w-full">
      <div class="flex-1 min-w-[200px]">
        <label
          for="batchYear"
          class="block text-xs font-bold uppercase tracking-wider text-gray-500 mb-2"
          >Year</label
        >
        <select
          id="batchYear"
          bind:value={batchYear}
          required
          class="w-full bg-black/50 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
        >
          <option value="">Select Year</option>
          {#each Array.from({ length: 77 }, (_, i) => new Date().getFullYear() + 1 - i) as year}
            <option value={year}>{year}</option>
          {/each}
        </select>
      </div>

      <div class="flex-1 min-w-[200px]">
        <label
          for="batchSeason"
          class="block text-xs font-bold uppercase tracking-wider text-gray-500 mb-2"
          >Season</label
        >
        <select
          id="batchSeason"
          bind:value={batchSeason}
          required
          class="w-full bg-black/50 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
        >
          <option value="">Select Season</option>
          <option value="WINTER">Winter</option>
          <option value="SPRING">Spring</option>
          <option value="SUMMER">Summer</option>
          <option value="FALL">Fall</option>
        </select>
      </div>

      <div class="flex-1 min-w-[200px]">
        <label
          for="batchFormat"
          class="block text-xs font-bold uppercase tracking-wider text-gray-500 mb-2"
          >Format</label
        >
        <select
          id="batchFormat"
          bind:value={batchFormat}
          required
          class="w-full bg-black/50 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
        >
          <option value="">Select Format</option>
          <option value="TV">TV</option>
          <option value="TV_SHORT">TV Short</option>
          <option value="MOVIE">Movie</option>
          <option value="SPECIAL">Special</option>
          <option value="OVA">OVA</option>
          <option value="ONA">ONA</option>
          <option value="MUSIC">Music</option>
          <option value="MANGA">Manga</option>
          <option value="NOVEL">Novel</option>
          <option value="ONE_SHOT">One Shot</option>
        </select>
      </div>
    </div>

    <div class="w-full pt-4 flex justify-end">
      <button
        type="submit"
        disabled={isGenerating}
        class="bg-anirank-primary hover:bg-blue-600 disabled:bg-gray-600 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 px-8 py-3 rounded-xl text-white font-bold transition-all hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-anirank-primary/20"
      >
        {#if isGenerating}
          <svg
            class="animate-spin h-5 w-5 text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
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
          Fetching & Saving...
        {:else}
          <span class="material-symbols-outlined">download</span>
          Fetch from AniList
        {/if}
      </button>
    </div>
  </form>
</div>

<!-- Search on Anilist -->
<div class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-xl">
  <div class="mb-4">
    <h2 class="text-lg font-semibold text-white flex items-center gap-2">
      <span class="material-symbols-outlined text-anirank-primary">search</span>
      Search & Import from AniList
    </h2>
    <p class="text-sm text-gray-400">
      Search for a specific anime by title to import it immediately.
    </p>
  </div>

  <form onsubmit={handleAnilistSearch} class="flex gap-3">
    <input
      type="text"
      bind:value={anilistQuery}
      placeholder="Search anime title on AniList..."
      aria-label="Search anime title on AniList"
      class="flex-1 bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-anirank-primary transition-colors"
    />
    <button
      type="submit"
      class="px-8 py-3 bg-anirank-primary hover:bg-anirank-primary/80 text-white font-bold rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] flex items-center gap-2 shadow-lg shadow-anirank-primary/20"
    >
      <span class="material-symbols-outlined">search</span>
      Search AniList
    </button>
  </form>
</div>

<!-- fetch from animethemes -->
<div class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-xl">
  <div class="mb-6">
    <h2 class="text-lg font-semibold text-white flex items-center gap-2">
      <span class="material-symbols-outlined text-anirank-primary"
        >cloud_download</span
      >
      Import from AnimeThemes (Hydrate)
    </h2>
    <p class="text-sm text-gray-400">
      Fetch music data from AnimeThemes and enrich it with AniList metadata.
    </p>
  </div>

  <form onsubmit={handleATHydration} class="flex flex-wrap gap-4 items-end">
    <div class="flex-1 min-w-[150px]">
      <label
        for="atYear"
        class="block text-xs font-bold uppercase tracking-wider text-gray-500 mb-2"
        >Year</label
      >
      <select
        id="atYear"
        bind:value={atYear}
        required
        class="w-full bg-black/50 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
      >
        <option value="">Select Year</option>
        {#each Array.from({ length: 77 }, (_, i) => new Date().getFullYear() + 1 - i) as year}
          <option value={year}>{year}</option>
        {/each}
      </select>
    </div>

    <div class="flex-1 min-w-[150px]">
      <label
        for="atSeason"
        class="block text-xs font-bold uppercase tracking-wider text-gray-500 mb-2"
        >Season</label
      >
      <select
        id="atSeason"
        bind:value={atSeason}
        required
        class="w-full bg-black/50 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
      >
        <option value="">Select Season</option>
        <option value="WINTER">Winter</option>
        <option value="SPRING">Spring</option>
        <option value="SUMMER">Summer</option>
        <option value="FALL">Fall</option>
      </select>
    </div>

    <button
      type="submit"
      disabled={isHydrating}
      class="bg-anirank-primary hover:bg-blue-600 disabled:bg-gray-600 cursor-pointer disabled:opacity-50 h-[46px] px-8 rounded-xl text-white font-bold transition-all flex items-center gap-2 shadow-lg shadow-anirank-primary/20"
    >
      {#if isHydrating}
        <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
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
        Hydrating...
      {:else}
        <span class="material-symbols-outlined">refresh</span>
        Start Hydration
      {/if}
    </button>
  </form>

  {#if isHydrating}
    <div class="mt-6 p-4 bg-anirank-primary/10 border border-anirank-primary/20 rounded-xl animate-in fade-in slide-in-from-top-2">
      <div class="flex items-center gap-3">
        <div class="shrink-0">
          <svg class="animate-spin h-5 w-5 text-anirank-primary" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-white truncate">
            {progressMessage}
          </p>
        </div>
      </div>
      <!-- Simple CSS progress bar (pulse) -->
      <div class="mt-3 w-full bg-white/5 rounded-full h-1.5 overflow-hidden">
        <div class="bg-anirank-primary h-full rounded-full animate-pulse w-full"></div>
      </div>
    </div>
  {/if}
</div>
