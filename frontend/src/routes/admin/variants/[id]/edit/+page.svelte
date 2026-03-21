<script lang="ts">
  import { goto } from "$app/navigation";
  import { configState as config } from "$lib/state/config.svelte";
  import api from "$lib/api";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";
  import type { PageData } from "./$types";

  let { data } = $props<{ data: PageData }>();
  const variant = data.variant;

  // Form State
  let song_id = $state<number | null>(variant.song_id);
  let version_number = $state(
    variant.version_number ? variant.version_number.toString() : "",
  );
  let slug = $state(variant.slug || "");
  let season_id = $state(variant.season_id || 0);
  let year_id = $state(variant.year_id || 0);
  let spoiler = $state(variant.spoiler || false);
  let status = $state(variant.status || false);

  // UI State
  let loading = $state(false);
  let errorMsg = $state("");

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;
    errorMsg = "";

    try {
      const payload = {
        song_id,
        version_number: version_number ? parseInt(version_number) : 0,
        slug,
        season_id: season_id > 0 ? season_id : null,
        year_id: year_id > 0 ? year_id : null,
        spoiler,
        status,
      };

      const res = await api.put(`/admin/variants/${data.variant.id}`, payload);

      if (res.status === 200) {
        goto("/admin/variants");
      }
    } catch (err: any) {
      console.error(err);
      errorMsg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        "An error occurred while updating the variant.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Variant | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/variants"
      aria-label="Back to Variants"
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
    <h1 class="text-3xl font-bold tracking-tight text-white">
      Edit Song Variant
    </h1>
  </div>
  <p class="text-gray-400 ml-10">
    Updating: <span class="font-medium text-gray-200"
      >Variant #{variant.id} (Song: {variant.song_id})</span
    >
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
          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
        /></svg
      >
      Variant Details
    </h2>

    <div class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label
            for="song_id"
            class="block text-sm font-medium text-gray-300 mb-1"
            >Song ID <span class="text-red-400">*</span></label
          >
          <input
            type="number"
            id="song_id"
            title="Song ID"
            bind:value={song_id}
            required
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all"
            placeholder="e.g. 1"
          />
        </div>

        <div class="pt-2 flex items-center">
          <label class="flex items-center gap-3 cursor-pointer">
            <div class="relative">
              <input
                type="checkbox"
                bind:checked={spoiler}
                class="sr-only peer"
              />
              <div
                class="w-11 h-6 bg-white/10 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-anirank-primary"
              ></div>
            </div>
            <span class="text-sm font-medium text-white">Contains Spoilers</span
            >
          </label>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label
            for="version_number"
            class="block text-sm font-medium text-gray-300 mb-1"
            >Version Number</label
          >
          <input
            type="number"
            id="version_number"
            title="Version Number"
            bind:value={version_number}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all"
            placeholder="Auto-calculates if empty (e.g. 1, 2)"
          />
        </div>

        <div>
          <label for="slug" class="block text-sm font-medium text-gray-300 mb-1"
            >Slug</label
          >
          <input
            type="text"
            id="slug"
            title="Slug"
            bind:value={slug}
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all font-mono text-sm"
            placeholder="Auto-calculates to v[version] if empty"
          />
        </div>
      </div>
    </div>
  </div>

  <!-- Status Management -->
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
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        /></svg
      >
      Visibility Status
    </h2>

    <StatusControl bind:status />
  </div>

  <!-- Taxonomies & Metadata -->
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
          d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
        /></svg
      >
      Variant Taxonomy
    </h2>
    <p class="text-sm text-gray-400 mb-4">
      You can optionally map this video to a specific timeframe if the visuals
      changed mid-season.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label for="year" class="block text-sm font-medium text-gray-300 mb-1"
          >Override Year</label
        >
        <select
          id="year"
          title="Year"
          bind:value={year_id}
          class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-all [&>option]:bg-anirank-card"
        >
          <option value={0}>Inherit from Song</option>
          {#each config.years as year}
            <option value={year.id}>{year.name}</option>
          {/each}
        </select>
      </div>

      <div>
        <label for="season" class="block text-sm font-medium text-gray-300 mb-1"
          >Override Season</label
        >
        <select
          id="season"
          title="Season"
          bind:value={season_id}
          class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:border-anirank-primary transition-all [&>option]:bg-anirank-card"
        >
          <option value={0}>Inherit from Song</option>
          {#each config.seasons as season}
            <option value={season.id}>{season.name}</option>
          {/each}
        </select>
      </div>
    </div>
  </div>

  <div class="flex items-center justify-end gap-3 pt-4 border-t border-white/5">
    <a
      href="/admin/variants"
      class="px-5 py-2.5 text-sm font-medium text-gray-300 hover:text-white bg-white/5 hover:bg-white/10 rounded-xl transition-colors"
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={loading || !song_id}
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
        Update Variant
      {/if}
    </button>
  </div>
</form>
