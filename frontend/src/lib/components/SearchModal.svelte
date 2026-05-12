<script lang="ts">
  import { onMount } from "svelte";
  import { fade, scale } from "svelte/transition";
  import api from "$lib/api";
  import {
    getSearchHistory,
    saveToSearchHistory,
    removeFromSearchHistory,
    clearSearchHistory,
    type HistoryItem,
  } from "$lib/utils/history";
  import { getSongName } from "$lib/song-utils";
  import { goto } from "$app/navigation";
  import Search from "lucide-svelte/icons/search";
  import X from "lucide-svelte/icons/x";
  import History from "lucide-svelte/icons/history";
  import Tv from "lucide-svelte/icons/tv";
  import User from "lucide-svelte/icons/user";
  import Music from "lucide-svelte/icons/music";
  import Frown from "lucide-svelte/icons/frown";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

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
  let recentSearches: HistoryItem[] = $state([]);

  onMount(() => {
    recentSearches = getSearchHistory();
  });

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
    if (e.key === "Enter" && query.trim().length >= 3) {
      recentSearches = saveToSearchHistory({
        id: `query:${query.trim().toLowerCase()}`,
        type: "query",
        label: query.trim(),
        slug: query.trim(),
      });
      handleSearch();
    }
    if (e.key === "Escape") {
      closeModal();
    }
  }

  function handleHistoryItemClick(item: HistoryItem) {
    if (item.type === "query") {
      query = item.label;
      recentSearches = saveToSearchHistory(item);
      handleSearch();
    } else {
      let url = "";
      switch (item.type) {
        case "anime":
          url = `/animes/${item.slug}`;
          break;
        case "artist":
          url = `/artists/${item.slug}`;
          break;
        case "user":
          url = `/users/${item.slug}`;
          break;
        case "song":
          url = `/songs/${item.animeSlug}/${item.slug}`;
          break;
      }
      if (url) {
        recentSearches = saveToSearchHistory(item);
        goto(url);
        closeModal();
      }
    }
  }

  function handleDeleteHistory(id: string) {
    recentSearches = removeFromSearchHistory(id);
  }

  function handleClearAllHistory() {
    clearSearchHistory();
    recentSearches = [];
  }

  function handleResultClick(item: HistoryItem) {
    recentSearches = saveToSearchHistory(item);
    closeModal();
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
        <Search class="text-primary" size={24} />

        <input
          type="text"
          bind:value={query}
          placeholder="Search animes, artists, users..."
          class="flex-1 bg-transparent border-none text-on-surface text-lg focus:outline-none placeholder:text-on-surface-variant/30 font-bold"
          aria-label="Search animes, artists, users, and songs"
        />
        {#if isLoading}
          <div
            class="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin"
          ></div>
        {:else if query.length > 0}
          <button
            onclick={() => (query = "")}
            class="w-8 h-8 rounded-full hover:bg-on-surface/5 flex items-center justify-center transition-colors text-on-surface-variant hover:text-on-surface"
            aria-label="Clear search query"
          >
            <X size={20} />
          </button>
        {/if}
      </div>

      <!-- Search Results -->
      <div class="overflow-y-auto flex-1 p-6 custom-scrollbar">
        {#if query.trim().length < 3}
          {#if recentSearches.length > 0}
            <div class="flex flex-col gap-4">
              <div class="flex items-center justify-between px-2">
                <h3
                  class="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em]"
                >
                  Recent Searches
                </h3>
                <button
                  onclick={handleClearAllHistory}
                  class="text-[10px] font-black text-primary uppercase tracking-widest hover:underline"
                >
                  Clear All
                </button>
              </div>
              <div class="flex flex-col gap-1">
                {#each recentSearches as item}
                  <div class="group flex items-center justify-between gap-2">
                    <button
                      onclick={() => handleHistoryItemClick(item)}
                      class="flex-1 flex items-center gap-3 p-2 rounded-sm hover:bg-surface-highest/50 transition-all text-left group"
                    >
                      <div
                        class="w-10 h-10 rounded-sm flex items-center justify-center bg-surface-lowest border border-outline-variant/10 overflow-hidden"
                      >
                        {#if item.image}
                          <OptimizedImage
                            src={item.image}
                            alt=""
                            class="w-full h-full object-cover"
                            sizes="48px"
                          />
                        {:else}
                          <div class="text-on-surface-variant/40 flex items-center justify-center">
                            {#if item.type === "query"}<History size={20} />
                            {:else if item.type === "anime"}<Tv size={20} />
                            {:else if item.type === "artist"}<User size={20} />
                            {:else if item.type === "user"}<User size={20} />
                            {:else if item.type === "song"}<Music size={20} />{/if}
                          </div>

                        {/if}
                      </div>
                      <div class="flex flex-col min-w-0">
                        <span
                          class="text-on-surface font-bold group-hover:text-primary transition-colors truncate"
                          >{item.label}</span
                        >
                        {#if item.description}
                          <span
                            class="text-[10px] text-on-surface-variant/50 uppercase font-black tracking-widest truncate"
                            >{item.description}</span
                          >
                        {/if}
                      </div>
                    </button>
                    <button
                      onclick={() => handleDeleteHistory(item.id)}
                      class="w-10 h-10 flex items-center justify-center text-on-surface-variant/40 hover:text-red-500 transition-colors"
                      title="Remove from history"
                    >
                      <X size={18} />

                    </button>
                  </div>
                {/each}
              </div>
            </div>
          {:else}
            <div class="text-center py-12 text-on-surface-variant">
              <Search class="mb-4 opacity-20" size={48} />

              <p class="text-sm font-bold uppercase tracking-widest opacity-50">
                Type at least 3 characters to search
              </p>
            </div>
          {/if}
        {:else if !isLoading && (results?.animes?.length || 0) === 0 && (results?.artists?.length || 0) === 0 && (results?.users?.length || 0) === 0 && (results?.songs?.length || 0) === 0}
          <div class="text-center py-12 text-on-surface-variant">
            <Frown class="mb-4 opacity-20" size={48} />

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
                      onclick={() =>
                        handleResultClick({
                          id: `anime:${anime.slug}`,
                          type: "anime",
                          label: anime.title,
                          description: `${anime.season?.name || ""} ${anime.year?.name || ""}`,
                          slug: anime.slug,
                          image: anime.cover_url,
                        })}
                      class="flex items-center gap-4 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                    >
                      <OptimizedImage
                        src={anime.cover_url}
                        sources={anime.cover_sources}
                        alt={anime.title}
                        class="w-12 h-16 object-cover rounded-sm bg-surface-lowest shadow-lg shadow-black/20"
                        sizes="48px"
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
                      onclick={() =>
                        handleResultClick({
                          id: `song:${song.slug}`,
                          type: "song",
                          label: getSongName(song),
                          description: `${song.type} - ${song.anime?.title || ""}`,
                          slug: song.slug,
                          animeSlug: song.anime?.slug,
                        })}
                      class="flex items-center gap-4 p-3 rounded-sm hover:bg-surface-highest/50 transition-all group"
                    >
                      <div
                        class="w-12 h-12 rounded-sm bg-surface-highest border border-outline-variant/10 flex items-center justify-center text-primary shadow-sm"
                      >
                        <Music size={24} />

                      </div>
                      <div class="flex flex-col">
                        <span
                          class="text-on-surface font-black group-hover:text-primary transition-colors line-clamp-1"
                          >{getSongName(song)}</span
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
                        class="flex items-center gap-3 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                        onclick={() =>
                          handleResultClick({
                            id: `artist:${artist.slug}`,
                            type: "artist",
                            label: artist.name,
                            slug: artist.slug,
                            image: artist.avatar_url,
                          })}
                      >
                        <div
                          class="w-10 h-10 rounded-full border-2 border-outline-variant/10 bg-surface-highest overflow-hidden flex items-center justify-center text-primary/50 group-hover:border-primary transition-all shadow-sm"
                        >
                          {#if artist.avatar_url}
                            <OptimizedImage
                              src={artist.avatar_url}
                              sources={artist.avatar_sources}
                              alt={artist.name}
                              class="w-full h-full object-cover"
                              sizes="40px"
                            />
                          {:else}
                            <User size={20} />

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
                        onclick={() =>
                          handleResultClick({
                            id: `user:${user.slug}`,
                            type: "user",
                            label: user.name,
                            slug: user.slug,
                            image: user.avatar_url,
                          })}
                        class="flex items-center gap-3 p-2 rounded-sm hover:bg-surface-highest/50 transition-all group"
                      >
                        <div
                          class="w-10 h-10 rounded-full border-2 border-outline-variant/10 bg-surface-highest overflow-hidden flex items-center justify-center text-primary/50 group-hover:border-primary transition-all shadow-sm"
                        >
                          {#if user.avatar_url}
                            <OptimizedImage
                              src={user.avatar_url}
                              sources={user.avatar_sources}
                              alt={user.name}
                              class="w-full h-full object-cover"
                              sizes="40px"
                            />
                          {:else}
                            <User size={20} />

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
