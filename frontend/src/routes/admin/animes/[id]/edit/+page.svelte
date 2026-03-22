<script lang="ts">
  import { goto } from "$app/navigation";
  import { configState as config } from "$lib/state/config.svelte";
  import TagsInput from "$lib/components/admin/TagsInput.svelte";
  import api from "$lib/api";
  import type { PageData } from "./$types";

  let { data } = $props<{ data: PageData }>();
  // svelte-ignore state_referenced_locally
  const anime = data.anime;

  // Form State
  let title = $state(anime.title || "");
  let description = $state(anime.description || "");
  let status = $state(!!anime.status);
  let year_id = $state(anime.year_id || 0);
  let season_id = $state(anime.season_id || 0);
  let format_id = $state(anime.format_id || 0);
  let anilist_id = $state<number | null>(anime.anilist_id || null);

  // Tags State
  let studiosString = $state(
    anime.studios ? anime.studios.map((s: any) => s.name).join(", ") : "",
  );
  let producersString = $state(
    anime.producers ? anime.producers.map((s: any) => s.name).join(", ") : "",
  );
  let genresString = $state(
    anime.genres ? anime.genres.map((s: any) => s.name).join(", ") : "",
  );

  // Files
  let coverFile: File | null = $state(null);
  let bannerFile: File | null = $state(null);

  // Existing images
  let coverPreview = $state(anime.cover_url || "");
  let bannerPreview = $state(anime.banner_url || "");

  // UI State
  let loading = $state(false);
  let errorMsg = $state("");

  // Handlers for file inputs
  function onCoverChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      coverFile = target.files[0];
      coverPreview = URL.createObjectURL(coverFile);
    }
  }

  function onBannerChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      bannerFile = target.files[0];
      bannerPreview = URL.createObjectURL(bannerFile);
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;
    errorMsg = "";

    try {
      const formData = new FormData();
      formData.append("title", title);
      formData.append("description", description);
      formData.append("status", status ? "1" : "0");
      if (year_id > 0) formData.append("year_id", year_id.toString());
      if (season_id > 0) formData.append("season_id", season_id.toString());
      if (format_id > 0) formData.append("format_id", format_id.toString());
      if (anilist_id) formData.append("anilist_id", anilist_id.toString());

      // Relations
      formData.append("studios", studiosString);
      formData.append("producers", producersString);
      formData.append("genres", genresString);

      if (coverFile) {
        formData.append("cover", coverFile);
      }
      if (bannerFile) {
        formData.append("banner", bannerFile);
      }

      const res = await api.put(`/admin/animes/${anime.id}`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (res.status === 200) {
        goto("/admin/animes");
      }
    } catch (err: any) {
      console.error(err);
      errorMsg =
        err.response?.data?.message ||
        "An error occurred while updating the anime.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Anime | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/animes"
      title="Back to Animes List"
      aria-label="Back to Animes List"
      class="text-gray-400 hover:text-white transition-colors p-2 -ml-2 rounded-lg hover:bg-white/5"
    >
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 19l-7-7m0 0l7-7m-7 7h18"
        />
      </svg>
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-white">Edit Anime</h1>
  </div>
  <p class="text-gray-400 ml-10">
    Updating: <span class="font-medium text-gray-200">{anime.title}</span>
  </p>
</div>

{#if errorMsg}
  <div
    class="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl mb-6 flex gap-3"
  >
    <svg
      class="w-5 h-5 shrink-0"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
      /></svg
    >
    <p>{errorMsg}</p>
  </div>
{/if}

<form onsubmit={handleSubmit} class="space-y-6 max-w-4xl">
  <!-- General Info -->
  <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
    <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
      <svg
        class="w-5 h-5 text-gray-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        /></svg
      >
      General Information
    </h2>

    <div class="space-y-4">
      <div>
        <label for="title" class="block text-sm font-medium text-gray-300 mb-1"
          >Title <span class="text-red-400">*</span></label
        >
        <input
          type="text"
          id="title"
          bind:value={title}
          required
          class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary focus:ring-1 focus:ring-anirank-primary transition-all"
        />
      </div>

      <div>
        <label
          for="description"
          class="block text-sm font-medium text-gray-300 mb-1"
          >Description</label
        >
        <textarea
          id="description"
          bind:value={description}
          rows="4"
          class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary focus:ring-1 focus:ring-anirank-primary transition-all"
        ></textarea>
      </div>
    </div>
  </div>

  <!-- Taxonomies & Metadata -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
      <div class="space-y-6">
        <div>
          <label
            for="studios"
            class="block text-sm font-medium text-gray-300 mb-2">Studios</label
          >
          <TagsInput
            endpoint="/admin/studios"
            bind:value={studiosString}
            placeholder="e.g. Mappa, Wit Studio"
            entityName="Studio"
          />
        </div>

        <div>
          <label
            for="producers"
            class="block text-sm font-medium text-gray-300 mb-2"
            >Producers</label
          >
          <TagsInput
            endpoint="/admin/producers"
            bind:value={producersString}
            placeholder="e.g. Aniplex, Kadokawa"
            entityName="Producer"
          />
        </div>

        <div>
          <label
            for="genres"
            class="block text-sm font-medium text-gray-300 mb-2">Genres</label
          >
          <TagsInput
            endpoint="/admin/genres"
            bind:value={genresString}
            placeholder="e.g. Action, Romance"
            entityName="Genre"
          />
        </div>
      </div>

      <div class="space-y-4">
        <div>
          <label for="year" class="block text-sm font-medium text-gray-300 mb-1"
            >Year</label
          >
          <select
            id="year"
            bind:value={year_id}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-all [&>option]:bg-anirank-card"
          >
            <option value={0}>Select Year</option>
            {#each config.years as year}
              <option value={year.id}>{year.name}</option>
            {/each}
          </select>
        </div>

        <div>
          <label
            for="season"
            class="block text-sm font-medium text-gray-300 mb-1">Season</label
          >
          <select
            id="season"
            bind:value={season_id}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-all [&>option]:bg-anirank-card"
          >
            <option value={0}>Select Season</option>
            {#each config.seasons as season}
              <option value={season.id}>{season.name}</option>
            {/each}
          </select>
        </div>

        <div>
          <label
            for="format"
            class="block text-sm font-medium text-gray-300 mb-1">Format</label
          >
          <select
            id="format"
            bind:value={format_id}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-all [&>option]:bg-anirank-card"
          >
            <option value={0}>Select Format</option>
            {#each config.formats as format}
              <option value={format.id}>{format.name}</option>
            {/each}
          </select>
        </div>
      </div>
    </div>

    <!-- External & States -->
    <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
      <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
        <svg
          class="w-5 h-5 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
          /></svg
        >
        Connections & Status
      </h2>

      <div class="space-y-6">
        <div>
          <label
            for="anilist_id"
            class="block text-sm font-medium text-gray-300 mb-1"
            >Anilist ID</label
          >
          <input
            type="number"
            id="anilist_id"
            bind:value={anilist_id}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all"
          />
        </div>

        <div class="pt-2">
          <label class="flex items-center gap-3 cursor-pointer">
            <div class="relative">
              <input
                type="checkbox"
                bind:checked={status}
                class="sr-only peer"
              />
              <div
                class="w-11 h-6 bg-white/10 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-anirank-primary"
              ></div>
            </div>
            <span class="text-sm font-medium text-white"
              >Publish Anime (Active)</span
            >
          </label>
          <p class="text-xs text-gray-500 mt-2">
            If unpublished, it will remain as a draft exclusively in the admin
            panel.
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Assets -->
  <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
    <h2 class="text-xl font-semibold text-white mb-6 flex items-center gap-2">
      <svg
        class="w-5 h-5 text-gray-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
        /></svg
      >
      Media Assets
    </h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Cover -->
      <div>
        <label for="cover" class="block text-sm font-medium text-gray-300 mb-2"
          >Cover Image</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="cover"
            class="flex flex-col items-center justify-center w-full h-40 border-2 border-white/10 border-dashed rounded-xl cursor-pointer bg-white/5 hover:bg-white/10 hover:border-white/20 transition-all overflow-hidden relative"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-4 text-center z-10"
            >
              <svg
                class="w-8 h-8 mb-3 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                /></svg
              >
              <p class="mb-1 text-sm text-gray-300">
                <span class="font-semibold">Click to upload</span>
              </p>
              <p class="text-xs text-gray-500">PNG, JPG up to 2MB</p>
            </div>
            <!-- PREVIEW -->
            {#if coverPreview}
              <div
                class="absolute inset-0 z-0 opacity-80 select-none bg-black/50"
              >
                <span
                  class="absolute inset-0 flex items-center justify-center text-white bg-black/60 z-10 font-medium opacity-0 hover:opacity-100 transition-opacity"
                  >Change Image</span
                >
                <img
                  src={coverPreview}
                  alt="Preview"
                  class="w-full h-full object-cover"
                />
              </div>
            {/if}
            <input
              id="cover"
              type="file"
              accept="image/*"
              class="hidden"
              onchange={onCoverChange}
            />
          </label>
        </div>
      </div>

      <!-- Banner -->
      <div>
        <label for="banner" class="block text-sm font-medium text-gray-300 mb-2"
          >Banner</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="banner"
            class="flex flex-col items-center justify-center w-full h-40 border-2 border-white/10 border-dashed rounded-xl cursor-pointer bg-white/5 hover:bg-white/10 hover:border-white/20 transition-all overflow-hidden relative"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-4 text-center z-10"
            >
              <svg
                class="w-8 h-8 mb-3 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                /></svg
              >
              <p class="mb-1 text-sm text-gray-300">
                <span class="font-semibold">Click to upload</span>
              </p>
              <p class="text-xs text-gray-500">Wide aspect ratio</p>
            </div>
            <!-- PREVIEW -->
            {#if bannerPreview}
              <div
                class="absolute inset-0 z-0 opacity-80 select-none bg-black/50"
              >
                <span
                  class="absolute inset-0 flex items-center justify-center text-white bg-black/60 z-10 font-medium opacity-0 hover:opacity-100 transition-opacity"
                  >Change Image</span
                >
                <img
                  src={bannerPreview}
                  alt="Preview"
                  class="w-full h-full object-cover"
                />
              </div>
            {/if}
            <input
              id="banner"
              type="file"
              accept="image/*"
              class="hidden"
              onchange={onBannerChange}
            />
          </label>
        </div>
      </div>
    </div>
  </div>

  <div class="flex items-center justify-end gap-3 pt-4 border-t border-white/5">
    <a
      href="/admin/animes"
      class="px-5 py-2.5 text-sm font-medium text-gray-300 hover:text-white bg-white/5 hover:bg-white/10 rounded-xl transition-colors"
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={loading || !title}
      class="px-5 py-2.5 text-sm font-medium text-white bg-anirank-primary hover:bg-blue-600 rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
    >
      {#if loading}
        <svg
          class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
          fill="none"
          viewBox="0 0 24 24"
          ><circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle><path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path></svg
        >
        Saving...
      {:else}
        Update Anime
      {/if}
    </button>
  </div>
</form>
