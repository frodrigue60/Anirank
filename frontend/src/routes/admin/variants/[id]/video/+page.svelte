<script lang="ts">
  import api from "$lib/api";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { goto } from "$app/navigation";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";
  import {
    fetchVariantVideoStorageCheck,
    normalizeStoragePath,
  } from "$lib/admin/videoStorage";
  import {
    appendVideoMetadataToFormData,
    metadataFormsEqual,
    metadataPreviewLabel,
    metadataTargetKey,
    videoToMetadataForm,
    VIDEO_OVERLAP_OPTIONS,
    VIDEO_RESOLUTION_OPTIONS,
    VIDEO_SOURCE_OPTIONS,
    DEFAULT_VIDEO_METADATA,
    type VideoMetadataForm,
  } from "$lib/admin/videoMetadata";
  import { getVideoTagText } from "$lib/song-utils";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let variant = $state(data.variant);
  let song = $derived(variant?.song || {});

  let loading = $state(false);
  let errorMsg = $state("");
  let successMsg = $state("");
  let storageCheckLoading = $state(false);
  let storageExists = $state<Record<string, boolean>>({});

  // svelte-ignore state_referenced_locally
  let status = $state(variant?.status || false);

  let videoFile: File | null = $state(null);
  // svelte-ignore state_referenced_locally
  let embedCode = $state(variant?.video?.embed_code || variant?.video?.embed_url || "");
  let uploadProgress = $state(0);
  let metadataTarget = $state("new");
  let metadataForm = $state<VideoMetadataForm>(
    videoToMetadataForm(variant?.video || variant?.videos?.[0]),
  );

  function findVideoByTarget(target: string) {
    return (variant?.videos || []).find((video) => metadataTargetKey(video) === target);
  }

  function baselineMetadata(): VideoMetadataForm {
    if (metadataTarget === "new") {
      return { ...DEFAULT_VIDEO_METADATA };
    }
    return videoToMetadataForm(findVideoByTarget(metadataTarget));
  }

  function loadMetadataForTarget(target: string) {
    metadataTarget = target;
    metadataForm =
      target === "new"
        ? { ...DEFAULT_VIDEO_METADATA }
        : videoToMetadataForm(findVideoByTarget(target));
  }

  function handleSourceChange(value: string) {
    metadataForm = {
      ...metadataForm,
      source: value,
      is_bd: value === "BD" ? true : metadataForm.is_bd,
    };
  }

  async function checkVideoStorage() {
    if (!variant?.id) return;
    storageCheckLoading = true;
    try {
      storageExists = await fetchVariantVideoStorageCheck(variant.id);
    } catch (err: any) {
      console.error(err);
      storageExists = {};
    } finally {
      storageCheckLoading = false;
    }
  }

  onMount(() => {
    checkVideoStorage();
  });

  async function handleSave() {
    loading = true;
    errorMsg = "";
    successMsg = "";
    uploadProgress = 0;

    try {
      const formData = new FormData();
      if (videoFile) {
        formData.append("video", videoFile);
      } else if (embedCode) {
        formData.append("embed", embedCode);
      }

      formData.append("status", status ? "true" : "false");

      const metadataChanged = !metadataFormsEqual(metadataForm, baselineMetadata());
      const hasUpload = Boolean(videoFile || embedCode.trim());
      const statusChanged = status !== variant.status;
      const metadataOnly =
        !hasUpload &&
        metadataChanged &&
        (metadataTarget !== "new" ||
          (variant.videos?.length === 1 && metadataTarget === "new"));

      if (!hasUpload && !metadataChanged && !statusChanged) {
        errorMsg =
          "Provide a video file, embed code, metadata changes, or change the status.";
        loading = false;
        return;
      }

      if (!hasUpload && metadataChanged && metadataTarget === "new" && (variant.videos?.length ?? 0) > 1) {
        errorMsg =
          "Select an existing video in metadata target before saving metadata only.";
        loading = false;
        return;
      }

      appendVideoMetadataToFormData(formData, metadataForm, {
        metadataTarget,
        metadataOnly,
      });

      const res = await api.put(
        `/admin/variants/${variant.id}/video`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
          onUploadProgress: (progressEvent) => {
            if (progressEvent.total) {
              uploadProgress = Math.round(
                (progressEvent.loaded * 100) / progressEvent.total
              );
            }
          },
        },
      );

      variant = res.data.data;
      successMsg = "Video asset updated successfully!";
      await checkVideoStorage();

      // Redirect back to Anime Hub -> Song detail after short delay for feedback
      setTimeout(() => {
        if (song.anime_id && song.id) {
          goto(`/admin/animes/${song.anime_id}/songs/${song.id}`);
        } else {
          // Fallback if relation somehow missed
          goto(`/admin/songs/${variant.song_id}`);
        }
      }, 1500);

      // Clear file selection after upload
      videoFile = null;
      const fileInput = document.getElementById(
        "formFileBanner",
      ) as HTMLInputElement;
      if (fileInput) fileInput.value = "";
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to update video asset";
    } finally {
      loading = false;
    }
  }

  function handleFileChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      videoFile = target.files[0];
      embedCode = ""; // Clear embed if file is selected
    }
  }
</script>

<div class="space-y-6">
  <div class="mb-6">
    <h2 class="text-xl font-bold text-on-surface">Video Source Management</h2>
    <p class="text-xs text-on-surface-variant/40">
      Configure the streaming URL and player settings for this version.
    </p>
  </div>

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

  {#if successMsg}
    <div
      transition:slide
      class="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl text-emerald-500 text-sm flex items-center gap-3"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M5 13l4 4L19 7"
        /></svg
      >
      {successMsg}
    </div>
  {/if}

  <div
    class="bg-zinc-900/50 border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8 max-w-4xl"
  >
    <div class="space-y-8">
      <!-- Registered Videos -->
      {#if variant.videos && variant.videos.length > 0}
        <div
          class="space-y-4 bg-zinc-950/50 p-6 rounded-2xl border border-blue-500/20 shadow-inner"
        >
          <div class="flex items-center justify-between gap-2">
            <h3
              class="text-[10px] font-black text-blue-400 uppercase tracking-[0.2em] flex items-center gap-2"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse"></span>
              Registered Videos ({variant.videos.length})
            </h3>
            <button
              type="button"
              onclick={checkVideoStorage}
              disabled={storageCheckLoading}
              class="text-[9px] font-black uppercase tracking-widest text-zinc-500 hover:text-zinc-300 disabled:opacity-50 flex items-center gap-1"
              title="Verificar archivos en R2/S3"
            >
              <span class="material-symbols-outlined text-xs {storageCheckLoading ? 'animate-spin' : ''}">sync</span>
              {storageCheckLoading ? "Verificando…" : "Verificar R2"}
            </button>
          </div>

          <div class="grid grid-cols-1 gap-2">
            {#each variant.videos as vid, i}
              {@const vidLabel = getVideoTagText(vid, i)}
              <div class="flex items-start gap-3 p-3 bg-black/40 border border-zinc-800/80 rounded-2xl">
                {#if vid.type === "file"}
                  <div class="flex items-center justify-center w-8 h-8 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shrink-0">
                    <span class="material-symbols-outlined text-base">movie</span>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-1.5 flex-wrap">
                      <span class="text-[11px] font-bold text-zinc-200 uppercase tracking-normal">{vidLabel}</span>
                      <span class="bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider">File</span>
                      {#if storageCheckLoading}
                        <span class="bg-zinc-800 text-zinc-500 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider flex items-center gap-0.5">
                          <span class="material-symbols-outlined text-[10px] animate-spin">progress_activity</span>
                          R2
                        </span>
                      {:else if vid.video_src}
                        {@const storageKey = normalizeStoragePath(vid.video_src)}
                        {#if storageExists[storageKey] === true}
                          <span class="bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider flex items-center gap-0.5" title="Archivo presente en R2/S3">
                            <span class="material-symbols-outlined text-[10px]">check_circle</span>
                            En R2
                          </span>
                        {:else if storageExists[storageKey] === false}
                          <span class="bg-red-500/10 text-red-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider flex items-center gap-0.5" title="Ruta en DB pero archivo no encontrado en R2/S3">
                            <span class="material-symbols-outlined text-[10px]">cancel</span>
                            No en R2
                          </span>
                        {/if}
                      {/if}
                    </div>
                    <p class="text-[9px] font-mono text-zinc-500 truncate mt-1 select-all">{vid.video_src}</p>
                    <button
                      type="button"
                      class="mt-2 text-[9px] font-black uppercase tracking-widest text-blue-400 hover:text-blue-300"
                      onclick={() => loadMetadataForTarget(metadataTargetKey(vid))}
                    >
                      Edit metadata
                    </button>
                  </div>
                {:else}
                  <div class="flex items-center justify-center w-8 h-8 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20 shrink-0">
                    <span class="material-symbols-outlined text-base">code</span>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-1.5 flex-wrap">
                      <span class="text-[11px] font-bold text-zinc-200 uppercase tracking-normal">{vidLabel}</span>
                      <span class="bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider">Embed</span>
                    </div>
                    <p class="text-[9px] font-mono text-zinc-500 truncate mt-1 select-all">{vid.embed_code || vid.embed_url}</p>
                    <button
                      type="button"
                      class="mt-2 text-[9px] font-black uppercase tracking-widest text-blue-400 hover:text-blue-300"
                      onclick={() => loadMetadataForTarget(metadataTargetKey(vid))}
                    >
                      Edit metadata
                    </button>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {:else if variant.video}
        <div
          class="space-y-4 bg-zinc-950/50 p-6 rounded-2xl border border-blue-500/20 shadow-inner"
        >
          <h3
            class="text-[10px] font-black text-blue-400 uppercase tracking-[0.2em] flex items-center gap-2"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse"
            ></span>
            Active Resource
          </h3>
          <div class="flex items-center gap-4 text-sm">
            <div
              class="px-3 py-1 bg-blue-500/10 rounded-lg border border-blue-500/20 text-blue-400 font-mono text-[10px] font-bold tracking-widest"
            >
              {variant.video.type.toUpperCase()}
            </div>
            <div class="text-zinc-400 truncate max-w-md italic text-xs">
              {variant.video.embed_code || variant.video.embed_url || variant.video.video_src || variant.video.local_url}
            </div>
          </div>
        </div>
      {/if}

      <!-- Status Management -->
      <div
        class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50"
      >
        <label
          for="status"
          class="text-[10px] font-black text-zinc-400 uppercase tracking-widest flex items-center gap-2"
        >
          <svg
            class="w-4 h-4 text-amber-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
            /><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
            /></svg
          >
          Visibility Status
        </label>
        <StatusControl bind:status />
      </div>

      <!-- File Upload Section -->
      <div
        class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50 hover:bg-zinc-950/40 transition-colors"
      >
        <label
          for="formFileBanner"
          class="text-[10px] font-black text-zinc-400 uppercase tracking-widest flex items-center gap-2"
        >
          <svg
            class="w-4 h-4 text-blue-500"
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
          Update Video Source (Local File)
        </label>

        <div
          class="mt-2 flex justify-center px-6 pt-5 pb-6 border-2 border-zinc-800 border-dashed rounded-2xl hover:border-blue-500/50 transition-colors group relative"
        >
          <div class="space-y-1 text-center">
            <svg
              class="mx-auto h-12 w-12 text-zinc-600 group-hover:text-blue-400 transition-colors"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-8l-4-4m0 0L8 8m4-4v12"
              />
            </svg>
            <div class="flex text-sm text-zinc-400">
              <label
                for="formFileBanner"
                class="relative cursor-pointer bg-transparent rounded-md font-bold text-blue-500 hover:text-blue-400 transition-colors"
              >
                <span>{videoFile ? videoFile.name : "Select a video file"}</span
                >
                <input
                  id="formFileBanner"
                  name="video"
                  type="file"
                  class="sr-only"
                  accept="video/mp4,video/webm"
                  onchange={handleFileChange}
                />
              </label>
              <p class="pl-1">or drag and drop</p>
            </div>
            <p class="text-[10px] text-zinc-500 font-medium italic">
              MP4 or WEBM recommended (Max 200MB)
            </p>
          </div>
        </div>
      </div>

      <div class="relative py-4 flex items-center">
        <div class="flex-grow border-t border-zinc-800"></div>
        <span
          class="flex-shrink mx-4 text-zinc-700 text-[10px] font-black uppercase tracking-[0.3em]"
          >OR USE CLOUD SOURCE</span
        >
        <div class="flex-grow border-t border-zinc-800"></div>
      </div>

      <!-- Embed Section -->
      <div
        class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50"
      >
        <label
          for="embed"
          class="text-[10px] font-black text-zinc-400 uppercase tracking-widest flex items-center gap-2"
        >
          <svg
            class="w-4 h-4 text-emerald-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
            /></svg
          >
          Direct Link / Embed Code
        </label>
        <input
          type="text"
          bind:value={embedCode}
          id="embed"
          class="block w-full bg-zinc-900 border border-zinc-800 text-on-surface rounded-xl px-4 py-3 focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all text-sm font-mono placeholder:text-zinc-600"
          placeholder="Paste embed code or direct video URL..."
          oninput={() => {
            if (embedCode) videoFile = null;
          }}
        />
        <p class="text-[10px] text-zinc-500 italic">
          Supports direct .mp4/webm links or iframe embed snippets.
        </p>
      </div>

      <!-- Video Metadata -->
      <div
        class="space-y-4 bg-zinc-950/30 p-6 rounded-2xl border border-zinc-800/50"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <label
            for="metadata-target"
            class="text-[10px] font-black text-zinc-400 uppercase tracking-widest"
          >
            Video Metadata
          </label>
          <span class="text-[10px] font-bold text-primary">
            Preview: {metadataPreviewLabel(metadataForm)}
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <label for="metadata-target" class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Apply to</label>
            <select
              id="metadata-target"
              class="block w-full bg-zinc-900 border border-zinc-800 text-on-surface rounded-xl px-3 py-2.5 text-sm"
              value={metadataTarget}
              onchange={(e) => loadMetadataForTarget((e.currentTarget as HTMLSelectElement).value)}
            >
              <option value="new">New upload / embed</option>
              {#each variant.videos || [] as vid, i}
                <option value={metadataTargetKey(vid)}>
                  {getVideoTagText(vid, i)}
                </option>
              {/each}
            </select>
          </div>

          <div class="space-y-2">
            <label for="video-source" class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Source</label>
            <select
              id="video-source"
              class="block w-full bg-zinc-900 border border-zinc-800 text-on-surface rounded-xl px-3 py-2.5 text-sm"
              value={metadataForm.source}
              onchange={(e) => handleSourceChange((e.currentTarget as HTMLSelectElement).value)}
            >
              {#each VIDEO_SOURCE_OPTIONS as option}
                <option value={option}>{option}</option>
              {/each}
            </select>
          </div>

          <div class="space-y-2">
            <label for="video-resolution" class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Resolution</label>
            <select
              id="video-resolution"
              class="block w-full bg-zinc-900 border border-zinc-800 text-on-surface rounded-xl px-3 py-2.5 text-sm"
              bind:value={metadataForm.resolution}
            >
              {#each VIDEO_RESOLUTION_OPTIONS as option}
                <option value={option}>{option === 0 ? "None" : `${option}p`}</option>
              {/each}
            </select>
          </div>

          <div class="space-y-2">
            <label for="video-overlap" class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Overlap</label>
            <select
              id="video-overlap"
              class="block w-full bg-zinc-900 border border-zinc-800 text-on-surface rounded-xl px-3 py-2.5 text-sm"
              bind:value={metadataForm.overlap}
            >
              {#each VIDEO_OVERLAP_OPTIONS as option}
                <option value={option}>{option}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="flex flex-wrap gap-4 pt-1">
          <label class="inline-flex items-center gap-2 text-xs text-zinc-300">
            <input type="checkbox" bind:checked={metadataForm.is_nc} class="rounded border-zinc-700 bg-zinc-900" />
            NC
          </label>
          {#if metadataForm.source !== "BD"}
            <label class="inline-flex items-center gap-2 text-xs text-zinc-300">
              <input type="checkbox" bind:checked={metadataForm.is_bd} class="rounded border-zinc-700 bg-zinc-900" />
              BD
            </label>
          {/if}
          <label class="inline-flex items-center gap-2 text-xs text-zinc-300">
            <input type="checkbox" bind:checked={metadataForm.is_uncensored} class="rounded border-zinc-700 bg-zinc-900" />
            Uncensored
          </label>
          <label class="inline-flex items-center gap-2 text-xs text-zinc-300">
            <input type="checkbox" bind:checked={metadataForm.is_subbed} class="rounded border-zinc-700 bg-zinc-900" />
            Subbed
          </label>
          <label class="inline-flex items-center gap-2 text-xs text-zinc-300">
            <input type="checkbox" bind:checked={metadataForm.is_lyrics} class="rounded border-zinc-700 bg-zinc-900" />
            Lyrics
          </label>
        </div>
      </div>

      <!-- Progress Bar -->
      {#if loading && videoFile}
        <div class="space-y-2" transition:slide>
          <div class="flex justify-between items-center text-xs text-zinc-400 font-medium">
            <span>Uploading "{videoFile.name}"</span>
            <span class="font-bold text-blue-400">{uploadProgress}%</span>
          </div>
          <div class="w-full bg-zinc-800 rounded-full h-2 overflow-hidden border border-zinc-700/50">
            <div 
              class="bg-blue-500 h-full rounded-full transition-all duration-300 ease-out" 
              style="width: {uploadProgress}%"
            ></div>
          </div>
        </div>
      {/if}

      <!-- Save Button -->
      <div class="pt-4">
        <button
          onclick={handleSave}
          disabled={loading}
          class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-on-surface font-black py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-3 text-[10px] uppercase tracking-widest group"
        >
          {#if loading}
            <svg
              class="animate-spin h-4 w-4 text-on-surface"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {#if videoFile}
              UPLOADING VIDEO ({uploadProgress}%)...
            {:else}
              SYNCHRONIZING ASSET...
            {/if}
          {:else}
            <svg
              class="w-4 h-4 group-hover:scale-110 transition-transform"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"
              /></svg
            >
            COMMIT VIDEO CHANGES
          {/if}
        </button>
      </div>
    </div>
  </div>
</div>
