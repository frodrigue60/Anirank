<script lang="ts">
  import type { PageData } from "./$types";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import AnimeCard from "$lib/components/AnimeCard.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data }: { data: PageData } = $props();
  let producer = $derived(data.producer);

  let searchTerm = $state("");
  let selectedSort = $state("");

  // svelte-ignore state_referenced_locally
  let animes = $state(data.animes?.data || []);
  // svelte-ignore state_referenced_locally
  let paginationMeta = $state(data.animes || { current_page: 1, last_page: 1 });
  let loading = $state(false);

  // Sync state with URL params
  $effect(() => {
    searchTerm = data.params?.name || "";
    selectedSort = data.params?.sort || "";

    // Reset infinite scroll on data change (filters)
    if (
      data.animes &&
      (data.animes.pagination?.current_page === 1 ||
        data.animes.current_page === 1)
    ) {
      animes = data.animes.data;
      paginationMeta = data.animes;
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

    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true, noScroll: true });
  }

  async function loadMore() {
    const nextUrl =
      paginationMeta.links?.next ||
      (paginationMeta.pagination?.has_more
        ? `/producers/${data.producer.slug}?page=${(paginationMeta.pagination?.current_page || 1) + 1}`
        : null);
    if (loading || !nextUrl) return;

    loading = true;
    try {
      const response = await api.get(nextUrl);

      // The backend now returns the flattened paginated object with "data" key
      const newAnimesData = response.data.data;

      if (newAnimesData?.data) {
        animes = [...animes, ...newAnimesData.data];
        paginationMeta = newAnimesData;
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

  const sortOptions = [
    { value: "name_desc", label: "Alphabetical (Z-A)" },
    { value: "most_themes", label: "Most Themes" },
    { value: "least_themes", label: "Least Themes" },
  ];
</script>

<SEO
  title="{producer.name} Anime Production"
  description="Explore the catalog of anime produced by {producer.name} and discover their theme songs on AniRank."
  type="profile"
/>

<main class="w-full max-w-[1440px] mx-auto px-6 py-8 space-y-4">
  <div class="text-center">
    <h1 class="text-4xl md:text-5xl font-black text-on-surface mb-4">
      {producer.name}
    </h1>
    <p class="text-on-surface-variant">Animes produced by {producer.name}</p>
  </div>

  <!-- Filter Bar -->
  <section
    class="relative z-40 bg-surface-container p-4 rounded-md border border-white/5 shadow-2xl mx-auto"
  >
    <div class="flex flex-col gap-6">
      <div class="grid grid-cols-1 md:grid-cols-12 gap-4 items-end">
        <!-- Search -->
        <div class="md:col-span-9 relative group">
          <label
            for="anime-search"
            class="block text-[10px] uppercase font-black text-on-surface-variant mb-2 ml-1 tracking-widest"
          >
            Search Anime
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant/50 group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="anime-search"
              bind:value={searchTerm}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm pl-12 pr-6 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-on-surface-variant/30 transition-all"
              placeholder="Search animes by title..."
              type="text"
            />
          </div>
        </div>

        <!-- Sort -->
        <div class="md:col-span-3 flex flex-col gap-2">
          <label
            for="sort-select"
            class="block text-[10px] uppercase font-black text-on-surface-variant mb-2 ml-1 tracking-widest"
          >
            Sort By
          </label>
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
    </div>
  </section>

  <!-- Title and Count -->
  <div class="flex items-center justify-between">
    <div>
      <h2
        class="text-2xl font-bold flex items-center gap-3 text-on-surface-variant/80"
      >
        <span class="w-2 h-8 bg-on-surface-variant/80 rounded-full"></span>
        Animes
        {#if data.animes?.pagination?.total > 0}
          ({data.animes.pagination.total.toLocaleString()})
        {/if}
      </h2>
    </div>
  </div>

  {#if animes.length > 0}
    <div
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6"
    >
      {#each animes as anime}
        <AnimeCard {anime} />
      {/each}
    </div>

    <InfiniteScroll
      hasMore={paginationMeta.links?.next ||
        paginationMeta.pagination?.has_more ||
        paginationMeta.current_page < paginationMeta.last_page}
      {loading}
      onLoadMore={loadMore}
    />
  {:else}
    <div
      class="text-center py-20 bg-surface-container/30 rounded-md border border-white/5"
    >
      <span
        class="material-symbols-outlined text-6xl text-on-surface-variant opacity-20 mb-4 block"
        >movie</span
      >
      <h3 class="text-xl font-bold text-on-surface-variant opacity-50 mb-2">
        No animes found
      </h3>
      <p class="text-on-surface-variant opacity-30 text-sm">
        There are no series associated with this producer.
      </p>
    </div>
  {/if}
</main>
