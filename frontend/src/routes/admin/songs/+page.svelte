<script lang="ts">
  import { goto } from "$app/navigation";
  import { getSongName } from "$lib/song-utils";
  import api from "$lib/api";
  import AutocompleteAnime from "$lib/components/admin/AutocompleteAnime.svelte";
  import { configState } from "$lib/state/config.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import Plus from "lucide-svelte/icons/plus";
  import Search from "lucide-svelte/icons/search";
  import Video from "lucide-svelte/icons/video";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import ChevronLeft from "lucide-svelte/icons/chevron-left";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import AlertCircle from "lucide-svelte/icons/alert-circle";

  let { data } = $props();

  let songs = $derived(data.songs);
  let pagination = $derived(data.pagination);
  let filters = $derived(data.filters);

  // svelte-ignore state_referenced_locally
  let searchQuery = $state(data.filters.search || "");
  // svelte-ignore state_referenced_locally
  let animeIdInput = $state(data.filters.anime || "");
  // svelte-ignore state_referenced_locally
  let statusFilter = $state(data.filters.status || "");
  // svelte-ignore state_referenced_locally
  let typeFilter = $state(data.filters.type || "");

  $effect(() => {
    searchQuery = data.filters.search || "";
    animeIdInput = data.filters.anime || "";
    statusFilter = data.filters.status || "";
    typeFilter = data.filters.type || "";
  });

  function handleSearch() {
    goto(`/admin/songs?search=${searchQuery}&anime=${animeIdInput}&status=${statusFilter}&type=${typeFilter}&page=1`, { keepFocus: true });
  }

  function changePage(newPage: number) {
    if (newPage >= 1 && newPage <= pagination.last_page) {
      goto(`/admin/songs?search=${searchQuery}&anime=${animeIdInput}&status=${statusFilter}&type=${typeFilter}&page=${newPage}`);
    }
  }

  async function handleStatusChange(id: number, currentStatus: boolean) {
    try {
      await api.patch(`/admin/songs/${id}/status`);
      toastState.addToast("Song status updated", "success");
      goto(window.location.pathname + window.location.search, { invalidateAll: true });
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to update status", "error");
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Are you sure you want to delete this song?")) return;
    try {
      await api.delete(`/admin/songs/${id}`);
      toastState.addToast("Song deleted successfully", "success");
      goto(window.location.pathname + window.location.search, { invalidateAll: true });
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to delete song", "error");
    }
  }

  let selectedIds = $state<number[]>([]);

  function toggleSelectAll(e: Event) {
    const checked = (e.target as HTMLInputElement).checked;
    if (checked) {
      selectedIds = songs.map((s: any) => s.id);
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
  <title>Songs Catalog | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Songs Catalog
    </h1>
    <p class="text-on-surface-variant/70">
      Manage anime themes (Openings, Endings, Inserts) and associations.
    </p>
  </div>

  <a
    href="/admin/songs/create"
    class="px-4 py-2 bg-primary hover:bg-primary-container text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
  >
    <Plus size={20} />

    New Song
  </a>
</div>

<!-- Filters Bar -->
<div
  class="flex flex-wrap gap-4 items-end bg-surface-container/30 p-4 rounded-2xl border border-outline-variant mb-6"
>
  <!-- Search -->
  <div class="flex-1 min-w-[200px]">
    <label
      for="search"
      class="block text-xs font-medium text-on-surface-variant/40 mb-1 uppercase tracking-wider"
      >Search Title</label
    >
    <div class="relative">
      <Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant/40" />

      <input
        id="search"
        type="text"
        bind:value={searchQuery}
        onkeydown={(e) => e.key === "Enter" && handleSearch()}
        placeholder="Filter by title..."
        class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2 pl-9 pr-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors text-sm"
      />
    </div>
  </div>

  <!-- Anime -->
  <div class="w-full sm:w-72">
    <label
      for="anime-filter"
      class="block text-xs font-medium text-on-surface-variant/40 mb-1 uppercase tracking-wider"
      >Anime</label
    >
    <AutocompleteAnime
      id="anime-filter"
      showLabel={false}
      bind:value={animeIdInput}
      onselect={() => handleSearch()}
      placeholder="Filter by anime..."
    />
  </div>

  <!-- Status -->
  <div class="w-full sm:w-40">
    <label
      for="status"
      class="block text-xs font-medium text-on-surface-variant/40 mb-1 uppercase tracking-wider"
      >Status</label
    >
    <select
      id="status"
      bind:value={statusFilter}
      onchange={() => handleSearch()}
      class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2 px-3 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors text-sm appearance-none cursor-pointer"
    >
      <option value="">All Status</option>
      <option value="true">Active Only</option>
      <option value="false">Inactive Only</option>
    </select>
  </div>

  <!-- Type -->
  <div class="w-full sm:w-40">
    <label
      for="type-filter"
      class="block text-xs font-medium text-on-surface-variant/40 mb-1 uppercase tracking-wider"
      >Type</label
    >
    <select
      id="type-filter"
      bind:value={typeFilter}
      onchange={() => handleSearch()}
      class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2 px-3 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors text-sm appearance-none cursor-pointer"
    >
      <option value="">All Types</option>
      {#each configState.songTypes as type}
        <option value={type.slug}>{type.name}</option>
      {/each}
    </select>
  </div>

  <button
    onclick={() => {
      searchQuery = "";
      animeIdInput = "";
      statusFilter = "";
      typeFilter = "";
      handleSearch();
    }}
    class="px-4 py-2 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-xl transition-colors border border-outline-variant text-sm font-medium h-[38px]"
  >
    Reset
  </button>
</div>

<!-- Table -->
<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
  <div class="p-4 border-b border-outline-variant flex items-center justify-between">
    <h2 class="text-xl font-semibold text-on-surface">Songs</h2>
    {#if selectedIds.length > 0}
      <span class="text-sm text-on-surface-variant/70">{selectedIds.length} selected</span>
    {/if}
  </div>

  <div class="overflow-x-auto">
    <table class="w-full text-left text-sm text-on-surface-variant">
      <thead
        class="text-xs text-on-surface-variant/70 uppercase bg-surface-highest border-b border-outline-variant"
      >
        <tr>
          <th class="px-6 py-4 font-semibold">
            <input
              type="checkbox"
              class="rounded border-outline-variant bg-surface-highest checked:bg-primary focus:ring-primary transition-all cursor-pointer"
              onchange={toggleSelectAll}
              checked={selectedIds.length === songs.length && songs.length > 0}
            />
          </th>
          <th class="px-6 py-4 font-semibold">Title (Romaji)</th>
          <th class="px-6 py-4 font-semibold">Type</th>
          <th class="px-6 py-4 font-semibold">Anime</th>
          <th class="px-6 py-4 font-semibold text-center">Status</th>
          <th class="px-6 py-4 font-semibold text-right">Views</th>
          <th class="px-6 py-4 font-semibold text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each songs as song}
          <tr class="hover:bg-white/2 transition-colors">
            <td class="px-6 py-4">
              <input
                type="checkbox"
                class="rounded border-outline-variant bg-surface-highest checked:bg-primary focus:ring-primary transition-all cursor-pointer"
                checked={selectedIds.includes(song.id)}
                onchange={() => toggleSelection(song.id)}
              />
            </td>
            <td class="px-6 py-4">
              <div
                class="font-medium text-on-surface line-clamp-1"
                title={getSongName(song)}
              >
                <a href="/admin/songs/{song.id}">
                  {getSongName(song)}
                </a>
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-2">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20"
                >
                  {song.type}
                  {song.theme_num}
                </span>
                {#if song.partial_artist_inactive}
                  <div class="group relative">
                    <AlertCircle size={16} class="text-amber-500" />

                    <div
                      class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-on-surface text-[10px] rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10 border border-outline-variant"
                    >
                      Partial Availability: Some artists are inactive
                    </div>
                  </div>
                {/if}
              </div>
            </td>
            <td class="px-6 py-4">
              <div
                class="line-clamp-1 text-on-surface-variant/70 max-w-[200px]"
                title={song.anime?.title}
              >
                {#if song.anime}
                  <a href={`/admin/animes/${song.anime.id}`}>
                    {song.anime.title}
                  </a>
                {:else}
                  <span class="text-on-surface-variant/40">Unknown Anime</span>
                {/if}
              </div>
            </td>
            <td class="px-6 py-4 text-center">
              {#if song.status}
                <button
                  onclick={() => handleStatusChange(song.id, song.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500/20 transition-colors"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Active
                </button>
              {:else}
                <button
                  onclick={() => handleStatusChange(song.id, song.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-orange-500/10 text-orange-400 border border-orange-500/20 hover:bg-orange-500/20 transition-colors"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-orange-400"></span> Inactive
                </button>
              {/if}
            </td>
            <td class="px-6 py-4 text-right">
              {song.views?.toLocaleString() || 0}
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2 text-lg">
                <a
                  href="/admin/songs/{song.id}/variants"
                  class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest rounded-lg transition-colors"
                  title="Variants"
                >
                  <Video size={18} />
                </a>
                <a
                  href="/admin/songs/{song.id}/edit"
                  class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest rounded-lg transition-colors"
                  title="Edit"
                >
                  <Edit2 size={18} />
                </a>
                <button
                  onclick={() => handleDelete(song.id)}
                  class="p-2 text-on-surface-variant/70 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors"
                  title="Delete"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="px-6 py-12 text-center text-on-surface-variant/40">
              No songs found.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if pagination?.last_page > 1}
    <div
      class="px-6 py-4 border-t border-outline-variant flex items-center justify-between"
    >
      <div class="text-sm text-on-surface-variant/70">
        Showing <span class="font-medium text-on-surface">{songs.length}</span> items
      </div>
      <div class="flex items-center gap-2">
        <button
          disabled={pagination.current_page === 1}
          onclick={() => changePage(pagination.current_page - 1)}
          aria-label="Previous Page"
          class="p-2 rounded-lg border border-outline-variant text-on-surface-variant/70 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-highest transition-colors"
        >
          <ChevronLeft size={16} />

        </button>
        <span class="text-sm text-on-surface-variant font-medium px-2"
          >Page {pagination.current_page} of {pagination.last_page}</span
        >
        <button
          disabled={pagination.current_page === pagination.last_page}
          onclick={() => changePage(pagination.current_page + 1)}
          aria-label="Next Page"
          class="p-2 rounded-lg border border-outline-variant text-on-surface-variant/70 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-highest transition-colors"
        >
          <ChevronRight size={16} />

        </button>
      </div>
    </div>
  {/if}
</div>
