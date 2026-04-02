<script lang="ts">
  import api from "$lib/api";

  let { type, initialItems = [] } = $props<{
    type: "genres" | "formats" | "seasons" | "years";
    initialItems?: any[];
  }>();

  let items = $state<any[]>([]);
  let errorMsg = $state("");

  // Modal State
  let showModal = $state(false);
  let modalMode = $state<"create" | "edit">("create");
  let editingItem = $state<any>(null);

  // Form State
  let formName = $state("");
  let formSlug = $state(""); // only for genres/formats
  let formCurrent = $state(false); // only for seasons/years
  let isSubmitting = $state(false);

  // Update items if initialItems changes
  $effect(() => {
    items = [...initialItems];
  });

  function openCreate() {
    modalMode = "create";
    editingItem = null;
    formName = "";
    formSlug = "";
    formCurrent = false;
    showModal = true;
  }

  function openEdit(item: any) {
    modalMode = "edit";
    editingItem = item;
    formName = item.name;
    formSlug = item.slug || "";
    formCurrent = !!item.current;
    showModal = true;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    isSubmitting = true;
    errorMsg = "";

    const payload: any = { name: formName };
    if (type === "genres" || type === "formats") {
      payload.slug = formSlug || undefined;
    } else {
      payload.current = formCurrent;
    }

    try {
      let res;
      if (modalMode === "create") {
        res = await api.post(`/admin/taxonomies/${type}`, payload);
        if (res.data && res.data.data) {
          items = [res.data.data, ...items];
        }
      } else {
        res = await api.put(
          `/admin/taxonomies/${type}/${editingItem.id}`,
          payload,
        );
        if (res.data) {
          const index = items.findIndex((i) => i.id === editingItem.id);
          if (index !== -1) {
            // If we set this one to current, unset others locally
            if (formCurrent && (type === "seasons" || type === "years")) {
              items = items.map((i) => ({
                ...i,
                current: i.id === editingItem.id,
              }));
            }
            items[index] = { ...items[index], ...payload };
          }
        }
      }
      showModal = false;
    } catch (err: any) {
      console.error(err);
      errorMsg = err.response?.data?.message || "Operation failed";
    } finally {
      isSubmitting = false;
    }
  }

  async function toggleCurrent(id: number) {
    if (type !== "seasons" && type !== "years") return;

    try {
      const res = await api.patch(`/admin/taxonomies/${type}/${id}/current`);
      if (res.data) {
        items = items.map((i) => {
          if (i.id === id) {
            return { ...i, current: !i.current };
          }
          // The backend sets all others to false
          return { ...i, current: false };
        });
      }
    } catch (err: any) {
      console.error(err);
      alert(err.response?.data?.message || "Operation failed");
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Are you sure you want to delete this item?")) return;

    try {
      await api.delete(`/admin/taxonomies/${type}/${id}`);
      items = items.filter((i) => i.id !== id);
    } catch (err: any) {
      alert(err.response?.data?.message || "Delete failed");
    }
  }
</script>

<div
  class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden flex-1 bg-white/2"
>
  <div
    class="p-4 border-b border-white/5 flex justify-between items-center bg-white/5"
  >
    <h2 class="font-bold text-white capitalize">{type}</h2>
    <button
      onclick={openCreate}
      class="px-3 py-1.5 bg-anirank-primary hover:bg-blue-600 text-white text-xs font-semibold rounded-lg transition-colors shadow-lg shadow-anirank-primary/20"
    >
      Add New +
    </button>
  </div>

  <div class="overflow-x-auto">
    <table class="w-full text-left text-sm text-gray-300">
      <thead
        class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5"
      >
        <tr>
          <th class="px-6 py-3 font-semibold">ID</th>
          <th class="px-6 py-3 font-semibold">Name</th>
          <th class="px-6 py-3 font-semibold">Slug / Meta</th>
          <th class="px-6 py-3 font-semibold text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each items as item}
          <tr class="hover:bg-white/2 transition-colors">
            <!-- ID -->
            <td class="px-6 py-3 font-medium text-gray-500"
              >#{item.id || "-"}</td
            >
            <!-- Name -->
            <td class="px-6 py-3 font-medium text-white">{item.name}</td>
            <!-- Slug / Meta -->
            <td class="px-6 py-3 font-mono text-xs text-blue-400">
              {#if item.slug}
                /{item.slug}
              {:else if type === "seasons" || type === "years"}
                <button
                  onclick={() => toggleCurrent(item.id)}
                  class="{item.current
                    ? 'text-green-400'
                    : 'text-gray-500'} hover:text-white transition-colors cursor-pointer"
                >
                  Status: {item.current ? "Current" : "Past"}
                </button>
              {:else}
                -
              {/if}
            </td>
            <!-- Actions -->
            <td class="px-6 py-3 text-right">
              <div class="flex justify-end gap-3 text-lg">
                <button
                  onclick={() => openEdit(item)}
                  class="font-medium text-blue-400 hover:text-blue-300 transition-colors"
                  ><span class="material-symbols-outlined">edit</span></button
                >
                <button
                  onclick={() => handleDelete(item.id)}
                  class="font-medium text-red-400 hover:text-red-300 transition-colors"
                  ><span class="material-symbols-outlined">delete</span></button
                >
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="4" class="px-6 py-8 text-center text-gray-500">
              No data available for this taxonomy.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<!-- Modal -->
{#if showModal}
  <div
    class="fixed inset-0 z-100 flex items-center justify-center p-4 bg-black/80"
  >
    <div
      class="bg-anirank-card border border-white/10 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200"
    >
      <div
        class="px-6 py-4 border-b border-white/5 flex justify-between items-center"
      >
        <h3 class="text-lg font-bold text-white capitalize">
          {modalMode}
          {type.slice(0, -1)}
        </h3>
        <button
          onclick={() => (showModal = false)}
          class="text-gray-400 hover:text-white transition-colors"
        >
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <form onsubmit={handleSubmit} class="p-6 space-y-4">
        {#if errorMsg}
          <div class="p-3 bg-red-500/10 border border-red-500/20 rounded-xl">
            <p class="text-xs text-red-400">{errorMsg}</p>
          </div>
        {/if}

        <div>
          <label
            for="name"
            class="block text-xs font-semibold text-gray-400 uppercase tracking-widest mb-1.5"
            >Name</label
          >
          <input
            id="name"
            type="text"
            bind:value={formName}
            required
            placeholder="e.g. Action, Winter, 2024"
            class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all text-sm"
          />
        </div>

        {#if type === "genres" || type === "formats"}
          <div>
            <label
              for="slug"
              class="block text-xs font-semibold text-gray-400 uppercase tracking-widest mb-1.5"
              >Slug (Optional)</label
            >
            <input
              id="slug"
              type="text"
              bind:value={formSlug}
              placeholder="Leave empty to auto-generate"
              class="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 px-4 text-white placeholder-gray-500 focus:outline-none focus:border-anirank-primary transition-all font-mono text-sm"
            />
          </div>
        {:else}
          <div class="flex items-center gap-3 py-2">
            <input
              id="current"
              type="checkbox"
              bind:checked={formCurrent}
              class="w-4 h-4 rounded border-white/10 bg-white/5 text-anirank-primary focus:ring-anirank-primary focus:ring-offset-zinc-900"
            />
            <label for="current" class="text-sm font-medium text-gray-300"
              >Mark as Current</label
            >
          </div>
        {/if}

        <div class="pt-4 flex gap-3">
          <button
            type="button"
            onclick={() => (showModal = false)}
            class="flex-1 px-4 py-2.5 rounded-xl bg-white/5 hover:bg-white/10 text-gray-300 font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting || !formName}
            class="flex-1 px-4 py-2.5 rounded-xl bg-anirank-primary hover:bg-blue-600 text-white font-bold transition-all disabled:opacity-50 shadow-lg shadow-anirank-primary/20"
          >
            {isSubmitting ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
