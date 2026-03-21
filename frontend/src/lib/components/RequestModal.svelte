<script lang="ts">
  import { X, Music, Send, CheckCircle2, Loader2 } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    onClose: () => void;
  }

  let { show, onClose }: Props = $props();

  let title = $state("");
  let content = $state("");
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let errorMessage = $state("");

  async function handleSubmit() {
    if (!title.trim() || !content.trim()) {
      errorMessage = "Please fill in both title and description.";
      return;
    }

    isSubmitting = true;
    errorMessage = "";
    try {
      const response = await api.post("/user-requests", {
        title: title,
        content: content,
      });

      if (response.data.success) {
        isSuccess = true;
        setTimeout(() => {
          handleClose();
        }, 2000);
      } else {
        errorMessage = response.data.message || "Failed to submit request.";
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to submit request.";
    } finally {
      isSubmitting = false;
    }
  }

  function handleClose() {
    isSuccess = false;
    errorMessage = "";
    title = "";
    content = "";
    onClose();
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
            <Music size={14} />
            <p class="text-[10px] font-bold uppercase tracking-[0.2em]">
              Request Theme
            </p>
          </div>
          <h3 class="text-xl font-bold leading-tight tracking-tight">
            Missing a theme?
          </h3>
          <p class="text-xs text-white/50 mt-1">
            Let us know what song or anime is missing.
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
          <h4 class="text-lg font-bold">Request Sent!</h4>
          <p class="text-xs text-white/50 leading-relaxed">
            Your request has been submitted. We'll check it out!
          </p>
        </div>
      {:else}
        <div class="w-full space-y-4 text-left">
          <!-- Title Input -->
          <div>
            <label
              for="request-title"
              class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-2 px-1"
              >Anime/Song Title</label
            >
            <input
              id="request-title"
              type="text"
              bind:value={title}
              placeholder="e.g. Oshi no Ko - Idol"
              class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary/50 transition-all"
            />
          </div>

          <!-- Content Textarea -->
          <div>
            <label
              for="request-content"
              class="block text-[10px] font-bold uppercase tracking-wider text-white/40 mb-2 px-1"
              >Links or Info (Optional)</label
            >
            <textarea
              id="request-content"
              bind:value={content}
              placeholder="Add links or specific details about the request..."
              class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-primary/50 transition-all min-h-[100px] resize-none"
            ></textarea>
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
              Sending...
            {:else}
              Submit Request
              <Send size={16} />
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
