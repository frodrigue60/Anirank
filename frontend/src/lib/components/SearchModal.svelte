<script lang="ts">
  import { onMount } from "svelte";
  import { fade, scale } from "svelte/transition";
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
      class="absolute inset-0 bg-black/60 backdrop-blur-sm"
      onclick={closeModal}
      transition:fade={{ duration: 200 }}
    ></div>

    <div
      class="relative w-full max-w-2xl bg-surface-container border border-outline-variant/10 rounded-md shadow-2xl overflow-hidden flex flex-col max-h-[80vh]"
      transition:scale={{ duration: 300, start: 0.95 }}
    >
      <!-- Search Input Header -->
      <div
        class="p-6 border-b border-outline-variant/10 flex items-center gap-4 bg-surface-container/50 backdrop-blur-md"
      >
        <span class="material-symbols-outlined text-primary text-[24px]"
          >search</span
        >
        <input
          type="text"
          bind:value={query}
          placeholder="Search animes, artists, users..."
          class="flex-1 bg-transparent border-none text-on-surface text-lg focus:outline-none placeholder:text-on-surface-variant/30 font-bold"
        />
        {#if isLoading}
          <div
            class="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin"
          ></div>
        {:else if query.length > 0}
          <button
            onclick={() => (query = "")}
            class="w-8 h-8 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
          >
            <span class="material-symbols-outlined text-[20px]">close</span>
          </button>
        {/if}
      </div>

      <!-- Search Results -->
      <div class="overflow-y-auto flex-1 p-6 custom-scrollbar">
        {#if query.trim().length < 3}
          <div class="text-center py-12 text-on-surface-variant">
            <span class="material-symbols-outlined text-[48px] mb-4 opacity-20"
              >search</span
            >
            <p class="text-sm font-bold uppercase tracking-widest opacity-50">
              Type at least 3 characters to search
            </p>
          </div>
        {:else if !isLoading && (results?.animes?.length || 0) === 0 && (results?.artists?.length || 0) === 0 && (results?.users?.length || 0) === 0 && (results?.songs?.length || 0) === 0}
          <div class="text-center py-12 text-on-surface-variant">
            <span class="material-symbols-outlined text-[48px] mb-4 opacity-20"
              >sentiment_dissatisfied</span
            >
            <p class="text-sm font-bold uppercase tracking-widest opacity-50">
              No results found for "{query}"
            </p>
          </div>
        {:else}
          <div class="flex flex-col gap-8">
            <!-- Animes -->
            {#if results.animes && results.animes.length > 0}
              <div>
                <h3
                  class="text-[10px] font-black text-primary uppercase tracking-[0.2em] mb-4 px-2"
                >
                  Animes
                </h3>
                <div class="flex flex-col gap-2">
                  {#each results.animes.slice(0, 5) as anime}
                    <a
                      href="/animes/{anime.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-4 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                    >
                      <img
                        src={anime.cover_url ||
                          "https://placehold.co/100x150/1e1e24/7f13ec?text=Anime"}
                        alt={anime.title}
                        title={anime.title}
                        class="w-12 h-16 object-cover rounded-md bg-surface-lowest shadow-lg shadow-black/20"
                      />
                      <div class="flex flex-col">
                        <span
                          class="text-on-surface font-black group-hover:text-primary transition-colors line-clamp-1"
                          >{anime.title}</span
                        >
                        {#if anime.year || anime.season}
                          <span
                            class="text-on-surface-variant/60 text-xs font-medium"
                            >{anime.season?.name} {anime.year?.name}</span
                          >
                        {/if}
                      </div>
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Songs -->
            {#if results.songs && results.songs.length > 0}
              <div>
                <h3
                  class="text-[10px] font-black text-primary uppercase tracking-[0.2em] mb-4 px-2"
                >
                  Songs
                </h3>
                <div class="flex flex-col gap-2">
                  {#each results.songs.slice(0, 5) as song}
                    <a
                      href="/songs/{song.anime?.slug}/{song.slug}"
                      onclick={closeModal}
                      class="flex items-center gap-4 p-3 rounded-md hover:bg-surface-highest/50 transition-all group"
                    >
                      <div
                        class="w-12 h-12 rounded-sm bg-surface-highest border border-outline-variant/10 flex items-center justify-center text-primary shadow-sm"
                      >
                        <span class="material-symbols-outlined text-[24px]"
                          >music_note</span
                        >
                      </div>
                      <div class="flex flex-col">
                        <span
                          class="text-on-surface font-black group-hover:text-primary transition-colors line-clamp-1"
                          >{song.song_romaji ||
                            song.song_en ||
                            "Untitled Song"}</span
                        >
                        {#if song.anime}
                          <span
                            class="text-on-surface-variant/60 text-xs font-medium"
                            >{song.type} - {song.anime.title}</span
                          >
                        {/if}
                      </div>
                    </a>
                  {/each}
                </div>
              </div>
            {/if}

            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
              <!-- Artists -->
              {#if results.artists && results.artists.length > 0}
                <div>
                  <h3
                    class="text-[10px] font-black text-primary uppercase tracking-[0.2em] mb-4 px-2"
                  >
                    Artists
                  </h3>
                  <div class="flex flex-col gap-2">
                    {#each results.artists.slice(0, 5) as artist}
                      <a
                        href="/artists/{artist.slug}"
                        onclick={closeModal}
                        class="flex items-center gap-3 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                      >
                        <div
                          class="w-10 h-10 rounded-full border-2 border-outline-variant/10 bg-surface-highest overflow-hidden flex items-center justify-center text-primary/50 group-hover:border-primary transition-all shadow-sm"
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
                          class="text-on-surface font-bold group-hover:text-primary transition-colors"
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
                    class="text-[10px] font-black text-primary uppercase tracking-[0.2em] mb-4 px-2"
                  >
                    Users
                  </h3>
                  <div class="flex flex-col gap-2">
                    {#each results.users.slice(0, 5) as user}
                      <a
                        href="/users/{user.slug}"
                        onclick={closeModal}
                        class="flex items-center gap-3 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                      >
                        <div
                          class="w-10 h-10 rounded-full border-2 border-outline-variant/10 bg-surface-highest overflow-hidden flex items-center justify-center text-primary/50 group-hover:border-primary transition-all shadow-sm"
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
                          class="text-on-surface font-bold group-hover:text-primary transition-colors"
                          >{user.name}</span
                        >
                      </a>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>
          </div>
        {/if}
      </div>

      <div
        class="p-4 bg-surface-container border-t border-outline-variant/10 flex items-center justify-between text-[10px] text-on-surface-variant font-black uppercase tracking-widest"
      >
        <span class="flex items-center gap-2 opacity-50"
          ><kbd
            class="bg-surface-highest px-2 py-1 rounded-sm border border-outline-variant/10"
            >ESC</kbd
          > to close</span
        >
        <span class="flex items-center gap-2 opacity-50"
          >Powered by <span class="text-primary">Anirank</span></span
        >
      </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: var(--color-outline-variant, rgba(255, 255, 255, 0.1));
    border-radius: 2px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: var(--color-primary);
  }
</style>
