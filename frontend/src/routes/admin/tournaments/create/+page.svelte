<script lang="ts">
  import { goto } from '$app/navigation';
  import api from '$lib/api';

  let name = '';
  let description = '';
  let size = 16;
  let type_filter = '';
  let loading = false;

  async function handleSubmit() {
    if (!name || !size) return;
    
    loading = true;
    try {
      // Note: We need a POST /admin/tournaments endpoint.
      // Based on the handler, I might have forgotten the Create endpoint in delivery?
      // Let me check the handler again.
      await api.post('/admin/tournaments', {
        name,
        description,
        size: Number(size),
        type_filter: type_filter || null
      });
      goto('/admin/tournaments');
    } catch (error) {
      console.error('Error creating tournament:', error);
      alert('Failed to create tournament. Make sure the backend endpoint exists.');
    } finally {
      loading = false;
    }
  }
</script>

<div class="max-w-2xl mx-auto p-6">
  <div class="mb-8">
    <h1 class="text-2xl font-bold">Create New Tournament</h1>
    <p class="opacity-60">Set up a new bracket-style competition. It will start in 'draft' mode.</p>
  </div>

  <form on:submit|preventDefault={handleSubmit} class="space-y-6">
    <div class="space-y-2">
      <label for="name" class="block font-bold opacity-80">Tournament Name</label>
      <input 
        id="name"
        type="text" 
        bind:value={name}
        placeholder="e.g. Best Opening Spring 2024"
        class="w-full bg-surface-highest border border-outline-variant rounded-lg p-3 outline-none focus:border-primary/30 focus:bg-surface-highest"
        required
      />
    </div>

    <div class="space-y-2">
      <label for="desc" class="block font-bold opacity-80">Description</label>
      <textarea 
        id="desc"
        bind:value={description}
        placeholder="Describe the tournament..."
        class="w-full bg-surface-highest border border-outline-variant rounded-lg p-3 outline-none focus:border-primary/30 focus:bg-surface-highest h-32"
      ></textarea>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-2">
        <label for="size" class="block font-bold opacity-80">Size (Bracket Depth)</label>
        <select 
          id="size"
          bind:value={size}
          class="w-full bg-surface-highest border border-outline-variant rounded-lg p-3 outline-none focus:border-primary/30 focus:bg-surface-highest"
        >
          <option value={8}>8 Songs</option>
          <option value={16}>16 Songs</option>
          <option value={32}>32 Songs</option>
          <option value={64}>64 Songs</option>
        </select>
      </div>

      <div class="space-y-2">
        <label for="filter" class="block font-bold opacity-80">Theme Type Filter</label>
        <select 
          id="filter"
          bind:value={type_filter}
          class="w-full bg-surface-highest border border-outline-variant rounded-lg p-3 outline-none focus:border-primary/30 focus:bg-surface-highest"
        >
          <option value="">All Types</option>
          <option value="OP">Openings Only</option>
          <option value="ED">Endings Only</option>
        </select>
      </div>
    </div>

    <div class="pt-4 flex gap-4">
      <button 
        type="submit" 
        class="flex-1 bg-primary text-on-surface font-bold py-3 rounded-xl hover:opacity-90 disabled:opacity-50"
        disabled={loading}
      >
        {loading ? 'Creating...' : 'Create Tournament'}
      </button>
      <a 
        href="/admin/tournaments" 
        class="flex-1 bg-surface-highest text-center font-bold py-3 rounded-xl hover:bg-white/20"
      >
        Cancel
      </a>
    </div>
  </form>
</div>
