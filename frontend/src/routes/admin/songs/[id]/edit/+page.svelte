<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { configState as config } from "$lib/state/config.svelte";
  import { getSongName } from "$lib/song-utils";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import type { PageData } from "./$types";
  import ArtistTagsInput from "$lib/components/admin/ArtistTagsInput.svelte";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";

  let { data } = $props<{ data: PageData }>();
  // svelte-ignore state_referenced_locally
  const song = data.song;

  // Form State
  let anime_id = $state<number | null>(song.anime_id);
  let selectedAnimeTitle = $state(song.anime?.title || "");
  let type_id = $state(song.type_id || 0);
  let type = $state(song.type || "OP");

  $effect(() => {
    if (type_id) {
      const st = config.songTypes.find((t) => t.id === type_id.toString());
      if (st) type = st.slug;
    }
  });
  let theme_num = $state(song.theme_num || "");
  let song_romaji = $state(song.song_romaji || "");
  let song_en = $state(song.song_en || "");
  let song_jp = $state(song.song_jp || "");
  let season_id = $state(song.season_id || 0);
  let year_id = $state(song.year_id || 0);
  let status = $state(song.status);

  // Parse artist names instead of IDs for the tags input
  let artistsString = $state(
    Array.isArray(song.artists)
      ? song.artists.map((a: any) => a.name).join(", ")
      : "",
  );

  // Search State
  let searchQuery = $state("");
  let searchResults = $state<any[]>([]);
  // svelte-ignore state_referenced_locally
  let comments: Comment[] = $state(data.comments);
  let showResults = $state(false);
  let isSearching = $state(false);
  let searchTimeout: any;

  // Search Handlers
  async function handleSearch() {
    if (searchQuery.length < 3) {
      searchResults = [];
      showResults = false;
      return;
    }

    isSearching = true;
    try {
      const res = await api.get(
        `/admin/animes?search=${encodeURIComponent(searchQuery)}`,
      );
      searchResults = res.data.data || [];
      showResults = true;
    } catch (err) {
      console.error("Search error:", err);
      searchResults = [];
    } finally {
      isSearching = false;
    }
  }

  function debounceSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(handleSearch, 300);
  }

  function selectAnime(anime: any) {
    anime_id = anime.id;
    selectedAnimeTitle = anime.title;
    season_id = anime.season_id || 0;
    year_id = anime.year_id || 0;
    searchQuery = "";
    searchResults = [];
    showResults = false;
  }

  function clearAnime() {
    anime_id = null;
    selectedAnimeTitle = "";
    season_id = 0;
    year_id = 0;
    theme_num = "";
  }

  // Pre-fill from URL params (e.g., if user wants to change anime via URL)
  $effect(() => {
    const urlAnimeId = page.url.searchParams.get("anime") || page.url.searchParams.get("anime_id");
    if (urlAnimeId && !anime_id) {
      const id = parseInt(urlAnimeId);
      if (!isNaN(id)) {
        fetchAnimeDetails(id);
      }
    }
  });

  async function fetchAnimeDetails(id: number) {
    loading = true;
    try {
      const res = await api.get(`/admin/animes/${id}`);
      if (res.data && res.data.data) {
        const anime = res.data.data;
        anime_id = anime.id;
        selectedAnimeTitle = anime.title;
        season_id = anime.season_id || 0;
        year_id = anime.year_id || 0;
      }
    } catch (err) {
      console.error("Error fetching anime details:", err);
      errorMsg = "Failed to load anime details from URL parameter.";
    } finally {
      loading = false;
    }
  }

  // UI State
  let loading = $state(false);
  let errorMsg = $state("");

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;
    errorMsg = "";

    try {
      const payload = {
        anime_id,
        type,
        type_id: type_id || 0,
        theme_num: theme_num ? theme_num.toString() : "",
        song_romaji,
        song_en,
        song_jp,
        season_id: season_id && season_id > 0 ? season_id : 0,
        year_id: year_id && year_id > 0 ? year_id : 0,
        artists_string: artistsString,
        status,
      };

      const res = await api.put(`/admin/songs/${song.id}`, payload);

      if (res.status === 200 || res.data.success) {
        toastState.addToast(res.data.message || "Song updated successfully!", "success");
        goto(`/admin/songs/${song.id}`);
      }
    } catch (err: any) {
      console.error(err);
      const msg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        "An error occurred while updating the song.";
      
      errorMsg = msg;
      toastState.addToast(msg, "error");
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Song | Admin</title>
</svelte:head>

<div class="mb-6">
  <h2 class="text-xl font-bold text-on-surface">Edit Song Information</h2>
  <p class="text-xs text-on-surface-variant/40">Update titles, artists, and general classification.</p>
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
          d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"
        /></svg
      >
      Song Details
    </h2>

    <div class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label
            for="anime_search"
            class="block text-sm font-medium text-zinc-400 uppercase tracking-widest mb-2"
            >Anime <span class="text-red-400">*</span></label
          >
          <div class="relative group">
            {#if anime_id}
              <div class="relative">
                <div
                  class="w-full bg-blue-500/10 border border-blue-500/30 rounded-xl py-3 px-4 flex items-center justify-between"
                >
                  <div class="flex items-center gap-3">
                    <svg
                      class="w-5 h-5 text-blue-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                      />
                    </svg>
                    <span class="text-blue-100 font-bold text-sm tracking-tight"
                      >{selectedAnimeTitle}</span
                    >
                  </div>
                  <button
                    type="button"
                    onclick={clearAnime}
                    title="Clear selection"
                    class="text-blue-400/60 hover:text-blue-400 transition-colors p-1"
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
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            {:else}
              <div class="relative">
                <input
                  type="text"
                  id="anime_search"
                  bind:value={searchQuery}
                  oninput={debounceSearch}
                  onfocus={() => (showResults = searchResults.length > 0)}
                  onblur={() => setTimeout(() => (showResults = false), 200)}
                  placeholder="Search for an anime..."
                  class="w-full bg-surface-highest border border-outline-variant rounded-xl py-3 pl-11 pr-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none transition-all group-hover:border-outline-variant"
                />
                <svg
                  class="absolute left-4 top-3.5 w-5 h-5 text-on-surface-variant/40"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  /></svg
                >
                {#if isSearching}
                  <div class="absolute right-4 top-3.5">
                    <div
                      class="w-5 h-5 border-2 border-primary/30 border-t-anirank-primary rounded-full animate-spin"
                    ></div>
                  </div>
                {/if}
              </div>

              <!-- Required hidden input for native form validation -->
              <input type="hidden" bind:value={anime_id} required />

              {#if showResults && searchResults.length > 0}
                <div
                  class="absolute z-50 w-full mt-2 bg-[#1a1c23] border border-outline-variant rounded-xl shadow-2xl overflow-hidden max-h-60 overflow-y-auto custom-scrollbar"
                >
                  {#each searchResults as result}
                    <button
                      type="button"
                      class="w-full text-left px-4 py-3 hover:bg-surface-highest transition-colors flex items-center gap-3 border-b border-outline-variant last:border-0"
                      onclick={() => selectAnime(result)}
                    >
                      {#if result.cover_url}
                        <img
                          src={result.cover_url || result.cover}
                          alt={result.title}
                          class="w-8 h-12 object-cover rounded bg-surface-highest"
                        />
                      {:else}
                        <div
                          class="w-8 h-12 rounded bg-surface-highest shrink-0"
                        ></div>
                      {/if}
                      <div>
                        <div
                          class="text-on-surface font-medium text-sm line-clamp-1"
                        >
                          {result.title}
                        </div>
                        <div class="text-on-surface-variant/40 text-xs mt-0.5">
                          {result.format?.name || "Unknown Format"} • {result
                            .season?.name || "Unknown"}
                          {result.year?.name || ""}
                        </div>
                      </div>
                    </button>
                  {/each}
                </div>
              {:else if showResults && searchQuery.length >= 3 && !isSearching}
                <div
                  class="absolute z-50 w-full mt-2 bg-[#1a1c23] border border-outline-variant rounded-xl shadow-2xl p-4 text-center text-on-surface-variant/70 text-sm"
                >
                  No animes found.
                </div>
              {/if}
            {/if}
          </div>
        </div>

        <div>
          <label for="type" class="block text-sm font-medium text-on-surface-variant mb-1"
            >Song Type <span class="text-red-400">*</span></label
          >
          <select
            id="type_id"
            title="Song Type"
            bind:value={type_id}
            required
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
          >
            <option value={0} disabled selected>Select a type...</option>
            {#each config.songTypes as t}
              <option value={t.id}>{t.name}</option>
            {/each}
          </select>
          <input type="hidden" name="type" bind:value={type} />
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label
            for="theme_num"
            class="block text-sm font-medium text-on-surface-variant mb-1"
            >Theme Number</label
          >
          <input
            type="number"
            id="theme_num"
            title="Theme Number"
            bind:value={theme_num}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
            placeholder="e.g. 1"
          />
          <p class="text-xs text-on-surface-variant/40 mt-1">
            Leave empty to auto-generate based on existing songs.
          </p>
        </div>

        <div>
          <label
            for="artists"
            class="block text-sm font-medium text-on-surface-variant mb-1"
            >Artist Names</label
          >
          <ArtistTagsInput bind:value={artistsString} />
          <p class="text-xs text-on-surface-variant/40 mt-1">
            Type a name and press Enter, or paste a comma-separated list.
          </p>
        </div>
      </div>

      <div>
        <label
          for="song_romaji"
          class="block text-sm font-medium text-on-surface-variant mb-1"
          >Title (Romaji) <span class="text-red-400">*</span></label
        >
        <input
          type="text"
          id="song_romaji"
          title="Title Romaji"
          bind:value={song_romaji}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
          placeholder="e.g. Gurenge"
        />
      </div>

      <div>
        <label
          for="song_en"
          class="block text-sm font-medium text-on-surface-variant mb-1"
          >Title (English)</label
        >
        <input
          type="text"
          id="song_en"
          title="Title English"
          bind:value={song_en}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
          placeholder="e.g. Red Lotus"
        />
      </div>

      <div>
        <label
          for="song_jp"
          class="block text-sm font-medium text-on-surface-variant mb-1"
          >Title (Japanese)</label
        >
        <input
          type="text"
          id="song_jp"
          title="Title Japanese"
          bind:value={song_jp}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
          placeholder="e.g. 紅蓮華"
        />
      </div>
      <StatusControl bind:status={status} />
    </div>
  </div>

  <!-- Taxonomies & Metadata -->
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
      Song Origin Taxonomy
    </h2>
    <p class="text-sm text-on-surface-variant/70 mb-4">
      You can optionally specify a Year or Season different from the base Anime,
      like for ending songs introduced mid-season.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label for="year" class="block text-sm font-medium text-on-surface-variant mb-1"
          >Override Year</label
        >
        <select
          id="year"
          title="Year"
          bind:value={year_id}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
        >
          <option value={0}>Inherit from Anime</option>
          {#each config.years as year}
            <option value={year.id}>{year.name}</option>
          {/each}
        </select>
      </div>

      <div>
        <label for="season" class="block text-sm font-medium text-on-surface-variant mb-1"
          >Override Season</label
        >
        <select
          id="season"
          title="Season"
          bind:value={season_id}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
        >
          <option value={0}>Inherit from Anime</option>
          {#each config.seasons as season}
            <option value={season.id}>{season.name}</option>
          {/each}
        </select>
      </div>
    </div>
  </div>

  <div class="flex items-center justify-end gap-3 pt-4 border-t border-outline-variant">
    <a
      href="/admin/songs"
      class="px-5 py-2.5 text-sm font-medium text-on-surface-variant hover:text-on-surface bg-surface-highest hover:bg-surface-highest rounded-xl transition-colors"
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={loading || !anime_id || (!song_romaji && !song_en && !song_jp)}
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
        Update Song
      {/if}
    </button>
  </div>
</form>
