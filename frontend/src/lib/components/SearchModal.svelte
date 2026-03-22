<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";

  let { show = $bindable(false) } = $props();

  let query = $state("");
  // svelte-ignore state_referenced_locally
  let results: {
    animes: any[];
    artists: any[];
    users: any[];
    songs: any[];
    studios: any[];
  } = $state({
    animes: [],
    artists: [],
    users: [],
    songs: [],
    studios: [],
  });
  let isLoading = $state(false);
  let searchTimeout: ReturnType<typeof setTimeout>;

  function handleSearch() {
    if (searchTimeout) clearTimeout(searchTimeout);

    if (query.trim().length < 3) {
      results = { animes: [], artists: [], users: [], songs: [], studios: [] };
      return;
    }

    isLoading = true;
    searchTimeout = setTimeout(async () => {
      try {
        // Adjust endpoint and response structure as per actual API
        const response = await api.get(
          `/search?q=${encodeURIComponent(query)}`,
        );
        results = {
          animes: response.data.data.animes || [],
          artists: response.data.data.artists || [],
          users: response.data.data.users || [],
          songs: response.data.data.songs || [],
          studios: response.data.data.studios || [],
        };
      } catch (error) {
        console.error("Search failed:", error);
        results = {
          animes: [],
          artists: [],
          users: [],
          songs: [],
          studios: [],
        };
      } finally {
        isLoading = false;
      }
    }, 300); // 300ms debounce
  }

  // Effect-like behavior for reacting to query changes
  $effect(() => {
    handleSearch();
  });

  function closeModal() {
    show = false;
    query = "";
    results = { animes: [], artists: [], users: [], songs: [], studios: [] };
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      closeModal();
    }
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if show}
  <div
    class="fixed inset-0 z-100 flex items-start justify-center pt-20 px-4 sm:px-0"
  >
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="absolute inset-0 bg-background-dark/80 backdrop-blur-sm"
      onclick={closeModal}
    ></div>

    <div
      class="relative w-full max-w-2xl bg-surface-dark border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[80vh]"
    >
      <!-- Search Input Header -->
      <div class="p-4 border-b border-white/5 flex items-center gap-4">
        <span class="material-symbols-outlined text-white/40text-[24px]"
          >search</span
        >
        <input
          type="text"
          bind:value={query}
          placeholder="Search animes, artists, users..."
          class="flex-1 bg-transparent border-none text-white text-lg focus:outline-none placeholder:text-white/30"
        />
        {#if isLoading}
          <div
            class="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin"
          ></div>
        {:else if query.length > 0}
          <button
            onclick={() => (query = "")}
            class="text-white/40 hover:text-white transition-colors"
          >
            <span class="material-symbols-outlined text-[20px]">close</span>
          </button>
        {/if}
      </div>

      <!-- Search Results -->
      <div class="overflow-y-auto flex-1 p-4">
        {#if query.trim().length < 3}
          <div class="text-center py-8 text-white/40">
            <span class="material-symbols-outlined text-[48px] mb-2 opacity-50"
              >search</span
            >
            <p>Type at least 3 characters to search</p>
          </div>
        {:else if !isLoading && (results?.animes?.length || 0) === 0 && (results?.artists?.length || 0) === 0 && (results?.users?.length || 0) === 0 && (results?.songs?.length || 0) === 0}
          <div class="text-center py-8 text-white/40">
            <span class="material-symbols-outlined text-[48px] mb-2 opacity-50"
              >sentiment_dissatisfied</span
            >
            <p>No results found for "{query}"</p>
          </div>
        {:else}
          <div class="flex flex-col gap-6">
            <!-- Animes -->
            {#if results.animes && results.animes.length > 0}
              <div>
                <h3
                  class="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 px-1"
                >
                  Animes
                </h3>
                <div class="flex flex-col gap-1">
                  {#each results.animes.slice(0, 5) as anime}
                    <a
                      href="/animes/{anime.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                      <img
                        src={anime.cover_url ||
                          "https://placehold.co/100x150/1e1e24/7f13ec?text=Anime"}
                        alt={anime.title}
                        title={anime.title}
                        class="w-10 h-14 object-cover rounded bg-surface-darker"
                      />
                      <div class="flex flex-col">
                        <span
                          class="text-white font-medium group-hover:text-primary transition-colors line-clamp-1"
                          >{anime.title}</span
                        >
                        {#if anime.year || anime.season}
                          <span class="text-white/40 text-xs line-clamp-1"
                            >{anime.season?.name} {anime.year?.name}</span
                          >
                        {/if}
                      </div>
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Artists -->
            {#if results.artists && results.artists.length > 0}
              <div>
                <h3
                  class="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 px-1"
                >
                  Artists
                </h3>
                <div class="flex flex-col gap-1">
                  {#each results.artists.slice(0, 5) as artist}
                    <a
                      href="/artists/{artist.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                      <div
                        class="w-10 h-10 rounded-full bg-surface-darker overflow-hidden flex items-center justify-center text-primary/50"
                      >
                        {#if artist.avatar_url}
                          <img
                            src={artist.avatar_url}
                            alt={artist.name}
                            title={artist.name}
                            class="w-full h-full object-cover"
                          />
                        {:else}
                          <span class="material-symbols-outlined text-[20px]"
                            >person</span
                          >
                        {/if}
                      </div>
                      <span
                        class="text-white font-medium group-hover:text-primary transition-colors"
                        >{artist.name}</span
                      >
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Users -->
            {#if results.users && results.users.length > 0}
              <div>
                <h3
                  class="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 px-1"
                >
                  Users
                </h3>
                <div class="flex flex-col gap-1">
                  {#each results.users.slice(0, 5) as user}
                    <a
                      href="/users/{user.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                      <div
                        class="w-10 h-10 rounded-full bg-surface-darker overflow-hidden flex items-center justify-center text-primary/50"
                      >
                        {#if user.avatar_url}
                          <img
                            src={user.avatar_url}
                            alt={user.name}
                            title={user.name}
                            class="w-full h-full object-cover"
                          />
                        {:else}
                          <span class="material-symbols-outlined text-[20px]"
                            >person</span
                          >
                        {/if}
                      </div>
                      <span
                        class="text-white font-medium group-hover:text-primary transition-colors"
                        >{user.name}</span
                      >
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Songs -->
            {#if results.songs && results.songs.length > 0}
              <div>
                <h3
                  class="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 px-1"
                >
                  Songs
                </h3>
                <div class="flex flex-col gap-1">
                  {#each results.songs.slice(0, 5) as song}
                    <a
                      href="/songs/{song.anime?.slug}/{song.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                      <div
                        class="w-10 h-10 rounded bg-surface-darker flex items-center justify-center text-primary"
                      >
                        <span class="material-symbols-outlined text-[20px]"
                          >music_note</span
                        >
                      </div>
                      <div class="flex flex-col">
                        <span
                          class="text-white font-medium group-hover:text-primary transition-colors line-clamp-1"
                          >{song.song_romaji ||
                            song.song_en ||
                            "Untitled Song"}</span
                        >
                        {#if song.anime}
                          <span class="text-white/40 text-xs line-clamp-1"
                            >{song.type} - {song.anime.title}</span
                          >
                        {/if}
                      </div>
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Studios -->
            {#if results.studios && results.studios.length > 0}
              <div>
                <h3
                  class="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 px-1"
                >
                  Studios
                </h3>
                <div class="flex flex-col gap-1">
                  {#each results.studios.slice(0, 5) as studio}
                    <a
                      href="/studios/{studio.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors group"
                    >
                      <div
                        class="w-10 h-10 rounded bg-white/5 flex items-center justify-center overflow-hidden"
                      >
                        {#if studio.logo_url}
                          <img
                            src={studio.logo_url}
                            alt={studio.name}
                            title={studio.name}
                            class="w-full h-full object-contain"
                          />
                        {:else}
                          <span
                            class="material-symbols-outlined text-[20px] text-white/20"
                            >business</span
                          >
                        {/if}
                      </div>
                      <span
                        class="text-white font-medium group-hover:text-primary transition-colors"
                        >{studio.name}</span
                      >
                    </a>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <div
        class="p-3 bg-surface-darker border-t border-white/5 flex items-center justify-between text-xs text-white/40"
      >
        <span class="flex items-center gap-1"
          ><kbd class="bg-white/10 px-1.5 py-0.5 rounded border border-white/10"
            >ESC</kbd
          > to close</span
        >
        <span class="flex items-center gap-1"
          >Powered by <span class="font-bold text-primary">Anirank</span></span
        >
      </div>
    </div>
  </div>
{/if}
