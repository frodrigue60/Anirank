<script lang="ts">
  import type { PageData } from "./$types";

  let { data } = $props<{ data: PageData }>();
  let artist = $derived(data.artist);
  let songs = $derived(data.songs);
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import Pencil from "lucide-svelte/icons/pencil";
  import Music from "lucide-svelte/icons/music";
  import Play from "lucide-svelte/icons/play";
  import Heart from "lucide-svelte/icons/heart";
  import Music2 from "lucide-svelte/icons/music-2";
  import User from "lucide-svelte/icons/user";
</script>

<svelte:head>
  <title>{artist?.name || "Artist Details"} | Admin</title>
</svelte:head>

<div class="max-w-5xl mx-auto pb-20">
  <!-- Nav Header -->
  <div class="mb-8 flex items-center justify-between">
    <div class="flex items-center gap-4">
      <button
        onclick={() => history.back()}
        class="p-2 hover:bg-surface-highest rounded-xl text-on-surface-variant/70 transition-colors"
        aria-label="Back"
      >
        <ArrowLeft size={24} />
      </button>
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
          Artist Profile
        </h1>
        <p class="text-on-surface-variant/70">View detailed information and associations.</p>
      </div>
    </div>
    {#if artist?.id}
      <a
        href={`/admin/artists/${artist.id}/edit`}
        class="px-4 py-2 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-xl border border-outline-variant transition-colors shadow-sm flex items-center gap-2 font-medium"
      >
        <Pencil size={16} />
        Edit Artist
      </a>
    {/if}
  </div>

  {#if !artist}
    <div
      class="bg-surface-container border border-outline-variant rounded-3xl p-12 text-center"
    >
      <p class="text-on-surface-variant/70">Artist not found or error loading data.</p>
      <button
        onclick={() => history.back()}
        class="mt-4 text-primary hover:underline"
      >
        Go back
      </button>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <!-- Left side: Artist Profile Card -->
      <div class="md:col-span-1">
        <div class="bg-surface-container border border-outline-variant rounded-3xl p-6 flex flex-col items-center shadow-md">
          <div class="w-40 h-40 rounded-full border-4 border-outline-variant overflow-hidden mb-6 bg-surface-highest shadow-xl relative">
            {#if artist.avatar_url}
              <img
                src={artist.avatar_url}
                alt={artist.name}
                class="w-full h-full object-cover"
              />
            {:else}
              <div class="w-full h-full flex items-center justify-center text-on-surface-variant/40">
                <User size={64} />
              </div>
            {/if}
          </div>
          <h2 class="text-2xl font-bold text-center text-on-surface wrap-break-word drop-shadow-sm w-full">
            {artist.name}
          </h2>
          {#if artist.name_jp}
            <p class="text-on-surface-variant/70 mt-1 mb-4 text-center">{artist.name_jp}</p>
          {/if}

          <!-- Badge Status -->
          <div class="mt-4">
            <span
              class="px-3 py-1 text-xs font-bold rounded-full {artist.status ? 'bg-green-500/20 text-green-400 border border-green-500/30' : 'bg-red-500/20 text-red-400 border border-red-500/30'}"
            >
              {artist.status ? "ACTIVE" : "INACTIVE"}
            </span>
          </div>

          <!-- Metadata -->
          <div class="w-full mt-8 border-t border-outline-variant pt-6 space-y-4">
            <div class="flex justify-between items-center text-sm">
              <span class="text-on-surface-variant/40">Slug</span>
              <span class="text-on-surface-variant font-medium bg-surface-highest px-2 py-0.5 rounded">{artist.slug}</span>
            </div>
            <div class="flex justify-between items-center text-sm">
              <span class="text-on-surface-variant/40">AniList ID</span>
              <span class="text-on-surface-variant font-medium">
                {#if artist.anilist_id}
                  <a href="https://anilist.co/staff/{artist.anilist_id}" target="_blank" rel="noreferrer" class="text-primary hover:underline">
                    {artist.anilist_id}
                  </a>
                {:else}
                  <span class="text-gray-600">-</span>
                {/if}
              </span>
            </div>
            <div class="flex justify-between items-center text-sm">
              <span class="text-on-surface-variant/40">AnimeThemes ID</span>
              <span class="text-on-surface-variant font-medium">
                {#if artist.animethemes_id}
                  {artist.animethemes_id}
                {:else}
                  <span class="text-gray-600">-</span>
                {/if}
              </span>
            </div>
            <div class="flex justify-between items-center text-sm">
              <span class="text-on-surface-variant/40">Songs Overview</span>
              <span class="text-on-surface-variant font-medium">{artist.songs_count || 0} track{artist.songs_count === 1 ? '' : 's'}</span>
            </div>
            <div class="flex justify-between items-center text-sm">
              <span class="text-on-surface-variant/40">Followers</span>
              <span class="text-on-surface-variant font-medium">{artist.favorites_count || 0}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Right side: Latest Songs -->
      <div class="md:col-span-2">
        <div class="bg-surface-container border border-outline-variant rounded-3xl p-6 shadow-md h-full">
          <div class="flex justify-between items-center mb-6">
            <h3 class="text-xl font-bold text-on-surface flex items-center gap-2">
               <Music size={20} class="text-primary" />
               Recent Songs
            </h3>
            <span class="text-sm text-on-surface-variant/40 bg-surface-highest px-3 py-1 rounded-full border border-outline-variant">Top 10 Latest</span>
          </div>

          {#if songs && songs.length > 0}
            <div class="space-y-3">
              {#each songs as song}
                <div class="p-4 bg-surface-highest hover:bg-surface-highest rounded-2xl border border-outline-variant transition-colors flex items-center justify-between group">
                  <div class="flex flex-col gap-1 overflow-hidden">
                    <span class="text-on-surface font-medium truncate w-full group-hover:text-primary transition-colors">
                      {song.name || song.song_romaji}
                    </span>
                    <span class="text-xs text-on-surface-variant/70 capitalize flex items-center gap-1.5">
                      <span class="inline-block w-2 h-2 rounded-full {song.type_name === 'OP' ? 'bg-blue-400' : song.type_name === 'ED' ? 'bg-purple-400' : 'bg-gray-400'}"></span>
                      {song.type_name || "OST"} • {song.anime?.title || "Unknown Anime"}
                    </span>
                  </div>
                  <div class="flex items-center gap-4 shrink-0 pl-4">
                    <span class="text-xs text-on-surface-variant/40 font-medium flex items-center gap-1">
                        <Play size={12} />
                       {song.views || 0}
                    </span>
                    <span class="text-xs text-on-surface-variant/40 font-medium flex items-center gap-1">
                        <Heart size={12} />
                       {song.favorites_count || 0}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="h-40 border-2 border-dashed border-outline-variant rounded-2xl flex flex-col items-center justify-center text-center">
               <Music2 size={48} class="text-gray-600 mb-2" />
               <p class="text-on-surface-variant/70 font-medium">No songs available for this artist.</p>
               <p class="text-sm text-gray-600 mt-1">They might have no tracks associated yet.</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
