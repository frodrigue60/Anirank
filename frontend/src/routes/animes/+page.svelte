<script lang="ts">
  import { fade } from "svelte/transition";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState } from "$lib/state/config.svelte";
  import AnimeCard from "$lib/components/AnimeCard.svelte";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import {
    Search,
    Calendar,
    CloudSun,
    Layout,
    SortDesc,
    Trash2,
  } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let searchTerm = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedType = $state("");
  let selectedSort = $state("");

  let animes = $state(data.animes?.data || []);
  let currentPage = $state(data.animes?.current_page || 1);
  let lastPage = $state(data.animes?.last_page || 1);
  let loading = $state(false);

  $effect(() => {
    searchTerm = data.params.name || "";
    selectedYear = data.params.year_id || "";
    selectedSeason = data.params.season_id || "";
    selectedType = data.params.type || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (data.animes && data.animes.current_page === 1) {
      animes = data.animes.data;
      currentPage = data.animes.current_page;
      lastPage = data.animes.last_page;
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
          page: nextPage
        }
      });

      if (response.data.data) {
        animes = [...animes, ...response.data.data];
        currentPage = response.data.current_page;
        lastPage = response.data.last_page;
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
  const typeOptions = $derived([
    { value: "any", label: "Any" },
    ...configState.formats.map((f) => ({
      value: f.id.toString(),
      label: f.name,
    })),
  ]);
  const sortOptions = [
    { value: "any", label: "Any" },
    { value: "most_themes", label: "Most Themes" },
    { value: "least_themes", label: "Least Themes" },
    { value: "latest", label: "Recently Added" },
    { value: "title", label: "Alphabetical" },
  ];
</script>

<SEO 
  title="Browse Anime" 
  description="Browse and search the extensive catalog of anime to find your favorite theme songs on AniRank." 
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <div class="flex flex-col gap-10">
    <!-- Header/Filters Section -->
    <section
      class="relative z-40 flex flex-col gap-4 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl"
    >
      <div>
        <!-- search -->
        <div class="relative group">
          <label
            for="anime-search"
            class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
          >
            Search Anime
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="anime-search"
              bind:value={searchTerm}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
              placeholder="What are you looking for?"
              type="text"
            />
          </div>
        </div>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-4 gap-4 items-end">
        <!-- Year Select -->
        <div class="">
          <CustomSelect
            label="Year"
            bind:value={selectedYear}
            options={yearOptions}
            placeholder="Any Year"
            icon={Calendar}
            onchange={updateFilters}
          />
        </div>

        <!-- Season Select -->
        <div class="">
          <CustomSelect
            label="Season"
            bind:value={selectedSeason}
            options={seasonOptions}
            placeholder="Any Season"
            icon={CloudSun}
            onchange={updateFilters}
          />
        </div>

        <!-- Type Select -->
        <div class="">
          <CustomSelect
            label="Type"
            bind:value={selectedType}
            options={typeOptions}
            placeholder="Any Type"
            icon={Layout}
            onchange={updateFilters}
          />
        </div>

        <!-- Sort & Clear -->
        <div class="flex items-end gap-2">
          <div class="w-full">
            <CustomSelect
              label="Sort"
              bind:value={selectedSort}
              options={sortOptions}
              placeholder="Sort By"
              icon={SortDesc}
              onchange={updateFilters}
            />
          </div>
        </div>
      </div>
    </section>

    <!-- Results Section -->
    <section class="mt-4">
      {#if animes.length > 0}
        <div
          class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6"
        >
          {#each animes as anime (anime.id)}
            <div in:fade={{ duration: 300 }}>
              <AnimeCard {anime} />
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
          class="flex flex-col items-center justify-center py-20 bg-surface-dark/30 rounded-3xl border border-dashed border-white/5"
        >
          <Search size={48} class="text-white/10 mb-4" />
          <h3 class="text-xl font-bold text-white mb-2">No results found</h3>
          <p class="text-white/40 max-w-xs text-center">
            Try adjusting your filters or search terms to find what you're looking
            for.
          </p>
        </div>
      {/if}
    </section>
  </div>
</main>
