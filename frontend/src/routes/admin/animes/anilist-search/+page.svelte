<script lang="ts">
  import { goto } from "$app/navigation";
  import { fly } from "svelte/transition";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import type { PageData } from "./$types";
  import { toastState } from "$lib/state/toast.svelte";

  interface AnilistBatchImportError {
    anilist_id: number;
    message: string;
  }

  interface AnilistBatchImportData {
    requested: number;
    imported: number;
    failed: number;
    imported_ids: number[];
    errors: AnilistBatchImportError[];
  }

  let { data }: { data: PageData } = $props();

  interface Media {
    id: number;
    title: { romaji: string; english?: string; native?: string };
    coverImage: { extraLarge: string; large: string };
    bannerImage: string;
    season: string;
    seasonYear: number;
    format: string;
    status: string;
    description: string;
  }

  let results = $derived(data.results as Media[]);

  let searchInput = $state("");
  let formatInput = $state("");

  $effect(() => {
    searchInput = data.q;
    formatInput = data.format;
  });

  let batchSaving = $state(false);
  let savedIds = $state<Set<number>>(new Set());
  let selectedIds = $state<Set<number>>(new Set());

  const formats = [
    { value: "", label: "All Formats" },
    { value: "TV", label: "TV" },
    { value: "TV_SHORT", label: "TV Short" },
    { value: "MOVIE", label: "Movie" },
    { value: "SPECIAL", label: "Special" },
    { value: "OVA", label: "OVA" },
    { value: "ONA", label: "ONA" },
    { value: "MUSIC", label: "Music" },
  ];

  function handleSearch(e: Event) {
    e.preventDefault();
    if (!searchInput.trim()) return;
    const params = new URLSearchParams();
    params.set("q", searchInput.trim());
    if (formatInput) params.set("format", formatInput);

    goto(`/admin/animes/anilist-search?${params.toString()}`);
  }

  function toggleSelect(id: number) {
    if (batchSaving) return;
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
    selectedIds = new Set(selectedIds);
  }

  function toggleSelectAll() {
    if (batchSaving) return;
    const allImportable = results.filter((m: Media) => !savedIds.has(m.id));
    if (selectedIds.size === allImportable.length && allImportable.length > 0) {
      selectedIds = new Set();
    } else {
      selectedIds = new Set(allImportable.map((m: Media) => m.id));
    }
  }

  function formatBatchErrors(errors: AnilistBatchImportError[] | undefined) {
    if (!errors?.length) return "";
    const parts = errors
      .slice(0, 3)
      .map((e) => `#${e.anilist_id}: ${e.message || "unknown error"}`);
    const suffix = errors.length > 3 ? ` (+${errors.length - 3} more)` : "";
    return ` ${parts.join(" · ")}${suffix}`;
  }

  async function importSelected() {
    if (selectedIds.size === 0 || batchSaving) return;
    batchSaving = true;
    try {
      const ids = Array.from(selectedIds);
      const res = await api.post<{
        success?: boolean;
        data?: AnilistBatchImportData;
      }>("/admin/animes/batch-from-anilist", { anilist_ids: ids });

      const r = res.data?.data;
      if (!r) {
        toastState.addToast(
          "Invalid response from server (missing data).",
          "error",
          8000,
        );
        return;
      }

      if (r.imported_ids?.length) {
        savedIds = new Set([...savedIds, ...r.imported_ids]);
      }
      selectedIds = new Set();

      const errDetail = formatBatchErrors(r.errors);

      if (r.failed === 0 && r.imported > 0) {
        toastState.addToast(
          `${r.imported} anime(s) imported successfully.`,
          "success",
          5000,
        );
        setTimeout(() => goto("/admin/animes"), 1000);
        return;
      }

      if (r.imported > 0 && r.failed > 0) {
        toastState.addToast(
          `Partial import: ${r.imported} of ${r.requested} saved. ${r.failed} failed.${errDetail}`,
          "warning",
          10000,
        );
        return;
      }

      toastState.addToast(
        `Import failed: 0 of ${r.requested} saved.${errDetail || " Check server logs or AniList availability."}`,
        "error",
        12000,
      );
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to batch import animes"),
        "error",
        8000,
      );
    } finally {
      batchSaving = false;
    }
  }

  function formatLabel(season: string, year: number) {
    if (!season && !year) return "Unknown";
    return [season, year].filter(Boolean).join(" ");
  }

  const allImportable = $derived(
    results.filter((m: Media) => !savedIds.has(m.id)),
  );
  const allSelected = $derived(
    allImportable.length > 0 && selectedIds.size === allImportable.length,
  );
</script>

<svelte:head>
  <title>AniList Search — AniRank Admin</title>
</svelte:head>

<div class="flex items-center justify-between mb-6">
  <div class="flex items-center gap-3">
    <a
      href="/admin/animes"
      class="p-2 rounded-xl bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-colors border border-white/10"
    >
      <span class="material-symbols-outlined text-lg">arrow_back</span>
    </a>
    <div>
      <h1 class="text-2xl font-black text-white">AniList Search</h1>
      <p class="text-white/40 text-sm">
        Select an anime to import it into the database
      </p>
    </div>
  </div>

  {#if results.length > 0}
    <button
      onclick={toggleSelectAll}
      disabled={allImportable.length === 0}
      class="px-4 py-2 rounded-xl text-sm font-bold transition-colors border
        {allSelected
        ? 'bg-anirank-primary/20 text-anirank-primary border-anirank-primary/30'
        : 'bg-white/5 text-white/60 border-white/10 hover:bg-white/10 hover:text-white'}"
    >
      {allSelected ? "Deselect All" : "Select All"}
    </button>
  {/if}
</div>

<!-- Search bar -->
<form onsubmit={handleSearch} class="flex flex-col md:flex-row gap-3 mb-8">
  <div class="flex-1 relative">
    <span
      class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-white/30"
      >search</span
    >
    <input
      type="text"
      bind:value={searchInput}
      placeholder="Search anime on AniList..."
      class="w-full bg-white/5 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-anirank-primary transition-colors"
    />
  </div>

  <select
    bind:value={formatInput}
    class="bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:border-anirank-primary transition-colors appearance-none cursor-pointer min-w-[140px]"
  >
    {#each formats as fmt}
      <option value={fmt.value} class="bg-[#121214]">{fmt.label}</option>
    {/each}
  </select>

  <button
    type="submit"
    class="px-8 py-3 bg-anirank-primary hover:bg-anirank-primary/80 text-white font-bold rounded-xl transition-colors flex items-center justify-center gap-2 shadow-lg shadow-anirank-primary/10"
  >
    Apply
  </button>
</form>

{#if data.q && results.length === 0}
  <div class="text-center py-16 text-white/40">
    <span class="material-symbols-outlined text-5xl mb-3 block">search_off</span
    >
    No results found for "<span class="text-white/60">{data.q}</span>"
    {#if data.format}
      with format <span class="text-white/60">{data.format}</span>
    {/if}
  </div>
{:else if results.length > 0}
  <p class="text-white/40 text-sm mb-4">
    {results.length} results for "<span class="text-white/70">{data.q}</span>"
    {#if data.format}
      in <span class="text-white/70">{data.format}</span>
    {/if}
  </p>
  <div
    class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 mb-24"
  >
    {#each results as media (media.id)}
      {@const alreadySaved = savedIds.has(media.id)}
      {@const isSelected = selectedIds.has(media.id)}
      <div
        class="bg-anirank-card border rounded-2xl overflow-hidden group flex flex-col transition-all relative
          {isSelected
          ? 'border-anirank-primary ring-1 ring-anirank-primary'
          : 'border-white/5'}"
      >
        <!-- Selection checkbox (only if not saved) -->
        {#if !alreadySaved}
          <button
            onclick={() => toggleSelect(media.id)}
            class="absolute top-2 right-2 z-10 w-6 h-6 rounded-lg border transition-all flex items-center justify-center
              {isSelected
              ? 'bg-anirank-primary border-anirank-primary text-white'
              : 'bg-black/40 border-white/20 text-transparent hover:border-white/40'}"
            aria-label={isSelected ? "Deselect anime" : "Select anime"}
          >
            <span class="material-symbols-outlined text-sm font-bold"
              >check</span
            >
          </button>
        {/if}

        <!-- Cover -->
        <button
          class="relative aspect-2/3 overflow-hidden bg-white/5 cursor-pointer w-full text-left p-0 border-none"
          onclick={() => !alreadySaved && toggleSelect(media.id)}
          disabled={alreadySaved}
          type="button"
          aria-label="Toggle selection"
        >
          {#if media.coverImage?.extraLarge}
            <img
              src={media.coverImage.extraLarge}
              alt={media.title?.romaji}
              class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
          {:else}
            <div
              class="w-full h-full flex items-center justify-center text-white/20"
            >
              <span class="material-symbols-outlined text-4xl">image</span>
            </div>
          {/if}
          <!-- Format badge -->
          <span
            class="absolute top-2 left-2 px-2 py-0.5 bg-black/70 rounded text-[10px] font-bold text-white/80 uppercase tracking-wider"
          >
            {media.format ?? "?"}
          </span>

          {#if alreadySaved}
            <div
              class="absolute inset-0 bg-green-500/10 flex items-center justify-center"
            >
              <div
                class="bg-green-500 text-white px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest flex items-center gap-1 shadow-lg"
              >
                <span class="material-symbols-outlined text-sm">check</span>
                Imported
              </div>
            </div>
          {/if}
        </button>

        <!-- Info -->
        <div class="p-3 flex flex-col gap-1 flex-1">
          <p class="text-white font-bold text-sm leading-tight line-clamp-2">
            {media.title?.romaji ?? media.title?.english ?? "Unknown Title"}
          </p>
          <div class="flex items-center justify-between">
            <p class="text-white/40 text-xs">
              {formatLabel(media.season, media.seasonYear)}
            </p>
          </div>
        </div>
      </div>
    {/each}
  </div>

  <!-- Floating selection bar -->
  {#if selectedIds.size > 0}
    <div
      class="fixed bottom-8 left-1/2 -translate-x-1/2 z-40 w-full max-w-2xl px-4"
      transition:fly={{ y: 50, duration: 400 }}
    >
      <div
        class="bg-[#121214] border border-anirank-primary/30 rounded-2xl p-4 shadow-2xl flex items-center justify-between gap-4"
      >
        <div class="flex items-center gap-3">
          <div
            class="bg-anirank-primary text-white w-10 h-10 rounded-xl flex items-center justify-center font-black shadow-lg shadow-anirank-primary/20"
          >
            {selectedIds.size}
          </div>
          <div>
            <p class="text-white font-bold text-sm">Titles selected</p>
            <p
              class="text-white/40 text-[10px] uppercase tracking-wider font-bold"
            >
              Ready to import
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            onclick={() => (selectedIds = new Set())}
            disabled={batchSaving}
            class="px-4 py-2 rounded-xl text-white/60 hover:text-white hover:bg-white/5 transition-colors text-sm font-bold disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onclick={importSelected}
            disabled={batchSaving}
            class="px-6 py-2.5 bg-anirank-primary hover:bg-anirank-primary/80 text-white font-bold rounded-xl transition-all flex items-center gap-2 shadow-lg shadow-anirank-primary/10 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:scale-100"
          >
            {#if batchSaving}
              <svg
                class="animate-spin h-4 w-4"
                xmlns="http://www.w3.org/2000/svg"
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
              Saving...
            {:else}
              <span class="material-symbols-outlined text-lg">download</span>
              Import Selected
            {/if}
          </button>
        </div>
      </div>
    </div>
  {/if}
{/if}
