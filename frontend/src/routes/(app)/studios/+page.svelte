<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import CustomSelect from "$lib/components/CustomSelect.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search, SortDesc } from "lucide-svelte";

  let { data }: { data: any } = $props();

  // svelte-ignore state_referenced_locally
  let studios = $state(data.studios?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.studios?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.studios?.last_page || 1);
  let loading = $state(false);
  let params = $state({ name: "", sort: "" });

  $effect(() => {
    params.name = data.params.name || "";
    params.sort = data.params.sort || "";

    // Reset infinite scroll on data change (filters)
    if (data.studios && data.studios.current_page === 1) {
      studios = data.studios.data;
      currentPage = data.studios.current_page;
      lastPage = data.studios.last_page;
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    const setParam = (key: string, val: string) => {
      if (val && val !== "any") url.searchParams.set(key, val);
      else url.searchParams.delete(key);
    };

    if (params.name) url.searchParams.set("name", params.name);
    else url.searchParams.delete("name");

    setParam("sort", params.sort);

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true });
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/studios", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        studios = [...studios, ...response.data.data];
        currentPage = response.data.current_page;
        lastPage = response.data.last_page;
      }
    } catch (e) {
      console.error("Error loading more studios", e);
    } finally {
      loading = false;
    }
  }

  let searchTimeout: any;
  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      if (params.name.length === 0 || params.name.length >= 2) {
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

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <!-- Filter Row -->
  <section
    class="relative z-40 flex flex-col gap-4 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 backdrop-blur-md shadow-2xl mb-10"
  >
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">
      <!-- Search -->
      <div class="lg:col-span-8 relative group">
        <label
          for="studio-search"
          class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
        >
          Search Studio
        </label>
        <div class="relative">
          <span
            class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
          >
            <Search size={20} />
          </span>
          <input
            id="studio-search"
            type="text"
            bind:value={params.name}
            oninput={handleInput}
            onkeydown={handleKeydown}
            class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
            placeholder="Search animation studios (e.g. MAPPA, Ufotable)..."
          />
        </div>
      </div>

      <!-- Sort -->
      <div class="lg:col-span-4">
        <CustomSelect
          label="Sort"
          bind:value={params.sort}
          options={sortOptions}
          placeholder="Any"
          icon={SortDesc}
          onchange={updateFilters}
        />
      </div>
    </div>
  </section>

  <!-- Studio Grid (3 Columns) -->
  {#if studios.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each studios as studio}
        <a
          href="/studios/{studio.slug}"
          class="group relative overflow-hidden rounded-xl bg-slate-800 aspect-video border border-transparent hover:border-primary/50 transition-all cursor-pointer shadow-lg shadow-black/20"
        >
          <!-- Background Image (Banner) -->

          <div
            class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
            style="background-image: url('{studio.banner_url ??
              '/images/placeholders/default-banner.jpg'}'); filter:brightness(0.5)"
          ></div>

          <div
            class="absolute inset-0 bg-linear-to-t from-background-dark/95 via-background-dark/40 to-transparent"
          ></div>

          <!-- LOGO -->
          <!-- <div
            class="absolute inset-0 flex items-center justify-center p-6 opacity-20 group-hover:opacity-40 transition-opacity pointer-events-none"
          >
            {#if studio.logo_url}
              <img
                src={studio.logo_url}
                alt=""
                class="max-w-[120px] max-h-[80px] object-contain"
              />
            {/if}
          </div> -->

          <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col gap-1">
            <h3
              class="text-2xl font-bold text-white group-hover:text-primary transition-colors"
            >
              {studio.name}
            </h3>
            <!-- {#if studio.featured_title}
                <p class="text-slate-300 text-sm font-medium">
                  Featured: <span class="text-white"
                    >{studio.featured_title}</span
                  >
                </p>
              {/if} -->
            <div
              class="mt-4 flex items-center justify-between border-t border-white/10 pt-4"
            >
              <div class="flex flex-col">
                <span
                  class="text-[10px] text-slate-400 uppercase font-bold tracking-wider"
                  >Produced</span
                >
                <span class="text-white text-sm font-semibold"
                  >{studio.anime_count || 0} Series</span
                >
              </div>
            </div>
          </div>
        </a>
      {/each}
    </div>
  {:else}
    <div class="text-center py-20">
      <span
        class="material-symbols-outlined text-6xl text-slate-300 dark:text-slate-700 mb-4"
        >search_off</span
      >
      <p class="text-slate-500 dark:text-slate-400">
        No studios found matching your criteria.
      </p>
    </div>
  {/if}

  <InfiniteScroll
    hasMore={currentPage < lastPage}
    {loading}
    onLoadMore={loadMore}
  />
</main>

<style>
</style>
