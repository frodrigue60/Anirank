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
          class="relative w-full min-h-[400px] rounded-2xl overflow-hidden bg-surface-dark group"
        >
          <div
            class="absolute inset-0 bg-cover bg-center opacity-40 transition-transform duration-700 group-hover:scale-105 group-hover:opacity-50"
            style="background-image: url('{homeData.featured_song.anime
              ?.banner_url ?? '/images/placeholders/default-banner.jpg'}');"
          ></div>
          <div
            class="absolute inset-0 bg-linear-to-t from-background-dark via-background-dark/80 to-transparent sm:bg-linear-to-r"
          ></div>
          <div
            class="relative z-10 p-8 md:p-12 flex flex-col md:flex-row gap-8 items-center md:items-end w-full h-full"
          >
            <div class="relative shrink-0 hero-glow">
              <div
                class="w-48 h-48 md:w-64 md:h-64 rounded-xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden relative border border-white/10"
              >
                <img
                  alt="Cover art for {homeData.featured_song.anime?.title ||
                    'featured theme'}"
                  title="Cover art for {homeData.featured_song.anime?.title ||
                    'featured theme'}"
                  class="w-full h-full object-cover"
                  src={homeData.featured_song.anime?.thumbnail_url ??
                    "/images/placeholders/default.jpg"}
                />
                <div
                  class="absolute bottom-0 inset-x-0 bg-linear-to-t from-black/90 via-black/40 to-transparent p-4 pt-12"
                >
                  <div
                    class="flex items-center gap-1 text-yellow-400 font-bold text-lg drop-shadow-md"
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
              class="flex flex-col gap-4 text-center md:text-left flex-1 pb-4"
            >
              <div
                class="inline-flex items-center gap-2 self-center md:self-start bg-primary/20 text-primary border border-primary/30 px-3 py-1.5 rounded-full backdrop-blur-md shadow-lg shadow-primary/10"
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
                  class="text-3xl sm:text-4xl md:text-5xl font-black leading-tight tracking-tight text-white drop-shadow-lg line-clamp-2"
                >
                  {getSongName(homeData.featured_song)}
                </h1>
                <span class="font-medium text-white/90"
                  >{homeData.featured_song.artists
                    ?.map((a: any) => a.name)
                    .join(", ") ?? "Unknown Artist"}</span
                >
                <span class="font-bold text-primary"
                  >{homeData.featured_song.anime?.title}</span
                >
              </div>
              <div
                class="flex items-center gap-4 mt-4 justify-center md:justify-start"
              >
                <a
                  href="/songs/{homeData.featured_song.anime?.slug}/{homeData
                    .featured_song.slug}"
                  class="bg-primary hover:bg-primary-light text-white h-12 sm:h-14 px-8 sm:px-10 rounded-full font-bold flex items-center justify-center gap-2 transition-transform hover:scale-105 active:scale-95 shadow-[0_0_20px_rgba(127,19,236,0.3)]"
                  title="Play featured theme: {getSongName(
                    homeData.featured_song,
                  )}"
                  aria-label="Play featured theme"
                >
                  <span class="material-symbols-outlined filled text-[24px]"
                    >play_arrow</span
                  >
                  <span class="text-lg">Play Now</span>
                </a>
              </div>
            </div>
          </div>
        </section>
      {/if}

      <!-- Weekly Rankings -->
      <section>
        <div
          class="flex flex-col sm:flex-row sm:items-center justify-between mb-6 px-1 gap-4"
        >
          <h2 class="text-2xl font-bold tracking-tight flex items-center gap-2">
            <span class="material-symbols-outlined text-primary"
              >leaderboard</span
            >
            Weekly Rankings
          </h2>
          <a
            class="text-primary text-md hover:text-white/80 flex items-center gap-1"
            href="/songs/ranking"
            title="View full song rankings"
            aria-label="View all rankings"
            >View All <span class="material-symbols-outlined text-md"
              >arrow_forward</span
            ></a
          >
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Top Openings -->
          <div class="flex flex-col gap-4">
            <div class="flex items-center justify-between mb-2">
              <span
                class="text-white/50 text-xs font-bold uppercase tracking-wider"
                >Top Openings</span
              >
            </div>
            {#if homeData.weakly_ranking?.op?.length > 0}
              {#each homeData.weakly_ranking.op.slice(0, 3) as item, index}
                <a
                  href="/songs/{item.anime?.slug}/{item.slug}"
                  class="group relative bg-surface-darker p-4 rounded-xl hover:bg-surface-dark transition-colors border border-white/5 flex gap-4 items-center"
                  title="View details for {getSongName(item)}"
                >
                  <div
                    class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden"
                  >
                    <img
                      alt={getSongName(item)}
                      title={getSongName(item)}
                      class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                      src={item.anime?.thumbnail_url ??
                        "/images/placeholders/default.jpg"}
                    />
                    <div
                      class="absolute top-1 left-1 bg-primary text-white text-xs font-bold px-1.5 py-0.5 rounded shadow"
                    >
                      #{index + 1}
                    </div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between">
                      <h3
                        class="font-bold text-white truncate text-lg group-hover:text-primary transition-colors"
                      >
                        {getSongName(item)}
                      </h3>
                      <div
                        class="flex items-center text-xs gap-1 bg-surface-dark px-2 py-0.5 rounded text-yellow-400 font-bold shrink-0 ml-2"
                      >
                        <span class="material-symbols-outlined filled"
                          >star</span
                        >
                        {formatScore(item.average_rating)}
                      </div>
                    </div>
                    <p class="text-sm text-primary font-medium truncate">
                      {item.anime?.title}
                    </p>
                    <p class="text-xs text-white/50 truncate">
                      {item.artists?.map((a: any) => a.name).join(", ") ??
                        "Unknown Artist"}
                    </p>
                  </div>
                </a>
              {/each}
            {:else}
              <div
                class="flex h-32 flex-col items-center justify-center rounded-xl border border-dashed border-white/10 bg-surface-darker text-white/40"
              >
                <span class="text-sm">No rankings available yet</span>
              </div>
            {/if}
          </div>
          <!-- Top Endings -->
          <div class="flex flex-col gap-4">
            <div class="flex items-center justify-between mb-2">
              <span
                class="text-white/50 text-xs font-bold uppercase tracking-wider"
                >Top Endings</span
              >
            </div>
            {#if homeData.weakly_ranking?.ed?.length > 0}
              {#each homeData.weakly_ranking.ed.slice(0, 3) as item, index}
                <a
                  href="/songs/{item.anime?.slug}/{item.slug}"
                  class="group relative bg-surface-darker p-4 rounded-xl hover:bg-surface-dark transition-colors border border-white/5 flex gap-4 items-center"
                  title="View details for {getSongName(item)}"
                >
                  <div
                    class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden"
                  >
                    <img
                      alt={getSongName(item)}
                      title={getSongName(item)}
                      class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                      src={item.anime?.thumbnail_url ??
                        "/images/placeholders/default.jpg"}
                    />
                    <div
                      class="absolute top-1 left-1 bg-surface-dark text-white text-xs font-bold px-1.5 py-0.5 rounded shadow border border-white/10"
                    >
                      #{index + 1}
                    </div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between">
                      <h3
                        class="font-bold text-white truncate text-lg group-hover:text-primary transition-colors"
                      >
                        {getSongName(item)}
                      </h3>
                      <div
                        class="flex items-center gap-1 bg-surface-dark px-2 py-0.5 rounded text-yellow-400 text-xs font-bold shrink-0 ml-2"
                      >
                        <span
                          class="material-symbols-outlined filled text-[14px]"
                          >star</span
                        >
                        {formatScore(item.average_rating)}
                      </div>
                    </div>
                    <p class="text-sm text-primary font-medium truncate">
                      {item.anime?.title}
                    </p>
                    <p class="text-xs text-white/50 truncate">
                      {item.artists?.map((a: any) => a.name).join(", ") ??
                        "Unknown Artist"}
                    </p>
                  </div>
                </a>
              {/each}
            {:else}
              <div
                class="flex h-32 flex-col items-center justify-center rounded-xl border border-dashed border-white/10 bg-surface-darker text-white/40"
              >
                <span class="text-sm">No rankings available yet</span>
              </div>
            {/if}
          </div>
        </div>
      </section>

      <!-- Tabs Section (Carousels) -->
      <section class="pb-12">
        <div class="border-b border-white/5 mb-6 flex justify-between">
          <div class="flex gap-8">
            <button
              onclick={() => changeTab("recently")}
              class="pb-4 border-b-2 font-medium text-sm tracking-wide transition-colors {activeTab ===
              'recently'
                ? 'border-primary text-white font-bold'
                : 'border-transparent text-white/40 hover:text-white'}"
            >
              New Releases
            </button>
            <button
              onclick={() => changeTab("popular")}
              class="pb-4 border-b-2 font-medium text-sm tracking-wide transition-colors {activeTab ===
              'popular'
                ? 'border-primary text-white font-bold'
                : 'border-transparent text-white/40 hover:text-white'}"
            >
              Most Popular
            </button>
            <button
              onclick={() => changeTab("viewed")}
              class="pb-4 border-b-2 font-medium text-sm tracking-wide transition-colors {activeTab ===
              'viewed'
                ? 'border-primary text-white font-bold'
                : 'border-transparent text-white/40 hover:text-white'}"
            >
              Most Viewed
            </button>
          </div>
          <div class="flex items-center gap-2">
            <button
              onclick={scrollLeft}
              class="p-1 rounded-md border border-white/10 hover:bg-white/10 transition-colors flex items-center justify-center"
              title="Scroll left"
              aria-label="Scroll carousel to the left"
            >
              <span class="material-symbols-outlined text-md">arrow_back</span>
            </button>
            <button
              onclick={scrollRight}
              class="p-1 rounded-md border border-white/10 hover:bg-white/10 transition-colors flex items-center justify-center"
              title="Scroll right"
              aria-label="Scroll carousel to the right"
            >
              <span class="material-symbols-outlined text-md"
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
                  class="aspect-2/3 rounded-lg overflow-hidden mb-3 relative bg-surface-darker"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="w-full h-full object-cover group-hover:opacity-80 transition-opacity"
                    src={song.anime?.thumbnail_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 bg-black/70 backdrop-blur-sm px-1.5 py-0.5 rounded text-[10px] font-bold text-white border border-white/10 uppercase"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 bg-black/70 px-1.5 py-0.5 rounded text-xs font-bold text-yellow-400 flex items-center gap-0.5"
                  >
                    <span class="material-symbols-outlined filled">star</span>
                    {formatScore(song.average_rating)}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="text-white font-semibold text-sm truncate group-hover:text-primary transition-colors"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="text-white/40 text-xs truncate">
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
                  class="aspect-2/3 rounded-lg overflow-hidden mb-3 relative bg-surface-darker"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="w-full h-full object-cover group-hover:opacity-80 transition-opacity"
                    src={song.anime?.thumbnail_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 bg-black/70 backdrop-blur-sm px-1.5 py-0.5 rounded text-[10px] font-bold text-white border border-white/10 uppercase"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 bg-black/70 px-1.5 py-0.5 rounded text-xs font-bold text-red-400 flex items-center gap-0.5"
                  >
                    <span class="material-symbols-outlined filled"
                      >favorite</span
                    >
                    {song.favorites_count ?? "N/A"}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="text-white font-semibold text-sm truncate group-hover:text-primary transition-colors"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="text-white/40 text-xs truncate">
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
                  class="aspect-2/3 rounded-lg overflow-hidden mb-3 relative bg-surface-darker"
                >
                  <img
                    alt={getSongName(song)}
                    title={getSongName(song)}
                    class="w-full h-full object-cover group-hover:opacity-80 transition-opacity"
                    src={song.anime?.thumbnail_url ??
                      "/images/placeholders/default.jpg"}
                  />
                  <div
                    class="absolute top-2 right-2 bg-black/70 backdrop-blur-sm px-1.5 py-0.5 rounded text-[10px] font-bold text-white border border-white/10 uppercase"
                  >
                    {song.type ?? "type n/a"}
                  </div>
                  <div
                    class="absolute bottom-2 left-2 bg-black/70 px-1.5 py-0.5 rounded text-xs font-bold text-blue-400 flex items-center gap-0.5"
                  >
                    <span class="material-symbols-outlined filled"
                      >visibility</span
                    >
                    {song.views ?? "N/A"}
                  </div>
                </div>
                <div class="flex flex-col">
                  <h4
                    class="text-white font-semibold text-sm truncate group-hover:text-primary transition-colors"
                  >
                    {getSongName(song)}
                  </h4>
                  <p class="text-white/40 text-xs truncate">
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
      <div class="bg-surface-darker rounded-2xl p-6 border border-white/5">
        <div class="flex items-center justify-between mb-6">
          <h3 class="font-bold text-white text-lg">Featured Artists</h3>
          <!-- <a
                        class="text-primary text-xs font-bold uppercase tracking-wide hover:underline"
                        href="/artists">View All</a
                    > -->
        </div>
        {#if homeData.featured_artists?.length > 0}
          <div class="flex flex-col gap-5">
            {#each homeData.featured_artists as artist}
              <a
                href="/artists/{artist.slug}"
                class="flex items-center justify-between group"
                title="View artist profile: {artist.name}"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-12 h-12 rounded-full overflow-hidden border-2 border-transparent group-hover:border-primary transition-colors flex items-center justify-center bg-primary/20 text-primary"
                  >
                    {#if artist.avatar_url}
                      <img
                        alt={artist.name}
                        title={artist.name}
                        class="w-full h-full object-cover"
                        src={artist.avatar_url}
                      />
                    {:else}
                      <span class="material-symbols-outlined">person</span>
                    {/if}
                  </div>
                  <div>
                    <h4
                      class="text-sm font-bold text-white group-hover:text-primary transition-colors"
                    >
                      {artist.name}
                    </h4>
                  </div>
                </div>
                <button
                  class="w-8 h-8 rounded-full bg-white/5 hover:bg-white/10 flex items-center justify-center text-primary transition-colors shrink-0"
                  title="View artist profile"
                  aria-label="Go to {artist.name}'s profile"
                >
                  <span class="material-symbols-outlined text-[18px]"
                    >arrow_forward</span
                  >
                </button>
              </a>
            {/each}
          </div>
        {:else}
          <div class="text-sm text-white/40">No artists featured.</div>
        {/if}
      </div>
      <!-- Announcements -->
      <AnnouncementSidebar />
    </aside>
  {/if}
</main>
