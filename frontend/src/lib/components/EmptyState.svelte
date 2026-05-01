<script lang="ts">
  import { fade, scale } from "svelte/transition";
  import type { Component } from "svelte";

  interface Props {
    title: string;
    message?: string;
    icon?: Component;
    actionLabel?: string;
    onAction?: () => void;
  }

  let { 
    title, 
    message = "", 
    icon: Icon, 
    actionLabel = "", 
    onAction 
  }: Props = $props();
</script>

<div 
  class="flex flex-col items-center justify-center py-20 px-6 text-center animate-in fade-in zoom-in duration-700"
  in:fade={{ duration: 400 }}
>
  <div class="relative mb-6">
    <div class="absolute inset-0 bg-primary/5 blur-3xl rounded-full scale-150"></div>
    <div class="relative w-20 h-20 bg-surface-highest rounded-sm flex items-center justify-center text-primary/40 shadow-inner border border-white/5" in:scale={{ delay: 200, duration: 400, start: 0.8 }}>
      {#if Icon}
        <Icon size={40} strokeWidth={1.5} />
      {:else}
        <div class="w-10 h-10 border-4 border-dashed border-primary/20 rounded-sm"></div>
      {/if}
    </div>
  </div>

  <h3 class="text-xl font-bold text-on-surface tracking-tight mb-2" in:fade={{ delay: 300 }}>
    {title}
  </h3>
  
  {#if message}
    <p class="text-sm text-on-surface-variant max-w-xs leading-relaxed mb-8" in:fade={{ delay: 400 }}>
      {message}
    </p>
  {/if}

  {#if actionLabel && onAction}
    <button
      onclick={onAction}
      class="bg-primary hover:opacity-90 text-white px-8 py-3 rounded-sm font-black text-xs uppercase tracking-[0.2em] transition-all shadow-lg shadow-primary/20 active:scale-95"
      aria-label={actionLabel}
      in:fade={{ delay: 500 }}
    >
      {actionLabel}
    </button>
  {/if}
</div>
