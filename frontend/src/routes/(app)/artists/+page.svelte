<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search, SortDesc } from "lucide-svelte";
  import SEO from "$lib/components/SEO.svelte";

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
    { value: "any", label: "Any" },
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
      class="relative z-40 flex flex-col gap-4 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl"
    >
      <div class="flex items-end gap-4">
        <!-- Search -->
        <div class="relative group w-full">
          <label
            for="artist-search"
            class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
          >
            Search Artist
          </label>
          <div class="relative">
            <span
              class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
            >
              <Search size={20} />
            </span>
            <input
              id="artist-search"
              bind:value={searchQuery}
              oninput={handleInput}
              onkeydown={handleKeydown}
              class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
              placeholder="Who are you looking for?"
              type="text"
            />
          </div>
        </div>

        <!-- Sort -->
        <div class="w-50">
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

    <!-- Artists Grid -->
    <section class="">
      <div class="flex items-center justify-between mb-8">
        <h2 class="text-2xl font-bold flex items-center gap-3">
          <span class="w-2 h-8 bg-primary rounded-full"></span>
          Artists
          {#if data.artistsData?.pagination?.total > 0}
            <span class="text-white/30 font-normal text-lg ml-2"
              >({data.artistsData.pagination.total.toLocaleString()})</span
            >
          {/if}
        </h2>
      </div>

      {#if artists.length > 0}
        <div
          class="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-8"
        >
          {#each artists as artist}
            <div class="group flex flex-col items-center gap-4 cursor-pointer">
              <a
                href="/artists/{artist.slug}"
                class="relative w-full aspect-square rounded-full overflow-hidden card-shadow ring-4 ring-transparent group-hover:ring-primary/50 transition-all duration-300"
                title="View artist profile: {artist.name}"
              >
                <img
                  alt="Avatar for {artist.name}"
                  title="Avatar for {artist.name}"
                  class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                  src={artist.avatar_url || "/images/placeholders/default.jpg"}
                />
                <div
                  class="absolute inset-0 bg-linear-to-t from-primary/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
                ></div>
              </a>
              <div class="text-center">
                <a
                  href="/artists/{artist.slug}"
                  class="font-bold text-white group-hover:text-primary transition-colors text-lg line-clamp-1"
                  title="View artist profile: {artist.name}">{artist.name}</a
                >
                <div
                  class="mt-1.5 inline-flex items-center px-3 py-1 rounded-full bg-primary/20 text-primary text-[11px] font-bold border border-primary/20"
                >
                  {artist.enabled_songs || 0} Themes
                </div>
              </div>
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
          class="text-center py-20 bg-surface-darker/30 rounded-3xl border-2 border-dashed border-white/5"
        >
          <span
            class="material-symbols-outlined text-6xl text-white/10 mb-4 block"
            >person_off</span
          >
          <h3 class="text-xl font-bold text-white/40">No artists found</h3>
          <p class="text-white/20 mt-2">Try adjusting your search or filters</p>
        </div>
      {/if}
    </section>
  </div>
</main>
