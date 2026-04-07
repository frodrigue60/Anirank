<script lang="ts">
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { invalidateAll } from "$app/navigation";

  let { data } = $props<{ data: PageData }>();
  let variant = $derived(data.variant);

  async function handleStatusChange(variantId: number, currentStatus: boolean) {
    try {
      const res = await api.put(`/admin/variants/${variantId}`, {
        status: !currentStatus,
      });

      if (res.status === 200) {
        toastState.addToast("Variant status updated successfully", "success");
        invalidateAll();
      }
    } catch (err) {
      console.error(err);
      toastState.addToast("Failed to update status", "error");
    }
  }
</script>

<div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
  <!-- Main Column -->
  <div class="lg:col-span-3 space-y-6">
    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm flex flex-col items-center text-center">
        <span class="text-xs text-on-surface-variant/40 uppercase font-black tracking-widest mb-2">Visibility</span>
        <button
          onclick={() => handleStatusChange(variant.id, variant.status)}
          class="inline-flex items-center px-4 py-2 rounded-full text-xs font-bold transition-all border {variant.status 
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500 hover:text-on-surface' 
            : 'bg-red-500/10 text-red-400 border-red-500/20 hover:bg-red-500 hover:text-on-surface'}"
        >
          {variant.status ? 'ACTIVE' : 'INACTIVE'}
        </button>
      </div>
      
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm flex flex-col items-center text-center">
        <span class="text-xs text-on-surface-variant/40 uppercase font-black tracking-widest mb-2">Spoilers</span>
        <span class="text-sm font-bold {variant.spoiler ? 'text-amber-400' : 'text-on-surface-variant/70'}">
          {variant.spoiler ? 'YES (SPOILER)' : 'NO'}
        </span>
      </div>

      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm flex flex-col items-center text-center">
        <span class="text-xs text-on-surface-variant/40 uppercase font-black tracking-widest mb-2">Video</span>
        {#if variant.video}
          <span class="text-emerald-400 text-sm font-bold flex items-center gap-1">
             <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
             CONFIGURED
          </span>
        {:else}
          <span class="text-on-surface-variant/40 text-sm font-bold flex items-center gap-1">
             <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
             NOT CONFIGURED
          </span>
        {/if}
      </div>
    </div>

    <!-- Specifications -->
    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-on-surface mb-6 border-b border-outline-variant pb-2">Technical Specifications</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-8 text-sm">
        <div class="space-y-4">
          <div>
            <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Internal ID</span>
            <span class="text-gray-200 font-mono">#{variant.id}</span>
          </div>
          <div>
            <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Slug (URL Segment)</span>
            <span class="text-gray-200 uppercase font-bold text-primary">{variant.slug || "None"}</span>
          </div>
          <div>
            <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Version Number</span>
            <span class="text-gray-200">v{variant.version_number}</span>
          </div>
        </div>
        <div class="space-y-4">
          <div>
            <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Added Date</span>
            <span class="text-gray-200">{new Date(variant.created_at).toLocaleDateString()}</span>
          </div>
          <div>
            <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Last Modified</span>
            <span class="text-gray-200">{new Date(variant.updated_at).toLocaleDateString()}</span>
          </div>
          <div>
             <span class="text-on-surface-variant/40 block mb-1 uppercase text-[10px] font-black tracking-widest">Override Taxonomy</span>
             <span class="text-gray-200">
                {#if variant.season || variant.year}
                  {variant.season?.name || ""} {variant.year?.name || ""}
                {:else}
                  <span class="text-gray-600 italic">Inherited from Song</span>
                {/if}
             </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Video Preview / Call to action -->
    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm overflow-hidden relative group">
        <div class="absolute inset-0 bg-blue-500/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"></div>
        <div class="flex flex-col md:flex-row items-center justify-between gap-6">
            <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-xl bg-surface-highest flex items-center justify-center text-on-surface-variant/70 group-hover:scale-110 transition-transform">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                </div>
                <div>
                   <h3 class="text-on-surface font-bold">Video Source Management</h3>
                   <p class="text-xs text-on-surface-variant/40">Configure the streaming URL and player settings for this version.</p>
                </div>
            </div>
            <a 
                href="/admin/variants/{variant.id}/video" 
                class="px-5 py-2.5 bg-surface-highest hover:bg-surface-highest text-on-surface text-xs font-bold rounded-xl transition-all border border-outline-variant whitespace-nowrap"
            >
                {variant.video ? 'Manage Video' : 'Add Video Source'}
            </a>
        </div>
    </div>
  </div>

  <!-- Sidebar (Context) -->
  <div class="space-y-6">
    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 shadow-sm">
        <h3 class="text-xs font-black text-zinc-500 uppercase tracking-widest mb-4 border-b border-outline-variant pb-2">Parent Context</h3>
        <div class="space-y-4">
            {#if variant.song}
              <div>
                  <span class="text-[10px] text-gray-600 uppercase font-black tracking-tighter block mb-1">Song</span>
                  <a href="/admin/songs/{variant.song.id}" class="text-sm text-on-surface-variant hover:text-on-surface font-medium line-clamp-1">
                      {variant.song.song_romaji || "Song Detail"}
                  </a>
              </div>
            {/if}
            {#if variant.song?.anime}
              <div>
                  <span class="text-[10px] text-gray-600 uppercase font-black tracking-tighter block mb-1">Anime Series</span>
                  <a href="/admin/animes/{variant.song.anime.id}" class="text-sm text-on-surface-variant hover:text-on-surface font-medium line-clamp-1">
                      {variant.song.anime.title}
                  </a>
              </div>
            {/if}
        </div>
    </div>
  </div>
</div>
