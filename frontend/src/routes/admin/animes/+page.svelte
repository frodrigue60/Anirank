<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";

  import { onMount } from "svelte";

  let { data } = $props();

  let animes = $derived(data.animes);
  let pagination = $derived(data.pagination);
  let filters = $derived(data.filters);
  let years = $derived(data.years);
  let seasons = $derived(data.seasons);
  let formats = $derived(data.formats);

  let searchQuery = $state("");
  let selectedYear = $state("");
  let selectedSeason = $state("");
  let selectedFormat = $state("");
  let selectedStatus = $state("");

  onMount(async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get("code");
    if (code) {
      try {
        const response = await api.post("/auth/anilist/callback", { code });
        if (response.data.success) {
          toastState.addToast(
            "Anilist account linked successfully!",
            "success",
          );
          goto("/settings");
        }
      } catch (err: any) {
        toastState.addToast(
          getApiErrorMessage(err, "Failed to link Anilist account"),
          "error",
        );
      }
    }
  });

  $effect(() => {
    searchQuery = data.filters.search || "";
    selectedYear = data.filters.year || "";
    selectedSeason = data.filters.season || "";
    selectedFormat = data.filters.format || "";
    selectedStatus = data.filters.status || "";
  });

  function getQueryString(page: number = 1) {
    const params = new URLSearchParams();
    if (searchQuery) params.set("search", searchQuery);
    if (selectedYear) params.set("year", selectedYear);
    if (selectedSeason) params.set("season", selectedSeason);
    if (selectedFormat) params.set("format", selectedFormat);
    if (selectedStatus) params.set("status", selectedStatus);
    params.set("page", page.toString());
    return params.toString();
  }

  function handleSearch() {
    goto(`/admin/animes?${getQueryString(1)}`, { keepFocus: true });
  }

  function changePage(newPage: number) {
    if (newPage >= 1 && newPage <= pagination.last_page) {
      goto(`/admin/animes?${getQueryString(newPage)}`);
    }
  }

  async function handleStatusChange(
    id: number,
    currentStatus: boolean | number,
  ) {
    try {
      await api.patch(`/admin/animes/${id}/status`);
      // Update local state reactively
      animes = animes.map((a: any) => {
        if (a.id === id) {
          return { ...a, status: !currentStatus };
        }
        return a;
      });
      toastState.addToast("Anime status updated", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to update status"), "error");
    }
  }

  async function handleDelete(id: number, title: string) {
    if (!confirm(`Are you sure you want to delete "${title}"?`)) return;

    try {
      await api.delete(`/admin/animes/${id}`);
      animes = animes.filter((a: any) => a.id !== id);
      toastState.addToast(`Anime "${title}" deleted successfully`, "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to delete anime"), "error");
    }
  }

  // --- Multi-selection logic ---
  let selectedIds = $state<number[]>([]);

  function toggleSelectAll(e: Event) {
    const checked = (e.target as HTMLInputElement).checked;
    if (checked) {
      selectedIds = animes.map((a: any) => a.id);
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

  async function handleBatchDelete() {
    if (selectedIds.length === 0) return;
    if (
      !confirm(
        `Are you sure you want to delete ${selectedIds.length} selected animes?`,
      )
    )
      return;

    try {
      await api.post("/admin/animes/batch-delete", { ids: selectedIds });
      animes = animes.filter((a: any) => !selectedIds.includes(a.id));
      toastState.addToast(
        `${selectedIds.length} animes deleted successfully`,
        "success",
      );
      selectedIds = [];
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to delete animes"), "error");
    }
  }
</script>

<svelte:head>
  <title>Animes Catalog | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      Animes Catalog
    </h1>
    <p class="text-gray-400">
      Manage anime entries, statuses, and Anilist relationships.
    </p>
  </div>

  <a
    href="/admin/animes/create"
    class="px-4 py-2 bg-anirank-primary hover:bg-blue-600 text-white font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
  >
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M12 4v16m8-8H4"
      /></svg
    >
    New Anime
  </a>
</div>

<!-- Search -->
<div
  class="bg-anirank-card border border-white/5 rounded-2xl p-4 mb-6 flex flex-col sm:flex-row gap-4"
>
  <div class="relative flex-1">
    <input
      type="text"
      bind:value={searchQuery}
      onkeydown={(e) => e.key === "Enter" && handleSearch()}
      placeholder="Search by title..."
      class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    />
  </div>
  <div class="flex-1 min-w-[150px]">
    <select
      bind:value={selectedYear}
      onchange={handleSearch}
      class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Years</option>
      {#each years as y}
        <option value={y.id}>{y.name}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[150px]">
    <select
      bind:value={selectedSeason}
      onchange={handleSearch}
      class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Seasons</option>
      {#each seasons as s}
        <option value={s.id}>{s.name}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[150px]">
    <select
      bind:value={selectedFormat}
      onchange={handleSearch}
      class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Formats</option>
      {#each formats as f}
        <option value={f.id}>{f.name}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[150px]">
    <select
      bind:value={selectedStatus}
      onchange={handleSearch}
      class="w-full bg-white/5 border border-white/10 rounded-xl p-2 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Status</option>
      <option value="true">Active</option>
      <option value="false">Inactive</option>
    </select>
  </div>
  <div class="flex gap-2">
    <button
      onclick={handleSearch}
      class="px-6 py-2 bg-anirank-primary hover:bg-blue-600 text-white rounded-xl transition-all font-medium border border-white/10"
    >
      Filter
    </button>
  </div>
</div>

<!-- Table -->
<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden">
  <div class="p-4 border-b border-white/5 flex items-center justify-between">
    <div class="flex items-center gap-4">
      <h2 class="text-xl font-semibold text-white">Animes</h2>
      {#if selectedIds.length > 0}
        <div
          class="flex items-center gap-2 animate-in fade-in slide-in-from-left-2"
        >
          <span class="text-sm text-gray-400"
            >{selectedIds.length} selected</span
          >
          <button
            onclick={handleBatchDelete}
            class="px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-medium rounded-lg transition-colors border border-red-500/20 flex items-center gap-1.5"
          >
            <span class="material-symbols-outlined text-base">delete</span>
            Delete selected
          </button>
        </div>
      {/if}
    </div>
  </div>
  <div class="overflow-x-auto">
    <table class="w-full text-left text-sm text-gray-300">
      <thead
        class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5"
      >
        <tr>
          <th class="px-6 py-4 font-semibold text-gray-400">
            <input
              type="checkbox"
              class="rounded border-white/10 bg-white/5 checked:bg-anirank-primary focus:ring-anirank-primary transition-all cursor-pointer"
              onchange={toggleSelectAll}
              checked={selectedIds.length === animes.length &&
                animes.length > 0}
            />
          </th>
          <th class="px-6 py-4 font-semibold">Cover</th>
          <th class="px-6 py-4 font-semibold">Title</th>
          <th class="px-6 py-4 font-semibold">Songs</th>
          <th class="px-6 py-4 font-semibold text-center">Status</th>
          <th class="px-6 py-4 font-semibold">Format</th>
          <th class="px-6 py-4 font-semibold text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each animes as anime}
          <tr class="hover:bg-white/2 transition-colors">
            <td class="px-6 py-4 font-medium text-gray-500">
              <input
                type="checkbox"
                class="rounded border-white/10 bg-white/5 checked:bg-anirank-primary focus:ring-anirank-primary transition-all cursor-pointer"
                checked={selectedIds.includes(anime.id)}
                onchange={() => toggleSelection(anime.id)}
              />
            </td>
            <td class="px-6 py-4">
              <img
                src={anime.cover_url || "/images/placeholders/anime-cover.png"}
                alt="{anime.title} cover"
                class="w-10 h-14 object-cover rounded shadow-sm"
              />
            </td>
            <td class="px-6 py-4">
              <div
                class="font-medium text-white line-clamp-1"
                title={anime.title}
              >
                <a href="/admin/animes/{anime.id}">{anime.title}</a>
              </div>
              <div class="text-xs text-gray-500 mt-1 flex gap-2">
                {#if anime.year?.name}<span>Year: {anime.year.name}</span>{/if}
                {#if anime.season?.name}<span>Season: {anime.season.name}</span
                  >{/if}
                {#if anime.anilist_id}<span
                    class="text-blue-400 text-[10px] uppercase border border-blue-400/30 rounded px-1"
                    >{anime.anilist_id}</span
                  >{/if}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="text-xs text-gray-500 mt-1 flex gap-2">
                <span
                  class="text-blue-400 text-[10px] uppercase border border-blue-400/30 rounded px-1"
                  >{anime.songs_count}</span
                >
              </div>
            </td>
            <td class="px-6 py-4 text-center">
              {#if anime.status === true || anime.status === 1}
                <button
                  onclick={() => handleStatusChange(anime.id, anime.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Active
                </button>
              {:else}
                <button
                  onclick={() => handleStatusChange(anime.id, anime.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-orange-500/10 text-orange-400 border border-orange-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-orange-400"></span> Draft
                </button>
              {/if}
            </td>
            <td class="px-6 py-4">
              {anime.format?.name || "N/A"}
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2 text-lg">
                <a
                  href="/admin/songs?anime={anime.id}"
                  class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                  title="Songs"
                >
                  <span class="material-symbols-outlined">music_note</span>
                </a>
                <a
                  href="/admin/songs/create?anime={anime.id}"
                  class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                  title="Add Song"
                >
                  <span class="material-symbols-outlined">add</span>
                </a>
                <a
                  href="/admin/animes/{anime.id}/edit"
                  class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                  title="Edit"
                >
                  <span class="material-symbols-outlined">edit</span>
                </a>
                <button
                  onclick={() => handleDelete(anime.id, anime.title)}
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
            <td colspan="6" class="px-6 py-12 text-center text-gray-500">
              No animes found.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if pagination?.last_page > 1}
    <div
      class="px-6 py-4 border-t border-white/5 flex items-center justify-between"
    >
      <div class="text-sm text-gray-400">
        Showing <span class="font-medium text-white">{animes.length}</span> items
      </div>
      <div class="flex items-center gap-2">
        <button
          disabled={pagination.current_page === 1}
          onclick={() => changePage(pagination.current_page - 1)}
          aria-label="Previous Page"
          class="p-2 rounded-lg border border-white/10 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/5 transition-colors"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 19l-7-7 7-7"
            /></svg
          >
        </button>
        <span class="text-sm text-gray-300 font-medium px-2"
          >Page {pagination.current_page} of {pagination.last_page}</span
        >
        <button
          disabled={pagination.current_page === pagination.last_page}
          onclick={() => changePage(pagination.current_page + 1)}
          aria-label="Next Page"
          class="p-2 rounded-lg border border-white/10 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/5 transition-colors"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5l7 7-7 7"
            /></svg
          >
        </button>
      </div>
    </div>
  {/if}
</div>
