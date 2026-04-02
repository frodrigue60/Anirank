<script lang="ts">
  import { getSongName, getFormattedScore } from "$lib/song-utils";
  import { authState } from "$lib/state/auth.svelte";

  let { song } = $props();
</script>

<div
  class="group relative overflow-hidden rounded-2xl h-48 transition-all duration-300 border border-primary/10 bg-background-dark hover:shadow-2xl hover:shadow-primary/20 hover:border-primary/30 hover:-translate-y-1"
>
  <!-- Background Banner -->
  <div
    class="absolute inset-0 bg-cover bg-center transition-transform duration-700 group-hover:scale-105"
    style="background-image: url('{song.anime?.banner_url ||
      song.anime?.cover_url ||
      '/images/placeholders/default-banner.jpg'}'); filter: brightness(0.5);"
  ></div>
  <!-- Gradient Overlay -->
  <div
    class="absolute inset-0 bg-linear-to-r from-background-dark via-background-dark/65 to-transparent"
  ></div>

  <!-- Content -->
  <div class="relative h-full p-6 flex items-center justify-between">
    <div class="space-y-1 pr-4">
      <div class="flex items-center gap-2 mb-2">
        <div
          class="inline-flex items-center px-2 py-0.5 rounded bg-primary text-[10px] font-black text-white uppercase tracking-wider shadow-lg"
        >
          {song.type}
          {song.number || ""}
        </div>
        <div
          class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-black text-white uppercase tracking-wider shadow-lg {song.user_rating
            ? 'bg-green-500/50'
            : 'bg-slate-500/50 '}"
        >
          {song.user_rating ? "Rated" : "Not rated"}
        </div>
      </div>
      <h3
        class="text-2xl font-bold text-white group-hover:text-primary transition-colors line-clamp-1 drop-shadow-md"
      >
        <a
          href="/songs/{song.anime?.slug}/{song.slug}"
          class="hover:underline"
          title="View song details: {getSongName(song)}">{getSongName(song)}</a
        >
      </h3>
      <p class="text-slate-300 text-sm font-medium line-clamp-1">
        by
        {#if song.artists && song.artists.length > 0}
          {#each song.artists as artist, i}
            <span class="transition-colors"
              >{artist.name ?? artist.name_jp ?? "N/A"}</span
            >{#if i < song.artists.length - 1},
            {/if}
          {/each}
        {:else}
          Without artists
        {/if}
      </p>
      <p class="text-slate-500 text-xs italic mt-2 line-clamp-1">
        {song.anime?.title || "N/A"}
      </p>
    </div>

    <div class="flex flex-col items-end gap-2 shrink-0">
      <!-- Rating -->
      <div
        class="glass px-3 py-2 rounded-xl flex items-center gap-1.5 border-primary/20"
      >
        <span class="material-symbols-outlined text-primary text-sm filled"
          >star</span
        >
        <span class="text-white font-bold text-lg"
          >{getFormattedScore(
            song.average_rating,
            authState.user?.score_format,
          )}</span
        >
      </div>

      <!-- Play Button -->
      <div class="flex items-center gap-2 mt-4">
        <a
          href="/songs/{song.anime?.slug}/{song.slug}"
          class="flex items-center justify-center h-12 w-12 rounded-full bg-white/10 hover:bg-primary transition-all text-white border border-white/10 group-hover:border-primary/50 group-hover:scale-110 shadow-lg"
          title="Play theme: {getSongName(song)}"
        >
          <span class="material-symbols-outlined text-2xl">play_arrow</span>
        </a>
      </div>
    </div>
  </div>
</div>
