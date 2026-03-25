<script lang="ts">
  import { onMount } from "svelte";
  import { getSongName, getFormattedScore } from "$lib/song-utils";
  import { authState } from "$lib/state/auth.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";

  let {
    songs = [],
    loading = false,
    hasMore = false,
    onLoadMore = () => {},
    startIndex = 0,
    rankingType = "main",
  } = $props();

  function formatScore(score: string | number | undefined | null) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  function getMovement(current: number, previous: number | undefined | null) {
    if (previous === undefined || previous === null || previous === 0)
      return "new";
    if (previous > current) return "up";
    if (previous < current) return "down";
    return "stable";
  }
</script>

<div
  class="bg-surface-dark/30 border border-white/5 rounded-2xl overflow-hidden mb-8"
>
  <div
    class="grid grid-cols-[80px_1fr_120px_140px] gap-4 px-8 py-4 border-b border-white/5 text-[10px] font-black uppercase tracking-widest text-white/30 bg-surface-darker/50"
  >
    <div class="text-center">Rank</div>
    <div>Theme Info</div>
    <div class="text-center">Score</div>
    <div class="text-right">Actions</div>
  </div>

  <div class="flex flex-col">
    {#if songs.length > 0}
      {#each songs as item, index}
        {@const rank = startIndex + index + 1}
        {@const previousRank =
          rankingType === "seasonal"
            ? item.prev_seasonal_rank
            : item.prev_main_rank}
        {@const movement = getMovement(rank, previousRank)}
        <div
          class="ranking-row grid grid-cols-[80px_1fr_120px_140px] gap-4 px-8 py-5 items-center transition-colors border-b border-white/5 last:border-0 hover:bg-white/5"
        >
          <div class="flex flex-col items-center gap-1">
            <span
              class="text-2xl font-black leading-none {rank <= 3
                ? 'text-primary'
                : 'text-white/90'}">{rank.toString().padStart(2, "0")}</span
            >

            {#if movement === "up"}
              <div
                class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-green-500/10 text-green-400"
                title={`Previous rank: ${previousRank}`}
              >
                <span class="material-symbols-outlined text-[12px] font-black"
                  >arrow_drop_up</span
                >
                <span class="text-[9px] font-black">{previousRank! - rank}</span
                >
              </div>
            {:else if movement === "down"}
              <div
                class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-red-500/10 text-red-400"
                title={`Previous rank: ${previousRank}`}
              >
                <span class="material-symbols-outlined text-[12px] font-black"
                  >arrow_drop_down</span
                >
                <span class="text-[9px] font-black">{rank - previousRank!}</span
                >
              </div>
            {:else if movement === "stable"}
              <div class="text-white/10 flex items-center">
                <span class="material-symbols-outlined text-[14px]">remove</span
                >
              </div>
            {:else if movement === "new"}
              <span
                class="text-[8px] font-black text-primary px-1.5 py-0.5 bg-primary/10 rounded uppercase tracking-tighter border border-primary/20"
                >New</span
              >
            {/if}
          </div>
          <div class="flex items-center gap-6 min-w-0">
            <div
              class="w-16 h-16 rounded-lg overflow-hidden shrink-0 shadow-lg shadow-black/40 border border-white/10"
            >
              <img
                alt={getSongName(item)}
                title={getSongName(item)}
                class="w-full h-full object-cover"
                src={item.anime?.cover_url ??
                  "/images/placeholders/default.jpg"}
              />
            </div>
            <div class="min-w-0 flex flex-col">
              {#if true}
                {@const songName = getSongName(item)}
                {@const artistNames =
                  item.artists?.map((a: any) => a.name).join(", ") ?? "Unknown"}
                <h3
                  class="text-lg font-bold text-white truncate leading-tight"
                  title={songName}
                >
                  {songName}
                </h3>
                <span class="text-white/60 truncate" title={artistNames}>
                  {artistNames}
                </span>
                <span
                  class="text-primary font-bold truncate"
                  title={item.anime?.title}
                >
                  {item.anime?.title}
                </span>
              {/if}
            </div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-black text-white tracking-tight">
              {formatScore(item.average_rating)}
            </div>
            <div
              class="text-[10px] font-bold text-white/30 uppercase tracking-widest"
            >
              Avg Rating
            </div>
            <div>
              {#if item.user_rating}
                <div class="text-xs">
                  <span class="text-green-400/60">Rated</span>
                </div>
              {:else}
                <div class="text-xs">
                  <span class="text-white/30">Not rated</span>
                </div>
              {/if}
            </div>
          </div>
          <div class="flex items-center justify-end gap-2">
            <a
              href={`/songs/${item.anime?.slug}/${item.slug}`}
              class="w-10 h-10 rounded-full flex items-center justify-center bg-white/5 hover:bg-white/10 text-white transition-all shadow-lg hover:shadow-primary/20"
              title="Play theme"
            >
              <span class="material-symbols-outlined text-[20px]"
                >play_arrow</span
              >
            </a>
          </div>
        </div>
      {/each}

      <InfiniteScroll {hasMore} {loading} {onLoadMore} />
    {:else if !loading}
      <div
        class="flex flex-col items-center justify-center py-20 text-white/30"
      >
        <span class="material-symbols-outlined text-6xl mb-4">music_off</span>
        <p class="text-lg font-bold">No themes found</p>
      </div>
    {/if}
  </div>
</div>
