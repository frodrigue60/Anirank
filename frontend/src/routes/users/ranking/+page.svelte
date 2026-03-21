<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import UserRankingList from "$lib/components/UserRankingList.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import axios from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  let ranking = $state(data.ranking);
  let activeSort = $derived(data.sort);
  let loading = $state(false);

  $effect(() => {
    ranking = data.ranking;
  });

  async function changeSort(sort: string) {
    const url = new URL(page.url);
    url.searchParams.set("sort", sort);
    url.searchParams.set("page", "1");
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
      console.error("Error loading more users:", error);
    } finally {
      loading = false;
    }
  }
</script>

<SEO 
  title="User Leaderboard" 
  description="Check out the most active members of the AniRank community based on XP, ratings, and social interactions." 
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
    <div>
      <h1 class="text-4xl font-black tracking-tight mb-2 uppercase italic">
        User Leaderboard
      </h1>
      <p class="text-white/40 font-bold uppercase tracking-widest text-xs">
        The most active members of our community
      </p>
    </div>
    
    <div class="flex flex-col gap-3">
      <div class="flex items-center p-1 bg-surface-darker border border-white/5 rounded-xl w-fit">
        <button
          onclick={() => changeSort("xp")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeSort === 'xp' ? 'bg-primary text-white shadow-lg' : 'text-white/40 hover:text-white'}"
        >
          Top XP
        </button>
        <button
          onclick={() => changeSort("ratings")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeSort === 'ratings' ? 'bg-primary text-white shadow-lg' : 'text-white/40 hover:text-white'}"
        >
          Top Ratings
        </button>
        <button
          onclick={() => changeSort("comments")}
          class="px-6 py-2.5 rounded-lg font-bold text-sm transition-all {activeSort === 'comments' ? 'bg-primary text-white shadow-lg' : 'text-white/40 hover:text-white'}"
        >
          Top Social
        </button>
      </div>
    </div>
  </div>

  <UserRankingList
    users={ranking?.data ?? []}
    startIndex={0}
    sort={activeSort}
  />

  <InfiniteScroll
    hasMore={!!ranking?.next_page_url}
    {loading}
    onLoadMore={loadMore}
  />
</main>
