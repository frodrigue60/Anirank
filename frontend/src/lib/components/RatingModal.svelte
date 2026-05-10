<script lang="ts">
  import { authState } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import X from "lucide-svelte/icons/x";
import Star from "lucide-svelte/icons/star";
import Send from "lucide-svelte/icons/send";
import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
import Loader2 from "lucide-svelte/icons/loader-2";
import AlertCircle from "lucide-svelte/icons/alert-circle";
  import api from "$lib/api";
  import { getSongName } from "$lib/song-utils";
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
    // Optimistic Update
    const previousRating = song.user_rating;
    if (onRated) {
      onRated({ rating: value });
    }

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
        
        // Final update with server data (including new average)
        if (onRated) {
          const serverData = response.data.data || response.data;
          onRated({
            rating: serverData.rating ?? serverData.score ?? value,
            average: serverData.average ?? serverData.averageRating
          });
        }
        
        setTimeout(() => {
          handleClose();
        }, 1500);
      }
    } catch (e: any) {
      errorMessage = e.response?.data?.message || "Failed to submit rating";
      toastState.addToast(errorMessage, "error");
      
      // Rollback on error
      if (onRated) {
        onRated({ rating: previousRating });
      }
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
    class="fixed inset-0 z-100 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
    onclick={handleClose}
    transition:fade={{ duration: 200 }}
  >
    <!-- Modal Content -->
    <div
      class="modal-glass w-full max-w-md rounded-md overflow-hidden shadow-2xl p-4 sm:p-6 flex flex-col items-center text-center relative"
      onclick={(e) => e.stopPropagation()}
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Header -->
      <div class="w-full flex justify-between items-start mb-2 group">
        <div class="text-left">
          <div class="flex items-center gap-2 text-primary mb-1">
            <Star size={14} class="fill-current" />
            <p class="text-[10px] font-black uppercase tracking-[0.2em]">
              Rate this theme
            </p>
          </div>
          <h3
            class="text-2xl font-bold leading-tight tracking-tight text-on-surface"
          >
            {getSongName(song)}
          </h3>
          <p class="text-xs text-on-surface-variant mt-1 font-medium">
            {song.post?.title || "Anime"} • {song.type}
            {song.theme_num || ""}
          </p>
        </div>
        <button
          onclick={handleClose}
          class="w-10 h-10 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
        >
          <X size={20} />
        </button>
      </div>

      {#if isSuccess}
        <div class="py-20 flex flex-col items-center space-y-4" in:scale>
          <div
            class="w-20 h-20 bg-green-500/10 rounded-full flex items-center justify-center text-green-500 mb-2 shadow-inner"
          >
            <CheckCircle2 size={48} />
          </div>
          <h4 class="text-xl font-bold text-on-surface">Rating Submitted!</h4>
          <p class="text-xs text-on-surface-variant font-medium">
            Your score has been saved.
          </p>
        </div>
      {:else}
        <!-- Score Display -->
        <div class="mt-8 mb-4">
          <div
            class="score-text text-[84px] font-black leading-none tracking-tighter text-on-surface"
          >
            {displayValue}
          </div>
          <p
            class="text-[10px] text-primary font-black uppercase tracking-widest mt-4 bg-primary/10 px-4 py-1.5 rounded-full inline-block"
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
                      ? 'text-primary fill-primary drop-shadow-[0_0_12px_rgba(var(--color-primary-rgb),0.4)]'
                      : 'text-on-surface-variant/10'} transition-all active:scale-95"
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
                  class="rating-range w-full h-2 rounded-full appearance-none cursor-pointer focus:outline-none"
                />
                <div class="flex justify-between mt-6 px-1">
                  {#each [0, 25, 50, 75, 100] as mark}
                    <span
                      class="text-[10px] font-black tracking-tighter transition-colors {Math.abs(
                        value - mark,
                      ) < 5
                        ? 'text-primary'
                        : 'text-on-surface-variant/30'}">{mark}</span
                    >
                  {/each}
                </div>
              </div>

              {#if format === "POINT_10" || format === "POINT_10_DECIMAL"}
                <div class="grid grid-cols-5 gap-2 w-full px-2">
                  {#each tenGrid as val}
                    <button
                      onclick={() => (value = val * 10)}
                      class="py-3 rounded-sm border transition-all text-xs font-black {Math.round(
                        value / 10,
                      ) === val
                        ? 'bg-primary border-primary text-white shadow-lg shadow-primary/20 scale-105'
                        : 'bg-surface-highest border-outline-variant/10 text-on-surface-variant hover:text-on-surface hover:bg-surface-highest/80'}"
                    >
                      {val}
                    </button>
                  {/each}
                </div>
              {:else if format === "POINT_100"}
                <div class="grid grid-cols-4 gap-2 w-full px-2">
                  {#each hundredGrid as val}
                    <button
                      onclick={() => (value = val)}
                      class="py-3 rounded-sm border transition-all text-xs font-black {value ===
                      val
                        ? 'bg-primary border-primary text-white shadow-lg shadow-primary/20 scale-105'
                        : 'bg-surface-highest border-outline-variant/10 text-on-surface-variant hover:text-on-surface hover:bg-surface-highest/80'}"
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
          <p
            class="text-red-500 text-[10px] font-black uppercase tracking-wider mb-6 leading-tight flex items-center justify-center gap-2"
          >
            <AlertCircle size={14} />
            {errorMessage}
          </p>
        {/if}

        <!-- Actions -->
        <div class="w-full grid grid-cols-2 gap-4 mt-2">
          <button
            onclick={handleClose}
            class="bg-surface-highest hover:bg-surface-highest/80 text-on-surface-variant hover:text-on-surface py-4 rounded-sm font-black text-sm transition-all border border-outline-variant/10 active:scale-95"
          >
            Cancel
          </button>
          <button
            onclick={handleSubmit}
            disabled={isSubmitting}
            class="bg-primary hover:bg-primary/90 disabled:opacity-50 text-white py-4 rounded-sm font-black text-sm transition-all shadow-xl shadow-primary/20 flex items-center justify-center gap-2 active:scale-95"
          >
            {#if isSubmitting}
              <Loader2 class="animate-spin" size={18} />
              Saving...
            {:else}
              Submit Rating
              <Send size={18} />
            {/if}
          </button>
        </div>
      {/if}

      <p
        class="mt-8 text-[10px] text-on-surface-variant/40 font-bold uppercase tracking-widest italic"
      >
        {format === "POINT_5"
          ? "Score will be converted to decimal format."
          : "Granular scoring supported via the slider."}
      </p>
    </div>
  </div>
{/if}

<style lang="postcss">
  .modal-glass {
    background: var(--color-surface-container);
    border: 1px solid var(--color-outline-variant, rgba(255, 255, 255, 0.1));
  }

  .score-text {
    background: linear-gradient(
      135deg,
      var(--color-on-surface) 0%,
      var(--color-primary) 100%
    );
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    filter: drop-shadow(0 0 20px rgba(var(--color-primary-rgb), 0.2));
  }

  .rating-range::-webkit-slider-runnable-track {
    background: var(--color-surface-highest);
    height: 8px;
    border-radius: 4px;
    border: 1px solid var(--color-outline-variant);
  }

  .rating-range::-webkit-slider-thumb {
    -webkit-appearance: none;
    height: 24px;
    width: 24px;
    border-radius: 50%;
    background: var(--color-primary);
    cursor: pointer;
    margin-top: -9px;
    box-shadow: 0 4px 12px rgba(var(--color-primary-rgb), 0.4);
    border: 3px solid var(--color-surface-container);
    transition: all 0.2s ease;
  }

  .rating-range::-webkit-slider-thumb:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 16px rgba(var(--color-primary-rgb), 0.5);
  }

  .rating-range::-webkit-slider-thumb:active {
    transform: scale(0.95);
  }
</style>
