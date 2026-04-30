<script lang="ts">
  import Star from "lucide-svelte/icons/star";
  import { onMount, onDestroy } from "svelte";

  let {
    hasMore = false,
    loading = false,
    onLoadMore = () => {}
  } = $props();

  let observer: IntersectionObserver;
  let element: HTMLElement;

  let isIntersecting = $state(false);

  onMount(() => {
    observer = new IntersectionObserver(
      (entries) => {
        isIntersecting = entries[0].isIntersecting;
      },
      { threshold: 0.1, rootMargin: "200px" }
    );

    if (element) {
      observer.observe(element);
    }
  });

  let isProcessing = false;
  $effect(() => {
    if (isIntersecting && hasMore && !loading && !isProcessing) {
      isProcessing = true;
      onLoadMore();
      // Reset processing after a short delay or when loading state changes
      setTimeout(() => {
        isProcessing = false;
      }, 500);
    }
  });

  onDestroy(() => {
    if (observer) observer.disconnect();
  });
</script>

<div bind:this={element} class="w-full flex justify-center py-12 min-h-[100px]">
  {#if loading}
    <div class="flex flex-col items-center gap-4">
        <div class="w-10 h-10 border-4 border-white/5 border-t-primary rounded-full animate-spin shadow-2xl shadow-primary/20"></div>
        <span class="text-[10px] font-black uppercase tracking-[0.2em] text-white/30 animate-pulse">Synchronizing Data</span>
    </div>
  {:else if !hasMore}
    <div class="flex flex-col items-center gap-3 opacity-20 group">
        <div class="h-px w-24 bg-linear-to-r from-transparent via-white/50 to-transparent mb-2"></div>
        <Star size={20} class="transition-transform group-hover:scale-110" />
        <span class="text-[9px] font-black uppercase tracking-[0.3em]">End of Collection</span>
    </div>
  {/if}
</div>
