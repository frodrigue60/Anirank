<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import Search from "lucide-svelte/icons/search";
  import UserMinus from "lucide-svelte/icons/user-minus";
  import SEO from "$lib/components/SEO.svelte";
  import ArtistAvatarCard from "$lib/components/ArtistAvatarCard.svelte";

  let { data } = $props();

  let searchQuery = $state("");
  let selectedSort = $state("");

  // svelte-ignore state_referenced_locally
  let artists = $state(data.artistsData?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.artistsData?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.artistsData?.pagination?.last_page || 1);
  let loading = $state(false);

  // Sync state with URL params and data
  $effect(() => {
    searchQuery = data.params.name || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (
      data.artistsData &&
      Number(data.artistsData.pagination?.current_page) === 1
    ) {
      artists = data.artistsData.data;
      currentPage = Number(data.artistsData.pagination.current_page);
      lastPage = Number(data.artistsData.pagination.last_page);
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val && val !== "any") url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (searchQuery) url.searchParams.set("name", searchQuery);
    else url.searchParams.delete("name");

    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString());
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/artists", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        artists = [...artists, ...response.data.data];
        currentPage = Number(response.data.pagination.current_page);
        lastPage = Number(response.data.pagination.last_page);
      }
    } catch (e) {
      console.error("Error loading more artists", e);
    } finally {
      loading = false;
    }
  }

  let searchTimeout: any;
  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      if (searchQuery.length === 0 || searchQuery.length >= 2) {
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
    { value: "most_themes", label: "Most Themes" },
    { value: "least_themes", label: "Least Themes" },
    { value: "name_asc", label: "A - Z" },
    { value: "name_desc", label: "Z - A" },
  ];
</script>

<SEO
  title="Browse Artists"
  description="Discover artists and performers of your favorite anime theme songs on AniRank."
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <div class="flex flex-col gap-4">
    <!-- Search and Filters Section -->
    <section
      class="relative z-40 flex flex-col gap-4 bg-surface-container p-4 rounded-md border border-white/5 shadow-sm"
    >
      <div class="flex items-end gap-4">
        <!-- Search -->
        <div class="relative group w-full">
          <label
            for="artist-search"
            class="block text-[10px] uppercase font-black text-on-surface-variant mb-2 ml-1 tracking-widest"
          >
            Search Artist
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="artist-search"
              bind:value={searchQuery}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-container border border-on-surface-variant/20 rounded-sm pl-12 pr-6 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-on-surface-variant transition-all"
              placeholder="Who are you looking for?"
              type="text"
            />
          </div>
        </div>

        <!-- Sort -->
        <div class="w-full md:w-50 flex flex-col gap-2">
          <label
            for="sort-select"
            class="block text-[10px] uppercase font-black text-on-surface-variant mb-0 ml-1 tracking-widest"
          >
            Sort By
          </label>
          <select
            id="sort-select"
            bind:value={selectedSort}
            onchange={updateFilters}
            class="w-full h-12 bg-surface-container border border-on-surface-variant/20 rounded-sm px-4 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all cursor-pointer"
          >
            <option value="">Any</option>
            {#each sortOptions as option}
              <option value={option.value}>{option.label}</option>
            {/each}
          </select>
        </div>
      </div>
    </section>

    <!-- Artists Grid -->
    <section class="">
      <div class="flex items-center justify-between mb-8">
        <h2
          class="text-2xl font-bold flex items-center gap-3 text-on-surface-variant/80"
        >
          <span class="w-2 h-8 bg-on-surface-variant/80 rounded-full"></span>
          Artists
          {#if data.artistsData?.pagination?.total > 0}
            ({data.artistsData.pagination.total.toLocaleString()})
          {/if}
        </h2>
      </div>

      {#if artists.length > 0}
        <div
          class="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-8"
        >
          {#each artists as artist}
            <ArtistAvatarCard {artist} />
          {/each}
        </div>

        <InfiniteScroll
          hasMore={currentPage < lastPage}
          {loading}
          onLoadMore={loadMore}
        />
      {:else}
        <div
          class="text-center py-20 bg-surface-container/30 rounded-md border-2 border-dashed border-white/5"
        >
          <UserMinus size={64} class="text-on-surface-variant opacity-20 mb-4 block mx-auto" />
          <h3 class="text-xl font-bold text-on-surface-variant">
            No artists found
          </h3>
          <p class="text-on-surface-variant opacity-40 mt-2">
            Try adjusting your search or filters
          </p>
        </div>
      {/if}
    </section>
  </div>
</main>
