<script lang="ts">
  import { authState } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import { X, Star, Send, CheckCircle2, Loader2 } from "lucide-svelte";
  import api from "$lib/api";
  import { fade, scale } from "svelte/transition";

  interface Props {
    show: boolean;
    song: any;
    onClose: () => void;
    onRated?: (newData: any) => void;
  }

  let { show, song, onClose, onRated }: Props = $props();

  let value = $state(0); // Default value (0-100)
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let errorMessage = $state("");

  // Detect format
  let format = $derived(authState.user?.score_format || "POINT_10_DECIMAL");

  // Display values based on format
  let displayValue = $derived.by(() => {
    switch (format) {
      case "POINT_100":
        return Math.round(value);
      case "POINT_10":
        return Math.round(value / 10);
      case "POINT_10_DECIMAL":
        return (value / 10).toFixed(1);
      case "POINT_5":
        return (value / 20).toFixed(1);
      default:
        return (value / 10).toFixed(1);
    }
  });

  async function handleSubmit() {
    isSubmitting = true;
    errorMessage = "";
    try {
      const response = await api.post(`/interactions/ratings`, {
        song_id: song.id,
        score: value,
      });
      if (
        response.status === 200 ||
        response.status === 201 ||
        response.data.success
      ) {
        isSuccess = true;
        toastState.addToast("Rating submitted successfully!", "success");
        if (onRated) onRated(response.data);
        setTimeout(() => {
          handleClose();
        }, 1500);
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to submit rating";
    } finally {
      isSubmitting = false;
    }
  }

  function handleClose() {
    isSuccess = false;
    errorMessage = "";
    onClose();
  }

  // Stars logic for POINT_5
  function setRatingFromStars(starIndex: number, isHalf: boolean) {
    value = (starIndex + (isHalf ? 0.5 : 1)) * 20;
  }

  // Grid buttons (for 10 or 100)
  const tenGrid = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
  const hundredGrid = [50, 75, 90, 100];

  // Pre-fill existing rating
  $effect(() => {
    if (show && song?.user_rating !== undefined) {
      value = song.user_rating;
    } else if (show) {
      value = 0;
    }
  });
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
      class="modal-glass w-full max-w-md rounded-[2.5rem] overflow-hidden shadow-[0_32px_64px_-12px_rgba(0,0,0,0.5)] p-10 flex flex-col items-center text-center relative"
      onclick={(e) => e.stopPropagation()}
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Header -->
      <div class="w-full flex justify-between items-start mb-2 group">
        <div class="text-left">
          <p
            class="text-[10px] font-bold uppercase tracking-[0.2em] text-primary mb-1"
          >
            Rate this theme
          </p>
          <h3 class="text-2xl font-bold leading-tight tracking-tight">
            {song.song_romaji || "Theme"}
          </h3>
          <p class="text-sm text-white/50">
            {song.post?.title || "Anime"} • {song.type}
            {song.theme_num || ""}
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-10 h-10 rounded-full hover:bg-white/5 flex items-center justify-center transition-colors text-white/40 hover:text-white"
        >
          <X size={20} />
        </button>
      </div>

      {#if isSuccess}
        <div class="py-20 flex flex-col items-center space-y-4" in:scale>
          <div
            class="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center text-green-500 mb-2"
          >
            <CheckCircle2 size={48} />
          </div>
          <h4 class="text-xl font-bold">Rating Submitted!</h4>
          <p class="text-white/50">Your score has been saved.</p>
        </div>
      {:else}
        <!-- Score Display -->
        <div class="mt-8 mb-4">
          <div
            class="text-[84px] font-bold leading-none tracking-tighter text-white drop-shadow-[0_0_15px_rgba(127,19,236,0.6)]"
          >
            {displayValue}
          </div>
          <p
            class="text-[10px] text-primary font-bold uppercase tracking-widest mt-2 bg-primary/10 px-3 py-1 rounded-full inline-block"
          >
            {format === "POINT_5"
              ? "Classic Stars"
              : format === "POINT_100"
                ? "Standard Score"
                : "Granular Rating"}
          </p>
        </div>

        <!-- Controls -->
        <div class="w-full my-8">
          {#if format === "POINT_5"}
            <!-- Stars Component -->
            <div class="flex items-center justify-center gap-1">
              {#each Array(5) as _, i}
                <div class="relative group cursor-pointer p-1">
                  <Star
                    size={48}
                    class="{value / 20 > i
                      ? 'text-primary fill-primary drop-shadow-[0_0_8px_rgba(127,19,236,0.6)]'
                      : 'text-white/10'} transition-all"
                  />
                  <!-- Invisible overlays for half/full stars -->
                  <div
                    class="absolute inset-y-0 left-0 w-1/2"
                    onmouseenter={() => (value = (i + 0.5) * 20)}
                    onclick={() => setRatingFromStars(i, true)}
                  ></div>
                  <div
                    class="absolute inset-y-0 right-0 w-1/2"
                    onmouseenter={() => (value = (i + 1) * 20)}
                    onclick={() => setRatingFromStars(i, false)}
                  ></div>
                </div>
              {/each}
            </div>
          {:else}
            <!-- Slider + Grid -->
            <div class="space-y-8">
              <div class="px-2">
                <input
                  type="range"
                  min="0"
                  max="100"
                  bind:value
                  class="rating-range w-full h-2.5 bg-white/10 rounded-full appearance-none cursor-pointer focus:outline-none"
                />
                <div class="flex justify-between mt-5 px-1">
                  <span class="text-[11px] font-bold text-white/30">0</span>
                  <span class="text-[11px] font-bold text-white/30">25</span>
                  <span class="text-[11px] font-bold text-white/30">50</span>
                  <span class="text-[11px] font-bold text-white/30">75</span>
                  <span class="text-[11px] font-bold text-white/30">100</span>
                </div>
              </div>

              {#if format === "POINT_10" || format === "POINT_10_DECIMAL"}
                <div class="grid grid-cols-5 gap-2 w-full">
                  {#each tenGrid as val}
                    <button
                      onclick={() => (value = val * 10)}
                      class="py-2 rounded-full border transition-all text-xs font-bold {Math.round(
                        value / 10,
                      ) === val
                        ? 'bg-primary/20 border-primary/40 text-primary'
                        : 'bg-white/5 border-white/5 text-white/60 hover:bg-white/15'}"
                    >
                      {val}
                    </button>
                  {/each}
                </div>
              {:else if format === "POINT_100"}
                <div class="grid grid-cols-4 gap-2 w-full">
                  {#each hundredGrid as val}
                    <button
                      onclick={() => (value = val)}
                      class="py-2.5 rounded-xl border transition-all text-sm font-bold {value ===
                      val
                        ? 'bg-primary/20 border-primary/40 text-primary'
                        : 'bg-white/5 border-white/5 text-white/60 hover:bg-white/15'}"
                    >
                      {val}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        </div>

        {#if errorMessage}
          <p class="text-red-500 text-xs mb-4 font-medium">
            {errorMessage}
          </p>
        {/if}

        <!-- Actions -->
        <div class="w-full grid grid-cols-2 gap-4 mt-4">
          <button
            onclick={handleClose}
            class="bg-white/5 hover:bg-white/10 text-white/70 hover:text-white py-4 rounded-2xl font-bold text-sm transition-all border border-white/5 active:scale-95"
          >
            Cancel
          </button>
          <button
            onclick={handleSubmit}
            disabled={isSubmitting}
            class="bg-primary hover:bg-primary/80 disabled:opacity-50 text-white py-4 rounded-2xl font-bold text-sm transition-all shadow-xl shadow-primary/30 flex items-center justify-center gap-2 active:scale-95"
          >
            {#if isSubmitting}
              <Loader2 class="animate-spin" size={18} />
              Submitting...
            {:else}
              Submit Rating
              <Send size={18} />
            {/if}
          </button>
        </div>
      {/if}

      <p class="mt-8 text-[11px] text-white/30 font-medium italic">
        {format === "POINT_5"
          ? "Score will be converted to decimal format."
          : "Granular scoring supported via the slider."}
      </p>
    </div>
  </div>
{/if}

<style lang="postcss">
  .modal-glass {
    background: rgba(25, 16, 34, 0.85);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .rating-range::-webkit-slider-runnable-track {
    background: linear-gradient(to right, #7f13ec, #a855f7);
    height: 8px;
    border-radius: 4px;
  }

  .rating-range::-webkit-slider-thumb {
    -webkit-appearance: none;
    height: 24px;
    width: 24px;
    border-radius: 50%;
    background: #ffffff;
    cursor: pointer;
    margin-top: -8px;
    box-shadow: 0 0 15px rgba(127, 19, 236, 0.5);
    border: 2px solid #7f13ec;
    transition: transform 0.1s ease;
  }

  .rating-range::-webkit-slider-thumb:active {
    transform: scale(1.1);
  }
</style>
