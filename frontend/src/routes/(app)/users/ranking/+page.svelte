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

  // Rate Limiting State
  let clickCount = $state(0);
  let lastClickTime = $state(0);
  let isRateLimited = $state(false);
  let rateLimitResetTime = $state(0);

  $effect(() => {
    // Reset on data change (sort change)
    if (
      data.ranking &&
      (data.ranking.pagination?.current_page === 1 ||
        data.ranking.current_page === 1)
    ) {
      ranking = data.ranking;
    }
  });

  async function changeSort(sort: string) {
    if (loading || activeSort === sort || isRateLimited) return;
    
    // Simple Rate Limiting Logic (Max 10 clicks per 5 seconds)
    const now = Date.now();
    if (now - lastClickTime < 5000) {
      clickCount++;
      if (clickCount >= 10) {
        isRateLimited = true;
        rateLimitResetTime = now + 30000; // 30 seconds penalty
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

    loading = true;
    const url = new URL(page.url);
    url.searchParams.set("sort", sort);
    url.searchParams.set("page", "1");
    
    try {
      await goto(url.toString(), { keepFocus: true });
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    const nextUrl =
      ranking?.links?.next ||
      (ranking?.pagination?.has_more
        ? `/users/ranking?sort=${activeSort}&page=${(ranking?.pagination?.current_page || 1) + 1}`
        : null);
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
        links: nextData.links || nextData,
      };
    } catch (error) {
      console.error("Error loading more users:", error);
    } finally {
      loading = false;
    }
  }

  // Calculate startIndex correctly for rank display
  // In our standardized API, current_page is in ranking.pagination.current_page or ranking.current_page
  let currentPage = $derived(
    ranking?.pagination?.current_page || ranking?.current_page || 1,
  );
  let perPage = $derived(
    ranking?.pagination?.per_page || ranking?.per_page || 24,
  );
</script>

<SEO
  title="User Leaderboard"
  description="Check out the most active members of the AniRank community based on XP, ratings, and social interactions."
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-8">
  <div
    class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10"
  >
    <div>
      <h1
        class="text-4xl font-black tracking-tight mb-2 uppercase italic text-on-surface flex items-center gap-4"
      >
        User Leaderboard
      </h1>
      <p
        class="text-on-surface-variant font-bold uppercase tracking-widest text-xs"
      >
        The most active members of our community
      </p>
    </div>

    <div class="flex flex-col gap-3">
      <div
        class="flex items-center p-1 bg-surface-highest border border-white/5 rounded-md w-fit"
      >
        <button
          onclick={() => changeSort("xp")}
          disabled={loading || isRateLimited}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed {activeSort ===
          'xp'
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface-variant hover:text-on-surface'}"
        >
          Top XP
        </button>
        <button
          onclick={() => changeSort("ratings")}
          disabled={loading || isRateLimited}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed {activeSort ===
          'ratings'
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface-variant hover:text-on-surface'}"
        >
          Top Ratings
        </button>
        <button
          onclick={() => changeSort("comments")}
          disabled={loading || isRateLimited}
          class="px-6 py-2.5 rounded-sm font-bold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed {activeSort ===
          'comments'
            ? 'bg-primary text-white shadow-lg'
            : 'text-on-surface-variant hover:text-on-surface'}"
        >
          Top Social
        </button>
      </div>

      {#if isRateLimited}
        <p class="text-[10px] text-red-500 font-bold uppercase tracking-widest animate-pulse">
          Cooling down: Too many requests. Please wait 30s.
        </p>
      {/if}
    </div>
  </div>

  <UserRankingList
    users={ranking?.data ?? []}
    startIndex={(currentPage - 1) * perPage}
    sort={activeSort}
  />

  <InfiniteScroll
    hasMore={ranking?.links?.next ||
      ranking?.pagination?.has_more ||
      currentPage < (ranking?.pagination?.last_page || ranking?.last_page || 1)}
    {loading}
    onLoadMore={loadMore}
  />
</main>
