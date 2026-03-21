<script lang="ts">
  import api from "$lib/api";

  let { data } = $props();
  let reports = $state(data.reports || []);
  let status = $state("pending");
  let isLoading = $state(false);

  async function loadReports(newStatus: string) {
    status = newStatus;
    isLoading = true;
    try {
      const resp = await api.get(`/admin/songs/reports?status=${status === 'resolved' ? 'fixed' : 'pending'}`);
      reports = resp.data.data;
    } catch (err) {
      console.error("Error loading reports:", err);
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    reports = data.reports || [];
  });

  let loadingDelete = $state<number | null>(null);

  async function deleteReport(id: number) {
    if (!confirm("Are you sure you want to delete this report?")) return;

    loadingDelete = id;
    try {
      await api.delete(`/admin/songs/reports/${id}`);
      reports = reports.filter((r: any) => r.id !== id);
    } catch (err) {
      console.error("Error deleting report:", err);
      alert("Failed to delete report.");
    } finally {
      loadingDelete = null;
    }
  }
</script>

<svelte:head>
  <title>User Reports | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      User Reports
    </h1>
    <p class="text-gray-400">
      Review reported content, comments, and community guideline violations.
    </p>
  </div>
</div>

<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden min-h-[400px]">
  <div
    class="p-4 border-b border-white/5 flex gap-2 overflow-x-auto custom-scrollbar"
  >
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'pending' ? 'bg-white/10 text-white' : 'hover:bg-white/5 text-gray-400'}"
      onclick={() => loadReports('pending')}
      >Pending</button
    >
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'resolved' ? 'bg-white/10 text-white' : 'hover:bg-white/5 text-gray-400'}"
      onclick={() => loadReports('resolved')}
      >Resolved</button
    >
  </div>

  <table class="w-full text-left text-sm text-gray-300">
    <thead
      class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5"
    >
      <tr>
        <th class="px-6 py-4 font-semibold">ID / Type</th>
        <th class="px-6 py-4 font-semibold">Target</th>
        <th class="px-6 py-4 font-semibold w-1/3">Reason</th>
        <th class="px-6 py-4 font-semibold">Reporter</th>
        <th class="px-6 py-4 font-semibold text-right">Actions</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-white/5 relative">
      {#if isLoading}
        <tr>
          <td colspan="5" class="px-6 py-24 text-center text-gray-500">
             <div class="flex flex-col items-center justify-center gap-3">
               <div class="w-8 h-8 border-2 border-white/10 border-t-blue-500 rounded-full animate-spin"></div>
               <span class="text-sm font-medium">Fetching {status} reports...</span>
             </div>
          </td>
        </tr>
      {:else}
        {#each reports as rpt}
          <tr class="hover:bg-white/[0.02] transition-colors">
            <td class="px-6 py-4">
              <div class="font-medium text-white mb-1">#{rpt.id}</div>
              <div
                class="text-xs font-semibold px-2 py-0.5 rounded-full w-fit bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 capitalize"
              >
                {rpt.status}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="font-mono text-xs text-blue-400 mb-1">Song Theme</div>
              <div class="text-xs font-bold text-white mb-0.5">
                {rpt.song?.song_romaji || rpt.song?.song_en || rpt.song?.song_jp || "Unknown Song"}
              </div>
              <div class="text-[10px] text-gray-500">
                ID: {rpt.song_id}
              </div>
              <a
                href="/admin/songs/reports/{rpt.id}"
                class="mt-2 text-xs text-gray-400 hover:text-white underline decoration-white/20 underline-offset-2"
                >Read report</a
              >
            </td>
            <td class="px-6 py-4 text-gray-400 text-sm">
              {rpt.content}
            </td>
            <td class="px-6 py-4">
              <a
                href="/users/{rpt.user?.slug}"
                class="text-white hover:text-blue-400 font-medium transition-colors"
                target="_blank">{rpt.user?.name}</a
              >
              <div class="text-xs text-gray-500 mt-1">
                {new Date(rpt.created_at).toLocaleDateString()}
              </div>
            </td>
            <td class="px-6 py-4 text-right">
              <div
                class="flex items-center justify-end gap-2 text-xs font-medium"
              >
                <button
                  class="px-3 py-1.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 rounded-lg transition-colors border border-rose-500/20 disabled:opacity-50"
                  title="Take action on content"
                  onclick={() => deleteReport(rpt.id)}
                  disabled={loadingDelete === rpt.id}
                >
                  {loadingDelete === rpt.id ? "Deleting..." : "Delete Target"}
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="5" class="px-6 py-24 text-center text-gray-500">
              <div class="flex flex-col items-center justify-center gap-2">
                <span class="material-symbols-outlined text-4xl opacity-20">inventory_2</span>
                <p>No {status} reports found.</p>
              </div>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
