<script lang="ts">
  import { onMount } from 'svelte';
  import api from '$lib/api';
  import { toastState } from '$lib/state/toast.svelte';
  import Pencil from "lucide-svelte/icons/pencil";
  import Trash2 from "lucide-svelte/icons/trash-2";

  interface Announcement {
    id: number;
    title: string;
    type: string;
    priority: number;
    is_active: boolean;
    starts_at: string | null;
    ends_at: string | null;
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
      const index = announcements.findIndex(a => a.id === id);
      if (index !== -1) {
        announcements[index].is_active = !announcements[index].is_active;
      }
      toastState.addToast("Status updated", "success");
    } catch (error: any) {
      console.error('Error toggling announcement:', error);
      toastState.addToast(error.response?.data?.message || 'Failed to toggle status', "error");
    }
  }


  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this announcement?')) return;
    
    try {
      await api.delete(`/admin/announcements/${id}`);
      announcements = announcements.filter(a => a.id !== id);
      toastState.addToast("Announcement deleted", "success");
    } catch (error: any) {
      console.error('Error deleting announcement:', error);
      toastState.addToast(error.response?.data?.message || 'Failed to delete announcement', "error");
    }
  }
</script>

<div class="p-6">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-2xl font-bold">Manage Announcements</h1>
    <a href="/admin/announcements/create" class="bg-primary hover:opacity-90 text-on-surface px-4 py-2 rounded-lg font-bold">
      + New Announcement
    </a>
  </div>

  {#if loading}
    <div class="text-center py-20 opacity-50">Loading announcements...</div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-outline-variant uppercase text-xs tracking-widest text-on-surface/50">
            <th class="py-4 font-bold">Status</th>
            <th class="py-4 font-bold">Title</th>
            <th class="py-4 font-bold">Type</th>
            <th class="py-4 font-bold">Priority</th>
            <th class="py-4 font-bold">Active Period</th>
            <th class="py-4 font-bold text-right pr-6">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each announcements as a}
            <tr class="border-b border-outline-variant hover:bg-surface-highest transition-colors">
              <td class="py-4">
                <button 
                  onclick={() => toggleActive(a.id)}
                  title={a.is_active ? 'Deactivate announcement' : 'Activate announcement'}
                  aria-label={a.is_active ? 'Deactivate announcement' : 'Activate announcement'}
                  class="w-10 h-6 rounded-full relative transition-colors {a.is_active ? 'bg-primary' : 'bg-surface-highest'}"
                >
                  <div class="absolute top-1 w-4 h-4 rounded-full bg-white transition-all {a.is_active ? 'right-1' : 'left-1'}"></div>
                </button>
              </td>
              <td class="py-4 font-bold">{a.title}</td>
              <td class="py-4">
                <span class="px-2 py-1 rounded text-[10px] uppercase font-bold bg-surface-highest border border-outline-variant">{a.type}</span>
              </td>
              <td class="py-4 font-mono">{a.priority}</td>
              <td class="py-4 text-[10px] uppercase font-bold tracking-tighter">
                {#if !a.starts_at && !a.ends_at}
                  <span class="text-primary/50 italic">No limit</span>
                {:else}
                   <div class="flex flex-col opacity-60">
                     <span>S: {a.starts_at ? new Date(a.starts_at).toLocaleDateString() : '∞'}</span>
                     <span>E: {a.ends_at ? new Date(a.ends_at).toLocaleDateString() : '∞'}</span>
                   </div>
                {/if}
              </td>
              <td class="py-4 text-right pr-6">
                <div class="flex items-center justify-end gap-3">
                  <a 
                    href="/admin/announcements/{a.id}/edit" 
                    title="Edit announcement"
                    aria-label="Edit announcement"
                    class="text-on-surface/50 hover:text-on-surface transition-colors"
                  >
                    <Pencil size={18} />
                  </a>
                  <button 
                    onclick={() => handleDelete(a.id)}
                    title="Delete announcement"
                    aria-label="Delete announcement"
                    class="text-red-500/50 hover:text-red-500 transition-colors"
                  >
                    <Trash2 size={18} />
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
