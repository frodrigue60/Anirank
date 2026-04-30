<script lang="ts">
  import X from "lucide-svelte/icons/x";
import Plus from "lucide-svelte/icons/plus";
import Music from "lucide-svelte/icons/music";
import Loader2 from "lucide-svelte/icons/loader-2";
import Check from "lucide-svelte/icons/check";
import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
import ListMusic from "lucide-svelte/icons/list-music";;
  import api from "$lib/api";
  import { fade, scale, slide } from "svelte/transition";
  import CreatePlaylistModal from "./CreatePlaylistModal.svelte";

  interface Playlist {
    id: string;
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
      playlists = response.data.data;
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
    class="fixed inset-0 z-100 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
    onclick={handleClose}
    transition:fade={{ duration: 200 }}
  >
    <!-- Modal Content -->
    <div
      class="modal-glass w-full max-w-sm rounded-md overflow-hidden shadow-2xl p-10 flex flex-col items-center text-center relative max-h-[90vh]"
      onclick={(e) => e.stopPropagation()}
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Header -->
      <div class="w-full flex justify-between items-start mb-6">
        <div class="text-left">
          <div class="flex items-center gap-2 text-primary mb-1">
            <span class="material-symbols-outlined text-[14px]"
              >library_music</span
            >
            <p class="text-[10px] font-black uppercase tracking-[0.2em]">
              Add to Collection
            </p>
          </div>
          <h3
            class="text-xl font-bold leading-tight tracking-tight text-on-surface"
          >
            Your Playlists
          </h3>
          <p class="text-xs text-on-surface-variant mt-1">
            Select where to add this theme.
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
        >
          <X size={18} />
        </button>
      </div>

      <div
        class="w-full flex-1 overflow-y-auto max-h-[350px] pr-2 custom-scrollbar space-y-2 text-left mb-6"
      >
        {#if isLoading}
          <div
            class="py-12 flex flex-col items-center justify-center text-on-surface-variant/20"
          >
            <Loader2 class="animate-spin mb-2" size={32} />
            <span class="text-[10px] font-black uppercase tracking-widest"
              >Loading...</span
            >
          </div>
        {:else if playlists.length === 0}
          <div
            class="py-12 flex flex-col items-center justify-center text-center space-y-4"
          >
            <div
              class="w-16 h-16 bg-surface-highest rounded-full flex items-center justify-center text-on-surface-variant/20"
            >
              <Music size={32} />
            </div>
            <div>
              <p class="text-sm font-bold text-on-surface/80">
                No playlists found
              </p>
              <p class="text-xs text-on-surface-variant">
                Create your first collection below.
              </p>
            </div>
          </div>
        {:else}
          {#each playlists as playlist}
            <button
              onclick={() => toggleSong(playlist)}
              class="w-full group flex items-center justify-between p-4 rounded-sm bg-surface-highest/40 border border-outline-variant/5 hover:bg-surface-highest/80 hover:border-outline-variant/10 transition-all text-left"
            >
              <div class="flex items-center gap-4">
                <div
                  class="w-10 h-10 rounded-sm bg-surface-lowest flex items-center justify-center text-on-surface-variant/40 group-hover:text-primary transition-colors"
                >
                  <span class="material-symbols-outlined text-[20px]"
                    >list_alt</span
                  >
                </div>
                <div>
                  <h4
                    class="text-sm font-bold text-on-surface group-hover:text-primary transition-colors"
                  >
                    {playlist.name}
                  </h4>
                  <p
                    class="text-[10px] text-on-surface-variant font-black uppercase tracking-wider"
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
                  class="w-6 h-6 rounded-full border border-outline-variant/20 group-hover:border-primary/50 flex items-center justify-center transition-colors"
                >
                  <Plus
                    size={14}
                    class="text-on-surface-variant/20 group-hover:text-primary"
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
            class="bg-primary/10 border border-primary/20 rounded-sm p-3 flex items-center gap-3 text-primary"
            transition:slide
          >
            <CheckCircle2 size={16} />
            <span class="text-xs font-bold">{successMessage}</span>
          </div>
        {/if}

        <button
          onclick={() => (showCreateModal = true)}
          class="w-full bg-on-surface text-surface py-4 rounded-sm font-black text-sm transition-all flex items-center justify-center gap-2 hover:opacity-90 active:scale-95"
        >
          <span class="material-symbols-outlined text-[18px]">add</span>
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
    background: var(--color-surface-container);
    border: 1px solid var(--color-outline-variant, rgba(255, 255, 255, 0.1));
  }

  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: var(--color-outline-variant, rgba(255, 255, 255, 0.1));
    border-radius: 2px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: var(--color-primary);
  }
</style>
