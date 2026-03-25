<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  let { data } = $props();

  let eventFilter = $state(page.url.searchParams.get("event") || "");
  let typeFilter = $state(page.url.searchParams.get("resource") || page.url.searchParams.get("auditable_type") || "");
  let expandedRows = $state(new Set<number>());

  function handleFilter() {
    let query = `page=1`;
    if (eventFilter) query += `&event=${eventFilter}`;
    if (typeFilter) query += `&resource=${typeFilter}`;
    goto(`/admin/audit-logs?${query}`);
  }

  function changePage(newPage: number) {
    if (newPage >= 1 && newPage <= data.pagination.last_page) {
      let query = `page=${newPage}`;
      if (eventFilter) query += `&event=${eventFilter}`;
      if (typeFilter) query += `&resource=${typeFilter}`;
      goto(`/admin/audit-logs?${query}`);
    }
  }

  function toggleRow(id: number) {
    if (expandedRows.has(id)) {
      expandedRows.delete(id);
    } else {
      expandedRows.add(id);
    }
  }

  function formatJSON(jsonStr: string | null) {
    if (!jsonStr) return "N/A";
    try {
      const parsed = JSON.parse(jsonStr);
      return JSON.stringify(parsed, null, 2);
    } catch (e) {
      return jsonStr;
    }
  }

  const eventColors: Record<string, string> = {
    created: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    updated: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    deleted: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    status_toggled: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  };
</script>

<svelte:head>
  <title>Audit Logs | Admin</title>
</svelte:head>

<div class="mb-8">
  <h1 class="text-3xl font-bold tracking-tight text-white mb-1">Audit Logs</h1>
  <p class="text-gray-400">
    Track all administrative actions and changes across the platform.
  </p>
</div>

<!-- Filters -->
<div
  class="bg-anirank-card border border-white/5 rounded-2xl p-4 mb-6 grid grid-cols-1 sm:grid-cols-3 gap-4"
>
  <div>
    <label
      for="event"
      class="block text-xs font-semibold text-gray-500 uppercase mb-2"
      >Event</label
    >
    <select
      id="event"
      bind:value={eventFilter}
      onchange={handleFilter}
      class="w-full bg-white/5 border border-white/10 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Events</option>
      <option value="created">Created</option>
      <option value="updated">Updated</option>
      <option value="deleted">Deleted</option>
      <option value="status_toggled">Status Toggled</option>
    </select>
  </div>
  <div>
    <label
      for="type"
      class="block text-xs font-semibold text-gray-500 uppercase mb-2"
      >Resource Type</label
    >
    <select
      id="type"
      bind:value={typeFilter}
      onchange={handleFilter}
      class="w-full bg-white/5 border border-white/10 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
    >
      <option value="">All Types</option>
      <option value="user">User</option>
      <option value="anime">Anime</option>
      <option value="song">Song</option>
      <option value="variant">Variant</option>
      <option value="artist">Artist</option>
      <option value="taxonomy_year">Year</option>
      <option value="taxonomy_season">Season</option>
      <option value="taxonomy_format">Format</option>
      <option value="taxonomy_genre">Genre</option>
    </select>
  </div>
  <div class="flex items-end">
    <button
      onclick={() => {
        eventFilter = "";
        typeFilter = "";
        handleFilter();
      }}
      class="px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-xl transition-colors border border-white/10 w-full"
    >
      Clear Filters
    </button>
  </div>
</div>

<!-- Table -->
<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden">
  {#if data.error}
    <div class="p-8 text-center">
      <div
        class="inline-flex items-center gap-2 px-4 py-2 bg-rose-500/10 text-rose-400 border border-rose-500/20 rounded-xl mb-4"
      >
        <span class="material-symbols-outlined text-sm">error</span>
        <span class="text-sm font-medium">{data.error}</span>
      </div>
      <p class="text-gray-500 text-sm">
        Verifica tu conexión o permisos e intenta recargar la página.
      </p>
    </div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left text-sm text-gray-300">
        <thead
          class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5"
        >
          <tr>
            <th class="px-6 py-4 font-semibold w-10"></th>
            <th class="px-6 py-4 font-semibold">User (Staff)</th>
            <th class="px-6 py-4 font-semibold">Event</th>
            <th class="px-6 py-4 font-semibold">Resource</th>
            <th class="px-6 py-4 font-semibold">Date</th>
            <!-- <th class="px-6 py-4 font-semibold">IP Address</th> -->
            <th class="px-6 py-4 font-semibold">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/5">
          {#each data.logs as log}
            <tr
              class="hover:bg-white/2 transition-colors group cursor-pointer"
              onclick={() => toggleRow(log.id)}
            >
              <td class="px-6 py-4">
                <span
                  class="material-symbols-outlined transition-transform {expandedRows.has(
                    log.id,
                  )
                    ? 'rotate-90'
                    : ''}"
                >
                  chevron_right
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <span class="font-bold text-white"
                    >{log.user?.name || `User ID: ${log.user_id}`}</span
                  >
                </div>
              </td>
              <td class="px-6 py-4">
                <span
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border {eventColors[
                    log.event
                  ] || 'bg-gray-500/10 text-gray-400 border-gray-500/20'}"
                >
                  {log.event}
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="flex flex-col">
                  <span class="text-white font-medium capitalize"
                    >{log.auditable_type}</span
                  >
                  <span class="text-xs text-gray-500"
                    >ID: {log.auditable_id}</span
                  >
                </div>
              </td>
              <td class="px-6 py-4 text-gray-400">
                {new Date(log.created_at).toLocaleString()}
              </td>
              <!--  <td class="px-6 py-4 text-gray-500 font-mono text-xs">
              {log.ip_address || "Unknown"}
            </td> -->
              <td>
                <a
                  href="/admin/audit-logs/{log.id}"
                  class="px-4 py-2 bg-anirank-primary/10 hover:bg-anirank-primary/20 text-anirank-primary rounded-xl transition-colors border border-anirank-primary/20 text-xs font-bold flex items-center gap-2"
                >
                  <span class="material-symbols-outlined text-sm"
                    >visibility</span
                  >
                  View
                </a>
              </td>
            </tr>

            {#if expandedRows.has(log.id)}
              <tr class="bg-white/1">
                <td colspan="6" class="px-6 py-6 transition-all">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <h4
                        class="text-xs font-bold text-gray-500 uppercase mb-3"
                        id="old-values-{log.id}"
                      >
                        Old Values
                      </h4>
                      <pre
                        aria-labelledby="old-values-{log.id}"
                        class="bg-black/40 rounded-xl p-4 text-xs font-mono text-gray-400 overflow-x-auto max-h-60 custom-scrollbar border border-white/5 whitespace-pre-wrap">{formatJSON(
                          log.old_values,
                        )}</pre>
                    </div>
                    <div>
                      <h4
                        class="text-xs font-bold text-gray-500 uppercase mb-3"
                        id="new-values-{log.id}"
                      >
                        New Values
                      </h4>
                      <pre
                        aria-labelledby="new-values-{log.id}"
                        class="bg-black/40 rounded-xl p-4 text-xs font-mono text-emerald-400/80 overflow-x-auto max-h-60 custom-scrollbar border border-white/5 whitespace-pre-wrap">{formatJSON(
                          log.new_values,
                        )}</pre>
                    </div>
                  </div>
                  <div
                    class="mt-4 pt-4 border-t border-white/5 flex flex-col md:flex-row justify-between items-start md:items-center gap-4"
                  >
                    <div class="flex flex-col gap-1">
                      <span
                        class="text-[10px] text-gray-600 uppercase font-bold"
                        >Metadata</span
                      >
                      <span class="text-xs text-gray-500"
                        ><b class="text-gray-400">URL:</b> {log.url}</span
                      >
                      <span class="text-xs text-gray-500"
                        ><b class="text-gray-400">Agent:</b>
                        {log.user_agent}</span
                      >
                    </div>
                    <a
                      href="/admin/audit-logs/{log.id}"
                      class="px-4 py-2 bg-anirank-primary/10 hover:bg-anirank-primary/20 text-anirank-primary rounded-xl transition-colors border border-anirank-primary/20 text-xs font-bold flex items-center gap-2"
                    >
                      <span class="material-symbols-outlined text-sm"
                        >visibility</span
                      >
                      View Full Detail
                    </a>
                  </div>
                </td>
              </tr>
            {/if}
          {:else}
            <tr>
              <td colspan="6" class="px-6 py-12 text-center text-gray-500">
                No audit logs found matching your filters.
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    {#if data.pagination?.last_page > 1}
      <div
        class="px-6 py-4 border-t border-white/5 flex items-center justify-between"
      >
        <div class="text-sm text-gray-400">
          Showing <span class="font-medium text-white">{data.logs.length}</span> items
        </div>
        <div class="flex items-center gap-2">
          <button
            disabled={data.pagination.current_page === 1}
            onclick={() => changePage(data.pagination.current_page - 1)}
            aria-label="Página anterior"
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
            >Page {data.pagination.current_page} of {data.pagination.last_page}</span
          >
          <button
            disabled={data.pagination.current_page === data.pagination.last_page}
            onclick={() => changePage(data.pagination.current_page + 1)}
            aria-label="Página siguiente"
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
  {/if}
</div>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
    height: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 10px;
  }
</style>
