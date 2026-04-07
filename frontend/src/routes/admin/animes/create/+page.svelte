<script lang="ts">
  import { goto } from "$app/navigation";
  import { configState as config } from "$lib/state/config.svelte";
  import TagsInput from "$lib/components/admin/TagsInput.svelte";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";
  import api from "$lib/api";

  let { data } = $props();
  // Form State
  let title = $state("");
  let description = $state("");
  let status = $state(true);
  let year_id = $state(0);
  let season_id = $state(0);
  let format_id = $state(0);
  let anilist_id = $state<number | null>(null);

  // Tags State
  let studiosString = $state("");
  let producersString = $state("");
  let genresString = $state("");

  // Files
  let coverFile: File | null = $state(null);
  let bannerFile: File | null = $state(null);

  // UI State
  let loading = $state(false);
  let errorMsg = $state("");

  // Handlers for file inputs
  function onCoverChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      coverFile = target.files[0];
    }
  }

  function onBannerChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      bannerFile = target.files[0];
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

      const res = await api.post("/admin/animes", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (res.status === 201) {
        goto("/admin/animes");
      }
    } catch (err: any) {
      console.error(err);
      errorMsg =
        err.response?.data?.message ||
        "An error occurred while creating the anime.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Create Anime | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/animes"
      aria-label="Back to Animes"
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
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
    <h1 class="text-3xl font-bold tracking-tight text-on-surface">
      Create New Anime
    </h1>
  </div>
  <p class="text-on-surface-variant/70 ml-10">Add a new anime entry to the catalog.</p>
</div>

{#if errorMsg}
  <div
    class="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl mb-6 flex gap-3"
  >
    <svg
      class="w-5 h-5 flex-shrink-0"
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
  <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
    <h2 class="text-xl font-semibold text-on-surface mb-6 flex items-center gap-2">
      <svg
        class="w-5 h-5 text-on-surface-variant/70"
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
        <label for="title" class="block text-sm font-medium text-on-surface-variant mb-1"
          >Title <span class="text-red-400">*</span></label
        >
        <input
          type="text"
          id="title"
          bind:value={title}
          required
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none transition-all"
          placeholder="e.g. Shingeki no Kyojin"
        />
      </div>

      <div>
        <label
          for="description"
          class="block text-sm font-medium text-on-surface-variant mb-1"
          >Description</label
        >
        <textarea
          id="description"
          bind:value={description}
          rows="4"
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none transition-all"
          placeholder="Synopsis or brief description..."
        ></textarea>
      </div>
    </div>
  </div>

  <!-- Taxonomies & Metadata -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
      <h2 class="text-xl font-semibold text-on-surface mb-6 flex items-center gap-2">
        <svg
          class="w-5 h-5 text-on-surface-variant/70"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
          /></svg
        >
        Taxonomy
      </h2>

      <div class="space-y-6">
        <div>
          <label
            for="studios"
            class="block text-sm font-medium text-on-surface-variant mb-2">Studios</label
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
            class="block text-sm font-medium text-on-surface-variant mb-2"
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
            class="block text-sm font-medium text-on-surface-variant mb-2">Genres</label
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
          <label for="year" class="block text-sm font-medium text-on-surface-variant mb-1"
            >Year</label
          >
          <select
            id="year"
            bind:value={year_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
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
            class="block text-sm font-medium text-on-surface-variant mb-1">Season</label
          >
          <select
            id="season"
            bind:value={season_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
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
            class="block text-sm font-medium text-on-surface-variant mb-1">Format</label
          >
          <select
            id="format"
            bind:value={format_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
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
    <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
      <h2 class="text-xl font-semibold text-on-surface mb-6 flex items-center gap-2">
        <svg
          class="w-5 h-5 text-on-surface-variant/70"
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
            class="block text-sm font-medium text-on-surface-variant mb-1"
            >Anilist ID</label
          >
          <input
            type="number"
            id="anilist_id"
            title="Anilist ID"
            bind:value={anilist_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
            placeholder="e.g. 16498"
          />
          <p class="text-xs text-on-surface-variant/40 mt-1">
            Leave empty if it's a manual entry not tied to Anilist.
          </p>
        </div>

        <div class="pt-2">
          <StatusControl bind:status={status} />
        </div>
      </div>
    </div>
  </div>

  <!-- Assets -->
  <div class="bg-surface-container border border-outline-variant rounded-2xl p-6">
    <h2 class="text-xl font-semibold text-on-surface mb-6 flex items-center gap-2">
      <svg
        class="w-5 h-5 text-on-surface-variant/70"
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
      <!-- Thumbnail -->
      <div>
        <label for="cover" class="block text-sm font-medium text-on-surface-variant mb-2"
          >Thumbnail (Cover)</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="cover"
            class="flex flex-col items-center justify-center w-full h-40 border-2 border-outline-variant border-dashed rounded-xl cursor-pointer bg-surface-highest hover:bg-surface-highest hover:border-outline-variant transition-all overflow-hidden relative"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-4 text-center z-10"
            >
              <svg
                class="w-8 h-8 mb-3 text-on-surface-variant/70"
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
              <p class="mb-1 text-sm text-on-surface-variant">
                <span class="font-semibold">Click to upload</span>
              </p>
              <p class="text-xs text-on-surface-variant/40">PNG, JPG up to 2MB</p>
            </div>
            <!-- PREVIEW -->
            {#if coverFile}
              <div class="absolute inset-0 z-0 opacity-30 select-none">
                <span
                  class="absolute inset-0 flex items-center justify-center text-on-surface bg-black/50 z-10 font-medium"
                  >Selected</span
                >
                <img
                  src={URL.createObjectURL(coverFile)}
                  alt="Preview"
                  class="w-full h-full object-cover blur-sm"
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
        <label for="banner" class="block text-sm font-medium text-on-surface-variant mb-2"
          >Banner</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="banner"
            class="flex flex-col items-center justify-center w-full h-40 border-2 border-outline-variant border-dashed rounded-xl cursor-pointer bg-surface-highest hover:bg-surface-highest hover:border-outline-variant transition-all overflow-hidden relative"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-4 text-center z-10"
            >
              <svg
                class="w-8 h-8 mb-3 text-on-surface-variant/70"
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
              <p class="mb-1 text-sm text-on-surface-variant">
                <span class="font-semibold">Click to upload</span>
              </p>
              <p class="text-xs text-on-surface-variant/40">Wide aspect ratio</p>
            </div>
            <!-- PREVIEW -->
            {#if bannerFile}
              <div class="absolute inset-0 z-0 opacity-30 select-none">
                <span
                  class="absolute inset-0 flex items-center justify-center text-on-surface bg-black/50 z-10 font-medium"
                  >Selected</span
                >
                <img
                  src={URL.createObjectURL(bannerFile)}
                  alt="Preview"
                  class="w-full h-full object-cover blur-sm"
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

  <div class="flex items-center justify-end gap-3 pt-4 border-t border-outline-variant">
    <a
      href="/admin/animes"
      class="px-5 py-2.5 text-sm font-medium text-on-surface-variant hover:text-on-surface bg-surface-highest hover:bg-surface-highest rounded-xl transition-colors"
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={loading || !title}
      class="px-5 py-2.5 text-sm font-medium text-on-surface bg-primary hover:bg-primary-container rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
    >
      {#if loading}
        <svg
          class="animate-spin -ml-1 mr-2 h-4 w-4 text-on-surface"
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
        Create Anime
      {/if}
    </button>
  </div>
</form>
