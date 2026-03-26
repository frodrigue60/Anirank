<script lang="ts">
  import { fade } from "svelte/transition";
  import { Music } from "lucide-svelte";
  import SongCard from "$lib/components/SongCard.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let songs = $state<any[]>(data.songs?.data || []);
  // svelte-ignore state_referenced_locally
  let songsPage = $state(data.songs?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let songsLastPage = $state(data.songs?.pagination?.last_page || 1);
  let loading = $state(false);

  // svelte-ignore state_referenced_locally
  let _sourceSongs = data.songs?.data;

  // Sync state if navigation happens
  $effect(() => {
    if (_sourceSongs !== data.songs?.data && data.songs?.pagination?.current_page === 1) {
      _sourceSongs = data.songs?.data;
      songs = data.songs?.data || [];
      songsPage = data.songs?.pagination?.current_page || 1;
      songsLastPage = data.songs?.pagination?.last_page || 1;
    }
  });

  async function loadMoreSongs() {
    if (loading || songsPage >= songsLastPage || !data.profile) return;

    loading = true;
    try {
      const nextPage = songsPage + 1;
      const response = await api.post(`/users/favorites/themes`, {
        user_uuid: data.profile.uuid,
        page: nextPage,
      });

      const songsData = response.data.songs || response.data;
      if (songsData?.data) {
        songs = [...songs, ...songsData.data];
        songsPage = songsData.pagination?.current_page || songsData.current_page;
        songsLastPage = songsData.pagination?.last_page || songsLastPage;
      }
    } catch (e) {
      console.error("Error loading more songs", e);
    } finally {
      loading = false;
    }
  }
</script>

<section in:fade={{ duration: 200 }}>
  {#if data.profile}
    <div class="flex items-center justify-between mb-12">
      <div>
        <h3 class="text-3xl font-black text-white tracking-tight leading-tight">
          Favorite <span class="text-primary italic">Themes</span>
        </h3>
        <p class="text-white/40 mt-2 font-medium">
          All themes favorited by {data.profile.name}.
        </p>
      </div>
    </div>

    {#if songs.length > 0}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
        {#each songs as song}
          <SongCard {song} />
        {/each}
      </div>

      <InfiniteScroll
        hasMore={songsPage < songsLastPage}
        {loading}
        onLoadMore={loadMoreSongs}
      />
    {:else}
      <div
        class="py-20 flex flex-col items-center justify-center text-center opacity-40"
      >
        <Music size={80} strokeWidth={1} />
        <h2 class="text-2xl font-bold mt-6">No favorites found</h2>
      </div>
    {/if}
  {/if}
</section>
