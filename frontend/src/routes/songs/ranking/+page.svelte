<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import RankingList from "$lib/components/RankingList.svelte";
  import axios from "$lib/api";

  let { data } = $props();

  let ranking = $state<{ data: any[]; next_page_url?: string } | null>(null);
  let activeType = $derived(data.type);
  let loading = $state(false);

  // Reset ranking when data changes (e.g., when type changes)
  $effect(() => {
    ranking = data.ranking?.songs;
  });

  async function changeType(type: string) {
    const url = new URL(page.url);
    url.searchParams.set("type", type);
    url.searchParams.set("page", "1");
    // Trigger SvelteKit navigation which will reload data
    await goto(url.toString());
  }

  async function loadMore() {
    if (loading || !ranking?.next_page_url) return;

    loading = true;
    try {
      const response = await axios.get(ranking.next_page_url);
      const newData = response.data;

      ranking = {
        ...newData,
        data: [...ranking.data, ...newData.data],
      };
    } catch (error) {
      console.error("Error loading more songs:", error);
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Global Ranking - Anirank Leaderboard</title>
</svelte:head>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div
    class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10"
  >
    <div>
      <h1 class="text-4xl font-black tracking-tight mb-2 uppercase italic">
        Global Leaderboard
      </h1>
    </div>
    <div class="flex flex-col gap-3">
      <div
        class="flex items-center p-1 bg-surface-darker border border-white/5 rounded-xl w-fit"
      >
        <button
          onclick={() => changeType("all")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeType ===
            'all' || !activeType
            ? 'bg-primary text-white shadow-lg'
            : 'text-white/40 hover:text-white'}"
        >
          All
        </button>
        <button
          onclick={() => changeType("OP")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeType ===
          'OP'
            ? 'bg-primary text-white shadow-lg'
            : 'text-white/40 hover:text-white'}"
        >
          Openings
        </button>
        <button
          onclick={() => changeType("ED")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeType ===
          'ED'
            ? 'bg-primary text-white shadow-lg'
            : 'text-white/40 hover:text-white'}"
        >
          Endings
        </button>
      </div>
    </div>
  </div>

  <RankingList
    songs={ranking?.data ?? []}
    {loading}
    hasMore={!!ranking?.next_page_url}
    onLoadMore={loadMore}
    startIndex={0}
  />
</main>
