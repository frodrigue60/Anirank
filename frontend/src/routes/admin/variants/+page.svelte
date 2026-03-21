<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import type { PageData } from "./$types";
  import AutocompleteAnime from "$lib/components/admin/AutocompleteAnime.svelte";

  let { data } = $props<{ data: PageData }>();
  let variants = $derived(data.data);
  let meta = $derived(data.meta);

  let animeIdInput = $state("");
  let statusFilter = $state("");

  $effect(() => {
    animeIdInput = meta?.anime || "";
    statusFilter = meta?.status !== undefined ? String(meta.status) : "";
  });

  function handleSearch(e?: Event) {
    if (e) e.preventDefault();
    const url = new URL($page.url);

    // Anime Filter
    if (animeIdInput.trim()) {
      url.searchParams.set("anime", animeIdInput.trim());
    } else {
      url.searchParams.delete("anime");
    }

    // Standardize: always remove legacy tags if any
    url.searchParams.delete("anime_id");
    url.searchParams.delete("anime-id");

    // Status
    if (statusFilter !== "") {
      url.searchParams.set("status", statusFilter);
    } else {
      url.searchParams.delete("status");
    }

    url.searchParams.set("page", "1");
    goto(url.toString());
  }

  function changePage(newPage: number) {
    if (newPage < 1 || newPage > (meta?.total_pages || 1)) return;
    const url = new URL($page.url);
    url.searchParams.set("page", newPage.toString());
    goto(url.toString());
  }

  // Helper to format source type
  function getSourceType(variant: any) {
    if (variant.video?.type === "embed") return "Embed";
    if (variant.video?.type === "file") return "Direct File";
    if (variant.video?.embed_url) return "Embed";
    if (variant.video?.local_url) return "Direct File";
    return "None";
  }
</script>

<svelte:head>
  <title>Song Variants | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      Song Variants
    </h1>
    <p class="text-gray-400">Manage video sources and versions for songs.</p>
  </div>

  <a
    href="/admin/variants/create"
    class="px-4 py-2 bg-anirank-primary hover:bg-blue-600 text-white font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
  >
    <span class="material-symbols-outlined">add</span>
    New Variant
  </a>
</div>

<div
  class="mb-6 flex flex-col sm:flex-row gap-4 justify-between items-start sm:items-center"
>
  <!-- Filters -->
  <div class="flex flex-wrap gap-4 items-end w-full">
    <!-- Anime Autocomplete -->
    <div class="w-full sm:w-72">
      <AutocompleteAnime
        bind:value={animeIdInput}
        onselect={() => handleSearch()}
        placeholder="Filter by anime..."
      />
    </div>

    <!-- Status -->
    <div class="w-full sm:w-36">
      <label
        for="status"
        class="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wider"
        >Status</label
      >
      <select
        id="status"
        bind:value={statusFilter}
        onchange={handleSearch}
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
        handleSearch(null as any);
      }}
      class="pb-3 text-gray-500 hover:text-white transition-colors text-sm"
    >
      Reset
    </button>
  </div>
</div>

<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden">
  <div class="overflow-x-auto">
    <table class="w-full text-left border-collapse">
      <thead>
        <tr
          class="bg-white/[0.02] border-b border-white/5 text-sm uppercase tracking-wider text-gray-400 font-medium"
        >
          <th class="py-4 px-6">ID</th>
          <th class="py-4 px-6">Song ID</th>
          <th class="py-4 px-6">Ver / Slug</th>
          <th class="py-4 px-6">Source</th>
          <th class="py-4 px-6">Flags</th>
          <th class="py-4 px-6 text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each variants as variant (variant.id)}
          <tr class="hover:bg-white/[0.02] transition-colors group">
            <td class="py-4 px-6 whitespace-nowrap text-gray-400">
              #{variant.id}
            </td>
            <td class="py-4 px-6 whitespace-nowrap text-white font-medium">
              <a
                href="/admin/songs/{variant.song_id}/edit"
                class="hover:text-blue-400 hover:underline"
              >
                Song #{variant.song_id}
              </a>
            </td>
            <td class="py-4 px-6 whitespace-nowrap">
              <div class="text-white">v{variant.version_number}</div>
              <div class="text-xs text-gray-500">{variant.slug}</div>
            </td>
            <td class="py-4 px-6 whitespace-nowrap">
              <span
                class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-500/10 text-gray-400"
              >
                {getSourceType(variant)}
              </span>
            </td>
            <td class="py-4 px-6 whitespace-nowrap">
              {#if variant.spoiler}
                <span
                  class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-500/10 text-red-400"
                >
                  Spoiler
                </span>
              {/if}
            </td>
            <td class="py-4 px-6 whitespace-nowrap text-right">
              <a
                href="/admin/variants/{variant.id}/edit"
                class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors inline-block"
                title="Edit"
              >
                <svg
                  class="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  /></svg
                >
              </a>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="py-12 text-center text-gray-500">
              No variants found matching your criteria.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if meta && meta.total_pages > 1}
    <div
      class="border-t border-white/5 px-6 py-4 flex items-center justify-between"
    >
      <p class="text-sm text-gray-400">
        Showing page <span class="font-medium text-white"
          >{meta.current_page}</span
        >
        of <span class="font-medium text-white">{meta.total_pages}</span>
        ({meta.total_items} total)
      </p>
      <div class="flex items-center gap-2">
        <button
          onclick={() => changePage(meta.current_page - 1)}
          disabled={meta.current_page === 1}
          class="px-3 py-1.5 rounded-lg text-sm font-medium bg-white/5 text-gray-300 hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Previous
        </button>

        <div class="flex items-center gap-1 px-2">
          {#each Array.from( { length: Math.min(5, meta.total_pages) }, (_, i) => {
              // Simple logic for < 5 pages. For more, center around current_page
              let start = Math.max(1, meta.current_page - 2);
              if (start + 4 > meta.total_pages) start = Math.max(1, meta.total_pages - 4);
              return start + i;
            }, ) as pageNum}
            <button
              onclick={() => changePage(pageNum)}
              class="w-8 h-8 flex items-center justify-center rounded-lg text-sm font-medium transition-colors {pageNum ===
              meta.current_page
                ? 'bg-anirank-primary text-white'
                : 'text-gray-400 hover:bg-white/10 hover:text-white'}"
            >
              {pageNum}
            </button>
          {/each}
        </div>

        <button
          onclick={() => changePage(meta.current_page + 1)}
          disabled={meta.current_page === meta.total_pages}
          class="px-3 py-1.5 rounded-lg text-sm font-medium bg-white/5 text-gray-300 hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  {/if}
</div>
