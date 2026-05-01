<script lang="ts">
  import { fade } from "svelte/transition";
  import Search from "lucide-svelte/icons/search";
  import Music from "lucide-svelte/icons/music";
  import Plus from "lucide-svelte/icons/plus";
  import PlaylistCard from "$lib/components/PlaylistCard.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import CreatePlaylistModal from "$lib/components/CreatePlaylistModal.svelte";
  import EditPlaylistModal from "$lib/components/EditPlaylistModal.svelte";
  import api from "$lib/api";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let playlists = $state<any[]>(data.playlists);
  let searchQuery = $state("");
  let showCreateModal = $state(false);
  let isEditModalOpen = $state(false);
  let playlistToEdit = $state<any>(null);

  // Non-reactive guard for prop changes
  // svelte-ignore state_referenced_locally
  let _sourcePlaylists = data.playlists;

  // Sync state if navigation happens or props change
  $effect(() => {
    if (_sourcePlaylists !== data.playlists) {
      _sourcePlaylists = data.playlists;
      playlists = data.playlists;
    }
  });

  let isOwner = $derived(
    authState.user && data.profile && authState.user.uuid === data.profile.uuid,
  );

  // Sync private playlists on the client if owner and data wasn't already loaded as private
  let initializedOwnerFetch = $state(false);
  $effect(() => {
    if (isOwner && !initializedOwnerFetch && data.profile) {
      initializedOwnerFetch = true;
      api
        .get(`/users/${data.profile.slug}/playlists`)
        .then((res) => {
          if (res.data.playlists) {
            playlists = res.data.playlists;
          }
        })
        .catch((err) =>
          console.error("Could not fetch private playlists", err),
        );
    }
  });

  function openEditModal(playlist: any) {
    playlistToEdit = playlist;
    isEditModalOpen = true;
  }

  function onTogglePrivacy(id: number, is_public: boolean) {
    playlists = playlists.map((p) => (p.id === id ? { ...p, is_public } : p));
  }

  async function handleDelete(id: number) {
    if (!confirm("Are you sure you want to delete this playlist?")) return;
    try {
      await api.delete(`/playlists/${id}`);
      playlists = playlists.filter((p) => p.id !== id);
    } catch (e) {
      console.error("Failed to delete playlist", e);
    }
  }

  let filteredPlaylists = $derived(
    playlists.filter(
      (p: any) =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (p.description &&
          p.description.toLowerCase().includes(searchQuery.toLowerCase())),
    ),
  );
</script>

<section in:fade={{ duration: 200 }}>
  {#if data.profile}
    <div
      class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12"
    >
      <div>
        <h3 class="text-3xl font-black text-on-surface tracking-tight leading-tight">
          {data.profile.name}'s
          <span class="text-primary italic">Playlists</span>
        </h3>
        <p class="text-on-surface-variant mt-2 font-medium">
          Browse collections curated by {data.profile.name}.
        </p>
      </div>

      <div
        class="flex flex-col sm:flex-row items-center gap-4 w-full max-w-2xl"
      >
        <div class="relative flex-1 w-full">
          <span
            class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant/40"
          >
            <Search size={20} />
          </span>
          <input
            type="text"
            bind:value={searchQuery}
            placeholder="Search playlists..."
            class="w-full bg-surface-low border border-on-surface-variant/10 rounded-md py-4 pl-12 pr-6 text-on-surface placeholder:text-on-surface-variant/40 focus:outline-hidden focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all font-medium"
          />
        </div>

        {#if isOwner}
          <button
            onclick={() => (showCreateModal = true)}
            class="bg-primary hover:opacity-90 text-white px-8 py-4 rounded-sm text-sm font-black uppercase tracking-widest transition-all flex items-center gap-2 whitespace-nowrap shadow-lg shadow-primary/20 active:scale-95"
          >
            <Plus size={16} strokeWidth={3} />
            New Playlist
          </button>
        {/if}
      </div>
    </div>

    {#if filteredPlaylists.length > 0}
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredPlaylists as playlist (playlist.id)}
          <PlaylistCard
            {playlist}
            profile={data.profile}
            {openEditModal}
            {handleDelete}
            {onTogglePrivacy}
          />
        {/each}
      </div>
    {:else}
      <EmptyState
        title={searchQuery ? "No playlists found" : "No playlists yet"}
        message={searchQuery 
          ? `No playlists matching "${searchQuery}" found.` 
          : `${data.profile.name} hasn't made any public playlists yet.`}
        icon={Music}
        actionLabel={isOwner && !searchQuery ? "Create First Playlist" : searchQuery ? "Clear Search" : ""}
        onAction={() => {
          if (isOwner && !searchQuery) showCreateModal = true;
          else if (searchQuery) searchQuery = "";
        }}
      />
    {/if}
  {/if}
</section>

{#if showCreateModal}
  <CreatePlaylistModal
    show={showCreateModal}
    onClose={() => (showCreateModal = false)}
    onCreated={(newPlaylist) => {
      playlists = [newPlaylist, ...playlists];
      showCreateModal = false;
    }}
  />
{/if}

{#if isEditModalOpen}
  <EditPlaylistModal
    show={isEditModalOpen}
    playlist={playlistToEdit}
    onClose={() => (isEditModalOpen = false)}
    onUpdated={(updatedPlaylist) => {
      playlists = playlists.map((p) =>
        p.id === updatedPlaylist.id ? updatedPlaylist : p,
      );
      isEditModalOpen = false;
    }}
  />
{/if}
