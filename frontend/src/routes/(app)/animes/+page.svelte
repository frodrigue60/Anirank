<script lang="ts">
  import { fade } from "svelte/transition";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState } from "$lib/state/config.svelte";
  import AnimeCard from "$lib/components/AnimeCard.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let searchTerm = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedFormat = $state("");
  let selectedSort = $state("");
  let selectedGenre = $state("");

  // svelte-ignore state_referenced_locally
  let animes = $state(data.animes?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.animes?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.animes?.pagination?.last_page || 1);
  // svelte-ignore state_referenced_locally
  let totalAnimes = $state(data.animes?.pagination?.total || 0);
  let loading = $state(false);

  $effect(() => {
    searchTerm = data.params.name || "";
    selectedYear = data.params.year || "";
    selectedSeason = data.params.season || "";
    selectedFormat = data.params.format || "";
    selectedGenre = data.params.genre || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (data.animes && Number(data.animes.pagination?.current_page) === 1) {
      animes = data.animes.data;
      currentPage = Number(data.animes.pagination.current_page);
      lastPage = Number(data.animes.pagination.last_page);
      totalAnimes = Number(data.animes.pagination.total);
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
    setParam("format", selectedFormat);
    setParam("genre", selectedGenre);
    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString());
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/animes", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        animes = [...animes, ...response.data.data];
        currentPage = Number(response.data.pagination.current_page);
        lastPage = Number(response.data.pagination.last_page);
        totalAnimes = Number(response.data.pagination.total);
      }
    } catch (e) {
      console.error("Error loading more animes", e);
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

  // Mapeo de opciones para CustomSelect
  const yearOptions = $derived(
    configState.years.map((y) => ({
      value: y.slug,
      label: y.name,
    })),
  );
  const seasonOptions = $derived(
    configState.seasons.map((s) => ({
      value: s.slug,
      label: s.name,
    })),
  );
  const formatOptions = $derived(
    configState.formats.map((f) => ({
      value: f.slug,
      label: f.name,
    })),
  );
  const sortOptions = [
    { value: "most_themes", label: "Most Themes" },
    { value: "least_themes", label: "Least Themes" },
    { value: "latest", label: "Recently Added" },
    { value: "title", label: "Alphabetical" },
  ];
  const genreOptions = $derived(
    configState.genres.map((g) => ({
      value: g.slug,
      label: g.name,
    })),
  );
  let viewType = $state<"grid" | "card" | "list">("grid");
</script>

<SEO
  title="Browse Anime"
  description="Browse and search the extensive catalog of anime to find your favorite theme songs on AniRank."
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <div class="flex flex-col gap-4">
    <!-- Header/Filters Section -->
    <section
      class="relative z-40 flex flex-col gap-4 bg-surface-container p-4 rounded-md border border-white/5 shadow-sm"
    >
      <div
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 items-end"
      >
        <!-- Search -->
        <div class="relative group">
          <label
            for="anime-search"
            class="block text-[10px] uppercase font-black mb-2 ml-1 tracking-widest text-on-surface-variant"
          >
            Search Anime
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="anime-search"
              bind:value={searchTerm}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm pl-12 pr-6 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-on-surface-variant transition-all"
              placeholder="What are you looking for?"
              type="text"
            />
          </div>
        </div>
        <!-- Year Select -->
        <div class="flex flex-col gap-2">
          <label
            for="year-select"
            class="text-[10px] uppercase font-black mb-0 ml-1 tracking-widest text-on-surface-variant"
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
            class="text-[10px] uppercase font-black mb-0 ml-1 tracking-widest text-on-surface-variant"
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

        <!-- Format Select -->
        <div class="flex flex-col gap-2">
          <label
            for="format-select"
            class="text-[10px] uppercase font-black mb-0 ml-1 tracking-widest text-on-surface-variant"
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

        <!-- Genres Select -->
        <div class="flex flex-col gap-2">
          <label
            for="genre-select"
            class="text-[10px] uppercase font-black mb-0 ml-1 tracking-widest text-on-surface-variant"
            >Genres</label
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

        <!-- Sort Select -->
        <div class="flex flex-col gap-2">
          <label
            for="sort-select"
            class="text-[10px] uppercase font-black mb-0 ml-1 tracking-widest text-on-surface-variant"
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

    <!-- Results Section -->
    <section class="">
      <div class="flex items-center justify-between text-2xl gap-2 mb-4">
        <div>
          <!-- Results count -->
          {#if totalAnimes > 0}
            <h3
              class="text-xl font-bold flex items-center gap-3 text-on-surface-variant/80"
            >
              <span class="w-2 h-6 bg-on-surface-variant/80 rounded-full"
              ></span>
              Results ({totalAnimes.toLocaleString()})
            </h3>
          {/if}
        </div>

        <div class="flex items-center gap-2 bg-surface-highest p-1 rounded-md">
          <button
            onclick={() => (viewType = "grid")}
            class="flex items-center gap-2 p-1 rounded-sm {viewType === 'grid'
              ? 'text-surface/80 bg-primary'
              : 'text-on-surface-variant'}"
            title="Compact Grid"
          >
            <span class="material-symbols-outlined">grid_on</span>
          </button>
          <button
            onclick={() => (viewType = "card")}
            class="flex items-center gap-2 p-1 rounded-sm {viewType === 'card'
              ? 'text-surface/80 bg-primary'
              : 'text-on-surface-variant'}"
            title="Detailed Cards"
          >
            <span class="material-symbols-outlined">border_all</span>
          </button>
          <button
            onclick={() => (viewType = "list")}
            class="flex items-center gap-2 p-1 rounded-sm {viewType === 'list'
              ? 'text-surface/80 bg-primary'
              : 'text-on-surface-variant'}"
            title="List View"
          >
            <span class="material-symbols-outlined">list_alt</span>
          </button>
        </div>
      </div>

      {#if animes.length > 0}
        <div
          class="grid gap-6 {viewType === 'grid'
            ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6'
            : viewType === 'card'
              ? 'grid-cols-1 md:grid-cols-2 xl:grid-cols-3'
              : 'grid-cols-1'}"
        >
          {#each animes as anime (anime.slug)}
            <div in:fade={{ duration: 300 }}>
              <AnimeCard {anime} view={viewType} />
            </div>
          {/each}
        </div>

        <InfiniteScroll
          hasMore={currentPage < lastPage}
          {loading}
          onLoadMore={loadMore}
        />
      {:else}
        <div
          class="flex flex-col items-center justify-center py-20 bg-surface-dark/30 rounded-md border border-dashed border-white/5"
        >
          <Search size={48} class="text-white/10 mb-4" />
          <h3 class="text-xl font-bold text-white mb-2">No results found</h3>
          <p class="text-white/40 max-w-xs text-center">
            Try adjusting your filters or search terms to find what you're
            looking for.
          </p>
        </div>
      {/if}
    </section>
  </div>
</main>
