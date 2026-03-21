<script lang="ts">
  import type { PageData } from "./$types";
  import { getSongArtistNames, getFormattedScore } from "$lib/song-utils";
  import SEO from "$lib/components/SEO.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { page } from "$app/state";

  let { data }: { data: PageData } = $props();
  let anime = $derived(data.anime);

  function formatScore(score: number | string | null | undefined) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  let isExpanded = $state(false);
</script>

<SEO 
  title={anime.title}
  description={anime.description || ""}
  image={`${page.url.origin}/api/og/anime/${anime.slug}`}
  type="video.tv_show"
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start">
    <!-- Sidebar -->
    <aside class="lg:col-span-3 space-y-8">
      <div
        class="relative rounded-xl overflow-hidden shadow-2xl shadow-primary/10 border border-white/5 group"
      >
        <img
          alt={anime.title}
          class="w-full h-auto aspect-2/3 object-cover transition-transform duration-700 group-hover:scale-105"
          src={anime.cover_url}
        />
        <div
          class="absolute top-0 left-0 w-full h-full bg-linear-to-t from-background-dark/80 via-transparent to-transparent opacity-60"
        ></div>
        <!-- <div class="absolute bottom-4 left-4 right-4">
                    <button
                        class="w-full bg-primary hover:bg-primary-dark text-white font-bold py-2.5 rounded-lg transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 text-sm"
                    >
                        <span class="material-symbols-outlined text-lg"
                            >add</span
                        >
                        Add to List
                    </button>
                </div> -->
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
          <div class="flex items-center gap-2">
            <span class="material-symbols-outlined text-white/40">link</span>
            <a
              href="https://anilist.co/anime/{anime.anilist_id}"
              target="_blank"
              class="text-white/70 hover:text-white transition-colors"
              >View on Anilist</a
            >
          </div>
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-white/40 mb-0.5">Format</span>
            <span class="font-medium text-white"
              >{anime.format?.name || "Unknown"}</span
            >
          </div>
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-white/40 mb-0.5">Season</span>
            <span class="font-medium text-white"
              >{anime.season?.name || "Unknown"}
              {anime.year?.name || "Unknown"}</span
            >
          </div>

          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-white/40 mb-0.5">Studio</span>
            <div class="flex flex-wrap gap-1">
              {#if anime.studios?.length > 0}
                {#each anime.studios as studio, index}
                  <a
                    class="font-medium text-primary hover:underline"
                    href="/studios/{studio.slug}">{studio.name}</a
                  >
                  {index < anime.studios.length - 1 ? ", " : ""}
                {/each}
              {:else}
                <span class="font-medium text-white/60">Unknown</span>
              {/if}
            </div>
          </div>
          <div
            class="border-l-2 border-transparent hover:border-primary pl-3 transition-all duration-200"
          >
            <span class="block text-xs text-white/40 mb-0.5">Producers</span>
            <span class="font-medium text-white/80">
              {#if anime.producers?.length > 0}
                {#each anime.producers as producer, index}
                  <a
                    class="font-medium text-primary hover:underline"
                    href="/producers/{producer.slug}">{producer.name}</a
                  >
                  {index < anime.producers.length - 1 ? ", " : ""}
                {/each}
              {:else}
                <span class="font-medium text-white/60">Unknown</span>
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
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg bg-surface-dark hover:bg-surface-darker border border-white/5 hover:border-primary/30 transition-all text-sm group"
                href={link.url}
                target="_blank"
              >
                <span
                  class="material-symbols-outlined text-white/40 group-hover:text-primary transition-colors text-lg"
                  >language</span
                >
                <span class="text-white/70 group-hover:text-white"
                  >{link.name}</span
                >
                <span
                  class="material-symbols-outlined text-white/20 text-sm ml-auto"
                  >open_in_new</span
                >
              </a>
            {/each}
          </div>
        </div>
      {/if}
    </aside>

    <!-- Main Content -->
    <section class="lg:col-span-9 space-y-10">
      <div class="flex flex-col gap-6">
        <div
          class="flex flex-col md:flex-row md:items-end justify-between gap-4"
        >
          <div>
            <h1
              class="text-4xl md:text-6xl font-black tracking-tight text-white mb-2"
            >
              {anime.title}
            </h1>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          {#each anime.genres as genre}
            <span
              class="px-3 py-1 rounded border border-white/10 text-white/60 text-xs font-medium bg-surface-dark"
              >{genre.name}</span
            >
          {/each}
        </div>
      </div>

      <div class="space-y-4">
        <h2 class="text-xl font-bold flex items-center gap-3 text-white">
          <span class="w-1.5 h-6 bg-primary rounded-full"></span>
          Synopsis
        </h2>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="bg-[rgba(30,22,37,0.6)] backdrop-blur-lg border border-[rgba(127,19,236,0.15)] p-8 rounded-2xl cursor-pointer hover:bg-[rgba(40,30,50,0.7)] transition-all group relative overflow-hidden"
          onclick={() => (isExpanded = !isExpanded)}
        >
          <p
            class="text-white/80 leading-relaxed text-lg font-body transition-all duration-300 {isExpanded
              ? ''
              : 'line-clamp-4'}"
          >
            {@html anime.description || "No synopsis available."}
          </p>

          {#if !isExpanded && anime.description && anime.description.length > 200}
            <div
              class="absolute bottom-0 left-0 w-full h-16 bg-linear-to-t from-[rgba(30,22,37,0.95)] via-[rgba(30,22,37,0.7)] to-transparent flex items-end justify-center pb-2 transition-opacity"
            >
              <span
                class="text-primary font-bold text-xs uppercase tracking-widest flex items-center gap-1"
              >
                Read More
                <span class="material-symbols-outlined text-sm"
                  >expand_more</span
                >
              </span>
            </div>
          {/if}

          {#if isExpanded}
            <div class="mt-4 flex justify-center">
              <span
                class="text-primary/50 font-bold text-xs uppercase tracking-widest flex items-center gap-1 hover:text-primary transition-colors"
              >
                Show Less
                <span class="material-symbols-outlined text-sm"
                  >expand_less</span
                >
              </span>
            </div>
          {/if}
        </div>
      </div>

      <!-- Music Themes -->
      <div class="space-y-6 pt-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold flex items-center gap-3 text-white">
            <span class="w-1.5 h-6 bg-primary rounded-full"></span>
            Music Themes
          </h2>
        </div>
        <div
          class="bg-[rgba(30,22,37,0.6)] backdrop-blur-lg border border-[rgba(127,19,236,0.15)] rounded-2xl overflow-hidden"
        >
          <table class="w-full text-left border-collapse">
            <thead>
              <tr
                class="border-b border-white/5 text-xs uppercase tracking-widest text-white/30"
              >
                <th class="p-5 font-bold w-24">Type</th>
                <th class="p-5 font-bold">Song Title</th>
                <th class="p-5 font-bold">Artist</th>
                <th class="p-5 font-bold text-right">Avg Rating</th>
                <th class="p-5 font-bold w-16"></th>
              </tr>
            </thead>
            <tbody class="text-sm">
              {#each anime.songs as song}
                <tr
                  class="hover:bg-[rgba(127,19,236,0.05)] border-b border-white/5 group transition-colors"
                >
                  <td class="p-5">
                    <span
                      class="inline-flex items-center justify-center px-2.5 py-1 rounded {song.type ===
                      'OP'
                        ? 'bg-green-500/10 text-green-400 border-green-500/20'
                        : 'bg-blue-500/10 text-blue-400 border-blue-500/20'} border text-[10px] font-bold"
                      >{song.type}{song.theme_num}</span
                    >
                  </td>
                  <td class="p-5">
                    <div class="font-bold text-white text-base">
                      {song.song_romaji ||
                        song.song_en ||
                        song.song_jp ||
                        "N/A"}
                    </div>
                  </td>
                  <td class="p-5 text-white/70">
                    {getSongArtistNames(song.artists)}
                  </td>
                  <td class="p-5 text-right">
                    <div class="flex items-center justify-end gap-1.5">
                      <span
                        class="material-symbols-outlined text-yellow-400 text-sm filled"
                        >star</span
                      >
                      <span class="font-bold text-white text-lg"
                        >{formatScore(song.average_rating)}</span
                      >
                    </div>
                  </td>
                  <td class="p-5 text-right">
                    <a
                      href="/songs/{anime.slug}/{song.slug}"
                      class="w-8 h-8 rounded-full flex items-center justify-center text-white bg-primary hover:bg-primary/20 hover:text-white transition-colors"
                    >
                      <span class="material-symbols-outlined text-lg"
                        >play_arrow</span
                      >
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
