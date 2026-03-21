<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongName } from "$lib/song-utils";

  let { data } = $props<{ data: PageData }>();
  const song = data.song;

  // Helpers to structure URLs
  const videoUrl =
    song.song_variants && song.song_variants.length > 0
      ? song.song_variants[0].video_url
      : null;
</script>

<svelte:head>
  <title>
    {getSongName(song)} | Admin
    Songs
  </title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div class="flex items-center gap-4">
    <a
      href="/admin/songs"
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
      {getSongName(song)}
    </h1>
    <span
      class="inline-flex items-center px-2.5 py-1 rounded-md text-sm font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 ml-2"
    >
      {song.type}
      {song.theme_num}
    </span>
  </div>

  <div class="flex gap-2">
    <a
      href={`/admin/songs/${song.id}/edit`}
      class="px-4 py-2 bg-white/5 hover:bg-white/10 text-white font-medium rounded-xl transition-colors border border-white/10 flex items-center gap-2"
    >
      <svg
        class="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
        />
      </svg>
      Edit Song
    </a>
    <a
      href={`/admin/songs/${song.id}/variants`}
      class="px-4 py-2 bg-anirank-primary hover:bg-blue-600 text-white font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
    >
      Manage Variants
    </a>
  </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
  <!-- Main Content -->
  <div class="lg:col-span-2 space-y-6">
    <!-- Media Player -->
    {#if videoUrl}
      <div
        class="bg-black border border-white/5 rounded-2xl overflow-hidden shadow-xl aspect-video w-full flex items-center justify-center relative group"
      >
        <video controls class="w-full h-full object-contain" src={videoUrl}>
          Your browser does not support the video tag.
        </video>
      </div>
    {/if}

    <!-- Alternative Titles -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h2
        class="text-lg font-semibold text-white mb-4 border-b border-white/5 pb-2"
      >
        Titles
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-gray-500 block mb-1">Romaji</span>
          <span class="text-gray-200">{song.song_romaji || "-"}</span>
        </div>
        <div>
          <span class="text-gray-500 block mb-1">Japanese</span>
          <span class="text-gray-200">{song.song_jp || "-"}</span>
        </div>
        <div>
          <span class="text-gray-500 block mb-1">English</span>
          <span class="text-gray-200">{song.song_en || "-"}</span>
        </div>
      </div>
    </div>

    <!-- Artists -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <div
        class="flex items-center justify-between mb-4 border-b border-white/5 pb-2"
      >
        <h2 class="text-lg font-semibold text-white">Artists</h2>
      </div>
      {#if song.artists && song.artists.length > 0}
        <div class="flex flex-wrap gap-2">
          {#each song.artists as artist}
            <a
              href="/admin/artists/{artist.id}/edit"
              class="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 rounded-lg text-sm text-gray-300 hover:text-white transition-colors flex items-center gap-2"
            >
              {#if artist.avatar_url}
                <img
                  src={artist.avatar_url}
                  alt={artist.name}
                  class="w-5 h-5 rounded-full object-cover"
                />
              {/if}
              {artist.name}
            </a>
          {/each}
        </div>
      {:else}
        <p class="text-gray-500 text-sm italic">
          No artists associated with this song.
        </p>
      {/if}
    </div>

    <!-- Variants List -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h2
        class="text-lg font-semibold text-white mb-4 border-b border-white/5 pb-2"
      >
        Video Variants
      </h2>
      {#if song.song_variants && song.song_variants.length > 0}
        <div class="space-y-3">
          {#each song.song_variants as variant}
            <div
              class="flex items-center justify-between p-3 bg-white/5 rounded-xl border border-white/5"
            >
              <div class="flex items-center gap-3">
                <span
                  class="w-10 h-10 rounded-lg bg-black/40 flex items-center justify-center text-gray-400"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"
                    ><path
                      d="M2 6a2 2 0 012-2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6zm12.553 1.106A1 1 0 0014 8v4a1 1 0 00.553.894l2 1A1 1 0 0018 13V7a1 1 0 00-1.447-.894l-2 1z"
                    ></path></svg
                  >
                </span>
                <div>
                  <p class="text-sm font-medium text-white">{variant.type}</p>
                  <p class="text-xs text-gray-500">File attached</p>
                </div>
              </div>
              <a
                href={variant.video_url}
                target="_blank"
                rel="noopener noreferrer"
                class="text-xs bg-white/5 hover:bg-white/10 px-3 py-1.5 rounded-lg border border-white/10 text-gray-300"
              >
                Open File
              </a>
            </div>
          {/each}
        </div>
      {:else}
        <p class="text-gray-500 text-sm italic">No video variants uploaded.</p>
      {/if}
    </div>
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
    <!-- Anime Link -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-white uppercase tracking-wider mb-4 border-b border-white/10 pb-2"
      >
        Anime Series
      </h3>
      {#if song.anime}
        <a href="/admin/animes/{song.anime.id}" class="group block">
          <div
            class="relative w-full aspect-[16/9] rounded-xl overflow-hidden mb-3 bg-zinc-900 border border-white/5"
          >
            {#if song.anime.banner_url || song.anime.cover_url}
              <img
                src={song.anime.banner_url || song.anime.cover_url}
                alt={song.anime.title}
                class="w-full h-full object-cover opacity-70 group-hover:opacity-100 transition-opacity"
              />
            {:else}
              <div
                class="w-full h-full flex items-center justify-center text-zinc-700"
              >
                No Image
              </div>
            {/if}
          </div>
          <p
            class="text-white font-medium group-hover:text-anirank-primary transition-colors line-clamp-2"
          >
            {song.anime.title}
          </p>
        </a>
      {:else}
        <p class="text-gray-500 text-sm italic">No anime linked.</p>
      {/if}
    </div>

    <!-- Metadata & Taxonomy -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-white uppercase tracking-wider mb-4 border-b border-white/10 pb-2"
      >
        Details
      </h3>
      <dl class="space-y-4 text-sm">
        <div>
          <dt class="text-gray-500 mb-1">Season</dt>
          <dd class="text-gray-200 font-medium">
            {#if song.season || song.year}
              {song.season?.name || ""} {song.year?.name || ""}
            {:else}
              <span class="text-gray-600">-</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Added On</dt>
          <dd class="text-gray-200 font-medium">
            {new Date(song.created_at).toLocaleDateString()}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 mb-1">Last Updated</dt>
          <dd class="text-gray-200 font-medium">
            {new Date(song.updated_at).toLocaleDateString()}
          </dd>
        </div>
      </dl>
    </div>

    <!-- Metrics Stats -->
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 shadow-sm"
    >
      <h3
        class="text-sm font-semibold text-white uppercase tracking-wider mb-4 border-b border-white/10 pb-2"
      >
        Metrics
      </h3>
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-white/5 p-3 rounded-xl border border-white/5">
          <span class="block text-xs text-gray-500 mb-1">Total Views</span>
          <span class="block text-xl font-semibold text-white"
            >{song.views.toLocaleString()}</span
          >
        </div>
        <div class="bg-white/5 p-3 rounded-xl border border-white/5">
          <span class="block text-xs text-gray-500 mb-1">Avg Rating</span>
          <span class="block text-xl font-semibold text-white"
            >{song.average_rating > 0
              ? song.average_rating.toFixed(2)
              : "-"}</span
          >
        </div>
        <div class="bg-white/5 p-3 rounded-xl border border-white/5">
          <span class="block text-xs text-gray-500 mb-1">Likes</span>
          <span class="block text-xl font-semibold text-white"
            >{song.likes_count.toLocaleString()}</span
          >
        </div>
        <div class="bg-white/5 p-3 rounded-xl border border-white/5">
          <span class="block text-xs text-gray-500 mb-1">Dislikes</span>
          <span class="block text-xl font-semibold text-white"
            >{song.dislikes_count.toLocaleString()}</span
          >
        </div>
      </div>
    </div>
  </div>
</div>
