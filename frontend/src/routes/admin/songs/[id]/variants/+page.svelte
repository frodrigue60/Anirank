<script lang="ts">
  import api from "$lib/api";
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";
  import { goto } from "$app/navigation";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let song = $state(data.song);
  // svelte-ignore state_referenced_locally
  let variants = $state(data.song.song_variants || []);

  $effect(() => {
    song = data.song;
    variants = data.song.song_variants || [];
  });

  let hasValidData = $derived(
    (song?.season_id > 0 || song?.anime?.season_id > 0) &&
      (song?.year_id > 0 || song?.anime?.year_id > 0),
  );

  let loading = $state(false);
  let status = $state(true);
  let errorMsg = $state("");

  async function createVariant() {
    loading = true;
    errorMsg = "";
    try {
      await api.post("/admin/variants", {
        song_id: song.id,
        status,
      });

      const refresh = await api.get(`/admin/songs/${song.id}`);
      song = refresh.data.data;
      variants = song.song_variants || [];
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to create variant";
    } finally {
      loading = false;
    }
  }

  async function toggleVariantStatus(variant: any) {
    const prev = variant.status;
    variant.status = !variant.status;
    variants = [...variants]; // trigger reactivity
    try {
      await api.put(`/admin/variants/${variant.id}`, {
        ...variant,
        status: variant.status,
      });
    } catch (err: any) {
      variant.status = prev;
      variants = [...variants];
      errorMsg =
        err.response?.data?.message || "Failed to update variant status";
    }
  }

  async function deleteVariant(id: number) {
    if (
      !confirm("Delete this variant? All associated videos will be unlinked.")
    )
      return;

    loading = true;
    try {
      await api.delete(`/admin/variants/${id}`);
      variants = variants.filter((v: any) => v.id !== id);
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to delete variant";
    } finally {
      loading = false;
    }
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between gap-4">
    <div>
      <h2 class="text-xl font-bold text-on-surface">Song Variants</h2>
      <p class="text-xs text-on-surface-variant/40">
        Version control and video sources for this theme.
      </p>
    </div>
    {#if !hasValidData}
      <button
        disabled
        class="px-4 py-2 bg-zinc-800 text-zinc-500 border border-zinc-700 text-xs font-bold rounded-xl flex items-center gap-2 cursor-not-allowed"
        title="Song or Anime must have Season and Year assigned"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        Add New Variant
      </button>
    {:else}
      <button
        onclick={createVariant}
        disabled={loading}
        class="px-4 py-2 bg-primary hover:bg-primary-container text-on-surface text-xs font-bold rounded-xl transition-all flex items-center gap-2"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        Add New Variant
      </button>
    {/if}
  </div>

  {#if !hasValidData}
    <div
      transition:slide
      class="p-4 bg-amber-500/10 border border-amber-500/20 rounded-2xl text-amber-500 text-sm flex items-start gap-3"
    >
      <svg
        class="w-5 h-5 shrink-0 mt-0.5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
      <div>
        <h3
          class="font-bold text-xs uppercase tracking-widest mb-1 text-amber-500"
        >
          Missing Information
        </h3>
        <p class="text-amber-500/80">
          This song (or its parent anime) must have a Season and Year assigned
          before you can create variants. Please edit the song or anime info
          first.
        </p>
      </div>
    </div>
  {/if}

  {#if errorMsg}
    <div
      transition:slide
      class="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-500 text-sm flex items-center gap-3"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        /></svg
      >
      {errorMsg}
    </div>
  {/if}

  <div
    class="bg-zinc-900/50 border border-zinc-800 rounded-3xl shadow-xl overflow-hidden text-on-surface"
  >
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-zinc-950/50 border-b border-zinc-800">
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest"
              >ID</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest"
              >Variant Slug</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest"
              >Status</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center"
              >Video</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-right"
              >Actions</th
            >
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-800/50">
          {#each variants as variant (variant.id)}
            <tr class="hover:bg-zinc-800/30 transition-colors group">
              <td class="px-6 py-4 text-sm font-mono text-zinc-500"
                >#{variant.id}</td
              >
              <td class="px-6 py-4">
                <div class="flex flex-col">
                  <span
                    class="text-sm font-bold text-on-surface group-hover:text-blue-400 transition-colors uppercase"
                  >
                    {variant.slug}
                  </span>
                  <span
                    class="text-[10px] text-zinc-500 uppercase font-black tracking-widest mt-0.5"
                  >
                    Version #{variant.version_number}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4">
                <button
                  onclick={() => toggleVariantStatus(variant)}
                  class="flex items-center gap-2 group/toggle"
                  title="Toggle status"
                >
                  <div class="relative">
                    <div
                      class="w-9 h-5 rounded-full transition-colors {variant.status
                        ? 'bg-emerald-500'
                        : 'bg-zinc-700'}"
                    ></div>
                    <div
                      class="absolute top-[2px] {variant.status
                        ? 'left-[18px]'
                        : 'left-[2px]'} w-4 h-4 bg-white rounded-full transition-all shadow"
                    ></div>
                  </div>
                  <span
                    class="text-[10px] font-black uppercase tracking-widest {variant.status
                      ? 'text-emerald-400'
                      : 'text-zinc-500'}"
                  >
                    {variant.status ? "Active" : "Inactive"}
                  </span>
                </button>
              </td>
              <td class="px-6 py-4 text-center">
                {#if variant.video}
                  <a
                    href="/admin/variants/{variant.id}/video"
                    class="inline-flex items-center px-3 py-1 bg-emerald-500/10 text-emerald-500 text-[10px] font-black uppercase rounded-lg border border-emerald-500/20 hover:bg-emerald-500 hover:text-on-surface transition-all tracking-widest"
                  >
                    <svg
                      class="w-4 h-4 mr-1.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                      /><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      /></svg
                    >
                    CONFIGURED
                  </a>
                {:else}
                  <span
                    class="inline-flex items-center px-3 py-1 bg-zinc-800 text-zinc-500 text-[10px] font-black uppercase rounded-lg border border-zinc-700 tracking-widest"
                  >
                    <svg
                      class="w-4 h-4 mr-1.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      /></svg
                    >
                    NO VIDEO
                  </span>
                {/if}
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center justify-end gap-2">
                  <!-- Edit Variant -->
                  <a
                    href="/admin/variants/{variant.id}/edit"
                    class="p-2 bg-zinc-800 hover:bg-primary-container text-zinc-400 hover:text-on-surface rounded-lg transition-all border border-zinc-700 hover:border-blue-500"
                    title="Edit Variant"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      /></svg
                    >
                  </a>
                  <a
                    href="/admin/variants/{variant.id}/video"
                    class="p-2 bg-zinc-800 hover:bg-emerald-600 text-zinc-400 hover:text-on-surface rounded-lg transition-all border border-zinc-700 hover:border-emerald-500"
                    title="Manage Video"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                      /></svg
                    >
                  </a>
                  <button
                    onclick={() => deleteVariant(variant.id)}
                    class="p-2 bg-zinc-800 hover:bg-red-600 text-zinc-400 hover:text-on-surface rounded-lg transition-all border border-zinc-700 hover:border-red-500"
                    title="Delete Variant"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      /></svg
                    >
                  </button>
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="4" class="px-6 py-12 text-center">
                <div class="flex flex-col items-center">
                  <svg
                    class="w-12 h-12 text-zinc-700 mb-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    /></svg
                  >
                  <h3
                    class="text-on-surface font-bold uppercase tracking-widest text-sm"
                  >
                    No Variants Found
                  </h3>
                  <p class="text-zinc-500 text-xs mt-2">
                    Add a new version for this song to begin configuration.
                  </p>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

<style>
  .overflow-x-auto::-webkit-scrollbar {
    height: 6px;
  }
  .overflow-x-auto::-webkit-scrollbar-track {
    background: transparent;
  }
  .overflow-x-auto::-webkit-scrollbar-thumb {
    background: #27272a;
    border-radius: 10px;
  }
  .overflow-x-auto::-webkit-scrollbar-thumb:hover {
    background: #3f3f46;
  }
</style>
