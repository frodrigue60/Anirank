<script lang="ts">
  import { X, Save, Library, Loader2 } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    playlist: any;
    onClose: () => void;
    onUpdated?: (updatedPlaylist: any) => void;
  }

  let { show, playlist, onClose, onUpdated }: Props = $props();

  let name = $state("");
  let description = $state("");
  let is_public = $state(true);
  let isSubmitting = $state(false);
  let errorMessage = $state("");

  // Update internal state when playlist prop changes
  $effect(() => {
    if (playlist) {
      name = playlist.name || "";
      description = playlist.description || "";
      is_public = playlist.is_public ?? true;
    }
  });

  async function handleSubmit() {
    if (!name.trim()) {
      errorMessage = "Playlist name is required.";
      return;
    }

    isSubmitting = true;
    errorMessage = "";
    try {
      const response = await api.put(`/playlists/${playlist.id}`, {
        name: name,
        description: description,
        is_public: is_public,
      });

      if (response.status === 200 || response.data.playlist) {
        if (onUpdated)
          onUpdated(
            response.data.playlist || {
              ...playlist,
              name,
              description,
              is_public,
            },
          );
        onClose();
      } else {
        errorMessage = response.data.message || "Failed to update playlist.";
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to update playlist.";
    } finally {
      isSubmitting = false;
    }
  }

  function handleClose() {
    errorMessage = "";
    onClose();
  }
</script>

{#if show}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-110 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
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
            <Library size={14} />
            <p class="text-[10px] font-bold uppercase tracking-[0.2em]">
              Edit Collection
            </p>
          </div>
          <h3 class="text-xl font-bold leading-tight tracking-tight">
            Update Playlist
          </h3>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-white/5 flex items-center justify-center transition-colors text-white/40 hover:text-white"
        >
          <X size={18} />
        </button>
      </div>

      <div class="w-full space-y-4 text-left">
        <!-- Name Input -->
        <div>
          <label
            for="playlist-name"
            class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-2 px-1"
            >Name</label
          >
          <input
            id="playlist-name"
            type="text"
            bind:value={name}
            placeholder="Playlist Name"
            class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary/50 transition-all"
          />
        </div>

        <!-- Description Input -->
        <div>
          <label
            for="playlist-desc"
            class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-2 px-1"
            >Description (Optional)</label
          >
          <textarea
            id="playlist-desc"
            bind:value={description}
            placeholder="A brief description..."
            class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary/50 transition-all min-h-[80px] resize-none"
          ></textarea>
        </div>

        <!-- Privacy Select -->
        <div>
          <label
            for="playlist-privacy"
            class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-2 px-1"
            >Privacy</label
          >
          <div class="relative">
            <select
              id="playlist-privacy"
              bind:value={is_public}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary/50 transition-all appearance-none"
            >
              <option value={true} class="bg-surface-dark text-white"
                >Public - Visible to anyone</option
              >
              <option value={false} class="bg-surface-dark text-white"
                >Private - Only visible to you</option
              >
            </select>
            <span
              class="material-symbols-outlined absolute right-4 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none"
              >expand_more</span
            >
          </div>
        </div>

        {#if errorMessage}
          <p class="text-red-500 text-[10px] font-medium px-1 leading-tight">
            {errorMessage}
          </p>
        {/if}

        <!-- Actions -->
        <button
          onclick={handleSubmit}
          disabled={isSubmitting}
          class="w-full bg-primary hover:bg-primary/80 disabled:opacity-50 text-white py-4 rounded-xl font-bold text-sm transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 active:scale-95"
        >
          {#if isSubmitting}
            <Loader2 class="animate-spin" size={18} />
            Saving...
          {:else}
            Save Changes
            <Save size={16} />
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .modal-glass {
    background: rgba(25, 16, 34, 0.9);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
</style>
