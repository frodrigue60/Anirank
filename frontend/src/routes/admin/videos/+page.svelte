<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import AutocompleteAnime from "$lib/components/admin/AutocompleteAnime.svelte";
  import { getSongName } from "$lib/song-utils";

  let { data } = $props<{ data: PageData }>();
  let videos = $state<any[]>([]);
  let pagination = $derived(data.pagination);
  let filters = $derived(data.filters);

  let animeIdInput = $state("");
  let statusFilter = $state("");

  $effect(() => {
    videos = data.data;
    animeIdInput = filters?.anime || "";
    statusFilter = filters?.status !== undefined ? String(filters.status) : "";
  });

  async function handleStatusChange(id: number, currentStatus: boolean) {
    try {
      await api.patch(`/admin/variants/${id}/status`);
      // Update local state reactively
      videos = videos.map((v: any) => {
        if (v.id === id) {
          return { ...v, status: !currentStatus };
        }
        return v;
      });
      toastState.addToast("Video status updated", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        `Failed to update status: ${err.message || err}`,
        "error",
      );
    }
  }

  function handleSearch(e?: Event) {
    if (e) e.preventDefault();
    const url = new URL($page.url);

    if (animeIdInput) {
      url.searchParams.set("anime", animeIdInput);
    } else {
      url.searchParams.delete("anime");
    }

    url.searchParams.delete("anime_id");
    url.searchParams.delete("anime-id");

    if (statusFilter !== "") {
      url.searchParams.set("status", statusFilter);
    } else {
      url.searchParams.delete("status");
    }

    url.searchParams.set("page", "1");
    goto(url.toString());
  }

  function changePage(newPage: number) {
    if (newPage < 1 || newPage > (pagination?.last_page || 1)) return;
    const url = new URL($page.url);
    url.searchParams.set("page", newPage.toString());
    goto(url.toString());
  }

  function getSourceType(variant: any) {
    if (variant.video?.type === "embed") return "Embed";
    if (variant.video?.type === "file") return "Direct File";
    if (variant.video?.embed_url) return "Embed";
    if (variant.video?.local_url) return "Direct File";
    return "None";
  }
</script>

<svelte:head>
  <title>Video Management | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      Video Management
    </h1>
    <p class="text-gray-400">
      Review and moderate video sources across the platform.
    </p>
  </div>
</div>

<div class="mb-6 flex flex-col gap-6">
  <!-- Filters Bar -->
  <div
    class="flex flex-wrap gap-4 items-end bg-anirank-card/30 p-4 rounded-2xl border border-white/5"
  >
    <!-- Anime Autocomplete -->
    <div class="w-full sm:w-72">
      <AutocompleteAnime
        bind:value={animeIdInput}
        onselect={() => handleSearch()}
        placeholder="Filter by anime..."
      />
    </div>

    <!-- Status -->
    <div class="w-full sm:w-40">
      <label
        for="status"
        class="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wider"
        >Status</label
      >
      <select
        id="status"
        bind:value={statusFilter}
        onchange={() => handleSearch()}
        class="w-full bg-anirank-card border border-white/5 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary focus:ring-1 focus:ring-anirank-primary transition-all appearance-none cursor-pointer"
      >
        <option value="">All Status</option>
        <option value="true">Active Only</option>
        <option value="false">Inactive Only</option>
      </select>
    </div>

    <!-- Reset -->
    <button
      onclick={() => {
        animeIdInput = "";
        statusFilter = "";
        const url = new URL($page.url);
        url.search = "";
        goto(url.toString());
      }}
      class="h-10 px-4 text-gray-500 hover:text-white transition-colors text-sm flex items-center"
    >
      Reset Filters
    </button>
  </div>
</div>

<div
  class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden shadow-xl"
>
  <div class="overflow-x-auto">
    <table class="w-full text-left border-collapse">
      <thead>
        <tr
          class="bg-white/2 border-b border-white/5 text-[10px] uppercase font-black tracking-widest text-white/40"
        >
          <th class="py-4 px-6 md:px-8">ID</th>
          <th class="py-4 px-6">Source</th>
          <th class="py-4 px-6">Type / Vers</th>
          <th class="py-4 px-6">Anime / Song</th>
          <th class="py-4 px-6">Status</th>
          <th class="py-4 px-6 text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5 text-sm">
        {#each videos as variant (variant.id)}
          <tr class="hover:bg-white/2 transition-colors group">
            <!-- ID -->
            <td class="py-4 px-6 text-gray-500">#{variant.id}</td>
            <!-- Source -->
            <td class="py-4 px-6 whitespace-nowrap">
              <span
                class="inline-flex items-center px-2 py-1 rounded text-[10px] font-bold uppercase tracking-tighter {variant
                  .video?.type === 'embed'
                  ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                  : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'}"
              >
                {getSourceType(variant)}
              </span>
            </td>
            <!-- Type / Vers -->
            <td class="py-4 px-6 whitespace-nowrap">
              <div class="text-white font-medium">{variant.slug}</div>
              <div class="text-[11px] text-gray-500 uppercase tracking-widest">
                Version {variant.version_number}
              </div>
            </td>
            <!-- Anime / Song -->
            <td class="py-4 px-6">
              {#if variant.song?.anime}
                <div
                  class="text-white text-xs font-semibold truncate max-w-[200px]"
                >
                  {variant.song.anime.title}
                </div>
              {/if}
              <div
                class="text-[11px] text-gray-400 truncate max-w-[200px] flex flex-col"
              >
                {getSongName(variant.song)}
                {variant.song.anime.title}
              </div>
            </td>
            <!-- Status -->
            <td class="py-4 px-6 whitespace-nowrap">
              {#if variant.status === true || variant.status === 1}
                <button
                  onclick={() => handleStatusChange(variant.id, variant.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                >
                  <span
                    class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"
                  ></span> Published
                </button>
              {:else}
                <button
                  onclick={() => handleStatusChange(variant.id, variant.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-orange-500/10 text-orange-400 border border-orange-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-orange-400"></span> Draft
                </button>
              {/if}
            </td>
            <!-- Actions -->
            <td class="py-4 px-6 whitespace-nowrap text-right">
              <a
                href="/admin/variants/{variant.id}/edit"
                class="px-3 py-1.5 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors inline-block text-xs font-bold border border-white/5"
              >
                Manage
              </a>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="py-16 text-center">
              <div class="flex flex-col items-center gap-3">
                <div
                  class="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center text-gray-400"
                >
                  <svg
                    class="w-6 h-6"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    /></svg
                  >
                </div>
                <p class="text-gray-500 text-sm">
                  No videos found matching your criteria.
                </p>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if pagination && pagination.last_page > 1}
    <div
      class="border-t border-white/5 px-6 py-4 flex items-center justify-between bg-white/1"
    >
      <div class="text-xs text-gray-500 font-medium">
        Page <span class="text-white">{pagination.current_page}</span> of {pagination.last_page}
      </div>
      <div class="flex items-center gap-2">
        <button
          onclick={() => changePage(pagination.current_page - 1)}
          disabled={pagination.current_page === 1}
          aria-label="Previous Page"
          class="p-2 rounded-lg bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white disabled:opacity-30 transition-all"
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
        <button
          onclick={() => changePage(pagination.current_page + 1)}
          disabled={pagination.current_page === pagination.last_page}
          aria-label="Next Page"
          class="p-2 rounded-lg bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white disabled:opacity-30 transition-all"
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
