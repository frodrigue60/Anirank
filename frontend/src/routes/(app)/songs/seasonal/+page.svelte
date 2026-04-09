<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import RankingList from "$lib/components/RankingList.svelte";
  import axios from "$lib/api";

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
    const url = new URL(page.url);
    url.searchParams.set("type", type);
    url.searchParams.set("page", "1");
    await goto(url.toString());
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

<svelte:head>
  <title
    >Seasonal Ranking - {currentSeason?.name}
    {currentYear?.name} - Anirank</title
  >
</svelte:head>

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
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all {activeType ===
            'all' || !activeType
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface/80 hover:text-on-surface'}"
        >
          All
        </button>
        <button
          onclick={() => changeType("OP")}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all {activeType ===
          'OP'
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface/80 hover:text-on-surface'}"
        >
          Openings
        </button>
        <button
          onclick={() => changeType("ED")}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all {activeType ===
          'ED'
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface/80 hover:text-on-surface'}"
        >
          Endings
        </button>
      </div>
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
