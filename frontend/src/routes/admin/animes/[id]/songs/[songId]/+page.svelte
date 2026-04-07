<script lang="ts">
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getSongName } from "$lib/song-utils";
  import { goto } from "$app/navigation";
  import { getApiErrorMessage } from "$lib/api-errors";

  let { data } = $props<{ data: PageData }>();
  let song = $derived(data.song);
  let variants = $state<any[]>([]);
  let isCreating = $state(false);

  let hasValidData = $derived(
    (song?.season_id > 0 || (song?.anime?.season_id > 0)) &&
    (song?.year_id > 0 || (song?.anime?.year_id > 0))
  );

  $effect(() => {
    if (song && song.song_variants) {
      variants = song.song_variants;
    }
  });

  async function handleStatusChange(id: number, currentStatus: boolean) {
    try {
      await api.patch(`/admin/variants/${id}/status`);
      // Update local state reactively
      const index = variants.findIndex(v => v.id === id);
      if (index !== -1) {
        variants[index].status = !currentStatus;
      }
      toastState.addToast("Status updated successfully", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to update status"), "error");
    }
  }

  async function createVariant() {
    isCreating = true;
    try {
      const resp = await api.post("/admin/variants", {
        song_id: song.id,
        status: true,
      });
      const newVariantId = resp.data.data.id;
      toastState.addToast("Variant created successfully", "success");
      goto(`/admin/variants/${newVariantId}/video`);
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to create variant"), "error");
    } finally {
      isCreating = false;
    }
  }
</script>

<div class="mb-6">
    <a href="/admin/animes/{data.id}/songs" class="inline-flex items-center text-sm text-on-surface-variant/70 hover:text-on-surface transition-colors gap-1 mb-4">
        <span class="material-symbols-outlined text-sm">arrow_back</span>
        Back to Songs List
    </a>
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
            <h1 class="text-2xl font-bold text-on-surface max-w-2xl truncate">{getSongName(song)}</h1>
            <p class="text-sm text-on-surface-variant/70 mt-1 flex items-center gap-2">
                <span class="bg-blue-500/10 text-blue-400 px-2 py-0.5 rounded text-xs border border-blue-500/10 font-medium">{song.type} {song.theme_num}</span>
                <span class="font-mono text-xs bg-surface-highest px-2 py-0.5 rounded border border-outline-variant">ID: #{song.id}</span>
            </p>
        </div>
        <div class="flex flex-wrap gap-2">
            <a href="/admin/songs/{song.id}/edit" class="px-4 py-2 bg-surface-highest hover:bg-surface-highest border border-outline-variant text-on-surface rounded-lg text-sm transition-colors flex items-center gap-2">
                <span class="material-symbols-outlined text-sm">edit</span>
                Edit Song
            </a>
            {#if !hasValidData}
                <button disabled class="px-4 py-2 bg-surface-highest border border-outline-variant text-on-surface-variant/40 rounded-lg text-sm cursor-not-allowed flex items-center gap-2" title="Song or Anime must have Season and Year assigned">
                    <span class="material-symbols-outlined text-sm">add</span>
                    Add Variant
                </button>
            {:else}
                <button onclick={createVariant} disabled={isCreating} class="px-4 py-2 bg-primary hover:bg-primary-container text-on-surface rounded-lg text-sm transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
                    {#if isCreating}
                        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-on-surface" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Creating...
                    {:else}
                        <span class="material-symbols-outlined text-sm">add</span>
                        Add Variant
                    {/if}
                </button>
            {/if}
        </div>
    </div>
</div>

{#if !hasValidData}
<div class="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-4 flex items-start gap-3 mb-6">
  <span class="material-symbols-outlined text-amber-500 mt-0.5">warning</span>
  <div>
    <h3 class="text-amber-500 font-medium text-sm">Missing Information</h3>
    <p class="text-amber-400/80 text-sm mt-1">This song (or its parent anime) must have a Season and Year assigned before you can create variants. Please edit the song or anime info first.</p>
  </div>
</div>
{/if}

<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-sm shadow-black/20">
    <div class="p-6 border-b border-outline-variant">
        <h2 class="text-lg font-semibold text-on-surface">Song Variants</h2>
        <p class="text-sm text-on-surface-variant/70 mt-1">Different versions and video sources for this theme.</p>
    </div>
    
    <div class="overflow-x-auto">
        {#if variants && variants.length > 0}
            <table class="w-full text-left text-sm text-on-surface-variant">
                <thead class="text-xs text-on-surface-variant/70 uppercase bg-black/20 border-b border-outline-variant">
                    <tr>
                        <th class="px-6 py-4 font-semibold">Version</th>
                        <th class="px-6 py-4 font-semibold">Slug</th>
                        <th class="px-6 py-4 font-semibold">Status</th>
                        <th class="px-6 py-4 font-semibold text-right">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                    {#each variants as variant}
                        <tr class="hover:bg-white/2 transition-colors">
                            <td class="px-6 py-4">
                                <span class="bg-surface-highest text-on-surface font-mono text-xs px-2 py-1 rounded inline-flex items-center gap-1">
                                    <span class="text-on-surface-variant/40">v</span>{variant.version_number}
                                </span>
                            </td>
                            <td class="px-6 py-4">
                                <span class="text-xs font-mono bg-black/20 px-2 py-1 rounded text-on-surface-variant/70">{variant.slug}</span>
                            </td>
                            <td class="px-6 py-4">
                                <button
                                    onclick={() => handleStatusChange(variant.id, variant.status)}
                                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border transition-all duration-200 hover:scale-105 active:scale-95 {variant.status === true || variant.status === 1 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/20' : 'bg-orange-500/10 text-orange-400 border-orange-500/20 hover:bg-orange-500/20'}"
                                    title="Toggle status for variant {variant.slug}"
                                    aria-label="Toggle status for variant {variant.slug}"
                                >
                                    <div class="w-1.5 h-1.5 rounded-full mr-2 {variant.status === true || variant.status === 1 ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.5)]' : 'bg-orange-400 shadow-[0_0_8px_rgba(251,146,60,0.5)]'}"></div>
                                    {variant.status === true || variant.status === 1 ? "Published" : "Draft"}
                                </button>
                            </td>
                            <td class="px-6 py-4 text-right">
                                <div class="flex items-center justify-end gap-2">
                                  <a href="/admin/variants/{variant.id}/video" class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500 hover:text-on-surface transition-all border border-emerald-500/20 hover:border-emerald-500" title="Manage Video">
                                      <span class="material-symbols-outlined text-sm">video_settings</span>
                                  </a>
                                  <a href="/admin/variants/{variant.id}/edit" class="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-surface-highest text-on-surface-variant/70 hover:bg-surface-highest hover:text-on-surface transition-all border border-outline-variant hover:border-outline-variant" title="Edit Variant">
                                      <span class="material-symbols-outlined text-sm">edit</span>
                                  </a>
                                </div>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {:else}
            <div class="px-6 py-16 flex flex-col items-center justify-center text-center">
                <div class="w-16 h-16 bg-surface-highest rounded-2xl flex items-center justify-center mb-4 border border-outline-variant">
                    <span class="material-symbols-outlined text-3xl text-on-surface-variant/40">video_library</span>
                </div>
                <h3 class="text-on-surface font-medium mb-1">No variants configured</h3>
                <p class="text-sm text-on-surface-variant/40 max-w-sm mb-6">Create the first variant to configure video sources for this theme.</p>
                {#if !hasValidData}
                    <button disabled class="px-5 py-2.5 bg-surface-highest border border-outline-variant text-on-surface-variant/40 rounded-xl text-sm font-medium cursor-not-allowed flex items-center gap-2">
                        <span class="material-symbols-outlined text-sm">add</span>
                        Create First Variant
                    </button>
                {:else}
                    <button onclick={createVariant} disabled={isCreating} class="px-5 py-2.5 bg-primary hover:bg-primary-container text-on-surface rounded-xl text-sm font-medium transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2">
                        {#if isCreating}
                            <svg class="animate-spin h-4 w-4 text-on-surface" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            Creating First Variant...
                        {:else}
                            <span class="material-symbols-outlined text-sm">add</span>
                            Create First Variant
                        {/if}
                    </button>
                {/if}
            </div>
        {/if}
    </div>
</div>
