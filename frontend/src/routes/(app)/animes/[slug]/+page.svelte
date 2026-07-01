<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongName, getFormattedScore, buildSongPlayHref } from "$lib/song-utils";
  import SEO from "$lib/components/SEO.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { configState } from "$lib/state/config.svelte";
  import { createTrustedHTML } from "$lib/trusted";
  import { page } from "$app/state";
  import Globe from "lucide-svelte/icons/globe";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import ChevronDown from "lucide-svelte/icons/chevron-down";
  import ChevronUp from "lucide-svelte/icons/chevron-up";
  import Star from "lucide-svelte/icons/star";
  import Play from "lucide-svelte/icons/play";
  import Clapperboard from "lucide-svelte/icons/clapperboard";
  import { PUBLIC_API_URL } from "$lib/api";
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";
  import ArtistListCell from "$lib/components/ArtistListCell.svelte";
  import type { Song, SongVariant, SongVariantVideo } from "$lib/types/song";

  const FILTER_TYPE_ORDER = ["OP", "ED"];

  let { data }: { data: PageData } = $props();
  let anime = $derived(data.data);

  function formatScore(score: number | string | null | undefined) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  let isExpanded = $state(false);
  let filterType = $state("all");
  let sortBy = $state("theme_num");

  function getSongTypeSortOrder(type: string | undefined): number {
    switch ((type || "").toUpperCase()) {
      case "OP":
        return 0;
      case "ED":
        return 1;
      default:
        return 2;
    }
  }

  function compareSongsByType(a: { type?: string }, b: { type?: string }) {
    const typeOrder =
      getSongTypeSortOrder(a.type) - getSongTypeSortOrder(b.type);
    if (typeOrder !== 0) return typeOrder;
    return (a.type || "").localeCompare(b.type || "");
  }

  let filteredAndSortedSongs = $derived.by(() => {
    let songs = anime.songs ? [...anime.songs] : [];

    // Filter
    if (filterType !== "all") {
      songs = songs.filter(
        (s) => s.type?.toLowerCase() === filterType.toLowerCase(),
      );
    }

    // Sort: OP → ED → otros tipos, luego por theme_num o score
    songs.sort((a, b) => {
      const typeOrder = compareSongsByType(a, b);
      if (typeOrder !== 0) return typeOrder;

      if (sortBy === "theme_num") {
        return (a.theme_num || "").localeCompare(b.theme_num || "", undefined, {
          numeric: true,
        });
      }

      return (b.average_rating || 0) - (a.average_rating || 0);
    });

    return songs;
  });

  let songTypeCounts = $derived.by(() => {
    const counts: Record<string, number> = { all: anime.songs?.length ?? 0 };
    for (const song of anime.songs ?? []) {
      const type = (song.type || "").toUpperCase();
      counts[type] = (counts[type] ?? 0) + 1;
    }
    return counts;
  });

  let orderedFilterTypes = $derived.by(() => {
    return [...configState.songTypes].sort((a, b) => {
      const aIdx = FILTER_TYPE_ORDER.indexOf(a.slug.toUpperCase());
      const bIdx = FILTER_TYPE_ORDER.indexOf(b.slug.toUpperCase());
      const aOrder = aIdx === -1 ? FILTER_TYPE_ORDER.length + 1 : aIdx;
      const bOrder = bIdx === -1 ? FILTER_TYPE_ORDER.length + 1 : bIdx;
      if (aOrder !== bOrder) return aOrder - bOrder;
      return a.slug.localeCompare(b.slug);
    });
  });

  let songGroups = $derived.by(() => {
    const songs = filteredAndSortedSongs;
    if (songs.length === 0) return [];

    if (filterType !== "all") {
      return [
        {
          id: filterType,
          label: getGroupLabel(filterType),
          songs,
        },
      ];
    }

    const openings = songs.filter((s) => s.type?.toUpperCase() === "OP");
    const endings = songs.filter((s) => s.type?.toUpperCase() === "ED");
    const others = songs.filter((s) => {
      const type = s.type?.toUpperCase() ?? "";
      return type !== "OP" && type !== "ED";
    });

    const groups: { id: string; label: string; songs: Song[] }[] = [];
    if (openings.length > 0) {
      groups.push({ id: "op", label: "Openings", songs: openings });
    }
    if (endings.length > 0) {
      groups.push({ id: "ed", label: "Endings", songs: endings });
    }
    if (others.length > 0) {
      groups.push({ id: "other", label: "Other Themes", songs: others });
    }
    return groups;
  });

  function getGroupLabel(type: string): string {
    switch (type.toUpperCase()) {
      case "OP":
        return "Openings";
      case "ED":
        return "Endings";
      case "INS":
        return "Insert Songs";
      default:
        return type.toUpperCase();
    }
  }

  function typeBadgeClass(type: string | undefined): string {
    switch ((type || "").toUpperCase()) {
      case "OP":
        return "bg-green-500/10 text-green-400 border-green-500/40";
      case "ED":
        return "bg-blue-500/10 text-blue-400 border-blue-500/40";
      default:
        return "bg-primary/10 text-primary border-primary/30";
    }
  }

  function formatThemeRating(score: number | string | null | undefined): string | null {
    const raw = typeof score === "string" ? parseFloat(score) : score;
    if (raw == null || Number.isNaN(raw) || raw <= 0) return null;
    return formatScore(score);
  }

  function getPrimaryVideo(variant: SongVariant): SongVariantVideo | SongVariant {
    if (variant.videos?.length) return variant.videos[0];
    return variant;
  }

  function songPlayHref(songSlug: string, versionNumber: number) {
    return buildSongPlayHref(anime.slug, songSlug, versionNumber);
  }

  function formatResolution(resolution?: number) {
    if (!resolution || resolution <= 0) return null;
    return `${resolution}p`;
  }
</script>

<SEO
  title="{anime.title} - AniRank"
  description="Listen to and rate the openings and endings from {anime.title}. {anime.description ||
    ''}"
  image={`${PUBLIC_API_URL}/og/anime/${anime.slug}`}
  type="video.tv_show"
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start">
    <!-- Sidebar -->
    <aside class="lg:col-span-3 space-y-8">
      <div class="flex flex-col space-y-2">
        <div
          class="relative rounded-md overflow-hidden shadow-2xl shadow-primary/10 border border-white/5 group"
        >
          <OptimizedImage
            src={anime.cover_url}
            sources={anime.cover_sources}
            alt="Cover art for {anime.title}"
            class="w-full h-auto aspect-2/3 object-cover transition-transform duration-700 group-hover:scale-105"
            sizes="(max-width: 1024px) 100vw, 320px"
          />
          <div
            class="absolute top-0 left-0 w-full h-full bg-linear-to-t from-background-dark/80 via-transparent to-transparent opacity-60"
          ></div>
        </div>

        <div>
          <a
            href="https://anilist.co/anime/{anime.anilist_id}"
            target="_blank"
            title="Track {anime.title} on AniList"
            class="w-full bg-primary hover:bg-primary/80 text-white font-bold py-2.5 rounded-sm transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 text-sm"
          >
            Track on AniList
          </a>
        </div>
      </div>

      <div class="space-y-6">
        <div class="">
          <h3
            class="text-xs font-bold text-primary uppercase tracking-widest mb-1"
          >
            Information
          </h3>
        </div>
        <div class="space-y-5 text-sm">
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-on-surface mb-0.5">Format</span>
            <span class="font-medium text-on-surface-variant"
              >{anime.format?.name || "Unknown"}</span
            >
          </div>
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-on-surface mb-0.5">Season</span>
            <span class="font-medium text-on-surface-variant"
              >{anime.season?.name || "Unknown"}
              {anime.year?.name || "Unknown"}</span
            >
          </div>

          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-on-surface mb-0.5">Studio</span>
            <div class="flex flex-wrap gap-1">
              {#if anime.studios?.length > 0}
                {#each anime.studios as studio, index}
                  <a
                    class="font-medium text-primary hover:underline"
                    href="/studios/{studio.slug}"
                    title="View details for studio: {studio.name}"
                    >{studio.name}</a
                  >
                  {index < anime.studios.length - 1 ? ", " : ""}
                {/each}
              {:else}
                <span class="font-medium text-on-surface-variant">Unknown</span>
              {/if}
            </div>
          </div>
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-on-surface mb-0.5">Producers</span>
            <span class="font-medium text-on-surface-variant">
              {#if anime.producers?.length > 0}
                {#each anime.producers as producer, index}
                  <a
                    class="font-medium text-primary hover:underline"
                    href="/producers/{producer.slug}"
                    title="View details for producer: {producer.name}"
                    >{producer.name}</a
                  >
                  {index < anime.producers.length - 1 ? ", " : ""}
                {/each}
              {:else}
                <span class="font-medium text-on-surface-variant">Unknown</span>
              {/if}
            </span>
          </div>

          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-on-surface mb-0.5">Romaji</span>
            <span class="font-medium text-on-surface-variant"
              >{anime.title}</span
            >
          </div>

          {#if anime.title_english}
            <div
              class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
            >
              <span class="block text-xs text-on-surface mb-0.5">English</span>
              <span class="font-medium text-on-surface-variant"
                >{anime.title_english}</span
              >
            </div>
          {/if}

          {#if anime.title_native}
            <div
              class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
            >
              <span class="block text-xs text-on-surface mb-0.5">Native</span>
              <span class="font-medium text-on-surface-variant"
                >{anime.title_native}</span
              >
            </div>
          {/if}

          {#if anime.synonyms && anime.synonyms.length > 0}
            <div
              class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
            >
              <span class="block text-xs text-on-surface mb-0.5">Synonyms</span>
              <span class="font-medium text-on-surface-variant"
                >{anime.synonyms.join(", ")}</span
              >
            </div>
          {/if}
        </div>
      </div>

      {#if anime.external_links?.length > 0}
        <div class="space-y-4 pt-4 border-t border-white/5">
          <h3 class="text-xs font-bold text-primary uppercase tracking-widest">
            External Links
          </h3>
          <div class="flex flex-col gap-2">
            {#each anime.external_links as link}
              <a
                class="flex items-center gap-3 px-3 py-2.5 rounded-sm bg-surface-container hover:bg-surface-container-high border border-white/5 hover:border-primary/30 transition-all text-sm group"
                href={link.url}
                target="_blank"
                title="Visit {link.name}"
              >
                <Globe size={18} class="text-on-surface-variant group-hover:text-primary transition-colors" />

                <span class="text-on-surface-variant group-hover:text-primary"
                  >{link.name}</span
                >
                <ExternalLink size={14} class="text-on-surface-variant ml-auto" />

              </a>
            {/each}
          </div>
        </div>
      {/if}
    </aside>

    <!-- Main Content -->
    <section class="lg:col-span-9 space-y-10">
      <div class="flex flex-col gap-6">
        <div class="flex flex-col md:flex-row md:items-end justify-between">
          <div>
            <h1
              class="text-4xl md:text-6xl font-black tracking-tight text-on-surface"
            >
              {anime.title}
            </h1>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          {#each anime.genres as genre}
            <span
              class="px-3 py-1 rounded-full border border-white/10 text-on-surface-variant text-xs font-medium bg-surface-container"
              >{genre.name}</span
            >
          {/each}
        </div>
      </div>

      <div class="space-y-4">
        <h2 class="text-xl font-bold flex items-center gap-3 text-on-surface">
          <span class="w-1.5 h-6 bg-primary rounded-full"></span>
          Synopsis
        </h2>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="bg-surface-container border border-white/5 p-8 rounded-md cursor-pointer hover:bg-surface-container-high transition-all group relative overflow-hidden"
          onclick={() => (isExpanded = !isExpanded)}
        >
          <p
            class="text-on-surface-variant leading-relaxed text-lg font-body transition-all duration-300 {isExpanded
              ? ''
              : 'line-clamp-4'}"
          >
            {@html createTrustedHTML(anime.description || "No synopsis available.")}
          </p>

          {#if !isExpanded && anime.description && anime.description.length > 200}
            <div
              class="absolute bottom-0 left-0 w-full h-16 bg-linear-to-t from-surface-container via-surface-container-high to-transparent flex items-end justify-center pb-2 transition-opacity"
            >
              <span
                class="text-primary font-bold text-xs uppercase tracking-widest flex items-center gap-1"
              >
                Read More
                <ChevronDown size={14} />

              </span>
            </div>
          {/if}

          {#if isExpanded}
            <div class="mt-4 flex justify-center">
              <span
                class="text-primary/50 font-bold text-xs uppercase tracking-widest flex items-center gap-1 hover:text-primary transition-colors"
              >
                Show Less
                <ChevronUp size={14} />

              </span>
            </div>
          {/if}
        </div>
      </div>

      <!-- Music Themes -->
      <div class="space-y-6 pt-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold flex items-center gap-3 text-on-surface">
            <span class="w-1.5 h-6 bg-primary rounded-full"></span>
            Music Themes
          </h2>
        </div>

        <div class="flex flex-col md:flex-row items-center justify-between gap-4">
          <div class="flex flex-wrap items-center gap-1 bg-surface-container p-1 rounded-sm border border-outline-variant/30">
            <button
              onclick={() => (filterType = "all")}
              class="px-3 py-1.5 rounded-sm text-[10px] font-bold uppercase tracking-widest transition-colors {filterType ===
              'all'
                ? 'bg-primary text-white'
                : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-low'}"
            >
              All ({songTypeCounts.all})
            </button>
            {#each orderedFilterTypes as type}
              {@const count = songTypeCounts[type.slug.toUpperCase()] ?? 0}
              {#if count > 0}
                <button
                  onclick={() => (filterType = type.slug)}
                  class="px-3 py-1.5 rounded-sm text-[10px] font-bold uppercase tracking-widest transition-colors {filterType ===
                  type.slug
                    ? 'bg-primary text-white'
                    : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-low'}"
                >
                  {type.slug} ({count})
                </button>
              {/if}
            {/each}
          </div>

          <div class="flex items-center gap-3">
            <span
              class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest"
              >Sort By</span
            >
            <select
              bind:value={sortBy}
              class="bg-surface-container border border-outline-variant/30 text-on-surface text-xs font-bold py-2 px-4 rounded-sm outline-hidden focus:border-primary transition-colors"
            >
              <option value="theme_num">Theme Number</option>
              <option value="score">Avg Score</option>
            </select>
          </div>
        </div>

        {#if filteredAndSortedSongs.length === 0}
          <div
            class="rounded-sm border border-outline-variant/30 bg-surface-container px-6 py-10 text-center"
          >
            <p class="text-sm font-medium text-on-surface-variant">
              No themes match this filter.
            </p>
          </div>
        {:else}
          <div class="space-y-5">
            {#each songGroups as group}
              {#if filterType === "all" && songGroups.length > 1}
                <h3
                  class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/80 flex items-center gap-2"
                >
                  <span class="w-1 h-3 bg-primary rounded-full"></span>
                  {group.label}
                  <span class="text-on-surface-variant/50">({group.songs.length})</span>
                </h3>
              {/if}

              <ul class="space-y-2">
                {#each group.songs as song}
                  {@const variants = song.variants?.length ? song.variants : [{ version_number: 1 }]}
                  {@const rating = formatThemeRating(song.average_rating)}
                  <li>
                    <article
                      class="rounded-sm border border-outline-variant/25 bg-surface-container hover:bg-surface-low transition-colors overflow-hidden"
                    >
                      <div class="px-4 pt-3 pb-1 flex items-start justify-between gap-3">
                        <div class="flex items-start gap-3 min-w-0 flex-1">
                          <span
                            class="inline-flex items-center justify-center px-2 py-0.5 rounded-sm border text-[10px] font-bold shrink-0 mt-0.5 {typeBadgeClass(song.type)}"
                          >
                            {song.type}{song.theme_num}
                          </span>
                          <div class="min-w-0 flex-1">
                            <a
                              href="/animes/{anime.slug}/{song.slug}"
                              class="font-bold text-on-surface text-sm leading-snug hover:text-primary transition-colors block truncate"
                              title={getSongName(song)}
                            >
                              {getSongName(song)}
                            </a>
                            <div class="mt-0.5">
                              <ArtistListCell
                                artists={song.artists}
                                popoverId="artists-{song.id}"
                              />
                            </div>
                          </div>
                        </div>

                        <div
                          class="shrink-0 text-right"
                          title={rating ? "Average rating" : "No ratings yet"}
                        >
                          {#if rating}
                            <div class="flex items-center justify-end gap-1">
                              <Star size={12} class="text-yellow-400 fill-current" />
                              <span class="font-bold text-on-surface text-sm leading-none">
                                {rating}
                              </span>
                            </div>
                          {:else}
                            <span class="text-[10px] font-bold uppercase tracking-wider text-on-surface-variant/50">
                              Unrated
                            </span>
                          {/if}
                        </div>
                      </div>

                      <div class="px-4 pb-2.5">
                        {#each variants as variant, variantIndex}
                          {@const video = song.variants?.length ? getPrimaryVideo(variant as SongVariant) : null}
                          {@const isLast = variantIndex === variants.length - 1}
                          <div
                            class="flex items-center justify-between gap-3 {variantIndex === 0 ? 'pt-0.5' : 'pt-2 mt-1 border-t border-outline-variant/10'} {isLast ? '' : 'pb-0.5'}"
                          >
                            <div class="flex items-center gap-2.5 min-w-0 text-on-surface-variant/80 pl-1">
                              <span class="text-[10px] font-bold uppercase tracking-wider shrink-0">
                                v{variant.version_number || 1}
                              </span>
                              {#if variant.episodes}
                                <span
                                  class="flex items-center gap-1 text-[11px] font-medium min-w-0"
                                  title="Episodes where this version appears"
                                >
                                  <Clapperboard
                                    size={12}
                                    class="shrink-0 text-on-surface-variant/70"
                                  />
                                  <span class="truncate">{variant.episodes}</span>
                                </span>
                              {/if}
                            </div>

                            <div
                              class="flex items-center gap-1.5 shrink-0 bg-secondary-container/25 border border-outline-variant/30 rounded-full pl-0.5 pr-1.5 py-0.5"
                            >
                              <a
                                href={songPlayHref(
                                  song.slug,
                                  variant.version_number || 1,
                                )}
                                class="w-7 h-7 rounded-full flex items-center justify-center text-white bg-primary hover:bg-primary-container transition-colors"
                                title="Play {getSongName(song)} v{variant.version_number || 1}"
                                aria-label="Play {getSongName(song)} version {variant.version_number || 1}"
                              >
                                <Play size={14} class="fill-current ml-0.5" />
                              </a>
                              {#if video?.resolution && formatResolution(video.resolution)}
                                <span
                                  class="text-[9px] font-black uppercase tracking-wider text-on-surface-variant px-0.5"
                                >
                                  {formatResolution(video.resolution)}
                                </span>
                              {/if}
                              {#if video?.is_nc}
                                <span
                                  class="text-[8px] font-black uppercase tracking-wider bg-surface-highest text-on-surface-variant px-1 py-0.5 rounded-sm"
                                >
                                  NC
                                </span>
                              {/if}
                              {#if video?.is_bd}
                                <span
                                  class="text-[8px] font-black uppercase tracking-wider bg-surface-highest text-on-surface-variant px-1 py-0.5 rounded-sm"
                                >
                                  BD
                                </span>
                              {/if}
                            </div>
                          </div>
                        {/each}
                      </div>
                    </article>
                  </li>
                {/each}
              </ul>
            {/each}
          </div>
        {/if}
      </div>
    </section>
  </div>
</main>
