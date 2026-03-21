<script lang="ts">
  import { goto } from '$app/navigation';
  import api from '$lib/api';

  let announcement = $state({
    title: '',
    content: '',
    type: 'info',
    icon: 'info',
    url: '',
    priority: 0,
    is_active: true,
    starts_at: '',
    ends_at: ''
  });

  let loading = $state(false);
  let imageFile = $state<File | null>(null);
  let imagePreview = $state<string | null>(null);

  function handleFileChange(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0] || null;
    imageFile = file;
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => imagePreview = e.target?.result as string;
      reader.readAsDataURL(file);
    } else {
      imagePreview = null;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;

    try {
      const formData = new FormData();
      Object.entries(announcement).forEach(([key, value]) => {
        if (value !== null && value !== undefined) {
          formData.append(key, value.toString());
        }
      });

      if (imageFile) {
        formData.append('image_file', imageFile);
      }

      await api.post('/admin/announcements', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      goto('/admin/announcements');
    } catch (error) {
      console.error('Error creating announcement:', error);
      alert('Failed to create announcement.');
    } finally {
      loading = false;
    }
  }

  const types = ['info', 'success', 'warning', 'danger', 'event'];
</script>

<div class="p-6 max-w-4xl mx-auto">
  <div class="mb-8">
    <a href="/admin/announcements" class="text-primary hover:underline flex items-center gap-1 mb-2 text-sm">
      <span class="material-symbols-outlined text-sm">arrow_back</span> Back to List
    </a>
    <h1 class="text-2xl font-bold">Create New Announcement</h1>
  </div>

  <form onsubmit={handleSubmit} class="grid grid-cols-1 md:grid-cols-2 gap-6 bg-surface-darker p-8 rounded-2xl border border-white/5">
    <div class="md:col-span-2 space-y-2">
      <label for="title" class="text-xs font-bold uppercase tracking-widest text-white/50">Title</label>
      <input 
        id="title"
        bind:value={announcement.title}
        type="text" 
        required
        placeholder="Enter announcement title"
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="md:col-span-2 space-y-2">
      <label for="image" class="text-xs font-bold uppercase tracking-widest text-white/50">Background Image (S3 Upload)</label>
      {#if imagePreview}
        <div class="mb-2 relative rounded-lg overflow-hidden h-32 border border-white/10 group">
          <img src={imagePreview} alt="Preview" class="w-full h-full object-cover" />
          <button 
            type="button"
            onclick={() => { imageFile = null; imagePreview = null; }}
            class="absolute top-2 right-2 bg-black/60 hover:bg-red-500/80 text-white p-1 rounded-full transition-colors opacity-0 group-hover:opacity-100"
          >
            <span class="material-symbols-outlined text-sm">close</span>
          </button>
        </div>
      {/if}
      <input 
        id="image"
        type="file" 
        accept="image/*"
        onchange={handleFileChange}
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="md:col-span-2 space-y-2">
      <label for="content" class="text-xs font-bold uppercase tracking-widest text-white/50">Content (Optional)</label>
      <textarea 
        id="content"
        bind:value={announcement.content}
        rows="3"
        placeholder="Enter detailed description"
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      ></textarea>
    </div>

    <div class="space-y-2">
      <label for="type" class="text-xs font-bold uppercase tracking-widest text-white/50">Type</label>
      <select 
        id="type"
        bind:value={announcement.type}
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors appearance-none"
      >
        {#each types as t}
          <option value={t}>{t.toUpperCase()}</option>
        {/each}
      </select>
    </div>

    <div class="space-y-2">
      <label for="icon" class="text-xs font-bold uppercase tracking-widest text-white/50">Icon (Material Symbol)</label>
      <input 
        id="icon"
        bind:value={announcement.icon}
        type="text" 
        placeholder="e.g. info, star, event"
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="md:col-span-2 space-y-2">
      <label for="url" class="text-xs font-bold uppercase tracking-widest text-white/50">URL (Optional)</label>
      <input 
        id="url"
        bind:value={announcement.url}
        type="text" 
        placeholder="Internal link (/songs/...) or external URL"
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="space-y-2">
      <label for="priority" class="text-xs font-bold uppercase tracking-widest text-white/50">Priority</label>
      <input 
        id="priority"
        bind:value={announcement.priority}
        type="number" 
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="space-y-2">
      <label for="is_active" class="text-xs font-bold uppercase tracking-widest text-white/50">Active Status</label>
      <div class="flex items-center gap-3 py-3">
        <button 
          id="is_active"
          type="button"
          aria-label="Toggle active status"
          onclick={() => announcement.is_active = !announcement.is_active}
          class="w-12 h-6 rounded-full relative transition-colors {announcement.is_active ? 'bg-primary' : 'bg-white/10'}"
        >
          <div class="absolute top-1 w-4 h-4 rounded-full bg-white transition-all {announcement.is_active ? 'right-1' : 'left-1'}"></div>
        </button>
        <span class="text-sm font-medium">{announcement.is_active ? 'Enabled' : 'Disabled'}</span>
      </div>
    </div>

    <div class="space-y-2">
      <label for="starts_at" class="text-xs font-bold uppercase tracking-widest text-white/50">Starts At</label>
      <input 
        id="starts_at"
        bind:value={announcement.starts_at}
        type="datetime-local" 
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="space-y-2">
      <label for="ends_at" class="text-xs font-bold uppercase tracking-widest text-white/50">Ends At</label>
      <input 
        id="ends_at"
        bind:value={announcement.ends_at}
        type="datetime-local" 
        class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 focus:border-primary outline-none transition-colors"
      />
    </div>

    <div class="md:col-span-2 pt-6">
      <button 
        type="submit" 
        disabled={loading}
        class="w-full bg-primary hover:opacity-90 disabled:opacity-50 text-white font-bold py-4 rounded-xl transition-all shadow-lg shadow-primary/20"
      >
        {loading ? 'Creating...' : 'Create Announcement'}
      </button>
    </div>
  </form>
</div>
