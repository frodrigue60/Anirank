<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import UserRankingList from "$lib/components/UserRankingList.svelte";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import axios from "$lib/api";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let ranking = $state(data.ranking);
  let activeSort = $derived(data.sort);
  let loading = $state(false);

  $effect(() => {
    // Reset on data change (sort change)
    if (data.ranking && (data.ranking.pagination?.current_page === 1 || data.ranking.current_page === 1)) {
      ranking = data.ranking;
    }
  });

  async function changeSort(sort: string) {
    const url = new URL(page.url);
    url.searchParams.set("sort", sort);
    url.searchParams.set("page", "1");
    await goto(url.toString());
  }

  async function loadMore() {
    const nextUrl = ranking?.links?.next || (ranking?.pagination?.has_more ? `/users/ranking?sort=${activeSort}&page=${(ranking?.pagination?.current_page || 1) + 1}` : null);
    if (loading || !nextUrl) return;

    loading = true;
    try {
      const response = await axios.get(nextUrl);
      const nextData = response.data;
      // Merge data and update meta
      ranking = { 
        ...nextData, 
        data: [...ranking.data, ...nextData.data],
        pagination: nextData.pagination || nextData,
        links: nextData.links || nextData
      };
    } catch (error) {
      console.error("Error loading more users:", error);
    } finally {
      loading = false;
    }
  }

  // Calculate startIndex correctly for rank display
  // In our standardized API, current_page is in ranking.pagination.current_page or ranking.current_page
  let currentPage = $derived(ranking?.pagination?.current_page || ranking?.current_page || 1);
  let perPage = $derived(ranking?.pagination?.per_page || ranking?.per_page || 24);
</script>

<SEO 
  title="User Leaderboard" 
  description="Check out the most active members of the AniRank community based on XP, ratings, and social interactions." 
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
    <div>
      <h1 class="text-4xl font-black tracking-tight mb-2 uppercase italic text-white flex items-center gap-4">
        <span class="w-3 h-12 bg-primary transform -skew-x-12"></span>
        User Leaderboard
      </h1>
      <p class="text-white/40 font-bold uppercase tracking-widest text-xs ml-7">
        The most active members of our community
      </p>
    </div>
    
    <div class="flex flex-col gap-3">
      <div class="flex items-center p-1 bg-surface-darker/50 border border-white/5 rounded-xl w-fit backdrop-blur-md">
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
    startIndex={(currentPage - 1) * perPage}
    sort={activeSort}
  />

  <InfiniteScroll
    hasMore={ranking?.links?.next || ranking?.pagination?.has_more || (currentPage < (ranking?.pagination?.last_page || ranking?.last_page || 1))}
    {loading}
    onLoadMore={loadMore}
  />
</main>
