<script lang="ts">
  import {
    X,
    Plus,
    Music,
    Loader2,
    Check,
    CheckCircle2,
    ListMusic,
  } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale, slide } from "svelte/transition";
  import CreatePlaylistModal from "./CreatePlaylistModal.svelte";

  interface Playlist {
    id: number;
    name: string;
    description: string | null;
    song_count: number;
    contains_song: boolean;
  }

  interface Props {
    show: boolean;
    song: any;
    onClose: () => void;
  }

  let { show, song, onClose }: Props = $props();

  let playlists: Playlist[] = $state([]);
  let isLoading = $state(false);
  let errorMessage = $state("");
  let showCreateModal = $state(false);
  let successMessage = $state("");

  async function fetchPlaylists() {
    if (!song) return;
    isLoading = true;
    errorMessage = "";
    try {
      const response = await api.get(`/me/playlists?song_id=${song.id}`);
      playlists = response.data.playlists;
    } catch (e: any) {
      errorMessage = "Failed to load playlists.";
      console.error(e);
    } finally {
      isLoading = false;
    }
  }

  async function toggleSong(playlist: Playlist) {
    try {
      if (playlist.contains_song) {
        // Remove song from playlist
        await api.delete(`/playlists/${playlist.id}/songs/${song.id}`);
        playlist.contains_song = false;
        playlist.song_count = playlist.song_count - 1;
        successMessage = "Song removed from playlist";
      } else {
        // Add song to playlist
        await api.post(`/playlists/${playlist.id}/songs`, {
          song_id: song.id,
        });
        playlist.contains_song = true;
        playlist.song_count = playlist.song_count + 1;
        successMessage = "Song added to playlist";
      }
      setTimeout(() => {
        successMessage = "";
      }, 3000);
    } catch (e) {
      console.error("Error toggling song in playlist", e);
    }
  }

  function handleCreated(newPlaylist: any) {
    // Refresh playlists after creating one
    fetchPlaylists();
  }

  $effect(() => {
    if (show) {
      fetchPlaylists();
    }
  });

  function handleClose() {
    onClose();
    successMessage = "";
    errorMessage = "";
  }
</script>

{#if show}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-100 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
    onclick={handleClose}
    transition:fade={{ duration: 200 }}
  >
    <!-- Modal Content -->
    <div
      class="modal-glass w-full max-w-sm rounded-4xl overflow-hidden shadow-2xl p-8 flex flex-col items-center text-center relative"
      onclick={(e) => e.stopPropagation()}
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Header -->
      <div class="w-full flex justify-between items-start mb-6">
        <div class="text-left">
          <div class="flex items-center gap-2 text-primary mb-1">
            <ListMusic size={14} />
            <p class="text-[10px] font-bold uppercase tracking-[0.2em]">
              Add to Collection
            </p>
          </div>
          <h3 class="text-xl font-bold leading-tight tracking-tight">
            Your Playlists
          </h3>
          <p class="text-xs text-white/50 mt-1">
            Select where to add this theme.
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-white/5 flex items-center justify-center transition-colors text-white/40 hover:text-white"
        >
          <X size={18} />
        </button>
      </div>

      <div
        class="w-full flex-1 overflow-y-auto max-h-[350px] pr-2 custom-scrollbar space-y-2 text-left mb-6"
      >
        {#if isLoading}
          <div
            class="py-12 flex flex-col items-center justify-center text-white/20"
          >
            <Loader2 class="animate-spin mb-2" size={32} />
            <span class="text-xs font-bold uppercase tracking-widest"
              >Loading...</span
            >
          </div>
        {:else if playlists.length === 0}
          <div
            class="py-12 flex flex-col items-center justify-center text-center space-y-4"
          >
            <div
              class="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center text-white/20"
            >
              <Music size={32} />
            </div>
            <div>
              <p class="text-sm font-bold text-white/80">No playlists found</p>
              <p class="text-xs text-white/40">
                Create your first collection below.
              </p>
            </div>
          </div>
        {:else}
          {#each playlists as playlist}
            <button
              onclick={() => toggleSong(playlist)}
              class="w-full group flex items-center justify-between p-4 rounded-2xl bg-white/5 border border-white/5 hover:bg-white/10 hover:border-white/10 transition-all text-left"
            >
              <div class="flex items-center gap-4">
                <div
                  class="w-10 h-10 rounded-xl bg-surface-dark flex items-center justify-center text-white/40 group-hover:text-primary transition-colors"
                >
                  <ListMusic size={20} />
                </div>
                <div>
                  <h4
                    class="text-sm font-bold text-white group-hover:text-primary transition-colors"
                  >
                    {playlist.name}
                  </h4>
                  <p
                    class="text-[10px] text-white/40 font-medium uppercase tracking-wider"
                  >
                    {playlist.song_count} themes
                  </p>
                </div>
              </div>

              {#if playlist.contains_song}
                <div
                  class="w-6 h-6 rounded-full bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/20"
                  in:scale
                >
                  <Check size={14} strokeWidth={4} />
                </div>
              {:else}
                <div
                  class="w-6 h-6 rounded-full border border-white/10 group-hover:border-primary/50 flex items-center justify-center transition-colors"
                >
                  <Plus
                    size={14}
                    class="text-white/20 group-hover:text-primary"
                  />
                </div>
              {/if}
            </button>
          {/each}
        {/if}
      </div>

      <!-- Footer Action -->
      <div class="w-full space-y-4">
        {#if successMessage}
          <div
            class="bg-primary/10 border border-primary/20 rounded-xl p-3 flex items-center gap-3 text-primary"
            transition:slide
          >
            <CheckCircle2 size={16} />
            <span class="text-xs font-bold">{successMessage}</span>
          </div>
        {/if}

        <button
          onclick={() => (showCreateModal = true)}
          class="w-full bg-white text-black py-4 rounded-xl font-bold text-sm transition-all flex items-center justify-center gap-2 hover:bg-gray-200 active:scale-95"
        >
          <Plus size={16} />
          Create New Playlist
        </button>
      </div>
    </div>
  </div>
{/if}

<CreatePlaylistModal
  show={showCreateModal}
  onClose={() => (showCreateModal = false)}
  onCreated={handleCreated}
/>

<style lang="postcss">
  .modal-glass {
    background: rgba(25, 16, 34, 0.9);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: var(--primary);
  }
</style>
