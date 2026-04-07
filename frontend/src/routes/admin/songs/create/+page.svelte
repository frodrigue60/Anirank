<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { configState as config } from "$lib/state/config.svelte";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import ArtistTagsInput from "$lib/components/admin/ArtistTagsInput.svelte";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";

  // Form State
  let anime_id = $state<number | null>(null);
  let selectedAnimeTitle = $state("");
  let type = $state("OP");
  let theme_num = $state("");
  let song_romaji = $state("");
  let song_en = $state("");
  let song_jp = $state("");
  let status = $state(true);
  let season_id = $state(0);
  let year_id = $state(0);
  let artistsString = $state("");

  // Search State
  let searchQuery = $state("");
  let searchResults = $state<any[]>([]);
  let showResults = $state(false);
  let isSearching = $state(false);

  // UI State
  let loading = $state(false);
  let errorMsg = $state("");

  // Anime Search
  let searchTimeout: any;
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
      // /admin/animes returns { data: [...], meta: {...} }
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

  async function selectAnime(anime: any) {
    anime_id = anime.id;
    selectedAnimeTitle = anime.title;
    season_id = anime.season_id || 0;
    year_id = anime.year_id || 0;
    searchQuery = "";
    searchResults = [];
    showResults = false;

    // Auto-fill theme_num
    await fetchLatestNumber();
  }

  function clearAnime() {
    anime_id = null;
    selectedAnimeTitle = "";
    season_id = 0;
    year_id = 0;
    theme_num = "";
  }

  // Auto-numbering
  // No longer needed to trigger on change if we do it only on selection or type change
  $effect(() => {
    if (anime_id && type) {
      fetchLatestNumber();
    }
  });

  // Pre-fill from URL params
  $effect(() => {
    const urlAnimeId =
      page.url.searchParams.get("anime") ||
      page.url.searchParams.get("anime_id");
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

        // After selecting anime, fetch latest song number
        fetchLatestNumber();
      }
    } catch (err) {
      console.error("Error fetching anime details:", err);
      errorMsg = "Failed to load anime details from URL parameter.";
    } finally {
      loading = false;
    }
  }

  async function fetchLatestNumber() {
    if (!anime_id) return;
    try {
      const res = await api.get(
        `/admin/songs/latest-number?anime_id=${anime_id}&type=${type}`,
      );
      if (res.data && res.data.number !== undefined) {
        theme_num = res.data.number.toString();
      }
    } catch (err) {
      console.error("Error fetching latest number:", err);
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;
    errorMsg = "";

    try {
      // The backend now handles the comma-separated string directly
      const payload = {
        anime_id,
        type,
        theme_num, // Send as string/mixed, backend handles it
        song_romaji,
        song_en,
        song_jp,
        season_id: season_id && season_id > 0 ? season_id : 0,
        year_id: year_id && year_id > 0 ? year_id : 0,
        artists_string: artistsString, // Backend processes names
        status,
      };

      const res = await api.post("/admin/songs", payload);

      if (res.status === 201 || res.data.success) {
        toastState.addToast(
          res.data.message || "Song created successfully!",
          "success",
        );
        const urlAnimeId =
          page.url.searchParams.get("anime_id") ||
          page.url.searchParams.get("anime");
        if (urlAnimeId) {
          goto(`/admin/animes/${urlAnimeId}/songs`);
        } else {
          goto("/admin/songs");
        }
      }
    } catch (err: any) {
      console.error(err);
      const msg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        "An error occurred while creating the song.";

      errorMsg = msg;
      toastState.addToast(msg, "error");
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Create Song | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <!-- svelte-ignore a11y_consider_explicit_label -->
    <a
      href="/admin/songs"
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
      Create New Song
    </h1>
  </div>
  <p class="text-on-surface-variant/70 ml-10">
    Add a new opening, ending, or insert song to the catalog.
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
                  class="w-full bg-zinc-950/50 border border-zinc-800 rounded-xl py-3 px-11 text-on-surface placeholder-zinc-600 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/20 transition-all text-sm h-12"
                  placeholder="Search anime to link..."
                />
                <div
                  class="absolute inset-y-0 left-4 flex items-center pointer-events-none"
                >
                  {#if isSearching}
                    <div
                      class="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"
                    ></div>
                  {:else}
                    <svg
                      class="w-4 h-4 text-zinc-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                      />
                    </svg>
                  {/if}
                </div>
              </div>

              {#if showResults && searchResults.length > 0}
                <div
                  class="absolute z-50 w-full mt-2 bg-zinc-900 border border-zinc-800 rounded-2xl shadow-2xl overflow-hidden"
                  onblur={() => setTimeout(() => (showResults = false), 200)}
                >
                  <div class="max-h-60 overflow-y-auto">
                    {#each searchResults as anime}
                      <button
                        type="button"
                        onclick={() => selectAnime(anime)}
                        class="w-full text-left px-4 py-3 hover:bg-primary-container/20 hover:text-blue-400 transition-all flex items-center gap-3 border-b border-zinc-800/50 last:border-0"
                      >
                        <svg
                          class="w-4 h-4 text-zinc-500"
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
                        <div class="flex flex-col">
                          <span class="text-sm font-medium">{anime.title}</span>
                          {#if anime.year}<span
                              class="text-[10px] text-zinc-500"
                              >{anime.year.name}</span
                            >{/if}
                        </div>
                      </button>
                    {/each}
                  </div>
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
            id="type"
            title="Song Type"
            bind:value={type}
            required
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
          >
            <option value="OP">Opening (OP)</option>
            <option value="ED">Ending (ED)</option>
            <option value="INS">Insert Song (INS)</option>
          </select>
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
            bind:value={theme_num}
            class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
            placeholder="Leave empty to auto-calculate (e.g. 1, 2)"
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
          bind:value={song_jp}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface placeholder-gray-500 focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all"
          placeholder="e.g. 紅蓮の弓矢"
        />
      </div>
      <div>
        <label for="status" class="block text-sm font-medium text-on-surface-variant mb-1"
          >Status</label
        >
        <select
          id="status"
          bind:value={status}
          class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2.5 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-all [&>option]:bg-surface-container"
        >
          <option value={true}>Active</option>
          <option value={false}>Inactive</option>
        </select>
      </div>
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
        Create Song
      {/if}
    </button>
  </div>
</form>
