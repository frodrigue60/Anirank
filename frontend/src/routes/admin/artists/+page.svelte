<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { getAuthToken } from "$lib/state/auth.svelte";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";

  let { data } = $props();

  let artists = $derived(data.artists);
  let pagination = $derived(data.pagination);

  // svelte-ignore state_referenced_locally
  let searchQuery = $state(data.pagination.search || "");

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
    if (newPage >= 1 && newPage <= pagination.last_page) {
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
      goto(window.location.pathname + window.location.search, {
        invalidateAll: true,
      });
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to update status"), "error");
    }
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Are you sure you want to delete artist "${name}"?`)) return;
    try {
      await api.delete(`/admin/artists/${id}`);
      toastState.addToast("Artist deleted successfully", "success");
      goto(window.location.pathname + window.location.search, {
        invalidateAll: true,
      });
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to delete artist"), "error");
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

  let generatingAvatars = $state(false);
  let mergingArtists = $state(false);
  let recountingSongs = $state(false);
  let progressMessage = $state("");

  async function generateAvatares() {
    if (
      !confirm(
        "Are you sure you want to generate avatars for all artists without one? This will search AniList and fallback to UI-Avatars.",
      )
    )
      return;

    generatingAvatars = true;
    progressMessage = "Connecting to generation stream...";
    try {
      const response = await fetch(
        `${api.defaults.baseURL}/admin/artists/generate-avatars`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${getAuthToken()}`,
            "X-CSRF-Token": typeof document !== 'undefined' && document.cookie.includes('csrf_token=') ? `; ${document.cookie}`.split(`; csrf_token=`)[1].split(';')[0] : '',
          },
          body: JSON.stringify({
            artist_ids: selectedIds,
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const messages = chunk.split("\n\n");

          for (const msg of messages) {
            if (msg.startsWith("data: ")) {
              progressMessage = msg.replace("data: ", "");
            }
          }
        }
      }

      toastState.addToast("Batch avatar generation completed", "success");
      selectedIds = [];
      // Refresh list to show new avatars
      goto(window.location.pathname + window.location.search, {
        invalidateAll: true,
      });
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to start generation"), "error");
    } finally {
      generatingAvatars = false;
      progressMessage = "";
    }
  }

  async function mergeArtists() {
    if (
      !confirm(
        "Are you sure you want to merge artists with duplicate names? This process will move all songs and favorites to a single record and delete the duplicates. This action is irreversible.",
      )
    )
      return;

    mergingArtists = true;
    progressMessage = "Connecting to merge stream...";
    try {
      const response = await fetch(
        `${api.defaults.baseURL}/admin/artists/merge`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${getAuthToken()}`,
            "X-CSRF-Token": typeof document !== 'undefined' && document.cookie.includes('csrf_token=') ? `; ${document.cookie}`.split(`; csrf_token=`)[1].split(';')[0] : '',
          },
        },
      );

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const messages = chunk.split("\n\n");

          for (const msg of messages) {
            if (msg.startsWith("data: ")) {
              progressMessage = msg.replace("data: ", "");
            }
          }
        }
      }

      toastState.addToast("Artist merge completed successfully", "success");
      // Refresh to show updated counts
      goto(window.location.pathname + window.location.search, {
        invalidateAll: true,
      });
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to complete merge"), "error");
    } finally {
      mergingArtists = false;
      progressMessage = "";
    }
  }

  async function recountSongs() {
    if (
      !confirm(
        "Are you sure you want to recalculate song counters for all artists? This might take a few moments.",
      )
    )
      return;

    recountingSongs = true;
    progressMessage = "Connecting to recalculation stream...";
    try {
      const response = await fetch(
        `${api.defaults.baseURL}/admin/artists/recount-songs`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${getAuthToken()}`,
            "X-CSRF-Token": typeof document !== 'undefined' && document.cookie.includes('csrf_token=') ? `; ${document.cookie}`.split(`; csrf_token=`)[1].split(';')[0] : '',
          },
        },
      );

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const messages = chunk.split("\n\n");

          for (const msg of messages) {
            if (msg.startsWith("data: ")) {
              progressMessage = msg.replace("data: ", "");
            }
          }
        }
      }

      toastState.addToast("Artist song counters recalculated", "success");
      // Refresh to show updated counts
      goto(window.location.pathname + window.location.search, {
        invalidateAll: true,
      });
    } catch (err: any) {
      console.error(err);
      toastState.addToast(getApiErrorMessage(err, "Failed to recount songs"), "error");
    } finally {
      recountingSongs = false;
      progressMessage = "";
    }
  }
</script>

<svelte:head>
  <title>Artists Catalog | Admin</title>
</svelte:head>

<div class="mb-8 flex flex-col gap-4">
  <div class="me-auto">
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Artists Catalog
    </h1>
    <p class="text-on-surface-variant/70">
      Manage musical artists, bands, and their information.
    </p>
  </div>
  <div class="flex flex-col md:flex-row gap-4">
    <!-- recount songs -->
    <button
      onclick={recountSongs}
      disabled={recountingSongs}
      class="px-4 py-2 bg-primary hover:bg-primary-container disabled:opacity-50 disabled:cursor-not-allowed text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
    >
      {#if recountingSongs}
        <svg class="animate-spin h-5 w-5 text-on-surface" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        Recounting...
      {:else}
        Recount Songs
      {/if}
    </button>
    <!-- merge artists -->
    <button
      onclick={mergeArtists}
      disabled={mergingArtists}
      class="px-4 py-2 bg-primary hover:bg-primary-container disabled:opacity-50 disabled:cursor-not-allowed text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
    >
      {#if mergingArtists}
        <svg class="animate-spin h-5 w-5 text-on-surface" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        Merging...
      {:else}
        Merge Artists
      {/if}
    </button>

    <!-- generate avatares for all artists -->
    <button
      onclick={generateAvatares}
      disabled={generatingAvatars}
      class="px-4 py-2 bg-primary hover:bg-primary-container disabled:opacity-50 disabled:cursor-not-allowed text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
    >
      {#if generatingAvatars}
        <svg class="animate-spin h-5 w-5 text-on-surface" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        Generating...
      {:else if selectedIds.length > 0}
        Generate for {selectedIds.length} Selected
      {:else}
        Generate Avatares
      {/if}
    </button>
    <!-- create new artist -->
    <a
      href="/admin/artists/create"
      class="px-4 py-2 bg-primary hover:bg-primary-container text-on-surface font-medium rounded-xl transition-colors shadow-lg shadow-anirank-primary/20 flex items-center gap-2"
    >
      + New Artist
    </a>
  </div>
</div>

<!-- Filters & Search -->
<div
  class="bg-surface-container border border-outline-variant rounded-2xl p-4 mb-6 flex flex-col sm:flex-row gap-4"
>
  <div class="relative flex-1">
    <svg
      class="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant/40"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
      />
    </svg>
    <input
      type="text"
      bind:value={searchQuery}
      onkeydown={(e) => e.key === "Enter" && handleSearch()}
      placeholder="Search artist name..."
      class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2 pl-10 pr-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors"
    />
  </div>
  <button
    onclick={handleSearch}
    class="px-6 py-2 bg-primary hover:bg-primary-container text-on-surface rounded-xl transition-all font-medium border border-outline-variant"
  >
    Search
  </button>
</div>

{#if generatingAvatars || mergingArtists}
  <div
    class="mt-6 p-4 bg-primary/10 border border-primary/20 rounded-xl animate-in fade-in slide-in-from-top-2 mb-6"
  >
    <div class="flex items-center gap-3">
      <div class="shrink-0">
        <svg
          class="animate-spin h-5 w-5 text-primary"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-on-surface truncate">
          {progressMessage}
        </p>
      </div>
    </div>
    <!-- Simple CSS progress bar (pulse) -->
    <div class="mt-3 w-full bg-surface-highest rounded-full h-1.5 overflow-hidden">
      <div
        class="bg-primary h-full rounded-full animate-pulse w-full"
      ></div>
    </div>
  </div>
{/if}

<!-- Table -->
<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
  <div class="p-4 border-b border-outline-variant flex items-center justify-between">
    <h2 class="text-xl font-semibold text-on-surface">Artists</h2>
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
              checked={selectedIds.length === artists.length &&
                artists.length > 0}
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
                class="rounded border-outline-variant bg-surface-highest checked:bg-primary focus:ring-primary transition-all cursor-pointer"
                checked={selectedIds.includes(artist.id)}
                onchange={() => toggleSelection(artist.id)}
              />
            </td>
            <td class="px-6 py-4">
              <div
                class="w-10 h-10 rounded-full bg-surface-highest overflow-hidden border border-outline-variant"
              >
                {#if artist.avatar_url}
                  <img
                    src={artist.avatar_url}
                    alt={artist.name}
                    class="w-full h-full object-cover"
                  />
                {:else}
                  <div
                    class="w-full h-full flex items-center justify-center text-gray-600"
                  >
                    <svg
                      class="w-6 h-6"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                      />
                    </svg>
                  </div>
                {/if}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="font-medium line-clamp-1" title={artist.name}>
                <a
                  class="text-light hover:text-primary/80 transition-colors"
                  href="/admin/artists/{artist.id}">{artist.name}</a
                >
              </div>
              {#if artist.name_jp}
                <div class="text-xs text-on-surface-variant/40">{artist.name_jp}</div>
              {/if}
            </td>
            <td class="px-6 py-4">
              <span
                class="text-blue-400 text-xs font-semibold px-2 py-0.5 rounded-full bg-blue-400/10 border border-blue-400/20"
              >
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
                  class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest rounded-lg transition-colors"
                  title="Edit"
                >
                  <span class="material-symbols-outlined">edit</span>
                </a>
                <button
                  onclick={() => handleDelete(artist.id, artist.name)}
                  class="p-2 text-on-surface-variant/70 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors"
                  title="Delete"
                >
                  <span class="material-symbols-outlined">delete</span>
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="px-6 py-12 text-center text-on-surface-variant/40"
              >No artists found.</td
            >
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
        Showing <span class="font-medium text-on-surface">{artists.length}</span> items
      </div>
      <div class="flex items-center gap-2">
        <button
          disabled={pagination.current_page === 1}
          onclick={() => changePage(pagination.current_page - 1)}
          aria-label="Previous Page"
          class="p-2 rounded-lg border border-outline-variant text-on-surface-variant/70 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-highest transition-colors"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 19l-7-7 7-7"
            />
          </svg>
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
          <svg
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>
      </div>
    </div>
  {/if}
</div>
