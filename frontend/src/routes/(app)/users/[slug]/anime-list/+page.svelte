<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import api from "$lib/api";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import RequestModal from "$lib/components/RequestModal.svelte";
  import { Loader2, Plus, ExternalLink, MessageSquarePlus } from "lucide-svelte";
  import { authState } from "$lib/state/auth.svelte";

  let { data } = $props();

  let items = $state<any[]>([]);
  let page = $state(1);
  let total = $state(0);
  let hasMore = $state(true);
  let isLoading = $state(false);
  let status = $state("ALL");
  let isNotLinked = $state(false);
  let isFetching = false; // Internal non-reactive lock

  const isOwnProfile = $derived(
    authState.user && authState.user.uuid === data.profile?.uuid,
  );

  let showRequestModal = $state(false);
  let requestInitialTitle = $state("");
  let requestInitialContent = $state("");

  const statuses = [
    { value: "ALL", label: "All" },
    { value: "CURRENT", label: "Watching" },
    { value: "PLANNING", label: "Planning" },
    { value: "COMPLETED", label: "Completed" },
    { value: "DROPPED", label: "Dropped" },
    { value: "PAUSED", label: "Paused" },
  ];

  async function fetchItems(reset = false) {
    if (isFetching) return;
    if (!reset && !hasMore) return;

    if (reset) {
      items = [];
      page = 1;
      hasMore = true;
      isNotLinked = false;
    }

    isFetching = true;
    isLoading = true;
    try {
      const response = await api.get(`/users/${data.profile.slug}/anilist-list`, {
        params: {
          status: status === "ALL" ? "" : status,
          page: page,
          limit: 50,
        },
      });

      const newItems = response.data.data;
      items = reset ? newItems : [...items, ...newItems];
      total = response.data.pagination.total;
      hasMore = response.data.pagination.has_more;
      page++;
    } catch (error: any) {
      if (error.response?.status === 400) {
        isNotLinked = true;
      }
      console.error("Error fetching AniList list:", error);
      hasMore = false;
    } finally {
      isLoading = false;
      // Small timeout to allow DOM to settle before releasing the lock
      setTimeout(() => {
        isFetching = false;
      }, 100);
    }
  }

  function handleStatusChange(newStatus: string) {
    if (status === newStatus) return;
    status = newStatus;
    fetchItems(true);
  }

  function openRequest(animeTitle: string) {
    requestInitialTitle = animeTitle;
    requestInitialContent = `${authState.user?.name || "Un usuario"} solicita los temas de ${animeTitle}`;
    showRequestModal = true;
  }

  onMount(() => {
    fetchItems();
  });
</script>

<div class="space-y-8" in:fade>
  <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
    <div>
      <h2 class="text-2xl font-black text-white tracking-tight uppercase">AniList Sync</h2>
      <p class="text-sm text-white/50">Synchronized collection from AniList profile</p>
    </div>

    <!-- Status Tabs -->
    <div class="flex bg-white/5 p-1 rounded-2xl border border-white/10 overflow-x-auto no-scrollbar max-w-full">
      {#each statuses as s}
        <button
          onclick={() => handleStatusChange(s.value)}
          class="px-4 py-2 rounded-xl text-[10px] font-bold uppercase tracking-wider transition-all whitespace-nowrap {status === s.value ? 'bg-primary text-white shadow-lg shadow-primary/20' : 'text-white/40 hover:text-white/60'}"
        >
          {s.label}
        </button>
      {/each}
    </div>
  </div>

  {#if isNotLinked}
    <div class="py-20 flex flex-col items-center justify-center text-center space-y-6" in:fade>
      <div class="w-24 h-24 bg-primary/10 rounded-full flex items-center justify-center text-primary shadow-2xl shadow-primary/20">
        <span class="material-symbols-outlined text-4xl">link_off</span>
      </div>
      <div class="space-y-2 max-w-sm px-4">
        <h3 class="text-2xl font-black text-white uppercase tracking-tighter">Sync Required</h3>
        <p class="text-sm text-white/50 leading-relaxed uppercase tracking-tighter">
          This user has not linked their AniList account yet. Syncing allows you to cross-reference their AniList collection with our library.
        </p>
      </div>
      {#if isOwnProfile}
        <a href="/settings" class="bg-primary hover:bg-primary/80 text-white px-8 py-3.5 rounded-2xl font-black text-[10px] uppercase tracking-widest transition-all shadow-xl shadow-primary/20 hover:scale-105 active:scale-95">
          Link AniList Now
        </a>
      {/if}
    </div>
  {:else if items.length === 0 && !isLoading}
    <div class="py-20 flex flex-col items-center justify-center text-center space-y-4" in:fade>
      <div class="w-20 h-20 bg-white/5 rounded-full flex items-center justify-center text-white/20">
        <Plus size={40} />
      </div>
      <h3 class="text-xl font-bold text-white">No series found</h3>
      <p class="text-sm text-white/40 max-w-xs">No series were found in this category on the user's AniList profile.</p>
    </div>
  {:else}
    <!-- Grid -->
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 md:gap-6">
      {#each items as item (item.anilist_id)}
        <div class="group relative" in:fade>
          <div class="aspect-2/3 rounded-2xl md:rounded-3xl overflow-hidden bg-white/5 border border-white/10 transition-all duration-500 hover:scale-[1.02] hover:shadow-2xl hover:shadow-primary/20 relative">
            <!-- Cover Image with Grayscale logic -->
            <img
              src={item.cover_image}
              alt={item.title}
              class="w-full h-full object-cover transition-all duration-700 {item.is_in_db ? 'grayscale-0' : 'grayscale brightness-75 group-hover:grayscale-0 group-hover:brightness-100'}"
              loading="lazy"
            />

            <!-- Overlay for nonexistent -->
            {#if !item.is_in_db}
              <div class="absolute inset-0 bg-black/70 flex flex-col items-center justify-center p-4 opacity-100 group-hover:opacity-0 transition-opacity">
                <div class="bg-zinc-900 border border-white/10 rounded-full px-3 py-1 flex items-center gap-1.5 shadow-xl">
                  <div class="w-1.5 h-1.5 rounded-full bg-white/50"></div>
                  <span class="text-[10px] font-bold text-white uppercase tracking-wider">Not in DB</span>
                </div>
              </div>
            {/if}

            <!-- Bottom Actions Overlay -->
            <div class="absolute inset-0 bg-linear-to-t from-black/90 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-all duration-300 flex flex-col justify-end p-4">
              {#if item.is_in_db}
                <a
                  href={`/animes/${item.anime_slug}`}
                  class="w-full bg-primary hover:bg-primary/80 text-white py-2.5 rounded-xl text-[10px] font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-all shadow-lg shadow-primary/30"
                >
                  View Details
                  <ExternalLink size={14} />
                </a>
              {:else}
                <button
                  onclick={() => openRequest(item.title)}
                  class="w-full bg-zinc-800 hover:bg-zinc-700 border border-white/10 text-white py-2.5 rounded-xl text-[10px] font-bold uppercase tracking-wider flex items-center justify-center gap-2 transition-all shadow-xl"
                >
                  Request Series
                  <MessageSquarePlus size={14} />
                </button>
              {/if}
            </div>

            <!-- Status Badge in corner -->
            <div class="absolute top-3 right-3 px-2 py-1 rounded-lg bg-black/90 border border-white/10 text-[9px] font-bold text-white uppercase tracking-tighter opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap shadow-lg">
               {item.status}
            </div>
          </div>

          <div class="mt-3 px-1">
            <h4 class="text-sm font-bold text-white line-clamp-1 group-hover:text-primary transition-colors">{item.title}</h4>
            <div class="flex items-center gap-2 mt-1 text-[10px] font-bold text-white/30 uppercase tracking-widest">
              <span>{item.format}</span>
              {#if item.season_year}
                <span class="w-1 h-1 rounded-full bg-white/10"></span>
                <span>{item.season_year}</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>

    <!-- Infinite Scroll Trigger -->
    <InfiniteScroll
      {hasMore}
      loading={isLoading}
      onLoadMore={() => fetchItems()}
    />
  {/if}

  {#if isLoading && items.length > 0}
    <div class="flex justify-center py-10">
      <Loader2 class="animate-spin text-primary" size={32} />
    </div>
  {/if}
</div>

<RequestModal
  show={showRequestModal}
  onClose={() => (showRequestModal = false)}
  initialTitle={requestInitialTitle}
  initialContent={requestInitialContent}
/>

<style lang="postcss">
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
