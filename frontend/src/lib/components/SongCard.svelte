<script lang="ts">
  import { getSongName, getFormattedScore } from "$lib/song-utils";
  import { authState } from "$lib/state/auth.svelte";
  import { getSrcset } from "$lib/utils/image";

  let { song } = $props();
</script>

<div
  class="group relative overflow-hidden rounded-md h-48 transition-all duration-300 bg-surface-container shadow-sm hover:shadow-2xl hover:-translate-y-1"
>
  <!-- Background Banner -->
  <img
    src={song.anime?.banner_url ||
      song.anime?.cover_url ||
      "/images/placeholders/default-banner.jpg"}
    srcset={getSrcset(song.anime?.banner_sources ?? song.anime?.cover_sources)}
    sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 600px"
    alt=""
    class="absolute inset-0 w-full h-full object-cover object-center transition-transform duration-700 group-hover:scale-105 brightness-50"
    loading="lazy"
  />
  <!-- Gradient Overlay -->
  <div
    class="absolute inset-0 bg-linear-to-r from-surface-container via-surface-container/65 to-transparent"
  ></div>

  <!-- Content -->
  <div class="relative h-full p-6 flex items-center justify-between">
    <div class="space-y-1 pr-4 h-full flex flex-col justify-end">
      <div class="flex items-center gap-2 mb-2">
        <div
          class="inline-flex items-center px-2 py-0.5 rounded bg-primary text-[10px] font-black text-white uppercase tracking-wider shadow-lg"
        >
          {song.type}
          {song.number || ""}
        </div>
        <div
          class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-black text-white uppercase tracking-wider shadow-lg {song.user_rating
            ? 'bg-green-500'
            : 'bg-slate-400 '}"
        >
          {song.user_rating ? "Rated" : "Not rated"}
        </div>
      </div>
      <h3
        class="text-2xl font-bold text-on-surface group-hover:text-primary transition-colors line-clamp-1 drop-shadow-md"
      >
        <a
          href="/songs/{song.anime?.slug}/{song.slug}"
          class="hover:underline"
          title="View song details: {getSongName(song)}">{getSongName(song)}</a
        >
      </h3>
      <p class="text-on-surface-variant text-sm font-medium line-clamp-1">
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
      <p class="text-on-surface-variant text-xs italic mt-2 line-clamp-1">
        {song.anime?.title || "N/A"}
      </p>
    </div>

    <div class="flex flex-col items-end gap-2 shrink-0 h-full">
      <!-- Rating -->
      <div
        class="px-2 py-1 rounded-xl flex items-center gap-1.5 text-yellow-500 drop-shadow"
        title="Average rating"
      >
        <span class="material-symbols-outlined text-sm filled">star</span>
        <span class="font-bold text-lg"
          >{getFormattedScore(
            song.average_rating,
            authState.user?.score_format,
          )}</span
        >
      </div>

      <!-- Play Button -->
      <!-- <div class="flex items-center gap-2 mt-4">
        <a
          href="/songs/{song.anime?.slug}/{song.slug}"
          class="flex items-center justify-center h-12 w-12 rounded-full bg-secondary-container hover:bg-primary/80 transition-all text-on-surface border border-white/10 group-hover:border-primary/50 group-hover:scale-110 shadow-lg"
          title="Play theme: {getSongName(song)}"
        >
          <span class="material-symbols-outlined text-2xl">play_arrow</span>
        </a>
      </div> -->
    </div>
  </div>
</div>
