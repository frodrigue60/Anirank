<script lang="ts">
  import { authState } from "$lib/state/auth.svelte";

  let { status = $bindable(true) } = $props<{
    status: boolean;
  }>();

  // A user can edit the status if they are admin, editor or owner.
  const canEditStatus = $derived(authState.canPublish);


</script>

<div class="pt-2">
  <label class="flex items-center gap-3 {canEditStatus ? 'cursor-pointer' : 'cursor-not-allowed opacity-70'}">
    <div class="relative">
      <input
        type="checkbox"
        bind:checked={status}
        disabled={!canEditStatus}
        class="sr-only peer"
      />
      <div
        class="w-11 h-6 bg-surface-highest rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"
      ></div>
    </div>
    <span class="text-sm font-medium text-on-surface">
      {status ? 'Published (Active)' : 'Draft (Inactive)'}
    </span>
  </label>
  
  {#if !canEditStatus}
    <p class="text-[10px] text-amber-400/80 mt-2 flex items-center gap-1">
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      As a Creator, your content will be reviewed before publishing.
    </p>
  {:else}
    <p class="text-xs text-on-surface-variant/40 mt-2">
      If unpublished, it will remain as a draft exclusively in the admin panel.
    </p>
  {/if}
</div>
