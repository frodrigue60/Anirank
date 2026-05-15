<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import RankingList from "$lib/components/RankingList.svelte";
  import axios from "$lib/api";
  import { configState } from "$lib/state/config.svelte";
  import { PUBLIC_API_URL } from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let allSongs = $state<any[]>([]);
  let paginationMeta = $state<{
    next?: string | null;
    current_page: number;
    per_page: number;
  } | null>(null);
  let currentSeason = $derived(data.songsData?.current_season);
  let currentYear = $derived(data.songsData?.current_year);
  let activeType = $derived(data.type);
  let loading = $state(false);

  // Rate Limiting State
  let clickCount = $state(0);
  let lastClickTime = $state(0);
  let isRateLimited = $state(false);

  // Reset when data changes (e.g., type filter changed via SvelteKit navigation)
  $effect(() => {
    const songsData = data.songsData;
    if (songsData) {
      allSongs = songsData.data || [];
      paginationMeta = {
        next: songsData.links?.next,
        current_page: songsData.pagination?.current_page || 1,
        per_page: songsData.pagination?.per_page || 24,
      };
    } else {
      allSongs = [];
      paginationMeta = null;
    }
  });

  async function changeType(type: string) {
    if (loading || activeType === type || (type === "all" && !activeType) || isRateLimited) return;
    
    // Rate Limiting (10 clicks / 5 seconds)
    const now = Date.now();
    if (now - lastClickTime < 5000) {
      clickCount++;
      if (clickCount >= 10) {
        isRateLimited = true;
        setTimeout(() => {
          isRateLimited = false;
          clickCount = 0;
        }, 30000);
        return;
      }
    } else {
      clickCount = 1;
    }
    lastClickTime = now;

    const url = new URL(page.url);
    url.searchParams.set("type", type);
    url.searchParams.set("page", "1");
    
    loading = true;
    try {
      await goto(url.toString(), { keepFocus: true });
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (loading || !paginationMeta?.next) return;

    loading = true;
    try {
      const response = await axios.get(paginationMeta.next);
      const newSongs = response.data;
      if (newSongs?.data) {
        allSongs = [...allSongs, ...newSongs.data];
        paginationMeta = {
          next: newSongs.links?.next,
          current_page:
            newSongs.pagination?.current_page ||
            (paginationMeta?.current_page || 1) + 1,
          per_page: newSongs.pagination?.per_page || 24,
        };
      }
    } catch (error) {
      console.error("Error loading more songs:", error);
    } finally {
      loading = false;
    }
  }
</script>

<SEO 
  title={`${currentSeason?.name} ${currentYear?.name} Anime Song Ranking`}
  description={`Explore and rate the best anime openings and endings from the ${currentSeason?.name} ${currentYear?.name} season. Stay up to date with the latest hits and community favorites.`}
  image={`${PUBLIC_API_URL}/og/ranking/seasonal`}
  keywords={`${currentSeason?.name} anime songs, ${currentYear?.name} anime openings, seasonal anime music ranking, best anime themes ${currentYear?.name}`}
  jsonLd={{
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    "itemListElement": [
      { "@type": "ListItem", "position": 1, "name": "Home", "item": "https://anirank.work/" },
      { "@type": "ListItem", "position": 2, "name": "Seasonal Ranking", "item": "https://anirank.work/songs/seasonal" }
    ]
  }}
/>



<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div
    class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10"
  >
    <div>
      <h1
        class="text-4xl font-black tracking-tight mb-2 uppercase italic text-on-surface"
      >
        {currentSeason?.name ?? "Current"}
        {currentYear?.name} Ranking
      </h1>
    </div>
    <div class="flex flex-col gap-3">
      <div
        class="flex items-center p-1 bg-surface-highest border border-white/5 rounded-md w-fit"
      >
        <button
          onclick={() => changeType("all")}
          disabled={loading || isRateLimited}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed {activeType ===
            'all' || !activeType
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface/80 hover:text-on-surface'}"
        >
          All
        </button>
        {#each configState.songTypes as type}
          <button
            onclick={() => changeType(type.slug)}
            disabled={loading || isRateLimited}
            class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed {activeType === type.slug
              ? 'bg-primary text-white shadow-lg'
              : 'text-on-surface/80 hover:text-on-surface'}"
          >
            {type.name}s
          </button>
        {/each}
      </div>

      {#if isRateLimited}
        <p class="text-[10px] text-red-500 font-bold uppercase tracking-widest animate-pulse mt-1">
          Too many requests. Please wait 30s.
        </p>
      {/if}
    </div>
  </div>

  <RankingList
    songs={allSongs}
    {loading}
    hasMore={!!paginationMeta?.next}
    onLoadMore={loadMore}
    startIndex={0}
    rankingType="seasonal"
  />
</main>
