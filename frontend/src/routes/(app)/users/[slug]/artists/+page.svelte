<script lang="ts">
  import { fade } from "svelte/transition";
  import { User as UserIcon } from "lucide-svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let artists = $state<any[]>(data.artists?.data || []);
  // svelte-ignore state_referenced_locally
  let artistsPage = $state(data.artists?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let artistsLastPage = $state(data.artists?.last_page || 1);
  let loading = $state(false);

  // svelte-ignore state_referenced_locally
  let _sourceArtists = data.artists?.data;

  // Sync state if navigation happens
  $effect(() => {
    if (
      _sourceArtists !== data.artists?.data &&
      data.artists?.current_page === 1
    ) {
      _sourceArtists = data.artists?.data;
      artists = data.artists?.data || [];
      artistsPage = data.artists?.current_page || 1;
      artistsLastPage = data.artists?.last_page || 1;
    }
  });

  async function loadMoreArtists() {
    if (loading || artistsPage >= artistsLastPage || !data.profile) return;

    loading = true;
    try {
      const nextPage = artistsPage + 1;
      const response = await api.post(`/users/artists/favorites`, {
        user_id: data.profile.id,
        page: nextPage,
      });

      if (response.data.artists) {
        artists = [...artists, ...response.data.artists.data];
        artistsPage = response.data.artists.current_page;
        artistsLastPage = response.data.artists.last_page;
      }
    } catch (e) {
      console.error("Error loading more artists", e);
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
          Favorite <span class="text-primary italic">Artists</span>
        </h3>
        <p class="text-white/40 mt-2 font-medium">
          Artists curated by {data.profile.name}.
        </p>
      </div>
    </div>

    {#if artists.length > 0}
      <div
        class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-8 mb-12"
      >
        {#each artists as artist}
          <a
            href={`/artists/${artist.slug}`}
            class="flex flex-col items-center group cursor-pointer"
          >
            <div
              class="size-24 sm:size-32 rounded-full border-2 border-primary/30 p-1 group-hover:border-primary transition-all mb-4 bg-surface-darker/50"
            >
              <!-- svelte-ignore a11y_missing_attribute -->
              <img
                class="h-full w-full rounded-full object-cover grayscale group-hover:grayscale-0 transition-all duration-500"
                data-alt="Artist portrait"
                src={artist.avatar_url ||
                  "https://placehold.co/200x200/2a2136/white?text=Artist"}
              />
            </div>
            <h5
              class="text-slate-100 font-bold text-center text-sm md:text-base line-clamp-1"
            >
              {artist.name}
            </h5>
            <p class="text-slate-500 text-xs mt-1">
              {artist.songs_count || 0} Themes
            </p>
          </a>
        {/each}
      </div>

      <InfiniteScroll
        hasMore={artistsPage < artistsLastPage}
        {loading}
        onLoadMore={loadMoreArtists}
      />
    {:else}
      <div
        class="py-20 flex flex-col items-center justify-center text-center opacity-40"
      >
        <UserIcon size={80} strokeWidth={1} />
        <h2 class="text-2xl font-bold mt-6">No artists found</h2>
      </div>
    {/if}
  {/if}
</section>
