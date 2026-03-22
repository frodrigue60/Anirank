<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongName } from "$lib/song-utils";

  let { data } = $props<{ data: PageData }>();
  let song = $derived(data.song);

  const videoUrl = $derived(
    song.song_variants && song.song_variants.length > 0
      ? song.song_variants[0].video_url
      : null
  );
</script>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
  <!-- Main Content -->
  <div class="lg:col-span-2 space-y-6">
    <!-- Media Player -->
    {#if videoUrl}
      <div
        class="bg-black border border-white/5 rounded-2xl overflow-hidden shadow-xl aspect-video w-full flex items-center justify-center relative group"
      >
        <!-- svelte-ignore a11y_media_has_caption -->
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
  </div>

  <!-- Sidebar Column -->
  <div class="space-y-6">
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
