<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState } from "$lib/state/config.svelte";
  import SongCard from "$lib/components/SongCard.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let searchTerm = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedGenre = $state("");
  let selectedType = $state("");
  let selectedFormat = $state("");
  let selectedSort = $state("");

  // svelte-ignore state_referenced_locally
  let songs = $state(data.songs?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.songs?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.songs?.pagination?.last_page || 1);
  // svelte-ignore state_referenced_locally
  let totalSongs = $state(data.songs?.pagination?.total || 0);
  let loading = $state(false);

  // Sync state with URL params
  $effect(() => {
    searchTerm = data.params.name || "";
    selectedYear = data.params.year || "";
    selectedSeason = data.params.season || "";
    selectedGenre = data.params.genre || "";
    selectedType = data.params.type || "";
    selectedFormat = data.params.format || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (
      data.songs &&
      (data.songs.pagination?.current_page === 1 ||
        data.songs.current_page === 1)
    ) {
      songs = data.songs.data;
      currentPage = Number(data.songs.pagination.current_page);
      lastPage = Number(data.songs.pagination.last_page);
      totalSongs = Number(data.songs.pagination.total);
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val) url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (searchTerm) url.searchParams.set("name", searchTerm);
    else url.searchParams.delete("name");

    setParam("year", selectedYear);
    setParam("season", selectedSeason);
    setParam("genre", selectedGenre);
    setParam("type", selectedType);
    setParam("format", selectedFormat);
    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true, noScroll: true });
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/songs", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        songs = [...songs, ...response.data.data];
        currentPage = Number(response.data.pagination.current_page);
        lastPage = Number(response.data.pagination.last_page);
        totalSongs = Number(response.data.pagination.total);
      }
    } catch (e) {
      console.error("Error loading more songs", e);
    } finally {
      loading = false;
    }
  }

  let searchTimeout: any;
  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      if (searchTerm.length === 0 || searchTerm.length >= 2) {
        updateFilters();
      }
    }, 300);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      updateFilters();
    }
  }
  const yearOptions = $derived([
    ...configState.years.map((y) => ({
      value: y.slug,
      label: y.name,
    })),
  ]);
  const seasonOptions = $derived([
    ...configState.seasons.map((s) => ({
      value: s.slug,
      label: s.name,
    })),
  ]);
  const genreOptions = $derived([
    ...configState.genres.map((g) => ({
      value: g.slug,
      label: g.name,
    })),
  ]);
  const typeOptions = [
    { value: "OP", label: "Opening" },
    { value: "ED", label: "Ending" },
    { value: "INS", label: "Insert" },
    { value: "Other", label: "Other" },
  ];
  const sortOptions = [
    { value: "rating", label: "Top Rated" },
    { value: "rating_asc", label: "Least Rated" },
    { value: "recently_added", label: "Recently Added" },
    { value: "favorites", label: "Most Favorited" },
    { value: "views", label: "Most Viewed" },
  ];

  const formatOptions = $derived([
    ...configState.formats.map((f) => ({
      value: f.slug,
      label: f.name,
    })),
  ]);
</script>

<SEO
  title="Discover Anime Songs"
  description="Explore thousands of high-quality anime openings and endings. Filter by season, year, or search for your favorite tracks on AniRank."
/>

<main class="max-w-[1440px] mx-auto px-6 py-10 space-y-4">
  <!-- Filter Bar -->
  <section
    class="relative z-40 bg-surface-container p-4 rounded-md border border-white/5 shadow-sm"
  >
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <!-- Search -->
      <div class="relative group">
        <label
          for="song-search"
          class="block text-[10px] uppercase font-black text-on-surface-variant mb-2 ml-1 tracking-widest"
        >
          Search Theme
        </label>
        <div class="relative">
          <span
            class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors"
          >
            <Search size={20} />
          </span>
          <input
            id="song-search"
            bind:value={searchTerm}
            oninput={handleInput}
            onkeydown={handleKeydown}
            class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm pl-12 pr-6 text-sm text-on-surface focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-on-surface-variant transition-all"
            placeholder="Search theme or artist..."
            type="text"
          />
        </div>
      </div>
      <!-- Year Select -->
      <div class="flex flex-col gap-2">
        <label
          for="year-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Year</label
        >
        <select
          id="year-select"
          bind:value={selectedYear}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">All Years</option>
          {#each yearOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <!-- Season Select -->
      <div class="flex flex-col gap-2">
        <label
          for="season-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Season</label
        >
        <select
          id="season-select"
          bind:value={selectedSeason}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">All Seasons</option>
          {#each seasonOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <!-- Genre Select -->
      <div class="flex flex-col gap-2">
        <label
          for="genre-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Genre</label
        >
        <select
          id="genre-select"
          bind:value={selectedGenre}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">All Genres</option>
          {#each genreOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <!-- Type Select -->
      <div class="flex flex-col gap-2">
        <label
          for="type-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Type</label
        >
        <select
          id="type-select"
          bind:value={selectedType}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">All Types</option>
          {#each typeOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <!-- Format Select -->
      <div class="flex flex-col gap-2">
        <label
          for="format-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Format</label
        >
        <select
          id="format-select"
          bind:value={selectedFormat}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">All Formats</option>
          {#each formatOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <!-- Sort Select -->
      <div class="flex flex-col gap-2">
        <label
          for="sort-select"
          class="text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >Sort By</label
        >
        <select
          id="sort-select"
          bind:value={selectedSort}
          onchange={updateFilters}
          class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
        >
          <option value="">Any</option>
          {#each sortOptions as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
    </div>
  </section>

  <div class="flex items-center justify-between">
    <h3
      class="text-xl font-bold flex items-center gap-3 text-on-surface-variant/80"
    >
      <span class="w-2 h-6 bg-on-surface-variant/80 rounded-full"></span>
      {#if totalSongs > 0}
        Results ({totalSongs.toLocaleString()})
      {:else}
        No results found
      {/if}
    </h3>
  </div>

  <!-- Grid Results -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    {#if songs.length > 0}
      {#each songs as song}
        <SongCard {song} />
      {/each}
    {:else}
      <div
        class="lg:col-span-2 text-center py-24 bg-surface-container/30 rounded-3xl border-2 border-dashed border-white/5"
      >
        <span
          class="material-symbols-outlined text-6xl text-on-surface-variant opacity-20 mb-4 block"
          >music_off</span
        >
        <h3 class="text-xl font-bold text-on-surface-variant">
          No themes found
        </h3>
        <p class="text-on-surface-variant opacity-40 mt-2">
          Try adjusting your filters or search query
        </p>
      </div>
    {/if}
  </div>

  <InfiniteScroll
    hasMore={currentPage < lastPage}
    {loading}
    onLoadMore={loadMore}
  />
</main>

<style>
</style>
