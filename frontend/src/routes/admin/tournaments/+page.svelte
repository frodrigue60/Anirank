<script lang="ts">
  import { onMount } from 'svelte';
  import api from '$lib/api';
  import type { Tournament } from '$lib/types/tournament';

  let tournaments: Tournament[] = [];
  let loading = true;

  onMount(loadTournaments);

  async function loadTournaments() {
    loading = true;
    try {
      const response = await api.get('/admin/tournaments');
      tournaments = response.data.data;
    } catch (error) {
      console.error('Error fetching tournaments:', error);
    } finally {
      loading = false;
    }
  }

  async function handleSeed(id: number) {
    if (!confirm('Are you sure you want to seed this tournament? This will generate the initial bracket and make it active.')) return;
    
    try {
      await api.post(`/admin/tournaments/${id}/seed`);
      alert('Tournament seeded successfully!');
      loadTournaments();
    } catch (error) {
      console.error('Error seeding tournament:', error);
      alert('Failed to seed tournament.');
    }
  }

  async function handleAdvance(id: number) {
    if (!confirm('Are you sure you want to advance this tournament phase? This will close the current active matchups and move winners to the next round.')) return;
    
    try {
      await api.post(`/admin/tournaments/${id}/advance`);
      alert('Tournament phase advanced successfully!');
      loadTournaments();
    } catch (error) {
      console.error('Error advancing tournament:', error);
      alert('Failed to advance tournament phase.');
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this tournament? This will PERMANENTLY remove all matchups and votes.')) return;
    
    try {
      await api.delete(`/admin/tournaments/${id}`);
      alert('Tournament deleted successfully!');
      loadTournaments();
    } catch (error) {
      console.error('Error deleting tournament:', error);
      alert('Failed to delete tournament.');
    }
  }
</script>

<div class="admin-tournaments p-6">
  <div class="header flex justify-between items-center mb-8">
    <h1 class="text-2xl font-bold">Manage Tournaments</h1>
    <a href="/admin/tournaments/create" class="bg-primary hover:opacity-90 text-white px-4 py-2 rounded-lg font-bold">
      + New Tournament
    </a>
  </div>

  {#if loading}
    <div class="text-center py-20 opacity-50">Loading tournaments...</div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10">
            <th class="py-4 font-bold opacity-60">ID</th>
            <th class="py-4 font-bold opacity-60">Name</th>
            <th class="py-4 font-bold opacity-60">Status</th>
            <th class="py-4 font-bold opacity-60">Size</th>
            <th class="py-4 font-bold opacity-60">Created At</th>
            <th class="py-4 font-bold opacity-60">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each tournaments as t}
            <tr class="border-b border-white/5 hover:bg-white/5">
              <td class="py-4">#{t.id}</td>
              <td class="py-4">
                <strong>{t.name}</strong>
                <div class="text-xs opacity-50">{t.slug}</div>
              </td>
              <td class="py-4">
                <span class="status-pill {t.status}">{t.status}</span>
              </td>
              <td class="py-4">{t.size} songs</td>
              <td class="py-4 text-sm opacity-60">{new Date(t.created_at).toLocaleDateString()}</td>
              <td class="py-4">
                <div class="flex gap-2">
                  {#if t.status === 'draft'}
                    <a 
                      class="text-xs bg-green-600 px-3 py-1 rounded font-bold hover:bg-green-700"
                      href="/admin/tournaments/{t.id}/seed"
                    >
                      Seed & Start
                    </a>
                  {/if}
                  {#if t.status === 'active'}
                    <button 
                      class="text-xs bg-blue-600 px-3 py-1 rounded font-bold hover:bg-blue-700"
                      on:click={() => handleAdvance(t.id)}
                    >
                      Advance Phase
                    </button>
                  {/if}
                  {#if t.status === 'draft' || t.status === 'completed'}
                    <button 
                      class="text-xs bg-red-600/20 text-red-400 px-3 py-1 rounded font-bold hover:bg-red-600 hover:text-white transition-colors"
                      on:click={() => handleDelete(t.id)}
                    >
                      Delete
                    </button>
                  {/if}
                  <a href="/tournaments/{t.slug}" target="_blank" class="text-xs bg-white/10 px-3 py-1 rounded hover:bg-white/20">
                    Preview
                  </a>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .status-pill {
    font-size: 0.7rem;
    padding: 2px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    font-weight: 800;
  }
  .draft { background: rgba(255, 255, 255, 0.1); }
  .active { background: #ff4e50; }
  .completed { background: #4caf50; }
</style>
