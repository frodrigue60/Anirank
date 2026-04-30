<script lang="ts">
  import { toastState } from "$lib/state/toast.svelte";
  import { flip } from "svelte/animate";
  import { fade, fly } from "svelte/transition";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import Info from "lucide-svelte/icons/info";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import X from "lucide-svelte/icons/x";
</script>

<div
  class="fixed bottom-6 right-6 z-9999 flex flex-col gap-3 pointer-events-none"
>
  {#each toastState.toasts as toast (toast.id)}
    <div
      in:fly={{ y: 20, duration: 300 }}
      out:fade={{ duration: 200 }}
      animate:flip={{ duration: 300 }}
      class="pointer-events-auto min-w-[300px] max-w-md bg-[#1a1a1e] border border-white/10 rounded-2xl p-4 shadow-2xl flex items-start gap-3"
    >
      <div
        class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0
        {toast.type === 'success' ? 'bg-green-500/10 text-green-500' : ''}
        {toast.type === 'error' ? 'bg-red-500/10 text-red-500' : ''}
        {toast.type === 'info'
          ? 'bg-anirank-primary/10 text-anirank-primary'
          : ''}
        {toast.type === 'warning' ? 'bg-amber-500/10 text-amber-400' : ''}"
      >
        {#if toast.type === "success"}<CheckCircle2 size={20} />{/if}
        {#if toast.type === "error"}<AlertCircle size={20} />{/if}
        {#if toast.type === "info"}<Info size={20} />{/if}
        {#if toast.type === "warning"}<AlertTriangle size={20} />{/if}

      </div>

      <div class="flex-1 pt-0.5">
        <p class="text-white text-sm font-medium leading-tight">
          {toast.message}
        </p>
      </div>

      <button
        onclick={() => toastState.removeToast(toast.id)}
        class="p-1 rounded-lg hover:bg-white/5 text-white/20 hover:text-white transition-colors"
      >
        <X size={18} />
      </button>
    </div>
  {/each}
</div>
