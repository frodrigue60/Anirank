<script lang="ts">
  import { onMount } from 'svelte';
  import api from '$lib/api';
  import type { Tournament } from '$lib/types/tournament';
  import InfiniteScroll from '$lib/components/InfiniteScroll.svelte';
  import SEO from '$lib/components/SEO.svelte';

  let tournaments = $state<Tournament[]>([]);
  let currentPage = $state(1);
  let lastPage = $state(1);
  let loading = $state(true);

  onMount(async () => {
    try {
      const response = await api.get('/tournaments'); 
      if (response.data.data) {
        tournaments = response.data.data;
        currentPage = response.data.current_page || 1;
        lastPage = response.data.last_page || 1;
      } else {
        tournaments = response.data; // Fallback if not paginated
      }
    } catch (error) {
      console.error('Error fetching tournaments:', error);
    } finally {
      loading = false;
    }
  });

  async function loadMore() {
    if (loading || currentPage >= lastPage) return;

    loading = true;
    try {
      const nextPage = currentPage + 1;
      const response = await api.get('/tournaments', {
        params: { page: nextPage }
      });

      if (response.data.data) {
        tournaments = [...tournaments, ...response.data.data];
        currentPage = response.data.current_page;
        lastPage = response.data.last_page;
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

<div class="container py-8">
  <div class="header mb-8">
    <h1 class="text-3xl font-bold">Anime Theme Tournaments</h1>
    <p class="text-gray-400">Vote for your favorite openings and endings in our bracket-style tournaments.</p>
  </div>

  {#if tournaments.length === 0 && !loading}
    <div class="empty-state">
      <p>No tournaments currently available. Check back soon!</p>
    </div>
  {:else}
    <div class="tournament-grid">
      {#each tournaments as tournament}
        <a href="/tournaments/{tournament.slug}" class="tournament-card {tournament.status}">
          <div class="status-badge">{tournament.status}</div>
          <h2>{tournament.name}</h2>
          <p>{tournament.description || 'No description available.'}</p>
          <div class="meta">
            <span>Size: {tournament.size} songs</span>
            <span>Filter: {tournament.type_filter || 'All'}</span>
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
</div>

<style>
  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
  }

  .tournament-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 24px;
  }

  .tournament-card {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    padding: 24px;
    text-decoration: none;
    color: white;
    transition: transform 0.2s, border-color 0.2s;
    position: relative;
    overflow: hidden;
  }

  .tournament-card:hover {
    transform: translateY(-5px);
    border-color: var(--primary-color, #ff4e50);
  }

  .status-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    font-size: 0.7rem;
    padding: 4px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    font-weight: 800;
    background: rgba(255, 255, 255, 0.1);
  }

  .active .status-badge {
    background: #ff4e50;
  }

  .completed .status-badge {
    background: #4caf50;
  }

  h2 {
    font-size: 1.5rem;
    margin-bottom: 12px;
    padding-right: 60px;
  }

  p {
    font-size: 0.9rem;
    color: #aaa;
    margin-bottom: 20px;
    line-height: 1.5;
  }

  .meta {
    display: flex;
    justify-content: space-between;
    font-size: 0.8rem;
    opacity: 0.6;
  }
</style>
