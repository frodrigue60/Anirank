<script lang="ts">
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import api from "$lib/api";

  let { data } = $props();
  let years = $state<any[]>([]);

  // Sync prop to state
  $effect(() => {
    years = data.years || [];
  });

  // Modal State
  let editingYear = $state<any>(null);
  let showEditModal = $state(false);
  let isUpdating = $state(false);

  async function toggleStatus(year: any) {
    try {
      await api.patch(`/admin/taxonomies/years/${year.id}/current`);

      // Update local state: only one can be current
      const newCurrent = !year.current;
      years = years.map((y: any) => ({
        ...y,
        current:
          y.id === year.id ? newCurrent : newCurrent ? false : y.current,
      }));

      toastState.addToast(
        `Year ${year.name} set as ${newCurrent ? "current" : "default"}`,
        "success",
      );
    } catch (err) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update status"),
        "error",
      );
    }
  }

  function editYear(year: any) {
    editingYear = { ...year };
    showEditModal = true;
  }

  async function saveYear() {
    if (!editingYear) return;
    isUpdating = true;
    try {
      await api.put(`/admin/taxonomies/years/${editingYear.id}`, {
        name: editingYear.name,
        current: editingYear.current,
      });

      // Update local state
      years = years.map((y: any) => (y.id === editingYear.id ? editingYear : y));

      toastState.addToast("Year updated successfully", "success");
      showEditModal = false;
    } catch (err) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update year"),
        "error",
      );
    } finally {
      isUpdating = false;
    }
  }

  async function deleteYear(year: any) {
    if (!confirm(`Are you sure you want to delete the year ${year.name}?`))
      return;

    try {
      await api.delete(`/admin/taxonomies/years/${year.id}`);
      years = years.filter((y: any) => y.id !== year.id);
      toastState.addToast("Year deleted successfully", "success");
    } catch (err) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to delete year"),
        "error",
      );
    }
  }
</script>

<svelte:head>
  <title>Years | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Years Management
    </h1>
    <p class="text-on-surface-variant">Manage release years for content.</p>
  </div>
</div>

<div class="grid grid-cols-1 gap-6">
  <div
    class="bg-surface-container border border-outline-variant/10 rounded-3xl overflow-hidden shadow-xl"
  >
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-surface-highest/50 border-b border-outline-variant/10">
            <th
              class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant"
              >ID</th
            >
            <th
              class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant"
              >Release Year</th
            >
            <th
              class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant text-center"
              >Status</th
            >
            <th
              class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant text-right"
              >Actions</th
            >
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/5">
          {#each years as year (year.id)}
            <tr class="hover:bg-primary/5 transition-colors group">
              <td class="px-6 py-4">
                <span class="text-xs font-mono text-on-surface-variant"
                  >#{year.id}</span
                >
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div
                    class="w-10 h-10 rounded-xl bg-surface-highest flex items-center justify-center text-primary font-black shadow-inner"
                  >
                    {year.name && year.name.toString().length >= 2 ? year.name.toString().slice(-2) : '??'}
                  </div>
                  <span class="text-lg font-bold text-on-surface"
                    >{year.name}</span
                  >
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="flex justify-center">
                  <button
                    onclick={() => toggleStatus(year)}
                    class="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all {year.current
                      ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20'
                      : 'bg-on-surface-variant/10 text-on-surface-variant border border-transparent hover:border-on-surface-variant/20'}"
                  >
                    <span
                      class="material-symbols-outlined text-sm {year.current
                        ? 'filled'
                        : ''}"
                    >
                      {year.current ? "star" : "star_outline"}
                    </span>
                    <span class="text-[10px] font-bold uppercase"
                      >{year.current ? "Current" : "Default"}</span
                    >
                  </button>
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="flex justify-end gap-2">
                  <button
                    onclick={() => editYear(year)}
                    class="p-2 rounded-xl bg-primary/10 text-primary hover:bg-primary hover:text-white transition-all shadow-sm"
                    title="Edit Year"
                  >
                    <span class="material-symbols-outlined text-sm">edit</span>
                  </button>
                  <button
                    onclick={() => deleteYear(year)}
                    class="p-2 rounded-xl bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white transition-all shadow-sm"
                    title="Delete Year"
                  >
                    <span class="material-symbols-outlined text-sm">delete</span
                    >
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

<!-- Edit Modal -->
{#if showEditModal && editingYear}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in"
  >
    <div
      class="bg-surface-container border border-outline-variant/20 rounded-[2.5rem] w-full max-w-md shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
    >
      <div class="p-8">
        <div class="flex justify-between items-start mb-6">
          <div>
            <h3 class="text-2xl font-black text-on-surface">Edit Year</h3>
            <p class="text-on-surface-variant text-sm">Modify year parameters</p>
          </div>
          <button
            onclick={() => (showEditModal = false)}
            class="p-2 hover:bg-surface-highest rounded-full transition-colors text-on-surface-variant"
          >
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>

        <div class="space-y-6">
          <div class="space-y-2">
            <label
              for="editName"
              class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant ml-1"
              >Year Name</label
            >
            <input
              id="editName"
              type="text"
              bind:value={editingYear.name}
              placeholder="e.g. 2024"
              class="w-full bg-surface-low border border-outline-variant/20 rounded-2xl px-5 py-4 text-on-surface focus:outline-none focus:border-primary transition-all font-bold"
            />
          </div>

          <div class="flex items-center gap-3 p-4 bg-surface-low rounded-2xl border border-outline-variant/10">
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" bind:checked={editingYear.current} class="sr-only peer">
              <div class="w-11 h-6 bg-on-surface-variant/20 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:inset-s-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
            </label>
            <div>
              <span class="text-sm font-bold text-on-surface">Mark as Current</span>
              <p class="text-[10px] text-on-surface-variant">Sets this year as the primary system year.</p>
            </div>
          </div>
        </div>

        <div class="mt-8 grid grid-cols-2 gap-4">
          <button
            onclick={() => (showEditModal = false)}
            class="py-4 rounded-2xl text-on-surface font-bold hover:bg-surface-highest transition-all"
          >
            Cancel
          </button>
          <button
            onclick={saveYear}
            disabled={isUpdating}
            class="py-4 bg-primary text-white rounded-2xl font-bold shadow-lg shadow-primary/25 hover:scale-[1.02] active:scale-95 transition-all flex items-center justify-center gap-2"
          >
            {#if isUpdating}
              <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            {/if}
            Save Changes
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
