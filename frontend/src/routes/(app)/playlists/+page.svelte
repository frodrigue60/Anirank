<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import { Search, Music, User } from "lucide-svelte";

  let { data }: { data: any } = $props();

  // svelte-ignore state_referenced_locally
  let playlists = $state(data.playlists?.data || []);
  // svelte-ignore state_referenced_locally
  let currentPage = $state(data.playlists?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let lastPage = $state(data.playlists?.pagination?.last_page || 1);
  let loading = $state(false);
  let params = $state({ name: "" });

  $effect(() => {
    params.name = data.params.name || "";

    // Reset infinite scroll on data change (filters)
    if (data.playlists && data.playlists.pagination?.current_page === 1) {
      playlists = data.playlists.data;
      currentPage = data.playlists.pagination.current_page;
      lastPage = data.playlists.pagination.last_page;
    }
  });

  function updateFilters() {
    const url = new URL(page.url);

    if (params.name) url.searchParams.set("name", params.name);
    else url.searchParams.delete("name");

    url.searchParams.set("page", "1");
    goto(url.toString(), { keepFocus: true });
  }

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/playlists", {
        params: {
          ...data.params,
          page: nextPage,
        },
      });

      if (response.data.data) {
        playlists = [...playlists, ...response.data.data];
        currentPage = response.data.pagination.current_page;
        lastPage = response.data.pagination.last_page;
      }
    } catch (e) {
      console.error("Error loading more playlists", e);
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
</script>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <!-- Header -->
  <div class="mb-10 text-center lg:text-left">
    <h1 class="text-4xl font-black text-white mb-2 tracking-tight">
      Public Playlists
    </h1>
    <p class="text-slate-400 text-lg">Explore community curated collections</p>
  </div>

  <!-- Filter Row -->
  <section
    class="relative z-40 flex flex-col gap-4 bg-surface-dark/30 p-4 rounded-3xl border border-white/5 shadow-2xl mb-10"
  >
    <div class="relative group">
      <label
        for="playlist-search"
        class="block text-[10px] uppercase font-black text-white/40 mb-2 ml-1 tracking-widest"
      >
        Search Playlists
      </label>
      <div class="relative">
        <span
          class="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors"
        >
          <Search size={20} />
        </span>
        <input
          id="playlist-search"
          type="text"
          bind:value={params.name}
          oninput={handleInput}
          onkeydown={handleKeydown}
          class="w-full h-12 bg-surface-darker/50 border border-white/10 rounded-xl pl-12 pr-6 text-sm text-white focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 placeholder:text-white/20 transition-all"
          placeholder="Search playlists (e.g. My favorites, Top Openings)..."
        />
      </div>
    </div>
  </section>

  <!-- Playlist Grid -->
  {#if playlists.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each playlists as playlist}
        <a
          href="/playlists/{playlist.id}"
          class="group relative overflow-hidden rounded-xl bg-slate-800 aspect-video border border-transparent hover:border-primary/50 transition-all cursor-pointer shadow-lg shadow-black/20"
        >
          <!-- Background Image -->
          {#if playlist.banner_url}
            <div
              class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
              style="background-image: url('{playlist.banner_url}'); filter:brightness(0.5)"
            ></div>
          {:else}
            <div
              class="absolute inset-0 bg-slate-700 transition-transform duration-500 group-hover:scale-105"
            ></div>
          {/if}
          <div
            class="absolute inset-0 bg-linear-to-t from-background-dark/95 via-background-dark/40 to-transparent"
          ></div>

          <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col gap-1">
            <div class="flex items-center gap-2 mb-1">
              <User size={12} class="text-primary" />
              <span
                class="text-[10px] text-slate-400 uppercase font-black tracking-widest"
              >
                {playlist.user?.name || "User"}
              </span>
            </div>
            <h3
              class="text-2xl font-bold text-white group-hover:text-primary transition-colors"
            >
              {playlist.name}
            </h3>

            <div
              class="mt-4 flex items-center justify-between border-t border-white/10 pt-4"
            >
              <div class="flex items-center gap-2 text-white/60">
                <Music size={14} />
                <span class="text-sm font-semibold"
                  >{playlist.song_count || 0} Songs</span
                >
              </div>
              <span class="text-xs text-slate-500">Public Collection</span>
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
    <div class="text-center py-20">
      <Search size={48} class="text-white/10 mx-auto mb-4" />
      <p class="text-white/40">No playlists found matching your criteria.</p>
    </div>
  {/if}
</main>

<style>
</style>
