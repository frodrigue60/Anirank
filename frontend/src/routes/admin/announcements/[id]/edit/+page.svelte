<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import api from '$lib/api';

  let announcement = $state({
    id: 0,
    title: '',
    content: '',
    type: 'info',
    icon: 'info',
    url: '',
    priority: 0,
    is_active: true,
    starts_at: '',
    ends_at: '',
    image_url: ''
  });

  let imageFile = $state<File | null>(null);
  let loading = $state(true);
  let saving = $state(false);

  onMount(async () => {
    const id = page.params.id;
    try {
      const response = await api.get(`/admin/announcements/${id}`);
      const data = response.data.data;
      
      // Format dates for datetime-local input
      if (data.starts_at) data.starts_at = new Date(data.starts_at).toISOString().slice(0, 16);
      if (data.ends_at) data.ends_at = new Date(data.ends_at).toISOString().slice(0, 16);
      
      announcement = data;
    } catch (error) {
      console.error('Error loading announcement:', error);
      alert('Failed to load announcement.');
    } finally {
      loading = false;
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    saving = true;

    try {
      const formData = new FormData();
      Object.entries(announcement).forEach(([key, value]) => {
        if (value !== null && value !== undefined && key !== 'image_url') {
          formData.append(key, value.toString());
        }
      });

      if (imageFile) {
        formData.append('image_file', imageFile);
      }

      await api.put(`/admin/announcements/${announcement.id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      goto('/admin/announcements');
    } catch (error) {
      console.error('Error updating announcement:', error);
      alert('Failed to update announcement.');
    } finally {
      saving = false;
    }
  }

  const types = ['info', 'success', 'warning', 'danger', 'event'];
</script>

<div class="p-6 max-w-4xl mx-auto">
  <div class="mb-8">
    <a href="/admin/announcements" class="text-primary hover:underline flex items-center gap-1 mb-2 text-sm">
      <span class="material-symbols-outlined text-sm">arrow_back</span> Back to List
    </a>
    <h1 class="text-2xl font-bold">Edit Announcement</h1>
  </div>

  {#if loading}
    <div class="text-center py-20 opacity-50">Loading announcement details...</div>
  {:else}
    <form onsubmit={handleSubmit} class="grid grid-cols-1 md:grid-cols-2 gap-6 bg-surface-darker p-8 rounded-2xl border border-outline-variant">
      <div class="md:col-span-2 space-y-2">
        <label for="title" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Title</label>
        <input 
          id="title"
          bind:value={announcement.title}
          type="text" 
          required
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="md:col-span-2 space-y-2">
        <label for="image" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Background Image (S3 Upload)</label>
        {#if announcement.image_url}
          <div class="mb-2 relative rounded-lg overflow-hidden h-24 border border-outline-variant">
            <img src={announcement.image_url} alt="Current banner" class="w-full h-full object-cover" />
            <div class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
              <span class="text-xs font-bold text-on-surface uppercase tracking-widest">Current Image</span>
            </div>
          </div>
        {/if}
        <input 
          id="image"
          type="file" 
          accept="image/*"
          onchange={(e) => imageFile = (e.target as HTMLInputElement).files?.[0] || null}
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="md:col-span-2 space-y-2">
        <label for="content" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Content (Optional)</label>
        <textarea 
          id="content"
          bind:value={announcement.content}
          rows="3"
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        ></textarea>
      </div>

      <div class="space-y-2">
        <label for="type" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Type</label>
        <select 
          id="type"
          bind:value={announcement.type}
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors appearance-none"
        >
          {#each types as t}
            <option value={t}>{t.toUpperCase()}</option>
          {/each}
        </select>
      </div>

      <div class="space-y-2">
        <label for="icon" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Icon (Material Symbol)</label>
        <input 
          id="icon"
          bind:value={announcement.icon}
          type="text" 
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="md:col-span-2 space-y-2">
        <label for="url" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">URL (Optional)</label>
        <input 
          id="url"
          bind:value={announcement.url}
          type="text" 
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="space-y-2">
        <label for="priority" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Priority</label>
        <input 
          id="priority"
          bind:value={announcement.priority}
          type="number" 
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="space-y-2">
        <label for="is_active" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Active Status</label>
        <div class="flex items-center gap-3 py-3">
          <button 
            id="is_active"
            type="button"
            aria-label="Toggle active status"
            onclick={() => announcement.is_active = !announcement.is_active}
            class="w-12 h-6 rounded-full relative transition-colors {announcement.is_active ? 'bg-primary' : 'bg-surface-highest'}"
          >
            <div class="absolute top-1 w-4 h-4 rounded-full bg-white transition-all {announcement.is_active ? 'right-1' : 'left-1'}"></div>
          </button>
          <span class="text-sm font-medium">{announcement.is_active ? 'Enabled' : 'Disabled'}</span>
        </div>
      </div>

      <div class="space-y-2">
        <label for="starts_at" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Starts At</label>
        <input 
          id="starts_at"
          bind:value={announcement.starts_at}
          type="datetime-local" 
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="space-y-2">
        <label for="ends_at" class="text-xs font-bold uppercase tracking-widest text-on-surface/50">Ends At</label>
        <input 
          id="ends_at"
          bind:value={announcement.ends_at}
          type="datetime-local" 
          class="w-full bg-surface-highest border border-outline-variant rounded-lg px-4 py-3 focus:border-primary/30 focus:bg-surface-highest outline-none transition-colors"
        />
      </div>

      <div class="md:col-span-2 pt-6">
        <button 
          type="submit" 
          disabled={saving}
          class="w-full bg-primary hover:opacity-90 disabled:opacity-50 text-on-surface font-bold py-4 rounded-xl transition-all shadow-lg shadow-primary/20"
        >
          {saving ? 'Saving...' : 'Update Announcement'}
        </button>
      </div>
    </form>
  {/if}
</div>
