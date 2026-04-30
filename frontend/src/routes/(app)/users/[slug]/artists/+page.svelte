<script lang="ts">
  import { fade } from "svelte/transition";
  import UserIcon from "lucide-svelte/icons/user";;
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import api from "$lib/api";
  import ArtistAvatarCard from "$lib/components/ArtistAvatarCard.svelte";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let artists = $state<any[]>(data.artists?.data || []);
  // svelte-ignore state_referenced_locally
  let artistsPage = $state(data.artists?.pagination?.current_page || 1);
  // svelte-ignore state_referenced_locally
  let artistsLastPage = $state(data.artists?.pagination?.last_page || 1);
  let loading = $state(false);

  // svelte-ignore state_referenced_locally
  let _sourceArtists = data.artists?.data;

  // Sync state if navigation happens
  $effect(() => {
    if (
      _sourceArtists !== data.artists?.data &&
      data.artists?.pagination?.current_page === 1
    ) {
      _sourceArtists = data.artists?.data;
      artists = data.artists?.data || [];
      artistsPage = data.artists?.pagination?.current_page || 1;
      artistsLastPage = data.artists?.pagination?.last_page || 1;
    }
  });

  async function loadMoreArtists() {
    if (loading || artistsPage >= artistsLastPage || !data.profile) return;

    loading = true;
    try {
      const nextPage = artistsPage + 1;
      const response = await api.post(`/users/favorites/artists`, {
        user_uuid: data.profile.uuid || data.profile.id,
        page: nextPage,
      });

      const producersData = response.data.artists || response.data;
      if (producersData?.data) {
        artists = [...artists, ...producersData.data];
        artistsPage =
          producersData.pagination?.current_page || producersData.current_page;
        artistsLastPage =
          producersData.pagination?.last_page || artistsLastPage;
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
        <h3 class="text-3xl font-black text-on-surface tracking-tight leading-tight">
          Favorite <span class="text-primary italic">Artists</span>
        </h3>
        <p class="text-on-surface-variant mt-2 font-medium">
          Artists curated by {data.profile.name}.
        </p>
      </div>
    </div>

    {#if artists.length > 0}
      <div
        class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-8 mb-12"
      >
        {#each artists as artist}
          <ArtistAvatarCard {artist} />
        {/each}
      </div>

      <InfiniteScroll
        hasMore={artistsPage < artistsLastPage}
        {loading}
        onLoadMore={loadMoreArtists}
      />
    {:else}
      <div
        class="py-20 flex flex-col items-center justify-center text-center text-on-surface-variant/40"
      >
        <UserIcon size={80} strokeWidth={1} />
        <h2 class="text-2xl font-bold mt-6">No artists found</h2>
      </div>
    {/if}
  {/if}
</section>
