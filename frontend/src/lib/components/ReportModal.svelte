<script lang="ts">
  import { X, Flag, Send, CheckCircle2, Loader2 } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    song: any;
    onClose: () => void;
    onSuccess?: () => void;
  }

  let { show, song, onClose, onSuccess }: Props = $props();

  let title = $state("");
  let content = $state("");
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let errorMessage = $state("");

  const reportingReasons = [
    "Missing Video/Audio",
    "Incorrect Title/Artist",
    "Copyright Infringement",
    "Low Quality Audio/Video",
    "Wrong Anime/Theme Information",
    "Other",
  ];

  async function handleSubmit() {
    if (!title) {
      errorMessage = "Please select a reason for reporting.";
      return;
    }

    isSubmitting = true;
    errorMessage = "";
    try {
      const response = await api.post("/songs/reports", {
        song_id: song.id,
        title: title,
        content: content,
        source: "web",
      });

      if (
        response.data.success ||
        response.status === 201 ||
        response.status === 200
      ) {
        isSuccess = true;
        if (onSuccess) onSuccess();
        setTimeout(() => {
          handleClose();
        }, 2000);
      } else {
        errorMessage = response.data.message || "Failed to submit report.";
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to submit report.";
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
            <span class="material-symbols-outlined text-[14px]">flag</span>
            <p class="text-[10px] font-black uppercase tracking-[0.2em]">
              Report Issue
            </p>
          </div>
          <h3
            class="text-xl font-bold leading-tight tracking-tight text-on-surface"
          >
            Help us improve
          </h3>
          <p class="text-xs text-on-surface-variant mt-1">
            Tell us what's wrong with this theme.
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-8 h-8 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
        >
          <X size={18} />
        </button>
      </div>

      {#if isSuccess}
        <div class="py-12 flex flex-col items-center space-y-4" in:scale>
          <div
            class="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center text-green-500 mb-2 shadow-inner"
          >
            <CheckCircle2 size={36} />
          </div>
          <h4 class="text-lg font-bold text-on-surface">Report Received</h4>
          <p class="text-xs text-on-surface-variant leading-relaxed">
            Thank you for your feedback. We'll look into it soon!
          </p>
        </div>
      {:else}
        <div class="w-full space-y-4 text-left">
          <!-- Reason Select -->
          <div>
            <label
              for="report-title"
              class="block text-[10px] font-black uppercase tracking-wider text-on-surface-variant mb-2 px-1"
              >Reason</label
            >
            <div class="relative group">
              <select
                id="report-title"
                bind:value={title}
                class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-sm text-on-surface focus:outline-none focus:border-primary/30 transition-all appearance-none cursor-pointer group-hover:bg-surface-highest/80"
              >
                <option value="" disabled selected>Select a reason...</option>
                {#each reportingReasons as reason}
                  <option value={reason}>{reason}</option>
                {/each}
              </select>
              <div
                class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-on-surface-variant/40 group-hover:text-primary transition-colors"
              >
                <span class="material-symbols-outlined text-[18px]"
                  >expand_more</span
                >
              </div>
            </div>
          </div>

          <!-- Content Textarea -->
          <div>
            <label
              for="report-content"
              class="block text-[10px] font-black uppercase tracking-wider text-on-surface-variant mb-2 px-1"
              >Details (Optional)</label
            >
            <textarea
              required
              id="report-content"
              bind:value={content}
              placeholder="Please provide more information..."
              class="w-full bg-surface-highest border border-outline-variant/10 rounded-sm px-4 py-3 text-sm text-on-surface focus:outline-none focus:border-primary/30 transition-all min-h-[100px] resize-none hover:bg-surface-highest/80"
            ></textarea>
          </div>

          {#if errorMessage}
            <p
              class="text-red-500 text-[10px] font-bold px-1 leading-tight flex items-center gap-1.5"
            >
              <span class="material-symbols-outlined text-[14px]">error</span>
              {errorMessage}
            </p>
          {/if}

          <!-- Actions -->
          <button
            onclick={handleSubmit}
            disabled={isSubmitting}
            class="w-full bg-primary hover:bg-primary/90 disabled:opacity-50 text-white py-4 rounded-sm font-black text-sm transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 active:scale-95 mt-2"
          >
            {#if isSubmitting}
              <Loader2 class="animate-spin" size={18} />
              Submitting...
            {:else}
              Send Report
              <span class="material-symbols-outlined text-[18px]">send</span>
            {/if}
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style lang="postcss">
  .modal-glass {
    background: var(--color-surface-container);
    border: 1px solid var(--color-outline-variant, rgba(255, 255, 255, 0.1));
  }

  select option {
    background: var(--color-surface-container);
    color: var(--color-on-surface);
  }
</style>
