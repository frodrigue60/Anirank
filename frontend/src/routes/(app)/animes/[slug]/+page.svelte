<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongName, getSongArtistNames, getFormattedScore } from "$lib/song-utils";
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
  import { PUBLIC_API_URL } from "$lib/api";

  let { data }: { data: PageData } = $props();
  let anime = $derived(data.data);

  function formatScore(score: number | string | null | undefined) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  let isExpanded = $state(false);
  let filterType = $state("all");
  let sortBy = $state("theme_num");

  let filteredAndSortedSongs = $derived.by(() => {
    let songs = anime.songs ? [...anime.songs] : [];

    // Filter
    if (filterType !== "all") {
      songs = songs.filter(
        (s) => s.type?.toLowerCase() === filterType.toLowerCase(),
      );
    }

    // Sort
    songs.sort((a, b) => {
      if (sortBy === "theme_num") {
        // Sort by type first based on DB order
        if (a.type !== b.type) {
          const order: Record<string, number> = {};
          configState.songTypes.forEach((st, i) => {
            order[st.slug] = i + 1;
          });
          return (order[a.type] || 99) - (order[b.type] || 99);
        }
        // Then natural numeric sort by theme_num
        return (a.theme_num || "").localeCompare(b.theme_num || "", undefined, {
          numeric: true,
        });
      } else {
        return (b.average_rating || 0) - (a.average_rating || 0);
      }
    });

    return songs;
  });
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
          <img
            alt="Cover art for {anime.title}"
            title="Cover art for {anime.title}"
            class="w-full h-auto aspect-2/3 object-cover transition-transform duration-700 group-hover:scale-105"
            src={anime.cover_url}
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
          <div class="flex flex-wrap items-center gap-1 bg-surface-container p-1 rounded-sm border border-white/5">
            <button
              onclick={() => (filterType = "all")}
              class="px-4 py-1.5 rounded-sm text-[10px] font-bold uppercase tracking-widest transition-all {filterType ===
              'all'
                ? 'bg-primary text-white shadow-lg shadow-primary/20'
                : 'text-on-surface-variant hover:text-on-surface'}"
            >
              all
            </button>
            {#each configState.songTypes as type}
              <button
                onclick={() => (filterType = type.slug)}
                class="px-4 py-1.5 rounded-sm text-[10px] font-bold uppercase tracking-widest transition-all {filterType ===
                type.slug
                  ? 'bg-primary text-white shadow-lg shadow-primary/20'
                  : 'text-on-surface-variant hover:text-on-surface'}"
              >
                {type.slug}
              </button>
            {/each}
          </div>

          <div class="flex items-center gap-3">
            <span
              class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest"
              >Sort By</span
            >
            <select
              bind:value={sortBy}
              class="bg-surface-container border border-white/5 text-on-surface text-xs font-bold py-2 px-4 rounded-sm outline-hidden focus:border-primary/50 transition-colors"
            >
              <option value="theme_num">Theme Number</option>
              <option value="score">Avg Score</option>
            </select>
          </div>
        </div>

        <div
          class="bg-surface-container border border-white/5 rounded-md overflow-hidden"
        >
          <table class="w-full text-left border-collapse">
            <thead>
              <tr
                class="border-b border-primary/10 text-xs uppercase tracking-widest text-on-surface-variant"
              >
                <th class="p-5 font-bold w-24">Type</th>
                <th class="p-5 font-bold">Song Title</th>
                <th class="p-5 font-bold">Artist</th>
                <th class="p-5 font-bold text-right">Avg Rating</th>
                <th class="p-5 font-bold w-16"></th>
              </tr>
            </thead>
            <tbody class="text-sm">
              {#each filteredAndSortedSongs as song}
                <tr
                  class="hover:bg-primary/10 border-b border-primary/10 group transition-colors"
                >
                  <td class="p-5">
                    <span
                      class="inline-flex items-center justify-center px-2.5 py-1 rounded-sm {song.type ===
                      'OP'
                        ? 'bg-green-500/10 text-green-400 border-green-500/40'
                        : 'bg-blue-500/10 text-blue-400 border-blue-500/40'} border text-[10px] font-bold"
                      >{song.type}{song.theme_num}</span
                    >
                  </td>
                  <td class="p-5">
                    <div
                      class="font-bold text-on-surface text-base hover:text-primary transition-colors"
                    >
                      <a href="/songs/{anime.slug}/{song.slug}"
                        >{getSongName(song)}</a
                      >
                    </div>
                  </td>
                  <td class="p-5 text-on-surface-variant">
                    {getSongArtistNames(song.artists)}
                  </td>
                  <td class="p-5 text-right">
                    <div class="flex items-center justify-end gap-1.5">
                      <Star size={14} class="text-yellow-400 fill-current" />

                      <span class="font-bold text-on-surface text-lg"
                        >{formatScore(song.average_rating)}</span
                      >
                    </div>
                  </td>
                  <td class="p-5 text-right">
                    <a
                      href="/songs/{anime.slug}/{song.slug}"
                      class="w-8 h-8 rounded-full flex items-center justify-center text-on-surface bg-primary/20 hover:bg-primary hover:text-white/50 transition-colors"
                      title="Play theme: {getSongName(song)}"
                    >
                      <Play size={18} class="fill-current" />

                    </a>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</main>
