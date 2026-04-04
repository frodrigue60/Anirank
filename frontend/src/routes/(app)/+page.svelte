<script lang="ts">
  import HeroIndex from "$lib/components/HeroIndex.svelte";
  import ActivityFeed from "$lib/components/ActivityFeed.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import AnnouncementSidebar from "$lib/components/AnnouncementSidebar.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { getSongName, getFormattedScore } from "$lib/song-utils";

  let { data } = $props();
  let homeData = $derived(data.homeData);
  let activeTab = $state("recently"); // 'recently', 'popular', 'viewed'

  let carouselContainer: HTMLElement | undefined = $state();

  function scrollLeft() {
    if (carouselContainer) {
      carouselContainer.scrollBy({ left: -300, behavior: "smooth" });
    }
  }

  function scrollRight() {
    if (carouselContainer) {
      carouselContainer.scrollBy({ left: 300, behavior: "smooth" });
    }
  }

  function changeTab(tab: string) {
    activeTab = tab;
    if (carouselContainer) {
      carouselContainer.scrollTo({ left: 0, behavior: "smooth" });
    }
  }

  function formatScore(score: string | number | undefined | null) {
    return getFormattedScore(score as any, authState.user?.score_format);
  }

  const PUBLIC_API_URL =
    import.meta.env.VITE_API_URL || "http://localhost:8080/api";
</script>

<SEO
  title="AniRank - Discover and Rate Anime Themes"
  description="Discover, rate, and explore the best anime openings and endings. Join our community of anime music enthusiasts and stay up-to-date with seasonal rankings."
  image={`${PUBLIC_API_URL}/og/home`}
/>

<!-- {#if !authState.isAuthenticated && !authState.loading}
  <HeroIndex />
{/if} -->

<main
  class="flex-1 w-full max-w-[1440px] mx-auto grid grid-cols-1 lg:grid-cols-12 gap-8 px-6 py-8"
>
  {#if !homeData}
    <!-- Page loading skeleton -->
    <div class="lg:col-span-12 h-96 flex items-center justify-center">
      <div
        class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"
      ></div>
    </div>
  {:else}
    <div class="lg:col-span-9 flex flex-col gap-10">
      <!-- Featured Theme -->
      {#if homeData.featured_song}
        <section
          class="group relative min-h-[400px] w-full overflow-hidden rounded-2xl bg-surface-container"
        >
          <div
            class="absolute inset-0 bg-cover bg-center opacity-40 transition-transform duration-700 group-hover:scale-105 group-hover:opacity-50"
            style="background-image: url('{homeData.featured_song.anime
              ?.banner_url ?? '/images/placeholders/default-banner.jpg'}');"
          ></div>
          <div
            class="absolute inset-0 bg-linear-to-t from-surface via-surface/40 to-transparent sm:bg-linear-to-r sm:from-surface sm:via-surface/80 sm:to-transparent"
          ></div>
          <div
            class="relative z-10 p-8 md:p-12 flex flex-col md:flex-row gap-8 items-center md:items-end w-full h-full"
          >
            <div class="relative shrink-0 shadow-2xl">
              <div
                class="relative h-48 w-48 overflow-hidden rounded-xl border border-outline-variant/10 shadow-[0_20px_50px_rgba(0,0,0,0.3)] md:h-64 md:w-64"
              >
                <img
                  alt="Cover art for {homeData.featured_song.anime?.title ||
                    'featured theme'}"
                  title="Cover art for {homeData.featured_song.anime?.title ||
                    'featured theme'}"
                  class="h-full w-full object-cover"
                  src={homeData.featured_song.anime?.cover_url ??
                    "/images/placeholders/default.jpg"}
                />
                <div
                  class="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/90 via-black/40 to-transparent p-4 pt-12"
                >
                  <div
                    class="flex items-center gap-1 text-lg font-bold text-yellow-400 drop-shadow-md"
                  >
                    <span class="material-symbols-outlined filled">star</span>
                    <span
                      >{formatScore(
                        homeData.featured_song.average_rating,
                      )}</span
                    >
                  </div>
                </div>
              </div>
            </div>
            <div
              class="flex flex-1 flex-col gap-4 pb-4 text-center md:text-left"
            >
              <div
                class="inline-flex items-center gap-2 self-center rounded-full bg-primary/10 px-3 py-1.5 text-primary shadow-lg shadow-primary/5 md:self-start"
              >
                <span class="material-symbols-outlined text-[16px] filled"
                  >auto_awesome</span
                >
                <span class="text-xs font-bold uppercase tracking-wider"
                  >Featured Theme</span
                >
              </div>
              <div class="flex flex-col gap-1">
                <h1
                  class="line-clamp-2 text-3xl font-black leading-tight tracking-tight text-on-surface sm:text-4xl md:text-5xl"
                >
                  {getSongName(homeData.featured_song)}
                </h1>
                <span class="font-medium text-on-surface-variant"
                  >{homeData.featured_song.artists
                    ?.map((a: any) => a.name)
                    .join(", ") ?? "Unknown Artist"}</span
                >
                <span class="font-bold text-primary"
                  >{homeData.featured_song.anime?.title}</span
                >
              </div>
              <div
                class="mt-4 flex items-center justify-center gap-4 md:justify-start"
              >
                <a
                  href="/songs/{homeData.featured_song.anime?.slug}/{homeData
                    .featured_song.slug}"
                  class="flex h-12 items-center justify-center gap-2 rounded-full bg-linear-to-br from-primary to-primary-container px-8 text-lg font-bold text-white transition-all hover:scale-105 hover:shadow-lg hover:shadow-primary/20 active:scale-95 sm:h-14 sm:px-10"
                  title="Play featured theme: {getSongName(
                    homeData.featured_song,
                  )}"
                  aria-label="Play featured theme"
                >
                  <span class="material-symbols-outlined filled text-[24px]"
                    >play_arrow</span
                  >
                  <span>Play Now</span>
                </a>
              </div>
            </div>
          </div>
        </section>
      {/if}

      <!-- Weekly Rankings -->
      <section>
        <div
          class="mb-6 flex flex-col gap-4 px-1 sm:flex-row sm:items-center justify-between"
        >
          <h2 class="flex items-center gap-2 text-2xl font-bold tracking-tight">
            <span class="material-symbols-outlined text-primary"
              >leaderboard</span
            >
            <a
              class="text-on-surface hover:text-primary transition-colors"
              href="/songs/ranking"
              title="View full song rankings"
              aria-label="View all rankings">Weekly Anime Songs Ranking</a
            >
          </h2>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Top Openings -->
          <div class="flex flex-col gap-4">
            <div class="mb-2 flex items-center justify-between">
              <span
                class="text-xs font-bold uppercase tracking-wider text-on-surface-variant/60"
                >Top Openings</span
              >
            </div>
            {#if homeData.weakly_ranking?.op?.length > 0}
              {#each homeData.weakly_ranking.op.slice(0, 3) as item, index}
                <a
                  href="/songs/{item.anime?.slug}/{item.slug}"
                  class="group relative flex items-center gap-4 rounded-xl bg-surface-low p-4 transition-all hover:bg-surface-container"
                  title="View details for {getSongName(item)}"
                >
                  <div
                    class="relative h-20 w-20 shrink-0 overflow-hidden rounded-lg"
                  >
                    <img
                      alt={getSongName(item)}
                      title={getSongName(item)}
                      class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
                      src={item.anime?.cover_url ??
                        "/images/placeholders/default.jpg"}
                    />
                    <div
                      class="absolute top-1 left-1 rounded bg-black/60 px-1.5 py-0.5 text-xs font-bold text-white shadow backdrop-blur-sm"
                    >
                      #{index + 1}
                    </div>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between">
                      <h3
                        class="truncate text-lg font-bold text-on-surface transition-colors group-hover:text-primary"
                      >
                        {getSongName(item)}
                      </h3>
                      <div
                        class="ml-2 flex shrink-0 items-center gap-1 rounded-full bg-surface-highest px-2 py-0.5 text-xs font-bold text-rating-star"
                      >
                        <span class="material-symbols-outlined filled"
                          >star</span
                        >
                        {formatScore(item.average_rating)}
                      </div>
                    </div>
                    <p class="truncate text-sm font-medium text-primary">
                      {item.anime?.title}
                    </p>
                    <p class="truncate text-xs text-on-surface-variant">
                      {item.artists?.map((a: any) => a.name).join(", ") ??
                        "Unknown Artist"}
                    </p>
                  </div>
                </a>
              {/each}
            {:else}
              <div
                class="flex h-32 flex-col items-center justify-center rounded-xl bg-surface-low text-on-surface-variant/40"
              >
                <span class="text-sm">No rankings available yet</span>
              </div>
            {/if}
          </div>
          <!-- Top Endings -->
          <div class="flex flex-col gap-4">
            <div class="mb-2 flex items-center justify-between">
              <span
                class="text-xs font-bold uppercase tracking-wider text-on-surface-variant/60"
                >Top Endings</span
              >
            </div>
            {#if homeData.weakly_ranking?.ed?.length > 0}
              {#each homeData.weakly_ranking.ed.slice(0, 3) as item, index}
                <a
                  href="/songs/{item.anime?.slug}/{item.slug}"
                  class="group relative flex items-center gap-4 rounded-xl bg-surface-low p-4 transition-all hover:bg-surface-container"
                  title="View details for {getSongName(item)}"
                >
                  <div
                    class="relative h-20 w-20 shrink-0 overflow-hidden rounded-lg"
                  >
                    <img
                      alt={getSongName(item)}
                      title={getSongName(item)}
                      class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
                      src={item.anime?.cover_url ??
                        "/images/placeholders/default.jpg"}
                    />
                    <div
                      class="absolute top-1 left-1 rounded bg-black/60 px-1.5 py-0.5 text-xs font-bold text-white shadow backdrop-blur-sm"
                    >
                      #{index + 1}
                    </div>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between">
                      <h3
                        class="truncate text-lg font-bold text-on-surface transition-colors group-hover:text-primary"
                      >
                        {getSongName(item)}
                      </h3>
                      <div
                        class="ml-2 flex shrink-0 items-center gap-1 rounded-full bg-surface-highest px-2 py-0.5 text-xs font-bold text-rating-star"
                      >
                        <span class="material-symbols-outlined filled"
                          >star</span
                        >
                        {formatScore(item.average_rating)}
                      </div>
                    </div>
                    <p class="truncate text-sm font-medium text-primary">
                      {item.anime?.title}
                    </p>
                    <p class="truncate text-xs text-on-surface-variant">
                      {item.artists?.map((a: any) => a.name).join(", ") ??
                        "Unknown Artist"}
                    </p>
                  </div>
                </a>
              {/each}
            {:else}
              <div
                class="flex h-32 flex-col items-center justify-center rounded-xl bg-surface-low text-on-surface-variant/40"
              >
                <span class="text-sm">No rankings available yet</span>
              </div>
            {/if}
          </div>
        </div>
      </section>

      <!-- Tabs Section (Carousels) -->
      <section class="">
        <div
          class="flex justify-between border-b border-outline-variant/10 mb-4"
        >
          <div class="flex gap-8">
            <button
              onclick={() => changeTab("recently")}
              class="pb-4 text-sm font-medium tracking-wide transition-colors border-b-2 {activeTab ===
              'recently'
                ? 'border-primary text-on-surface font-bold'
                : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            >
              New Releases
            </button>
            <button
              onclick={() => changeTab("popular")}
              class="pb-4 text-sm font-medium tracking-wide transition-colors border-b-2 {activeTab ===
              'popular'
                ? 'border-primary text-on-surface font-bold'
                : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            >
              Most Popular
            </button>
            <button
              onclick={() => changeTab("viewed")}
              class="pb-4 text-sm font-medium tracking-wide transition-colors border-b-2 {activeTab ===
              'viewed'
                ? 'border-primary text-on-surface font-bold'
                : 'border-transparent text-on-surface-variant hover:text-on-surface'}"
            >
              Most Viewed
            </button>
          </div>
          <div class="flex items-center gap-2">
            <button
              onclick={scrollLeft}
              class="group flex h-10 w-10 items-center justify-center rounded-full border border-outline-variant/10 bg-surface-highest/30 transition-all hover:border-primary/20 hover:bg-primary/10 active:scale-90"
              title="Scroll left"
              aria-label="Scroll carousel to the left"
            >
              <span
                class="material-symbols-outlined text-on-surface-variant/60 transition-colors group-hover:text-primary"
                >arrow_back</span
              >
            </button>
            <button
              onclick={scrollRight}
              class="group flex h-10 w-10 items-center justify-center rounded-full border border-outline-variant/10 bg-surface-highest/30 transition-all hover:border-primary/20 hover:bg-primary/10 active:scale-90"
              title="Scroll right"
              aria-label="Scroll carousel to the right"
            >
              <span
                class="material-symbols-outlined text-on-surface-variant/60 transition-colors group-hover:text-primary"
                >arrow_forward</span
              >
            </button>
          </div>
        </div>

        <div
          bind:this={carouselContainer}
          class="flex overflow-x-auto gap-4 snap-x snap-mandatory pb-4"
          style="scrollbar-width: none;"
        >
          <!-- recently added -->
          {#if activeTab === "recently" && homeData.recently_added}
            {#each homeData.recently_added.slice(0, 10) as song}
              <a
                href="/songs/{song.anime?.slug}/{song.slug}"
                class="group cursor-pointer shrink-0 w-[140px] sm:w-[160px] md:w-[180px] snap-start"
                title="View theme: {getSongName(song)}"
              >
                <div
                  class="aspect-2/3 relative mb-3 overflow-hidden rounded-lg bg-surface-low"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="h-full w-full object-cover transition-opacity group-hover:opacity-80"
                    src={song.anime?.cover_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 rounded border border-outline-variant/10 bg-black/60 px-1.5 py-0.5 text-[10px] font-bold uppercase text-white"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 flex items-center gap-0.5 rounded bg-black/60 px-1.5 py-0.5 text-xs font-bold text-yellow-400"
                  >
                    <span class="material-symbols-outlined filled">star</span>
                    {formatScore(song.average_rating)}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="truncate text-sm font-semibold text-on-surface transition-colors group-hover:text-primary"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="truncate text-xs text-on-surface-variant">
                    {song.anime?.title}
                  </p>
                </div>
              </a>
            {/each}
          {/if}
          <!-- most popular -->
          {#if activeTab === "popular" && homeData.most_popular}
            {#each homeData.most_popular.slice(0, 10) as song}
              <a
                href="/songs/{song.anime?.slug}/{song.slug}"
                class="group cursor-pointer shrink-0 w-[140px] sm:w-[160px] md:w-[180px] snap-start"
                title="View theme: {getSongName(song)}"
              >
                <div
                  class="aspect-2/3 relative mb-3 overflow-hidden rounded-lg bg-surface-low"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="h-full w-full object-cover transition-opacity group-hover:opacity-80"
                    src={song.anime?.cover_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 rounded border border-outline-variant/10 bg-black/60 px-1.5 py-0.5 text-[10px] font-bold uppercase text-white"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 flex items-center gap-0.5 rounded bg-black/60 px-1.5 py-0.5 text-xs font-bold text-red-500"
                  >
                    <span class="material-symbols-outlined filled"
                      >favorite</span
                    >
                    {song.favorites_count ?? "N/A"}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="truncate text-sm font-semibold text-on-surface transition-colors group-hover:text-primary"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="truncate text-xs text-on-surface-variant">
                    {song.anime?.title}
                  </p>
                </div>
              </a>
            {/each}
          {/if}
          <!-- most viewed -->
          {#if activeTab === "viewed" && homeData.most_viewed}
            {#each homeData.most_viewed.slice(0, 10) as song}
              <a
                href="/songs/{song.anime?.slug}/{song.slug}"
                class="group cursor-pointer shrink-0 w-[140px] sm:w-[160px] md:w-[180px] snap-start"
                title="View theme: {getSongName(song)}"
              >
                <div
                  class="aspect-2/3 relative mb-3 overflow-hidden rounded-lg bg-surface-low"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="h-full w-full object-cover transition-opacity group-hover:opacity-80"
                    src={song.anime?.cover_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 rounded border border-outline-variant/10 bg-black/60 px-1.5 py-0.5 text-[10px] font-bold uppercase text-white"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 flex items-center gap-0.5 rounded bg-black/60 px-1.5 py-0.5 text-xs font-bold text-blue-500"
                  >
                    <span class="material-symbols-outlined filled"
                      >visibility</span
                    >
                    {song.views ?? "N/A"}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="truncate text-sm font-semibold text-on-surface transition-colors group-hover:text-primary"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="truncate text-xs text-on-surface-variant">
                    {song.anime?.title}
                  </p>
                </div>
              </a>
            {/each}
          {/if}
        </div>
      </section>

      <!-- Activity Feed -->
      <ActivityFeed recentOnly />
    </div>

    <!-- Right Column (Sidebar) -->
    <aside class="lg:col-span-3 flex flex-col gap-8">
      <!-- Featured Artists -->
      <div
        class="rounded-2xl bg-surface-low p-6 border border-outline-variant/10"
      >
        <div class="mb-6 flex items-center justify-between">
          <h3 class="text-lg font-bold text-on-surface">Featured Artists</h3>
        </div>
        {#if homeData.featured_artists?.length > 0}
          <div class="flex flex-col gap-1">
            {#each homeData.featured_artists as artist}
              <a
                href="/artists/{artist.slug}"
                class="group -mx-3 flex items-center justify-between rounded-xl px-3 py-2.5 transition-all hover:bg-surface-container"
                title="View artist profile: {artist.name}"
              >
                <div class="flex min-w-0 items-center gap-3">
                  <div
                    class="relative flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-full border border-outline-variant/10 bg-surface-container transition-colors group-hover:border-primary/50"
                  >
                    <img
                      alt={artist.name}
                      title={artist.name}
                      class="h-full w-full object-cover transition-transform group-hover:scale-110"
                      src={artist.avatar_url ||
                        "/images/placeholders/default.jpg"}
                    />
                  </div>
                  <div class="flex min-w-0 flex-col">
                    <h4
                      class="truncate text-sm font-bold text-on-surface transition-colors group-hover:text-primary"
                    >
                      {artist.name}
                    </h4>
                    <span
                      class="text-[10px] font-medium uppercase tracking-widest text-on-surface-variant/60"
                    >
                      {artist.enabled_songs || 0} Themes
                    </span>
                  </div>
                </div>
                <span
                  class="material-symbols-outlined translate-x-2 text-primary opacity-0 transition-all group-hover:translate-x-0 group-hover:opacity-100"
                >
                  chevron_right
                </span>
              </a>
            {/each}
          </div>
        {:else}
          <div class="text-sm text-on-surface-variant/40">
            No artists featured.
          </div>
        {/if}
      </div>
      <!-- Announcements -->
      <AnnouncementSidebar />
    </aside>
  {/if}
</main>
