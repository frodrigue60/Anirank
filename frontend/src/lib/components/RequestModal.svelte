<script lang="ts">
  import X from "lucide-svelte/icons/x";
import Music from "lucide-svelte/icons/music";
import Send from "lucide-svelte/icons/send";
import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
import Loader2 from "lucide-svelte/icons/loader-2";
import AlertCircle from "lucide-svelte/icons/alert-circle";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { toastState } from "$lib/state/toast.svelte";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    onClose: () => void;
    initialTitle?: string;
    initialContent?: string;
  }

  let {
    show,
    onClose,
    initialTitle = "",
    initialContent = "",
  }: Props = $props();

  let title = $state("");
  let content = $state("");

  // Update internal state when props change or modal opens
  $effect(() => {
    if (show) {
      title = initialTitle;
      content = initialContent;
    }
  });
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

      if (response.data.success || response.status === 201 || response.status === 200) {
        toastState.addToast("Request submitted successfully", "success");
        handleClose();
      } else {
        throw new Error(response.data.message || "Failed to submit request.");
      }
    } catch (e: any) {
      errorMessage = getApiErrorMessage(e, "Failed to submit request.");
      toastState.addToast(errorMessage, "error");
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
    class="fixed inset-0 z-100 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
    onclick={handleClose}
    transition:fade={{ duration: 200 }}
  >
    <!-- Modal Content -->
    <div
      class="modal-glass w-full max-w-sm rounded-md overflow-hidden shadow-2xl p-10 flex flex-col items-center text-center relative"
      onclick={(e) => e.stopPropagation()}
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Header -->
      <div class="w-full flex justify-between items-start mb-6">
        <div class="text-left">
          <div class="flex items-center gap-2 text-primary mb-1">
            <Music size={14} />

            <p class="text-[10px] font-black uppercase tracking-[0.2em]">
              Request Theme
            </p>
          </div>
          <h3
            class="text-xl font-bold leading-tight tracking-tight text-on-surface"
          >
            Missing a theme?
          </h3>
          <p class="text-xs text-on-surface-variant mt-1">
            Let us know what song or anime is missing.
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
        >
          <X size={18} />
        </button>
      </div>

        <div class="w-full space-y-4 text-left">
          <!-- Title Input -->
          <div>
            <label
              for="request-title"
              class="block text-[10px] font-black uppercase tracking-wider text-on-surface-variant mb-2 px-1"
              >Anime/Song Title</label
            >
            <input
              id="request-title"
              type="text"
              bind:value={title}
              placeholder="e.g. Oshi no Ko - Idol"
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-sm text-on-surface focus:outline-none focus:border-primary/30 transition-all hover:bg-surface-highest/80"
            />
          </div>

          <!-- Content Textarea -->
          <div>
            <label
              for="request-content"
              class="block text-[10px] font-black uppercase tracking-wider text-on-surface-variant mb-2 px-1"
              >Links or Info (Optional)</label
            >
            <textarea
              id="request-content"
              bind:value={content}
              placeholder="Add links or specific details about the request..."
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-sm text-on-surface focus:outline-none focus:border-primary/30 transition-all min-h-[100px] resize-none hover:bg-surface-highest/80"
            ></textarea>
          </div>

          {#if errorMessage}
            <p
              class="text-red-500 text-[10px] font-bold px-1 leading-tight flex items-center gap-1.5"
            >
              <AlertCircle size={14} />
              {errorMessage}
            </p>
          {/if}

          <!-- Actions -->
          <button
            onclick={handleSubmit}
            disabled={isSubmitting}
            class="w-full bg-primary hover:bg-primary/90 disabled:opacity-50 text-white py-4 rounded-sm font-black text-sm transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 active:scale-95"
          >
            {#if isSubmitting}
              <Loader2 class="animate-spin" size={18} />
              Sending...
            {:else}
              Submit Request
              <Send size={18} />
            {/if}
          </button>
        </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .modal-glass {
    background: var(--color-surface-container);
    border: 1px solid var(--color-outline-variant, rgba(255, 255, 255, 0.1));
  }
</style>
