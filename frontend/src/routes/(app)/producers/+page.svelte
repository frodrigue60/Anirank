<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

  let { data }: { data: any } = $props();

  let searchQuery = $state("");
  let selectedSort = $state("");

  // svelte-ignore state_referenced_locally
  let producers = $state(data.producers?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.producers?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.producers?.pagination?.last_page || 1);
  let loading = $state(false);

  // Sync state with URL params
  $effect(() => {
    searchQuery = data.params.name || "";
    selectedSort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (
      data.producers &&
      (data.producers.pagination?.current_page === 1 ||
        data.producers.current_page === 1)
    ) {
      producers = data.producers.data;
      currentPage = Number(data.producers.pagination.current_page);
      lastPage = Number(data.producers.pagination.last_page);
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val) url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (searchQuery) url.searchParams.set("name", searchQuery);
    else url.searchParams.delete("name");

    setParam("sort", selectedSort);

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true, noScroll: true });
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/producers", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        producers = [...producers, ...response.data.data];
        currentPage = Number(response.data.pagination.current_page);
        lastPage = Number(response.data.pagination.last_page);
      }
    } catch (e) {
      console.error("Error loading more producers", e);
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
    { value: "name_asc", label: "Alphabetical (A-Z)" },
    { value: "name_desc", label: "Alphabetical (Z-A)" },
    { value: "most_series", label: "Total Series (Most)" },
    { value: "least_series", label: "Total Series (Least)" },
  ];
</script>

<SEO
  title="Anime Producers"
  description="Discover the companies involved in anime production and explore their series catalog on AniRank."
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <!-- Filter Row -->
  <section
    class="relative z-40 flex flex-col gap-4 bg-surface-container p-4 rounded-md shadow-sm mb-10"
  >
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">
      <!-- Search -->
      <div class="lg:col-span-9 relative group">
        <label
          for="producer-search"
          class="block text-[10px] uppercase font-black text-on-surface-variant mb-2 ml-1 tracking-widest"
        >
          Search Producer
        </label>
        <div class="relative">
          <span
            class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors"
          >
            <Search size={20} />
          </span>
          <input
            id="producer-search"
            type="text"
            bind:value={searchQuery}
            oninput={handleInput}
            onkeydown={handleKeydown}
            class="w-full h-12 bg-surface-container border border-on-surface-variant/10 rounded-sm pl-12 pr-6 text-sm text-on-surface focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-on-surface-variant/30 transition-all"
            placeholder="Search animation producers (e.g. Shueisha, Aniplex)..."
          />
        </div>
      </div>

      <!-- Sort -->
      <div class="lg:col-span-3 flex flex-col gap-2">
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
  </section>

  <!-- Title and Count -->
  <div class="flex items-center justify-between mb-8">
    <h2
      class="text-2xl font-bold flex items-center gap-3 text-on-surface-variant/80"
    >
      <span class="w-2 h-8 bg-on-surface-variant/80 rounded-full"></span>
      Producers
      {#if data.producers?.pagination?.total > 0}
        ({data.producers.pagination.total.toLocaleString()})
      {/if}
    </h2>
  </div>

  <!-- Producer Grid (3 Columns) -->
  {#if producers.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each producers as producer}
        <a
          href="/producers/{producer.slug}"
          class="group relative overflow-hidden rounded-sm bg-surface border border-surface hover:border-primary/50 aspect-video transition-all cursor-pointer shadow-sm"
        >
          <!-- Background Image (Banner) -->
          <div
            class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
            style="background-image: url('{producer.banner_url ||
              '/images/placeholders/default-banner.jpg'}'); filter:brightness(0.5)"
          ></div>

          <div
            class="absolute inset-0 bg-linear-to-t from-surface-container via-surface-container/40 to-transparent"
          ></div>

          <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col gap-1">
            <h3
              class="text-2xl font-bold text-on-surface group-hover:text-primary transition-colors"
            >
              {producer.name}
            </h3>
            <div
              class="mt-4 flex items-center justify-between border-t border-white/10 pt-4"
            >
              <div class="flex flex-col">
                <span
                  class="text-[10px] text-on-surface-variant uppercase font-bold tracking-wider"
                  >Produced</span
                >
                <span class="text-on-surface text-sm font-semibold"
                  >{producer.anime_count || 0} Series</span
                >
              </div>
            </div>
          </div>
        </a>
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
      <span
        class="material-symbols-outlined text-6xl text-on-surface-variant opacity-20 mb-4 block"
        >search_off</span
      >
      <h3 class="text-xl font-bold text-on-surface-variant">
        No producers found
      </h3>
      <p class="text-on-surface-variant opacity-40 mt-2">
        Try adjusting your search or filters
      </p>
    </div>
  {/if}
</main>

<style>
</style>
