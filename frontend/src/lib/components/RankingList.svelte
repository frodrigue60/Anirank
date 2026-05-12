<script lang="ts">
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import { getSongName, getFormattedScore } from "$lib/song-utils";
  import { authState } from "$lib/state/auth.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import TrendingUp from "lucide-svelte/icons/trending-up";
  import TrendingDown from "lucide-svelte/icons/trending-down";
  import Minus from "lucide-svelte/icons/minus";
  import Play from "lucide-svelte/icons/play";
  import Music2 from "lucide-svelte/icons/music-2";

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
  class="bg-surface-container shadow-sm border border-primary/10 rounded-md overflow-x-auto mb-8"
>
  <table class="w-full table-fixed border-collapse">
    <thead>
      <tr
        class="border-b border-primary/10 bg-surface-highest text-[10px] font-black uppercase tracking-widest text-on-surface-variant"
      >
        <th class="w-16 sm:w-20 py-4 pl-4 sm:pl-8 pr-2 text-center font-black">Rank</th>
        <th class="py-4 px-2 text-left font-black">Theme Info</th>
        <th class="w-20 sm:w-[120px] py-4 px-2 text-center font-black">Score</th>
        <th class="w-14 sm:w-[80px] py-4 pl-2 pr-4 sm:pr-8 text-right font-black">Actions</th>
      </tr>
    </thead>

    <tbody class="divide-y divide-primary/10">
      {#if songs.length > 0}
        {#each songs as item, index}
          {@const rank = startIndex + index + 1}
          {@const previousRank =
            rankingType === "seasonal"
              ? item.prev_seasonal_rank
              : item.prev_main_rank}
          {@const movement = getMovement(rank, previousRank)}
          <tr class="transition-colors hover:bg-surface">
            <td class="py-5 pl-8 pr-2 text-center">
              <div class="flex flex-col items-center gap-1">
                <span
                  class="text-2xl font-black leading-none {rank <= 3
                    ? 'text-primary'
                    : 'text-on-surface/90'}"
                  >{rank.toString().padStart(2, "0")}</span
                >

                {#if movement === "up"}
                  <div
                    class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-green-500/10 text-green-400"
                    title={`Previous rank: ${previousRank}`}
                  >
                    <TrendingUp size={12} />

                    <span class="text-[9px] font-black"
                      >{previousRank! - rank}</span
                    >
                  </div>
                {:else if movement === "down"}
                  <div
                    class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-red-500/10 text-red-400"
                    title={`Previous rank: ${previousRank}`}
                  >
                    <TrendingDown size={12} />

                    <span class="text-[9px] font-black"
                      >{rank - previousRank!}</span
                    >
                  </div>
                {:else if movement === "stable"}
                  <div class="text-on-surface-variant/20 flex items-center">
                    <Minus size={14} />

                  </div>
                {:else if movement === "new"}
                  <span
                    class="text-[8px] font-black text-primary px-1.5 py-0.5 bg-primary/10 rounded uppercase tracking-tighter border border-primary/20"
                    >New</span
                  >
                {/if}
              </div>
            </td>

            <td class="py-5 px-2 overflow-hidden">
              <div class="flex items-center gap-6 min-w-0">
                <div
                  class="w-16 h-auto aspect-12/18 rounded-lg overflow-hidden shrink-0 shadow-lg shadow-black/40 border border-primary/10 relative"
                >
                  <img
                    alt={getSongName(item)}
                    title={getSongName(item)}
                    class="w-full h-full object-cover"
                    src={item.anime?.cover_url ??
                      "/images/placeholders/default.svg"}
                    loading="lazy"
                  />
                </div>
                <div class="min-w-0 flex flex-col flex-1">
                  {#if true}
                    {@const songName = getSongName(item)}
                    {@const artistNames =
                      item.artists?.map((a: any) => a.name).join(", ") ??
                      "Unknown"}
                    <h3
                      class="text-lg font-bold text-on-surface truncate leading-tight"
                      title={songName}
                    >
                      {songName}
                    </h3>
                    <span
                      class="text-on-surface/80 truncate block"
                      title={artistNames}
                    >
                      {artistNames}
                    </span>
                    <span
                      class="text-primary font-bold truncate block"
                      title={item.anime?.title}
                    >
                      {item.anime?.title}
                    </span>
                    {#if item.season && item.year}
                      <div class="flex items-center gap-1.5 mt-1">
                        <span
                          class="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest border transition-colors {rankingType ===
                            'seasonal' &&
                          (item.season?.id !== page.data.songsData?.current_season?.id ||
                            item.year?.id !== page.data.songsData?.current_year?.id)
                            ? 'bg-primary/20 border-primary text-primary shadow-sm shadow-primary/10'
                            : 'bg-surface-highest border-on-surface-variant/10 text-on-surface-variant/60'}"
                        >
                          {item.year.name} {item.season.name}
                        </span>
                      </div>
                    {/if}
                  {/if}
                </div>
              </div>
            </td>

            <td class="py-5 px-2 text-center">
              <div class="text-2xl font-black text-on-surface tracking-tight">
                {formatScore(item.average_rating)}
              </div>
              <div
                class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest"
              >
                Avg Rating
              </div>
              <div>
                {#if item.user_rating}
                  <div class="text-xs">
                    <span class="text-emerald-500/60 font-bold">Rated</span>
                  </div>
                {:else}
                  <div class="text-xs">
                    <span class="text-on-surface-variant/80"
                      >Not rated</span
                    >
                  </div>
                {/if}
              </div>
            </td>

            <td class="py-5 pl-2 pr-8 text-right">
              <div class="flex items-center justify-end gap-2">
                <a
                  href={`/songs/${item.anime?.slug}/${item.slug}`}
                  class="w-10 h-10 rounded-full flex items-center justify-center bg-primary/10 hover:bg-primary text-primary hover:text-white transition-all shadow-lg hover:shadow-primary/20"
                  title="Play theme"
                >
                  <Play class="fill-current" size={20} />

                </a>
              </div>
            </td>
          </tr>
        {/each}

        <tr>
          <td colspan="4" class="p-0">
            <InfiniteScroll {hasMore} {loading} {onLoadMore} />
          </td>
        </tr>
      {:else if !loading}
        <tr>
          <td colspan="4" class="py-20 text-on-surface-variant">
            <div class="flex flex-col items-center justify-center opacity-60">
              <Music2 size={60} class="mb-4" />

              <p class="text-lg font-bold">No themes found</p>
            </div>
          </td>
        </tr>
      {/if}
    </tbody>
  </table>
</div>
