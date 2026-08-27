<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import type { PageData } from "./$types";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import AutocompleteAnime from "$lib/components/admin/AutocompleteAnime.svelte";
  import { getSongName } from "$lib/song-utils";
  import Plus from "lucide-svelte/icons/plus";
  import Pencil from "lucide-svelte/icons/pencil";

  let { data } = $props<{ data: PageData }>();
  let variants = $state<any[]>([]);
  let pagination = $derived(data.pagination);
  let filters = $derived(data.filters);

  let animeIdInput = $state("");
  let statusFilter = $state("");

  $effect(() => {
    variants = data.data;
    animeIdInput = filters?.anime || "";
    statusFilter = filters?.status !== undefined ? String(filters.status) : "";
  });

  async function handleStatusChange(id: number, currentStatus: boolean) {
    try {
      await api.patch(`/admin/variants/${id}/status`);
      // Update local state reactively
      variants = variants.map((v: any) => {
        if (v.id === id) {
          return { ...v, status: !currentStatus };
        }
        return v;
      });
      toastState.addToast("Variant status updated", "success");
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to update status"), "error");
    }
  }

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
    if (newPage < 1 || newPage > (pagination?.last_page || 1)) return;
    const url = new URL($page.url);
    url.searchParams.set("page", newPage.toString());
    goto(url.toString());
  }

  // Helper to format source type
  function getSourceType(variant: any) {
    if (variant.video?.video_src || variant.video?.local_url || variant.video?.type === "file") {
      return "Direct File";
    }
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
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Song Variants
    </h1>
    <p class="text-on-surface-variant/70">Manage video sources and versions for songs.</p>
  </div>

  <a
    href="/admin/variants/create"
    class="px-4 py-2 bg-primary hover:bg-primary-container text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
  >
    <Plus size={20} />
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
        class="block text-xs font-medium text-on-surface-variant/40 mb-1 uppercase tracking-wider"
        >Status</label
      >
      <select
        id="status"
        bind:value={statusFilter}
        onchange={handleSearch}
        class="w-full bg-surface-container border border-outline-variant rounded-xl py-2 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none transition-all appearance-none cursor-pointer"
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
      class="pb-3 text-on-surface-variant/40 hover:text-on-surface transition-colors text-sm"
    >
      Reset
    </button>
  </div>
</div>

<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
  <div class="overflow-x-auto">
    <table class="w-full text-left border-collapse">
      <thead>
        <tr
          class="bg-white/2 border-b border-outline-variant text-sm uppercase tracking-wider text-on-surface-variant/70 font-medium"
        >
          <th class="py-4 px-6">ID</th>
          <th class="py-4 px-6">Song ID</th>
          <th class="py-4 px-6">Version</th>
          <th class="py-4 px-6">Anime / Song</th>
          <th class="py-4 px-6">Status</th>
          <th class="py-4 px-6 text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each variants as variant (variant.id)}
          <tr class="hover:bg-white/2 transition-colors group">
            <!-- ID -->
            <td class="py-4 px-6 whitespace-nowrap text-on-surface-variant/70">
              #{variant.id}
            </td>
            <!-- Song ID -->
            <td class="py-4 px-6 whitespace-nowrap text-on-surface font-medium">
              <a
                href="/admin/songs/{variant.song_id}/edit"
                class="hover:text-blue-400 hover:underline"
              >
                Song #{variant.song_id}
              </a>
            </td>
            <!-- Version / Slug -->
            <td class="py-4 px-6 whitespace-nowrap">
              <!-- <div class="text-on-surface">v{variant.version_number}</div> -->
              <div class="text-xs text-on-surface-variant/40">{variant.slug}</div>
            </td>
            <!-- Anime / Song -->
            <td class="py-4 px-6 whitespace-nowrap">
              <div class="flex flex-col max-w-[200px]">
                <div class="text-on-surface truncate">
                  {variant.song.anime.title}
                </div>
                <div class="text-xs text-on-surface-variant/40 truncate">
                  {getSongName(variant.song)}
                </div>
              </div>
            </td>
            <!-- Status -->
            <td class="py-4 px-6 whitespace-nowrap">
              {#if variant.status === true || variant.status === 1}
                <button
                  onclick={() => handleStatusChange(variant.id, variant.status)}
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Published
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
            <!-- Flags -->
            <!-- <td class="py-4 px-6 whitespace-nowrap">
              {#if variant.spoiler}
                <span
                  class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-500/10 text-red-400"
                >
                  Spoiler
                </span>
              {/if}
            </td> -->
            <!-- Actions -->
            <td class="py-4 px-6 whitespace-nowrap text-right">
              <a
                href="/admin/variants/{variant.id}/edit"
                class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest rounded-lg transition-colors inline-block"
                title="Edit"
              >
                <Pencil size={18} />
              </a>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="py-12 text-center text-on-surface-variant/40">
              No variants found matching your criteria.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if pagination && pagination.last_page > 1}
    <div
      class="border-t border-outline-variant px-6 py-4 flex items-center justify-between"
    >
      <p class="text-sm text-on-surface-variant/70">
        Showing page <span class="font-medium text-on-surface"
          >{pagination.current_page}</span
        >
        of <span class="font-medium text-on-surface">{pagination.last_page}</span>
        ({pagination.total} total)
      </p>
      <div class="flex items-center gap-2">
        <button
          onclick={() => changePage(pagination.current_page - 1)}
          disabled={pagination.current_page === 1}
          class="px-3 py-1.5 rounded-lg text-sm font-medium bg-surface-highest text-on-surface-variant hover:bg-surface-highest disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Previous
        </button>

        <div class="flex items-center gap-1 px-2">
          {#each Array.from( { length: Math.min(5, pagination.last_page) }, (_, i) => {
              // Simple logic for < 5 pages. For more, center around current_page
              let start = Math.max(1, pagination.current_page - 2);
              if (start + 4 > pagination.last_page) start = Math.max(1, pagination.last_page - 4);
              return start + i;
            }, ) as pageNum}
            <button
              onclick={() => changePage(pageNum)}
              class="w-8 h-8 flex items-center justify-center rounded-lg text-sm font-medium transition-colors {pageNum ===
              pagination.current_page
                ? 'bg-primary text-on-surface'
                : 'text-on-surface-variant/70 hover:bg-surface-highest hover:text-on-surface'}"
            >
              {pageNum}
            </button>
          {/each}
        </div>

        <button
          onclick={() => changePage(pagination.current_page + 1)}
          disabled={pagination.current_page === pagination.last_page}
          class="px-3 py-1.5 rounded-lg text-sm font-medium bg-surface-highest text-on-surface-variant hover:bg-surface-highest disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  {/if}
</div>
