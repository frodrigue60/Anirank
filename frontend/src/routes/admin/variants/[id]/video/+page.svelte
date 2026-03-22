<script lang="ts">
  import api from "$lib/api";
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";
  import { goto } from "$app/navigation";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let variant = $state(data.variant);
  // svelte-ignore state_referenced_locally
  let song = $state(variant?.song || {});

  let loading = $state(false);
  let errorMsg = $state("");
  let successMsg = $state("");

  // svelte-ignore state_referenced_locally
  let status = $state(variant?.status || false);

  let videoFile: File | null = $state(null);
  // svelte-ignore state_referenced_locally
  let embedCode = $state(variant?.video?.embed_url || "");

  async function handleSave() {
    loading = true;
    errorMsg = "";
    successMsg = "";

    try {
      const formData = new FormData();
      if (videoFile) {
        formData.append("video", videoFile);
      } else if (embedCode) {
        formData.append("embed", embedCode);
      }
      
      formData.append("status", status ? "true" : "false");

      if (!videoFile && !embedCode && status === variant.status) {
        errorMsg = "Please provide either a video file, an embed code, or change the status.";
        loading = false;
        return;
      }

      const res = await api.put(
        `/admin/variants/${variant.id}/video`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        },
      );

      variant = res.data.data;
      successMsg = "Video asset updated successfully!";

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

<div class="space-y-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/songs/{song.id}/variants"
      title="Back to Variant List"
      aria-label="Back to Variant List"
      class="text-gray-400 hover:text-white transition-colors p-2 -ml-2 rounded-lg hover:bg-white/5"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 19l-7-7m0 0l7-7m-7 7h18"
        /></svg
      >
    </a>
    <div>
      <h1 class="text-3xl font-bold text-white tracking-tight">
        Link Video Asset
      </h1>
      <p
        class="text-zinc-400 mt-1 uppercase text-[10px] font-black tracking-widest pl-1"
      >
        {song.song_romaji || song.song_jp || song.song_en} - {variant.slug}
      </p>
    </div>
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
    class="bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-3xl shadow-xl overflow-hidden p-8 max-w-4xl"
  >
    <div class="space-y-8">
      <!-- Current Video Preview if exists -->
      {#if variant.video}
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
              {variant.video.embed_url || variant.video.local_url}
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
              MP4 or WEBM recommended (Max 100MB)
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
          class="block w-full bg-zinc-900 border border-zinc-800 text-white rounded-xl px-4 py-3 focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all text-sm font-mono placeholder:text-zinc-600"
          placeholder="Paste embed code or direct video URL..."
          oninput={() => {
            if (embedCode) videoFile = null;
          }}
        />
        <p class="text-[10px] text-zinc-500 italic">
          Supports direct .mp4/webm links or iframe embed snippets.
        </p>
      </div>

      <!-- Save Button -->
      <div class="pt-4">
        <button
          onclick={handleSave}
          disabled={loading}
          class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-black py-4 px-6 rounded-2xl transition-all shadow-lg shadow-blue-900/20 active:scale-[0.98] flex items-center justify-center gap-3 text-[10px] uppercase tracking-widest group"
        >
          {#if loading}
            <svg
              class="animate-spin h-4 w-4 text-white"
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
            SYNCHRONIZING ASSET...
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
