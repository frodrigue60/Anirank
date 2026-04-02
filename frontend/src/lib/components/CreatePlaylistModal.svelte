<script lang="ts">
  import { X, Plus, CheckCircle2, Loader2, Library } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    onClose: () => void;
    onCreated?: (playlist: any) => void;
  }

  let { show, onClose, onCreated }: Props = $props();

  let name = $state("");
  let description = $state("");
  let is_public = $state(true);
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let errorMessage = $state("");

  async function handleSubmit() {
    if (!name.trim()) {
      errorMessage = "Playlist name is required.";
      return;
    }

    isSubmitting = true;
    errorMessage = "";
    try {
      const response = await api.post("/playlists", {
        name: name,
        description: description,
        is_public: is_public,
      });

      if (
        response.status === 201 ||
        response.status === 200 ||
        response.data.playlist
      ) {
        isSuccess = true;
        if (onCreated) onCreated(response.data.playlist || response.data.data);
        setTimeout(handleClose, 1500);
      } else {
        throw new Error("Failed to create playlist");
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to create playlist.";
    } finally {
      isSubmitting = false;
    }
  }

  function handleClose() {
    isSuccess = false;
    errorMessage = "";
    name = "";
    description = "";
    is_public = true;
    onClose();
  }
</script>

{#if show}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-110 flex items-center justify-center p-4 bg-black/80"
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
              New Collection
            </p>
          </div>
          <h3 class="text-xl font-bold leading-tight tracking-tight">
            Create Playlist
          </h3>
          <p class="text-xs text-white/50 mt-1">
            Organize your favorite themes.
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-white/5 flex items-center justify-center transition-colors text-white/40 hover:text-white"
        >
          <X size={18} />
        </button>
      </div>

      {#if isSuccess}
        <div class="py-12 flex flex-col items-center space-y-4" in:scale>
          <div
            class="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center text-green-500 mb-2"
          >
            <CheckCircle2 size={36} />
          </div>
          <h4 class="text-lg font-bold">Playlist Created</h4>
          <p class="text-xs text-white/50 leading-relaxed">
            Your new collection is ready!
          </p>
        </div>
      {:else}
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
              placeholder="My Awesome Playlist"
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
              placeholder="A brief description of your collection..."
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
              Creating...
            {:else}
              Create Playlist
              <Plus size={16} />
            {/if}
          </button>
        </div>
      {/if}
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
