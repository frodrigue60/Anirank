<script lang="ts">
  import api from "$lib/api";

  let { data } = $props();

  // Local State
  let items = $state<any[]>(data.genres || []);
  let errorMsg = $state("");

  // Modal State
  let showModal = $state(false);
  let modalMode = $state<"create" | "edit">("create");
  let editingItem = $state<any>(null);

  // Form State
  let formName = $state("");
  let formSlug = $state("");
  let isSubmitting = $state(false);

  function openCreate() {
    modalMode = "create";
    editingItem = null;
    formName = "";
    formSlug = "";
    errorMsg = "";
    showModal = true;
  }

  function openEdit(item: any) {
    modalMode = "edit";
    editingItem = item;
    formName = item.name;
    formSlug = item.slug || "";
    errorMsg = "";
    showModal = true;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    isSubmitting = true;
    errorMsg = "";

    const payload = {
      name: formName,
      slug: formSlug || undefined,
    };

    try {
      if (modalMode === "create") {
        const res = await api.post("/admin/taxonomies/genres", payload);
        if (res.data?.data) {
          items = [res.data.data, ...items];
        }
      } else {
        const res = await api.put(
          `/admin/taxonomies/genres/${editingItem.id}`,
          payload,
        );
        if (res.data) {
          const index = items.findIndex((i) => i.id === editingItem.id);
          if (index !== -1) {
            items[index] = { ...items[index], ...payload };
          }
        }
      }
      showModal = false;
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Failed to save genre";
    } finally {
      isSubmitting = false;
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Are you sure you want to delete this genre?")) return;

    try {
      await api.delete(`/admin/taxonomies/genres/${id}`);
      items = items.filter((i) => i.id !== id);
    } catch (err: any) {
      console.error(err);
      alert(err.response?.data?.message || "Delete failed");
    }
  }
</script>

<svelte:head>
  <title>Genres | Admin</title>
</svelte:head>

<div class="space-y-6 animate-fade-in">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-on-surface">Content Genres</h1>
      <p class="text-on-surface-variant/70 mt-1">Manage content categorization (Action, Drama, etc.)</p>
    </div>
    <button
      onclick={openCreate}
      class="px-4 py-2 bg-primary hover:bg-primary/90 text-on-surface font-medium rounded-xl transition-colors flex items-center gap-2"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      Add Genre
    </button>
  </div>

  <div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-left text-sm text-on-surface-variant">
        <thead class="bg-surface-highest text-on-surface-variant/70 text-xs uppercase tracking-wider">
          <tr>
            <th class="px-6 py-4 font-medium">ID</th>
            <th class="px-6 py-4 font-medium">Name</th>
            <th class="px-6 py-4 font-medium">Slug</th>
            <th class="px-6 py-4 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/5">
          {#each items as item}
            <tr class="hover:bg-white/2 transition-colors group">
              <td class="px-6 py-4 font-mono text-xs opacity-40">#{item.id}</td>
              <td class="px-6 py-4 font-medium text-on-surface">{item.name}</td>
              <td class="px-6 py-4">
                <span class="text-xs font-mono text-primary opacity-70">/{item.slug || 'auto'}</span>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    onclick={() => openEdit(item)}
                    aria-label="Edit genre"
                    title="Edit genre"
                    class="p-2 hover:bg-surface-highest text-on-surface-variant/70 hover:text-on-surface rounded-lg transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                  </button>
                  <button
                    onclick={() => handleDelete(item.id)}
                    aria-label="Delete genre"
                    title="Delete genre"
                    class="p-2 hover:bg-red-500/20 text-on-surface-variant/70 hover:text-red-400 rounded-lg transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="4" class="px-6 py-12 text-center text-on-surface-variant/40 italic">
                No genres found. Click "Add Genre" to create one.
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
    <div class="bg-surface-container border border-outline-variant rounded-2xl w-full max-w-md shadow-2xl overflow-hidden">
      <div class="px-6 py-4 border-b border-outline-variant flex justify-between items-center">
        <h2 class="text-lg font-bold text-on-surface">
          {modalMode === 'create' ? 'Add' : 'Edit'} Genre
        </h2>
        <button onclick={() => (showModal = false)} class="text-on-surface-variant/70 hover:text-on-surface" aria-label="Close modal">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form onsubmit={handleSubmit} class="p-6 space-y-4">
        {#if errorMsg}
          <div class="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-xs text-red-400">
            {errorMsg}
          </div>
        {/if}

        <div class="space-y-1.5">
          <label for="name" class="text-xs font-medium text-on-surface-variant/70 uppercase tracking-widest ml-1">Name</label>
          <input
            id="name"
            type="text"
            bind:value={formName}
            required
            placeholder="e.g. Action, Comedy, Drama"
            class="w-full bg-surface-highest border border-outline-variant rounded-xl px-4 py-2.5 text-on-surface focus:outline-none focus:border-primary/50 transition-all text-sm"
          />
        </div>

        <div class="space-y-1.5">
          <label for="slug" class="text-xs font-medium text-on-surface-variant/70 uppercase tracking-widest ml-1">Slug (Optional)</label>
          <input
            id="slug"
            type="text"
            bind:value={formSlug}
            placeholder="auto-generated if empty"
            class="w-full bg-surface-highest border border-outline-variant rounded-xl px-4 py-2.5 text-on-surface focus:outline-none focus:border-primary/50 transition-all font-mono text-sm"
          />
        </div>

        <div class="pt-4 flex gap-3">
          <button
            type="button"
            onclick={() => (showModal = false)}
            class="flex-1 px-4 py-2.5 rounded-xl bg-surface-highest hover:bg-surface-highest/80 text-on-surface-variant font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting || !formName}
            class="flex-1 px-4 py-2.5 rounded-xl bg-primary hover:bg-primary/90 text-on-surface font-bold transition-all disabled:opacity-50"
          >
            {isSubmitting ? 'Saving...' : 'Save Genre'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
