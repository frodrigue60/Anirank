<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";
  import type { Tournament } from "$lib/types/tournament";
  import InfiniteScroll from "$lib/components/InfiniteScroll.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import { Trophy, Users, Music } from "lucide-svelte";

  let tournaments = $state<Tournament[]>([]);
  let currentPage = $state(1);
  let lastPage = $state(1);
  let loading = $state(true);

  onMount(async () => {
    try {
      const response = await api.get("/tournaments");
      if (response.data.data) {
        tournaments = response.data.data;
        currentPage = response.data.pagination?.current_page || 1;
        lastPage = response.data.pagination?.last_page || 1;
      } else {
        tournaments = response.data; // Fallback if not paginated
      }
    } catch (error) {
      console.error("Error fetching tournaments:", error);
    } finally {
      loading = false;
    }
  });

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get("/tournaments", {
        params: { page: nextPage },
      });

      if (response.data.data) {
        tournaments = [...tournaments, ...response.data.data];
        currentPage = response.data.pagination.current_page;
        lastPage = response.data.pagination.last_page;
      }
    } catch (e) {
      console.error("Error loading more tournaments", e);
    } finally {
      loading = false;
    }
  }
</script>

<SEO
  title="Anime Theme Tournaments"
  description="Vote for your favorite anime openings and endings in our bracket-style tournaments on AniRank."
/>

<main class="flex-1 w-full max-w-[1440px] mx-auto px-6 py-12">
  <!-- Header -->
  <div class="mb-12 text-center lg:text-left">
    <h1
      class="text-4xl lg:text-5xl font-black text-on-surface mb-4 tracking-tight"
    >
      Anime Theme Tournaments
    </h1>
    <p class="text-on-surface-variant text-lg max-w-2xl">
      Vote for your favorite openings and endings in our community-curated
      bracket-style tournaments.
    </p>
  </div>

  {#if tournaments.length === 0 && !loading}
    <div
      class="text-center py-24 bg-surface-container/30 rounded-md border-2 border-dashed border-white/5"
    >
      <Trophy size={48} class="text-on-surface-variant/20 mx-auto mb-4" />
      <h3 class="text-xl font-bold text-on-surface-variant/50">
        No tournaments currently available
      </h3>
      <p class="text-on-surface-variant/30 mt-2">
        Check back soon for new events!
      </p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each tournaments as tournament}
        <a
          href="/tournaments/{tournament.slug}"
          class="group relative flex flex-col bg-surface-container rounded-sm border border-on-surface-variant/10 hover:border-primary/50 transition-all duration-300 overflow-hidden shadow-lg shadow-black/20"
        >
          <!-- Status Badge -->
          <div class="absolute top-4 right-4 z-10">
            <span
              class="px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest shadow-lg
              {tournament.status === 'active' ? 'bg-primary text-white' : ''}
              {tournament.status === 'completed'
                ? 'bg-green-500 text-white'
                : ''}
              {tournament.status === 'draft'
                ? 'bg-surface-highest text-on-surface-variant'
                : ''}"
            >
              {tournament.status}
            </span>
          </div>

          <div class="p-8 flex-1 flex flex-col">
            <h2
              class="text-2xl font-black text-on-surface group-hover:text-primary transition-colors mb-3 line-clamp-1"
            >
              {tournament.name}
            </h2>
            <p class="text-on-surface-variant text-sm line-clamp-2 mb-6 flex-1">
              {tournament.description || "No description available."}
            </p>

            <div
              class="flex items-center justify-between pt-6 border-t border-on-surface-variant/5 mt-auto"
            >
              <div class="flex items-center gap-4">
                <div class="flex flex-col">
                  <span
                    class="text-[10px] text-on-surface-variant uppercase font-black tracking-widest opacity-40 mb-1"
                    >Size</span
                  >
                  <div
                    class="flex items-center gap-1.5 text-on-surface font-bold text-sm"
                  >
                    <Music size={14} class="text-primary/60" />
                    {tournament.size} Songs
                  </div>
                </div>
                <div class="flex flex-col">
                  <span
                    class="text-[10px] text-on-surface-variant uppercase font-black tracking-widest opacity-40 mb-1"
                    >Filter</span
                  >
                  <div
                    class="flex items-center gap-1.5 text-on-surface font-bold text-sm"
                  >
                    <Users size={14} class="text-primary/60" />
                    {tournament.type_filter || "All"}
                  </div>
                </div>
              </div>

              <div
                class="w-10 h-10 rounded-full bg-surface-highest flex items-center justify-center group-hover:bg-primary group-hover:scale-110 transition-all"
              >
                <span
                  class="material-symbols-outlined text-on-surface group-hover:text-white transition-colors"
                  >chevron_right</span
                >
              </div>
            </div>
          </div>
        </a>
      {/each}
    </div>

    <InfiniteScroll
      hasMore={currentPage < lastPage}
      {loading}
      onLoadMore={loadMore}
    />
  {/if}
</main>
