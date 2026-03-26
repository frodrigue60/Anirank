<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState } from "$lib/state/config.svelte";
  import SongCard from "$lib/components/SongCard.svelte";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import {
    Search,
    Calendar,
    Cloud,
    Play,
    ListFilter,
    SortDesc,
  } from "lucide-svelte";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import SEO from "$lib/components/SEO.svelte";
  const PUBLIC_API_URL =
    import.meta.env.VITE_API_URL || "http://localhost:8080/api";

  let { data }: { data: any } = $props();

  let artist = $derived(data.artist);
  let searchTerm = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedType = $state("");
  let selectedSort = $state("");
  // svelte-ignore state_referenced_locally
  let isFavorited = $state(data.artist?.is_favorited || false);

  // Sync state with URL params
  $effect(() => {
    isFavorited = artist?.is_favorited || false;
    searchTerm = data.params?.name || "";
    selectedYear = data.params?.year_id || "";
    selectedSeason = data.params?.season_id || "";
    selectedType = data.params?.type || "";
    selectedSort = data.params?.sort || "";
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

    url.searchParams.delete("page");
    goto(url.toString(), { keepFocus: true, noScroll: true });
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
    { value: "any", label: "Any" },
    { value: "OP", label: "Opening" },
    { value: "ED", label: "Ending" },
  ];
  const sortOptions = [
    { value: "any", label: "Recently Added" },
    { value: "rating", label: "Top Rated" },
    { value: "rating_asc", label: "Least Rated" },
    { value: "most_popular", label: "Most Popular (Likes)" },
    { value: "most_views", label: "Most Viewed" },
    { value: "alphabetical", label: "Alphabetical (A-Z)" },
  ];

  function goToPage(p: number | string) {
    if (typeof p === "string") return;
    const url = new URL(page.url);
    url.searchParams.set("page", p.toString());
    goto(url.toString(), { keepFocus: true, noScroll: false });
  }

  // Helper for pagination numbers
  function getPaginationRange(current: number, last: number) {
    const delta = 2;
    const left = current - delta;
    const right = current + delta + 1;
    const range = [];
    const rangeWithDots = [];
    let l;

    for (let i = 1; i <= last; i++) {
      if (i === 1 || i === last || (i >= left && i < right)) {
        range.push(i);
      }
    }

    for (const i of range) {
      if (l) {
        if (i - l === 2) {
          rangeWithDots.push(l + 1);
        } else if (i - l !== 1) {
          rangeWithDots.push("...");
        }
      }
      rangeWithDots.push(i);
      l = i;
    }

    return rangeWithDots;
  }

  let paginationRange = $derived(
    data.songs && data.songs.current_page
      ? getPaginationRange(data.songs.current_page, data.songs.last_page)
      : [],
  );

  function formatScore(scoreStr: string) {
    if (!scoreStr) return "0.0";
    return scoreStr.split("/")[0];
  }

  function clearFilters() {
    searchTerm = "";
    selectedYear = "";
    selectedSeason = "";
    selectedType = "";
    selectedSort = "any";
    updateFilters();
  }

  async function toggleFavorite() {
    if (!artist) return;
    try {
      const response = await api.post("/interactions/favorites", {
        entity_id: artist.id,
        entity_type: "artist",
      });
      isFavorited = response.data.data.favorited;
      toastState.addToast(
        isFavorited ? "Added to favorites" : "Removed from favorites",
        "success",
      );
    } catch (err: any) {
      toastState.addToast(
        err.response?.data?.message || "Failed to update favorite",
        "error",
      );
    }
  }
</script>

{#if artist}
  <SEO
    title={`${artist.name} - Artist on AniRank`}
    description={`Explore all theme songs by ${artist.name} on AniRank. Discover their work, ratings, and rankings.`}
    image={`${PUBLIC_API_URL}/og/artist/${artist.slug}`}
    type="profile"
  />
{:else}
  <SEO title="Artist Not Found" />
{/if}

<main class="max-w-[1440px] mx-auto px-6 py-10">
  {#if !artist}
    <div class="text-center py-24">
      <h1 class="text-3xl font-bold text-white mb-4">Artist Not Found</h1>
      <a href="/artists" class="text-primary hover:underline">Back to Artists</a
      >
    </div>
  {:else}
    <!-- Hero Section -->
    <div class="mb-10 flex flex-col md:flex-row items-center gap-6">
      <div class="shrink-0 relative group">
        <div
          class="w-24 h-24 md:w-32 md:h-32 rounded-full overflow-hidden border-4 border-surface shadow-2xl relative z-10 group-hover:border-primary/50 transition-colors"
        >
          <img
            src={artist.avatar_url || "/images/placeholders/default.jpg"}
            alt="Avatar for {artist.name}"
            title="Avatar for {artist.name}"
            class="w-full h-full object-cover"
          />
        </div>
        <div
          class="absolute inset-0 bg-primary/20 blur-3xl -z-10 rounded-full group-hover:bg-primary/40 transition-colors"
        ></div>
      </div>
      <div>
        <h2
          class="text-2xl md:text-3xl font-black text-white mb-2 tracking-tight"
        >
          {artist.name}
        </h2>
        {#if artist.name_jp}
          <p class="text-white/60 text-lg mb-2">{artist.name_jp}</p>
        {/if}
        <p class="text-primary font-bold">Theme Songs Collection</p>
      </div>
      <div class="ms-auto">
        <button
          onclick={toggleFavorite}
          class="flex items-center gap-2 px-4 py-2 rounded-xl border border-white/10 transition-all active:scale-95 {isFavorited
            ? 'bg-red-500/20 border-red-500/50 text-red-500'
            : 'text-white hover:bg-white/5'}"
        >
          <span
            class="material-symbols-outlined {isFavorited
              ? 'fill-current'
              : ''}"
          >
            favorite
          </span>
          Favorite
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <section
      class="relative z-40 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl mb-12"
    >
      <div class="flex flex-col gap-6">
        <div class="relative group">
          <label
            for="theme-search"
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
              id="theme-search"
              bind:value={searchTerm}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
              placeholder="Search for a song or anime title..."
              type="text"
            />
          </div>
        </div>
        <div class="flex flex-col lg:grid lg:grid-cols-4 gap-4">
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
            label="Sort"
            bind:value={selectedSort}
            options={sortOptions}
            placeholder="Any"
            icon={SortDesc}
            onchange={updateFilters}
          />
        </div>
      </div>
    </section>

    <!-- Grid Results -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      {#if data.songs?.data && data.songs.data.length > 0}
        {#each data.songs.data as song}
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
          <button
            onclick={clearFilters}
            class="mt-6 text-primary hover:underline font-bold"
          >
            Clear all filters
          </button>
        </div>
      {/if}
    </div>

    <!-- Pagination -->
    {#if data.songs?.last_page > 1}
      <div class="mt-16 flex justify-center items-center gap-2">
        <button
          onclick={() => goToPage(data.songs.current_page - 1)}
          disabled={data.songs.current_page === 1}
          class="w-10 h-10 rounded-xl bg-background-dark/50 border border-white/5 flex items-center justify-center text-white hover:bg-surface-dark transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
          title="Go to previous page"
        >
          <span class="material-symbols-outlined">chevron_left</span>
        </button>

        {#each paginationRange as p}
          {#if p === "..."}
            <span class="text-white/20 px-2">...</span>
          {:else}
            <button
              onclick={() => goToPage(p)}
              class="w-10 h-10 rounded-xl font-bold text-sm transition-all {data
                .songs.current_page === p
                ? 'bg-primary text-white shadow-lg shadow-primary/30'
                : 'bg-background-dark/50 border border-white/5 text-white/60 hover:text-white hover:bg-surface-dark'}"
              title="Go to page {p}"
            >
              {p}
            </button>
          {/if}
        {/each}

        <button
          onclick={() => goToPage(data.songs.current_page + 1)}
          disabled={data.songs.current_page === data.songs.last_page}
          class="w-10 h-10 rounded-xl bg-background-dark/50 border border-white/5 flex items-center justify-center text-white hover:bg-surface-dark transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
          title="Go to next page"
        >
          <span class="material-symbols-outlined">chevron_right</span>
        </button>
      </div>
    {/if}
  {/if}
</main>

<style>
</style>
