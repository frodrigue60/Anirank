<script lang="ts">
  import SEO from "$lib/components/SEO.svelte";
  import snarkdown from "snarkdown";
  import { createTrustedHTML } from "$lib/trusted";
  import ArtistAvatarCard from "$lib/components/ArtistAvatarCard.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Heart from "lucide-svelte/icons/heart";
  import Mic2 from "lucide-svelte/icons/mic-2";
  import History from "lucide-svelte/icons/history";
  import Construction from "lucide-svelte/icons/construction";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import { PUBLIC_API_URL } from "$lib/api";
  import {
    getSongName,
    getSongArtistNames,
    getFormattedScore,
  } from "$lib/song-utils";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";

  let { data } = $props();

  const accentColor = $derived(data.profile?.profile_color || "#683bc9");
  const hasFavorites = $derived(
    (data.initialSongs?.length ?? 0) > 0 || (data.artists?.length ?? 0) > 0,
  );
  const recentRatings = $derived(data.ratingInsights?.recent_ratings ?? []);
  const scoreDistribution = $derived(
    data.ratingInsights?.score_distribution ?? [],
  );
  const totalRatings = $derived(data.ratingInsights?.total_ratings ?? 0);
  const scoreScale = $derived(
    data.profile?.score_format === "POINT_100"
      ? 100
      : data.profile?.score_format === "POINT_5"
        ? 5
        : 10,
  );

  function renderMarkdown(text: string | null | undefined) {
    if (!text) return "";
    return createTrustedHTML(snarkdown(text));
  }

  function formatRatedAt(value: string) {
    return new Intl.DateTimeFormat("en", {
      dateStyle: "medium",
    }).format(new Date(value));
  }
</script>

<SEO
  title={`${data.profile.name}'s Profile`}
  description={`Check out ${data.profile.name}'s anime theme song favorites and stats on AniRank.`}
  image={`${PUBLIC_API_URL}/og/user/${data.profile.slug}`}
  type="profile"
/>

<div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
  <!-- Left column -->
  <div class="lg:col-span-8 flex flex-col gap-8">
    <!-- Recent rating activity -->
    <section class="bg-surface-low rounded-md p-5 sm:p-6">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-1 h-5 rounded-sm bg-primary shrink-0" aria-hidden="true"></div>
        <h2 class="text-lg font-black tracking-tight text-on-surface uppercase">
          Recent Rating Activity
        </h2>
      </div>
      {#if recentRatings.length > 0}
        <div class="flex flex-col gap-3">
          {#each recentRatings as item (item.song.id)}
            <a
              href="/animes/{item.song.anime.slug}/{item.song.slug}"
              class="group bg-surface-container hover:bg-surface-highest rounded-md p-3 sm:p-4 flex items-center gap-4 transition-colors"
              title="View theme: {getSongName(item.song)}"
            >
              <div
                class="size-14 sm:size-16 rounded-md overflow-hidden bg-surface-highest shrink-0"
              >
                <OptimizedImage
                  src={item.song.anime?.cover_url}
                  alt=""
                  class="w-full h-full object-cover"
                  sizes="64px"
                />
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-xs text-on-surface-variant/80 truncate">
                  {item.song.anime?.title}
                </p>
                <h3
                  class="text-base font-bold text-on-surface tracking-tight truncate mt-0.5 group-hover:text-primary transition-colors"
                >
                  {getSongName(item.song)}
                </h3>
                <p class="text-xs text-on-surface-variant/80 truncate mt-0.5">
                  {getSongArtistNames(item.song.artists)} · {formatRatedAt(
                    item.rated_at,
                  )}
                </p>
              </div>
              <div
                class="min-w-16 rounded-md bg-surface-highest px-2 py-2 text-center shrink-0"
              >
                <span class="block text-lg font-black text-primary tabular-nums">
                  {getFormattedScore(
                    item.rating,
                    data.profile?.score_format,
                  )}
                </span>
                <span class="block text-[10px] font-bold text-on-surface-variant/80">
                  /{scoreScale}
                </span>
              </div>
            </a>
          {/each}
        </div>
      {:else}
        <p class="py-8 text-center text-sm text-on-surface-variant">
          This user has not rated any themes yet.
        </p>
      {/if}
    </section>

    <!-- Favorite themes (live data) -->
    {#if data.initialSongs && data.initialSongs.length > 0}
      <section>
        <div class="flex items-center justify-between gap-3 mb-4">
          <div class="flex items-center gap-3 min-w-0">
            <div
              class="w-1 h-5 rounded-sm shrink-0"
              style="background-color: {accentColor}"
              aria-hidden="true"
            ></div>
            <h2
              class="text-lg font-black tracking-tight text-on-surface uppercase flex items-center gap-2"
            >
              <Heart size={18} style="color: {accentColor}" aria-hidden="true" />
              Favorite Themes
            </h2>
          </div>
          <a
            href={`/users/${data.profile.slug}/favorites`}
            class="text-xs font-bold text-primary hover:underline inline-flex items-center gap-0.5 shrink-0"
          >
            View all
            <ChevronRight size={14} aria-hidden="true" />
          </a>
        </div>
        <div class="flex flex-col gap-3">
          {#each data.initialSongs.slice(0, 6) as song}
            <a
              href="/animes/{song.anime?.slug}/{song.slug}"
              class="group bg-surface-container hover:bg-surface-highest rounded-md p-3 sm:p-4 flex items-center gap-4 transition-colors"
              title="View theme: {getSongName(song)}"
            >
              <div
                class="relative size-14 sm:size-16 rounded-md overflow-hidden shrink-0 bg-surface-highest"
              >
                <OptimizedImage
                  src={song.anime?.cover_url}
                  sources={song.anime?.cover_sources}
                  alt=""
                  class="w-full h-full object-cover"
                  sizes="64px"
                />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 min-w-0">
                  {#if song.song_type?.name || song.type}
                    <span
                      class="px-2 py-0.5 text-[10px] font-bold rounded-sm bg-surface-highest text-primary shrink-0"
                    >
                      {song.song_type?.name || song.type}{song.theme_num || ""}
                    </span>
                  {/if}
                  {#if song.anime?.title}
                    <span
                      class="text-xs text-on-surface-variant/80 truncate"
                    >
                      {song.anime.title}
                    </span>
                  {/if}
                </div>
                <h3
                  class="text-base font-bold text-on-surface tracking-tight truncate mt-0.5 group-hover:text-primary transition-colors"
                >
                  {getSongName(song)}
                </h3>
                <p class="text-xs text-on-surface-variant/80 truncate mt-0.5">
                  {getSongArtistNames(song.artists)}
                </p>
              </div>
            </a>
          {/each}
        </div>
      </section>
    {/if}

    <!-- Top seasons — WIP -->
    <section
      class="relative flex flex-col gap-4"
      aria-label="Top rated seasons, work in progress"
    >
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-lg font-bold text-on-surface tracking-tight">
          Top Rated Seasons
        </h2>
        <span
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm bg-surface-highest text-[10px] font-black uppercase tracking-wider text-on-surface-variant"
        >
          <Construction size={12} aria-hidden="true" />
          WIP
        </span>
      </div>
      <p class="text-xs text-on-surface-variant/80 -mt-2">
        Needs seasonal rating aggregation API
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        {#each [1, 2, 3] as rank}
          <div
            class="bg-surface-low rounded-md p-5 border border-dashed border-outline-variant flex flex-col gap-3"
          >
            <div class="flex items-center justify-between">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/50"
              >
                Season
              </span>
              <span
                class="text-xs font-mono font-semibold text-on-surface-variant/40"
              >
                #{rank}
              </span>
            </div>
            <div class="h-5 w-2/3 rounded-sm bg-surface-highest"></div>
            <div class="h-3 w-1/2 rounded-sm bg-surface-highest"></div>
            <div class="h-8 w-16 rounded-sm bg-surface-highest mt-2"></div>
          </div>
        {/each}
      </div>
    </section>
  </div>

  <!-- Right column -->
  <aside class="lg:col-span-4 flex flex-col gap-6">
    <!-- Top animes — WIP -->
    <section
      class="relative bg-surface-low rounded-md p-5 border border-dashed border-outline-variant"
      aria-label="Top recent animes, work in progress"
    >
      <span
        class="absolute top-3 right-3 inline-flex items-center gap-1 px-2 py-0.5 rounded-sm bg-surface-highest text-[10px] font-black uppercase tracking-wider text-on-surface-variant"
      >
        <Construction size={12} aria-hidden="true" />
        WIP
      </span>
      <h3
        class="text-sm font-bold uppercase tracking-tight text-on-surface pr-14"
      >
        Top Recent Animes
      </h3>
      <p class="text-xs text-on-surface-variant/80 mt-1 mb-4">
        Needs AniList summary or ratings-by-anime API
      </p>
      <div class="flex flex-col gap-2">
        {#each [1, 2, 3] as _}
          <div class="flex items-center gap-3 p-2 rounded-md">
            <div
              class="size-10 rounded-md bg-surface-highest shrink-0"
              aria-hidden="true"
            ></div>
            <div class="flex-1 space-y-1.5">
              <div class="h-3 w-3/4 rounded-sm bg-surface-highest"></div>
              <div class="h-2 w-1/2 rounded-sm bg-surface-highest"></div>
            </div>
            <div class="h-5 w-8 rounded-sm bg-surface-highest"></div>
          </div>
        {/each}
      </div>
    </section>

    <!-- Score distribution -->
    <section class="bg-surface-low rounded-md p-5">
      <h3
        class="text-sm font-bold uppercase tracking-tight text-on-surface"
      >
        Score Distribution
      </h3>
      {#if totalRatings > 0 && scoreDistribution.length > 0}
        <div class="flex flex-col gap-2.5 mt-4">
          {#each scoreDistribution as row (row.label)}
            <div
              class="flex items-center gap-2 text-[11px]"
              aria-label="{row.label}: {row.count} ratings"
            >
              <span
                class="w-14 font-medium text-on-surface-variant shrink-0"
              >
                {row.label}
              </span>
              <div
                class="flex-1 h-2.5 rounded-sm bg-surface-highest overflow-hidden"
              >
                <div
                  class="h-full rounded-sm bg-primary"
                  style="width: {row.percentage}%"
                ></div>
              </div>
              <span class="w-8 text-right font-mono text-on-surface-variant">
                {row.count}
              </span>
            </div>
          {/each}
        </div>
      {:else}
        <p class="mt-4 text-xs text-on-surface-variant">
          No rating data available.
        </p>
      {/if}
    </section>

    <!-- Bio -->
    {#if data.profile.about}
      <section class="bg-surface-container rounded-md p-5 sm:p-6">
        <div class="flex items-center justify-between gap-2 mb-3">
          <span
            class="text-[11px] font-bold uppercase tracking-wider"
            style="color: {accentColor}"
          >
            About
          </span>
        </div>
        <div
          class="border-l-4 pl-3 text-sm leading-relaxed text-on-surface-variant markdown-content"
          style="border-color: {accentColor}"
        >
          {@html renderMarkdown(data.profile.about)}
        </div>
      </section>
    {/if}

    <!-- Favorite artists -->
    {#if data.artists && data.artists.length > 0}
      <section class="bg-surface-container rounded-md p-5 sm:p-6">
        <div class="flex items-center justify-between gap-3 mb-4">
          <h3
            class="text-sm font-bold uppercase tracking-tight text-on-surface flex items-center gap-2"
          >
            <Mic2 size={16} style="color: {accentColor}" aria-hidden="true" />
            Favorite Artists
          </h3>
          <a
            href={`/users/${data.profile.slug}/artists`}
            class="text-xs font-bold text-primary hover:underline"
          >
            View all
          </a>
        </div>
        <div class="grid grid-cols-3 gap-4">
          {#each data.artists.slice(0, 6) as artist}
            <ArtistAvatarCard {artist} />
          {/each}
        </div>
      </section>
    {/if}

    {#if !hasFavorites && !data.profile.about && recentRatings.length === 0}
      <EmptyState
        title="Quiet Profile"
        message="This user hasn't added favorites or a bio yet."
        icon={History as any}
      />
    {/if}
  </aside>
</div>
