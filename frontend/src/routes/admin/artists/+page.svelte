<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";

  let { data } = $props();

  let artists = $derived(data.artists);
  let meta = $derived(data.meta);

  // svelte-ignore state_referenced_locally
  let searchQuery = $state(data.meta.search || "");

  function getQueryString(page: number = 1) {
    const params = new URLSearchParams();
    if (searchQuery) params.set("search", searchQuery);
    params.set("page", page.toString());
    return params.toString();
  }

  function handleSearch() {
    goto(`/admin/artists?${getQueryString(1)}`, { keepFocus: true });
  }

  function changePage(newPage: number) {
    if (newPage >= 1 && newPage <= meta.total_pages) {
      goto(`/admin/artists?${getQueryString(newPage)}`);
    }
  }

  async function handleStatusChange(id: number, currentStatus: boolean) {
    try {
      await api.patch(`/admin/artists/${id}/status`);
      toastState.addToast("Artist status updated", "success");
      // Since it's derived from data, we might need to invalidate or update locally if data isn't re-fetched
      // For now, let's assume the user refresh or we can update local state if we had a non-derived version
      // But usually, we want to update the local list if possible.
      // Since 'artists' is $derived(data.artists), we can't mutate it directly.
      // However, we can use goto to refresh or just let the user see it on next load.
      // Better: use a local state for the list if we want immediate feedback.
      goto(window.location.pathname + window.location.search, { invalidateAll: true });
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to update status", "error");
    }
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Are you sure you want to delete artist "${name}"?`)) return;
    try {
      await api.delete(`/admin/artists/${id}`);
      toastState.addToast("Artist deleted successfully", "success");
      goto(window.location.pathname + window.location.search, { invalidateAll: true });
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to delete artist", "error");
    }
  }

  let selectedIds = $state<number[]>([]);

  function toggleSelectAll(e: Event) {
    const checked = (e.target as HTMLInputElement).checked;
    if (checked) {
      selectedIds = artists.map((a: any) => a.id);
    } else {
      selectedIds = [];
    }
  }

  function toggleSelection(id: number) {
    if (selectedIds.includes(id)) {
      selectedIds = selectedIds.filter((sid) => sid !== id);
    } else {
      selectedIds = [...selectedIds, id];
    }
  }
</script>

<svelte:head>
  <title>Artists Catalog | Admin</title>
</svelte:head>

<div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">Artists Catalog</h1>
    <p class="text-gray-400">Manage musical artists, bands, and their information.</p>
  </div>

  <a
    href="/admin/artists/create"
    class="px-4 py-2 bg-anirank-primary hover:bg-blue-600 text-white font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
  >
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
    </svg>
    New Artist
  </a>
</div>

<!-- Filters & Search -->
<div class="bg-anirank-card border border-white/5 rounded-2xl p-4 mb-6 flex flex-col sm:flex-row gap-4">
  <div class="relative flex-1">
    <svg class="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>
    <input
      type="text"
      bind:value={searchQuery}
      onkeydown={(e) => e.key === "Enter" && handleSearch()}
      placeholder="Search artist name..."
      class="w-full bg-white/5 border border-white/10 rounded-xl py-2 pl-10 pr-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    />
  </div>
  <button
    onclick={handleSearch}
    class="px-6 py-2 bg-anirank-primary hover:bg-blue-600 text-white rounded-xl transition-all font-medium border border-white/10"
  >
    Search
  </button>
</div>

<!-- Table -->
<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden">
  <div class="p-4 border-b border-white/5 flex items-center justify-between">
    <h2 class="text-xl font-semibold text-white">Artists</h2>
    {#if selectedIds.length > 0}
      <span class="text-sm text-gray-400">{selectedIds.length} selected</span>
    {/if}
  </div>

  <div class="overflow-x-auto">
    <table class="w-full text-left text-sm text-gray-300">
      <thead class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5">
        <tr>
          <th class="px-6 py-4 font-semibold">
            <input
              type="checkbox"
              class="rounded border-white/10 bg-white/5 checked:bg-anirank-primary focus:ring-anirank-primary transition-all cursor-pointer"
              onchange={toggleSelectAll}
              checked={selectedIds.length === artists.length && artists.length > 0}
            />
          </th>
          <th class="px-6 py-4 font-semibold">Avatar</th>
          <th class="px-6 py-4 font-semibold">Name</th>
          <th class="px-6 py-4 font-semibold">Songs</th>
          <th class="px-6 py-4 font-semibold text-center">Status</th>
          <th class="px-6 py-4 font-semibold text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each artists as artist}
          <tr class="hover:bg-white/2 transition-colors">
            <td class="px-6 py-4">
              <input
                type="checkbox"
                class="rounded border-white/10 bg-white/5 checked:bg-anirank-primary focus:ring-anirank-primary transition-all cursor-pointer"
                checked={selectedIds.includes(artist.id)}
                onchange={() => toggleSelection(artist.id)}
              />
            </td>
            <td class="px-6 py-4">
              <div class="w-10 h-10 rounded-full bg-white/5 overflow-hidden border border-white/10">
                {#if artist.avatar_url}
                  <img src={artist.avatar_url} alt={artist.name} class="w-full h-full object-cover" />
                {:else}
                  <div class="w-full h-full flex items-center justify-center text-gray-600">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
                {/if}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="font-medium text-white line-clamp-1" title={artist.name}>
                {artist.name}
              </div>
              {#if artist.name_jp}
                <div class="text-xs text-gray-500">{artist.name_jp}</div>
              {/if}
            </td>
            <td class="px-6 py-4">
              <span class="text-blue-400 text-xs font-semibold px-2 py-0.5 rounded-full bg-blue-400/10 border border-blue-400/20">
                {artist.songs_count} themes
              </span>
            </td>
            <td class="px-6 py-4 text-center">
              {#if artist.status}
                <button
                  onclick={() => handleStatusChange(artist.id, artist.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Active
                </button>
              {:else}
                <button
                  onclick={() => handleStatusChange(artist.id, artist.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-orange-500/10 text-orange-400 border border-orange-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-orange-400"></span> Inactive
                </button>
              {/if}
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2 text-lg">
                <a
                  href="/admin/artists/{artist.id}/edit"
                  class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                  title="Edit"
                >
                  <span class="material-symbols-outlined">edit</span>
                </a>
                <button
                  onclick={() => handleDelete(artist.id, artist.name)}
                  class="p-2 text-gray-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors"
                  title="Delete"
                >
                  <span class="material-symbols-outlined">delete</span>
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="px-6 py-12 text-center text-gray-500">No artists found.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if meta?.total_pages > 1}
    <div class="px-6 py-4 border-t border-white/5 flex items-center justify-between">
      <div class="text-sm text-gray-400">
        Showing <span class="font-medium text-white">{artists.length}</span> items
      </div>
      <div class="flex items-center gap-2">
        <button
          disabled={meta.current_page === 1}
          onclick={() => changePage(meta.current_page - 1)}
          aria-label="Previous Page"
          class="p-2 rounded-lg border border-white/10 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/5 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <span class="text-sm text-gray-300 font-medium px-2">Page {meta.current_page} of {meta.total_pages}</span>
        <button
          disabled={meta.current_page === meta.total_pages}
          onclick={() => changePage(meta.current_page + 1)}
          aria-label="Next Page"
          class="p-2 rounded-lg border border-white/10 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/5 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
    </div>
  {/if}
</div>
