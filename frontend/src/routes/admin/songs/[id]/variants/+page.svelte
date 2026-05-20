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
      await api.patch(`/admin/variants/${variant.id}/status`);
    } catch (err: any) {
      variant.status = prev;
      variants = [...variants];
      errorMsg =
        err.response?.data?.message || "Failed to update variant status";
    }
  }

  async function toggleVariantSpoiler(variant: any) {
    const prev = variant.spoiler;
    variant.spoiler = !variant.spoiler;
    variants = [...variants];
    try {
      await api.patch(`/admin/variants/${variant.id}/spoiler`);
    } catch (err: any) {
      variant.spoiler = prev;
      variants = [...variants];
      errorMsg = err.response?.data?.message || "Failed to toggle spoiler";
    }
  }

  async function toggleVariantNSFW(variant: any) {
    const prev = variant.nsfw;
    variant.nsfw = !variant.nsfw;
    variants = [...variants];
    try {
      await api.patch(`/admin/variants/${variant.id}/nsfw`);
    } catch (err: any) {
      variant.nsfw = prev;
      variants = [...variants];
      errorMsg = err.response?.data?.message || "Failed to toggle NSFW";
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

  async function detachVideo(variant: any, video: any) {
    if (
      !confirm(
        "Are you sure you want to detach this video? The video relation will be removed from the database, but the physical file in cloud storage will not be deleted.",
      )
    )
      return;

    loading = true;
    errorMsg = "";
    try {
      const src = video.video_src || "";
      const embed = video.embed_code || "";
      await api.delete(`/admin/variants/${variant.id}/video`, {
        params: {
          video_src: src,
          embed_code: embed,
          purge: false,
        },
      });

      // Update local state to reflect deletion
      if (variant.videos) {
        variant.videos = variant.videos.filter(
          (v: any) => v.video_src !== src || v.embed_code !== embed,
        );
      }
      if (
        variant.video &&
        variant.video.video_src === src &&
        variant.video.embed_code === embed
      ) {
        variant.video =
          variant.videos && variant.videos.length > 0
            ? variant.videos[0]
            : null;
      }
      variants = [...variants];
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to detach video";
    } finally {
      loading = false;
    }
  }

  async function deleteVideo(variant: any, video: any) {
    if (
      !confirm(
        "Are you sure you want to delete this video? This will remove the database relation AND permanently delete the file from the cloud storage. This action cannot be undone.",
      )
    )
      return;

    loading = true;
    errorMsg = "";
    try {
      const src = video.video_src || "";
      const embed = video.embed_code || "";
      await api.delete(`/admin/variants/${variant.id}/video`, {
        params: {
          video_src: src,
          embed_code: embed,
          purge: true,
        },
      });

      // Update local state to reflect deletion
      if (variant.videos) {
        variant.videos = variant.videos.filter(
          (v: any) => v.video_src !== src || v.embed_code !== embed,
        );
      }
      if (
        variant.video &&
        variant.video.video_src === src &&
        variant.video.embed_code === embed
      ) {
        variant.video =
          variant.videos && variant.videos.length > 0
            ? variant.videos[0]
            : null;
      }
      variants = [...variants];
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to delete video file";
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
          <tr class="bg-zinc-950/50 border-b border-zinc-800 hidden md:table-row">
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
              >Episodes</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest"
              >Status</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center"
              >Spoiler</th
            >
            <th
              class="px-6 py-4 text-xs font-bold text-zinc-400 uppercase tracking-widest text-center"
              >NSFW</th
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
            <tr class="hover:bg-zinc-800/30 transition-colors group flex flex-col md:table-row p-4 md:p-0">
              <!-- Row 1 (Mobile) / Cell 1 (Desktop) -->
              <td class="px-2 md:px-6 py-2 md:py-4 text-sm font-mono text-zinc-500 md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">ID</span>
                #{variant.id}
              </td>

              <td class="px-2 md:px-6 py-2 md:py-4 md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">Variant</span>
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

              <td class="px-2 md:px-6 py-2 md:py-4 md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">Episodes</span>
                <span class="text-xs font-bold text-zinc-300 bg-zinc-800 px-2 py-1 rounded-md border border-zinc-700">
                  {variant.episodes || "Full"}
                </span>
              </td>

              <td class="px-2 md:px-6 py-2 md:py-4 whitespace-nowrap md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">Status</span>
                <button
                  onclick={() => toggleVariantStatus(variant)}
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border transition-all duration-200 hover:scale-105 active:scale-95 {variant.status
                    ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/20'
                    : 'bg-orange-500/10 text-orange-400 border-orange-500/20 hover:bg-orange-500/20'}"
                  title="Toggle status"
                >
                  <div
                    class="w-1.5 h-1.5 rounded-full mr-2 {variant.status
                      ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.5)]'
                      : 'bg-orange-400 shadow-[0_0_8px_rgba(251,146,60,0.5)]'}"
                  ></div>
                  {variant.status ? "Active" : "Inactive"}
                </button>
              </td>

              <!-- Row 2 (Mobile) Container -->
              <td class="md:hidden" colspan="8">
                <div class="grid grid-cols-2 gap-4 mt-4 pt-4 border-t border-zinc-800">
                  <!-- Spoiler Toggle -->
                  <div class="space-y-1">
                    <span class="text-[10px] uppercase font-black text-zinc-600">Spoiler</span>
                    <button
                      onclick={() => toggleVariantSpoiler(variant)}
                      class="w-full flex items-center justify-center px-2.5 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest border transition-all {variant.spoiler
                        ? 'bg-red-500/10 text-red-400 border-red-500/20'
                        : 'bg-zinc-800 text-zinc-500 border-zinc-700'}"
                    >
                      {variant.spoiler ? "Spoiler" : "Clean"}
                    </button>
                  </div>

                  <!-- NSFW Toggle -->
                  <div class="space-y-1">
                    <span class="text-[10px] uppercase font-black text-zinc-600">NSFW</span>
                    <button
                      onclick={() => toggleVariantNSFW(variant)}
                      class="w-full flex items-center justify-center px-2.5 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest border transition-all {variant.nsfw
                        ? 'bg-fuchsia-500/10 text-fuchsia-400 border-fuchsia-500/20'
                        : 'bg-zinc-800 text-zinc-500 border-zinc-700'}"
                    >
                      {variant.nsfw ? "NSFW" : "Safe"}
                    </button>
                  </div>
                </div>
              </td>

              <!-- Desktop-only cells for Spoiler, NSFW, Video -->
              <td class="hidden md:table-cell px-6 py-4 text-center whitespace-nowrap">
                <button
                  onclick={() => toggleVariantSpoiler(variant)}
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border transition-all duration-200 hover:scale-105 active:scale-95 {variant.spoiler
                    ? 'bg-red-500/10 text-red-400 border-red-500/20 hover:bg-red-500/20'
                    : 'bg-zinc-800 text-zinc-500 border-zinc-700 hover:bg-zinc-700'}"
                  title="Toggle Spoiler"
                >
                  <div
                    class="w-1.5 h-1.5 rounded-full mr-2 {variant.spoiler
                      ? 'bg-red-400 shadow-[0_0_8px_rgba(239,68,68,0.5)]'
                      : 'bg-zinc-600'}"
                  ></div>
                  {variant.spoiler ? "Spoiler" : "Clean"}
                </button>
              </td>

              <td class="hidden md:table-cell px-6 py-4 text-center whitespace-nowrap">
                <button
                  onclick={() => toggleVariantNSFW(variant)}
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border transition-all duration-200 hover:scale-105 active:scale-95 {variant.nsfw
                    ? 'bg-fuchsia-500/10 text-fuchsia-400 border-fuchsia-500/20 hover:bg-fuchsia-500/20'
                    : 'bg-zinc-800 text-zinc-500 border-zinc-700 hover:bg-zinc-700'}"
                  title="Toggle NSFW"
                >
                  <div
                    class="w-1.5 h-1.5 rounded-full mr-2 {variant.nsfw
                      ? 'bg-fuchsia-400 shadow-[0_0_8px_rgba(192,38,211,0.5)]'
                      : 'bg-zinc-600'}"
                  ></div>
                  {variant.nsfw ? "NSFW" : "Safe"}
                </button>
              </td>

              <td class="px-2 md:px-6 py-2 md:py-4 md:text-center md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">Video</span>
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

              <td class="px-2 md:px-6 py-2 md:py-4 md:table-cell">
                <span class="md:hidden text-[10px] uppercase font-black text-zinc-600 block mb-1">Actions</span>
                <div class="flex items-center md:justify-end gap-2">
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

            <!-- Video Configuration Details -->
            <tr class="bg-zinc-900/30 border-t border-zinc-800/50">
              <td colspan="8" class="px-4 md:px-6 py-3">
                <div class="flex flex-col md:flex-row gap-4 md:gap-8 text-[10px] uppercase font-black tracking-widest">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1.5">
                      <span class="text-zinc-600">Embed Configuration</span>
                      {#if variant.embed_code || variant.video?.embed_code}
                        <span class="bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded text-[8px]">EMBED</span>
                      {/if}
                    </div>
                    <div class="bg-black/60 p-2.5 rounded-lg border border-zinc-800/50 font-mono text-[9px] text-zinc-500 break-all leading-relaxed line-clamp-1 hover:line-clamp-none transition-all cursor-text select-all">
                      {variant.embed_code || variant.video?.embed_code || "NOT CONFIGURED"}
                    </div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1.5">
                      <span class="text-zinc-600">Storage Reference</span>
                      {#if variant.video_src || variant.video?.video_src}
                        <span class="bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded text-[8px]">FILE</span>
                      {/if}
                    </div>
                    <div class="bg-black/60 p-2.5 rounded-lg border border-zinc-800/50 font-mono text-[9px] text-emerald-400/50 break-all truncate hover:text-emerald-400 transition-colors cursor-text select-all">
                      {variant.video_src || variant.video?.video_src || "NOT CONFIGURED"}
                    </div>
                  </div>
                </div>

                <!-- Video Manager -->
                <div class="mt-4 pt-4 border-t border-zinc-800/50">
                  <div class="text-[10px] uppercase font-black tracking-widest text-zinc-500 mb-3 flex items-center gap-1.5">
                    <span class="material-symbols-outlined text-sm">settings</span>
                    Video Manager
                  </div>
                  {#if variant.videos && variant.videos.length > 0}
                    <div class="grid grid-cols-1 gap-2">
                      {#each variant.videos as vid}
                        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 bg-black/40 border border-zinc-800/80 rounded-2xl transition-all hover:border-zinc-700/50">
                          <div class="flex items-center gap-3 min-w-0">
                            {#if vid.type === 'file'}
                              <div class="flex items-center justify-center w-8 h-8 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shrink-0">
                                <span class="material-symbols-outlined text-base">movie</span>
                              </div>
                              <div class="min-w-0 flex-1">
                                <div class="flex items-center gap-1.5 flex-wrap">
                                  <span class="text-[11px] font-bold text-zinc-200 uppercase tracking-normal">Cloud Storage File</span>
                                  <span class="bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider">{vid.source || 'TV'}</span>
                                  {#if vid.resolution > 0}
                                    <span class="bg-zinc-800 text-zinc-400 px-1 py-0.5 rounded text-[8px] font-mono">{vid.resolution}p</span>
                                  {/if}
                                  {#if vid.is_nc}
                                    <span class="bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded text-[8px] font-black">NC</span>
                                  {/if}
                                </div>
                                <p class="text-[9px] font-mono text-zinc-500 truncate mt-1 select-all">{vid.video_src}</p>
                              </div>
                            {:else}
                              <div class="flex items-center justify-center w-8 h-8 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20 shrink-0">
                                <span class="material-symbols-outlined text-base">code</span>
                              </div>
                              <div class="min-w-0 flex-1">
                                <div class="flex items-center gap-1.5 flex-wrap">
                                  <span class="text-[11px] font-bold text-zinc-200 uppercase tracking-normal">Embed / Iframe Code</span>
                                  <span class="bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider">Embed</span>
                                </div>
                                <p class="text-[9px] font-mono text-zinc-500 truncate mt-1 select-all">{vid.embed_code}</p>
                              </div>
                            {/if}
                          </div>
                          <div class="flex items-center gap-2 shrink-0 self-end sm:self-center">
                            <button
                              onclick={() => detachVideo(variant, vid)}
                              disabled={loading}
                              class="px-2.5 py-1.5 bg-amber-500/10 hover:bg-amber-500 text-amber-400 hover:text-zinc-950 text-[9px] font-black uppercase tracking-widest rounded-xl border border-amber-500/20 hover:border-amber-500 transition-all flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
                              title="Remove video relationship but keep file in cloud storage"
                            >
                              <span class="material-symbols-outlined text-xs">link_off</span>
                              Detach
                            </button>
                            <button
                              onclick={() => deleteVideo(variant, vid)}
                              disabled={loading}
                              class="px-2.5 py-1.5 bg-red-500/10 hover:bg-red-600 text-red-400 hover:text-on-surface text-[9px] font-black uppercase tracking-widest rounded-xl border border-red-500/20 hover:border-red-500 transition-all flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
                              title="Delete video relationship and purge file from cloud storage"
                            >
                              <span class="material-symbols-outlined text-xs">delete_forever</span>
                              Delete
                            </button>
                          </div>
                        </div>
                      {/each}
                    </div>
                  {#else}
                    <div class="flex items-center justify-center p-4 bg-black/20 border border-zinc-800/50 rounded-2xl text-zinc-600 text-xs">
                      No videos attached to this variant. Use the 'Manage Video' button above to configure one.
                    </div>
                  {/if}
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="8" class="px-6 py-12 text-center">
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
