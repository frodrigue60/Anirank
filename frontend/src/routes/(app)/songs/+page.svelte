<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState } from "$lib/state/config.svelte";
  import SongCard from "$lib/components/SongCard.svelte";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search, Calendar, Cloud, Play, ListFilter } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let searchTerm = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedType = $state("");
  let selectedSort = $state("");

  // svelte-ignore state_referenced_locally
  let songs = $state(data.songs?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.songs?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.songs?.pagination?.last_page || 1);
  let loading = $state(false);

  // Sync state with URL params
  $effect(() => {
    searchTerm = data.params.name || "";
    selectedYear = data.params.year_id || "";
    selectedSeason = data.params.season_id || "";
    selectedType = data.params.type || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (data.songs && (data.songs.pagination?.current_page === 1 || data.songs.current_page === 1)) {
      songs = data.songs.data;
      currentPage = Number(data.songs.pagination.current_page);
      lastPage = Number(data.songs.pagination.last_page);
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val && val !== "any") url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (searchTerm) url.searchParams.set("name", searchTerm);
    else url.searchParams.delete("name");

    setParam("year_id", selectedYear);
    setParam("season_id", selectedSeason);
    setParam("type", selectedType);
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
          page: nextPage
        }
      });
      
      if (response.data.data) {
        songs = [...songs, ...response.data.data];
        currentPage = Number(response.data.pagination.current_page);
        lastPage = Number(response.data.pagination.last_page);
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
    { value: "any", label: "Any Year" },
    ...configState.years.map((y) => ({
      value: y.id.toString(),
      label: y.name,
    })),
  ]);
  const seasonOptions = $derived([
    { value: "any", label: "Any Season" },
    ...configState.seasons.map((s) => ({
      value: s.id.toString(),
      label: s.name,
    })),
  ]);
  const typeOptions = [
    { value: "any", label: "Any Type" },
    { value: "OP", label: "Opening" },
    { value: "ED", label: "Ending" },
    { value: "INS", label: "Insert" },
    { value: "Other", label: "Other" },
  ];
  const sortOptions = [
    { value: "any", label: "Recently Added" },
    { value: "rating", label: "Top Rated" },
    { value: "rating_asc", label: "Least Rated" },
  ];
</script>

<SEO 
  title="Discover Anime Songs" 
  description="Explore thousands of high-quality anime openings and endings. Filter by season, year, or search for your favorite tracks on AniRank." 
/>

<main class="max-w-[1440px] mx-auto px-6 py-10">
  <!-- Filter Bar -->
  <section
    class="relative z-40 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl mb-12"
  >
    <div class="flex flex-col gap-6">
      <div class="relative group">
        <label
          for="song-search"
          class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
        >
          Search Theme
        </label>
        <div class="relative">
          <span
            class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
          >
            <Search size={20} />
          </span>
          <input
            id="song-search"
            bind:value={searchTerm}
            oninput={handleInput}
            onkeydown={handleKeydown}
            class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
            placeholder="Search for a song or anime..."
            type="text"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <CustomSelect
          label="Year"
          bind:value={selectedYear}
          options={yearOptions}
          placeholder="All Years"
          icon={Calendar}
          onchange={updateFilters}
        />

        <CustomSelect
          label="Season"
          bind:value={selectedSeason}
          options={seasonOptions}
          placeholder="All Seasons"
          icon={Cloud}
          onchange={updateFilters}
        />

        <CustomSelect
          label="Type"
          bind:value={selectedType}
          options={typeOptions}
          placeholder="All Types"
          icon={Play}
          onchange={updateFilters}
        />

        <CustomSelect
          label="Sort By"
          bind:value={selectedSort}
          options={sortOptions}
          placeholder="Any"
          icon={ListFilter}
          onchange={updateFilters}
        />
      </div>
    </div>
  </section>

  <!-- Results count -->
  {#if data.songs?.pagination?.total > 0}
    <div class="mb-8 flex items-center justify-between">
      <h3 class="text-xl font-bold flex items-center gap-3 text-white">
        <span class="w-2 h-6 bg-primary rounded-full"></span>
        Results ({data.songs.pagination.total.toLocaleString()})
      </h3>
    </div>
  {/if}

  <!-- Grid Results -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    {#if songs.length > 0}
      {#each songs as song}
        <SongCard {song} />
      {/each}
    {:else}
      <div
        class="lg:col-span-2 text-center py-24 bg-surface-darker/30 rounded-3xl border-2 border-dashed border-white/5"
      >
        <span
          class="material-symbols-outlined text-6xl text-white/10 mb-4 block"
          >music_off</span
        >
        <h3 class="text-xl font-bold text-white/40">No themes found</h3>
        <p class="text-white/20 mt-2">
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
