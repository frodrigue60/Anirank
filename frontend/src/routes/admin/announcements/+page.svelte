<script lang="ts">
  import { onMount } from 'svelte';
  import api from '$lib/api';

  interface Announcement {
    id: number;
    title: string;
    type: string;
    priority: number;
    is_active: boolean;
    created_at: string;
  }

  let announcements = $state<Announcement[]>([]);
  let loading = $state(true);

  onMount(loadAnnouncements);

  async function loadAnnouncements() {
    loading = true;
    try {
      const response = await api.get('/admin/announcements');
      announcements = response.data.data;
    } catch (error) {
      console.error('Error fetching announcements:', error);
    } finally {
      loading = false;
    }
  }

  async function toggleActive(id: number) {
    try {
      await api.patch(`/admin/announcements/${id}/toggle`);
      loadAnnouncements();
    } catch (error) {
      console.error('Error toggling announcement:', error);
      alert('Failed to toggle status.');
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this announcement?')) return;
    
    try {
      await api.delete(`/admin/announcements/${id}`);
      loadAnnouncements();
    } catch (error) {
      console.error('Error deleting announcement:', error);
      alert('Failed to delete announcement.');
    }
  }
</script>

<div class="p-6">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-2xl font-bold">Manage Announcements</h1>
    <a href="/admin/announcements/create" class="bg-primary hover:opacity-90 text-white px-4 py-2 rounded-lg font-bold">
      + New Announcement
    </a>
  </div>

  {#if loading}
    <div class="text-center py-20 opacity-50">Loading announcements...</div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-white/10 uppercase text-xs tracking-widest text-white/50">
            <th class="py-4 font-bold">Status</th>
            <th class="py-4 font-bold">Title</th>
            <th class="py-4 font-bold">Type</th>
            <th class="py-4 font-bold">Priority</th>
            <th class="py-4 font-bold">Created At</th>
            <th class="py-4 font-bold">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each announcements as a}
            <tr class="border-b border-white/5 hover:bg-white/5 transition-colors">
              <td class="py-4">
                <button 
                  onclick={() => toggleActive(a.id)}
                  title={a.is_active ? 'Deactivate announcement' : 'Activate announcement'}
                  aria-label={a.is_active ? 'Deactivate announcement' : 'Activate announcement'}
                  class="w-10 h-6 rounded-full relative transition-colors {a.is_active ? 'bg-primary' : 'bg-white/10'}"
                >
                  <div class="absolute top-1 w-4 h-4 rounded-full bg-white transition-all {a.is_active ? 'right-1' : 'left-1'}"></div>
                </button>
              </td>
              <td class="py-4 font-bold">{a.title}</td>
              <td class="py-4">
                <span class="px-2 py-1 rounded text-[10px] uppercase font-bold bg-white/5 border border-white/10">{a.type}</span>
              </td>
              <td class="py-4 font-mono">{a.priority}</td>
              <td class="py-4 text-sm opacity-60 font-mono">{new Date(a.created_at).toLocaleDateString()}</td>
              <td class="py-4">
                <div class="flex gap-3">
                  <a 
                    href="/admin/announcements/{a.id}/edit" 
                    title="Edit announcement"
                    aria-label="Edit announcement"
                    class="text-white/50 hover:text-white transition-colors"
                  >
                    <span class="material-symbols-outlined">edit</span>
                  </a>
                  <button 
                    onclick={() => handleDelete(a.id)}
                    title="Delete announcement"
                    aria-label="Delete announcement"
                    class="text-red-500/50 hover:text-red-500 transition-colors"
                  >
                    <span class="material-symbols-outlined">delete</span>
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
