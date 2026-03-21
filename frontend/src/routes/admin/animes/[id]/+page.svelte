<script lang="ts">
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { invalidateAll } from "$app/navigation";

  let { data } = $props<{ data: PageData }>();
  let anime = $derived(data.anime);
  let isSyncing = $state(false);

  async function updateAnimeData() {
    if (isSyncing) return;
    isSyncing = true;
    try {
      await api.post(`/admin/animes/${anime.id}/sync`);
      toastState.addToast("Anime data synchronized with AniList", "success");
      await invalidateAll();
    } catch (err: any) {
      console.error(err);
      toastState.addToast(err.message || "Failed to sync anime data", "error");
    } finally {
      isSyncing = false;
    }
  }
</script>

<svelte:head>
  <title>{anime.title} | Admin Animes</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/animes"
      aria-label="Back to Animes"
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
    <h1 class="text-3xl font-bold tracking-tight text-white line-clamp-1">
      {anime.title}
    </h1>
    {#if anime.anilist_id}
      <a
        href="https://anilist.co/anime/{anime.anilist_id}"
        target="_blank"
        rel="noopener noreferrer"
        class="p-2 ml-2 bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 rounded-lg transition-colors flex shrink-0"
        title="View on AniList"
      >
        <svg
          class="w-5 h-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
          <polyline points="15 3 21 3 21 9" />
          <line x1="10" y1="14" x2="21" y2="3" />
        </svg>
      </a>
    {/if}
  </div>
  <p class="text-gray-400 ml-10">Views details, assets and related songs.</p>
</div>

<!-- Hero Banner & Cover -->
<div
  class="relative w-full h-[300px] mb-20 rounded-3xl bg-zinc-900 border border-white/10 shadow-xl overflow-visible hidden md:block"
>
  <div class="absolute inset-0 overflow-hidden rounded-3xl bg-zinc-950">
    {#if anime.banner_url}
      <img
        src={anime.banner_url}
        alt="Banner"
        class="w-full h-full object-cover opacity-60"
      />
      <div
        class="absolute inset-0 bg-linear-to-t from-zinc-950 via-zinc-900/40 to-transparent"
      ></div>
    {/if}
  </div>

  <div
    class="absolute -bottom-16 left-10 flex items-end gap-6 z-10 w-full pr-20"
  >
    <!-- Cover -->
    <div
      class="w-48 h-72 rounded-2xl border-4 border-zinc-950 shadow-2xl bg-zinc-900 overflow-hidden shrink-0 relative"
    >
      {#if anime.cover_url}
        <img
          src={anime.cover_url}
          alt="Cover"
          class="w-full h-full object-cover"
        />
      {:else}
        <div
          class="w-full h-full flex items-center justify-center text-zinc-700"
        >
          No Cover
        </div>
      {/if}
      <div class="absolute top-3 right-3">
        <span
          class="px-2 py-1 text-xs font-bold rounded-lg {anime.status
            ? 'bg-green-500 text-white'
            : 'bg-zinc-800 text-zinc-400'} shadow-lg backdrop-blur-md"
        >
          {anime.status ? "ACTIVE" : "INACTIVE"}
        </span>
      </div>
    </div>
  </div>
</div>

<!-- Mobile simple header -->
<div class="flex md:hidden gap-6 mb-8">
  <div
    class="w-32 h-48 rounded-xl border-2 border-zinc-900 shadow-xl bg-zinc-900 overflow-hidden shrink-0 relative"
  >
    {#if anime.cover_url}
      <img
        src={anime.cover_url}
        alt="Cover"
        class="w-full h-full object-cover"
      />
    {:else}
      <div class="w-full h-full flex items-center justify-center text-zinc-700">
        No Cover
      </div>
    {/if}
    <div class="absolute top-2 right-2">
      <span
        class="px-2 py-1 text-[10px] font-bold rounded-lg {anime.status
          ? 'bg-green-500 text-white'
          : 'bg-zinc-800 text-zinc-400'} shadow-md"
      >
        {anime.status ? "ACT" : "INA"}
      </span>
    </div>
  </div>
  {#if anime.banner_url}
    <div
      class="flex-1 rounded-xl overflow-hidden bg-zinc-900 opacity-50 relative border border-white/5 hidden sm:block"
    >
      <img
        src={anime.banner_url}
        alt="Banner"
        class="w-full h-full object-cover"
      />
    </div>
  {/if}
</div>

<div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
  <!-- Main Column -->
  <div class="lg:col-span-3 space-y-8">
    <!-- Description -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h2 class="text-xl font-semibold text-white mb-4">Synopsis</h2>
      <div class="text-gray-300 leading-relaxed max-w-none text-sm space-y-4">
        {#if anime.description}
          {@html anime.description}
        {:else}
          <p class="text-gray-500 italic">No description provided.</p>
        {/if}
      </div>
    </div>

    <!-- Related Songs -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden shadow-sm"
    >
      <div
        class="p-6 border-b border-white/5 flex items-center justify-between"
      >
        <h2 class="text-xl font-semibold text-white">Featured Songs</h2>
        <a
          href="/admin/songs/create?anime_id={anime.id}"
          class="text-sm px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-lg transition-colors"
        >
          Add Song
        </a>
      </div>
      <div class="overflow-x-auto">
        {#if anime.songs && anime.songs.length > 0}
          <table class="w-full text-left text-sm text-gray-300">
            <thead
              class="text-xs text-gray-400 uppercase bg-black/20 border-b border-white/5"
            >
              <tr>
                <th class="px-6 py-4 font-semibold">Type</th>
                <th class="px-6 py-4 font-semibold">Title</th>
                <th class="px-6 py-4 font-semibold hidden md:table-cell"
                  >Artists</th
                >
                <th class="px-6 py-4 font-semibold text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5">
              {#each anime.songs as song}
                <tr class="hover:bg-white/2 transition-colors">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20"
                    >
                      {song.type}
                      {song.theme_num}
                    </span>
                  </td>
                  <td
                    class="px-6 py-4 font-medium text-white max-w-[200px] truncate"
                    title={song.song_romaji || song.song_jp || song.song_en}
                  >
                    <a
                      href="/admin/songs/{song.id}/edit"
                      class="hover:text-anirank-primary transition-colors"
                    >
                      {song.song_romaji || song.song_jp || song.song_en}
                    </a>
                  </td>
                  <td
                    class="px-6 py-4 text-xs text-gray-500 max-w-[150px] truncate hidden md:table-cell"
                    title={song.artists
                      ? song.artists.map((a: any) => a.name).join(", ")
                      : ""}
                  >
                    {song.artists
                      ? song.artists.map((a: any) => a.name).join(", ")
                      : "-"}
                  </td>
                  <td class="px-6 py-4 text-right">
                    <a
                      href="/admin/songs/{song.id}/edit"
                      class="text-gray-400 hover:text-white mr-3">Edit</a
                    >
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {:else}
          <div class="px-6 py-12 text-center text-gray-500">
            No songs registered for this anime yet.
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
    <div class="flex">
      <button
        onclick={updateAnimeData}
        disabled={isSyncing}
        class="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-lg transition-colors w-full flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {#if isSyncing}
          <div
            class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"
          ></div>
          Syncing...
        {:else}
          <span class="material-symbols-outlined text-sm">refresh</span>
          Update data
        {/if}
      </button>
    </div>
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-white uppercase tracking-wider mb-4 border-b border-white/10 pb-2"
      >
        Information
      </h3>

      <dl class="space-y-4 text-sm">
        <div>
          <dt class="text-gray-500 mb-1">Format</dt>
          <dd class="text-gray-200 font-medium">
            {anime.format?.name || "Unknown"}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Season</dt>
          <dd class="text-gray-200 font-medium">
            {anime.season?.name || "Unknown"}
            {anime.year?.name || ""}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Studios</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.studios && anime.studios.length > 0}
              {anime.studios.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Producers</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.producers && anime.producers.length > 0}
              {anime.producers.map((s: any) => s.name).join(", ")}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Genres</dt>
          <dd class="text-gray-200 font-medium">
            {#if anime.genres && anime.genres.length > 0}
              <div class="flex flex-wrap gap-1 mt-1">
                {#each anime.genres as genre}
                  <span
                    class="px-2 py-0.5 bg-white/5 border border-white/10 rounded text-xs text-gray-300"
                    >{genre.name}</span
                  >
                {/each}
              </div>
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
      </dl>
    </div>

    <!-- External Links -->
    {#if anime.external_links && anime.external_links.length > 0}
      <div
        class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm flex flex-col gap-2"
      >
        <h3
          class="text-sm font-semibold text-white uppercase tracking-wider mb-2 border-b border-white/10 pb-2"
        >
          External Links
        </h3>
        {#each anime.external_links as link}
          <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-2 p-2 hover:bg-white/5 rounded-lg transition-colors border border-transparent hover:border-white/10 group"
          >
            <span
              class="w-6 h-6 rounded bg-black/20 text-gray-400 flex items-center justify-center shrink-0 group-hover:text-blue-400 transition-colors"
            >
              <svg
                class="w-3.5 h-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                ><path
                  d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                ></path><polyline points="15 3 21 3 21 9"></polyline><line
                  x1="10"
                  y1="14"
                  x2="21"
                  y2="3"
                ></line></svg
              >
            </span>
            <span
              class="text-sm text-gray-300 group-hover:text-white transition-colors"
              >{link.name}</span
            >
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>
