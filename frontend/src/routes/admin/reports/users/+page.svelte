<script lang="ts">
  import api from "$lib/api";
  import { User, Flag, Trash2, CheckCircle, ExternalLink } from "lucide-svelte";
  import { fade } from "svelte/transition";
  import { toastState } from "$lib/state/toast.svelte";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let reports = $state(data.reports || []);
  let status = $state("pending");
  let isLoading = $state(false);
  let loadingDelete = $state<number | null>(null);
  let loadingResolve = $state<number | null>(null);

  async function loadReports(newStatus: string) {
    status = newStatus;
    isLoading = true;
    try {
      const resp = await api.get(`/admin/users/reports?status=${status === 'resolved' ? 'resolved' : 'pending'}`);
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

  async function deleteReport(id: number) {
    if (!confirm("Are you sure you want to delete this report?")) return;

    loadingDelete = id;
    try {
      const response = await api.delete(`/admin/users/reports/${id}`);
      if (response.data.success) {
        toastState.addToast(response.data.message || "Report deleted successfully", "success");
        reports = reports.filter((r: any) => r.id !== id);
      } else {
        toastState.addToast(response.data.message || "Failed to delete report", "error");
      }
    } catch (err: any) {
      console.error("Error deleting report:", err);
      const msg = err.response?.data?.message || "Failed to delete report.";
      toastState.addToast(msg, "error");
    } finally {
      loadingDelete = null;
    }
  }

  async function resolveReport(id: number) {
    loadingResolve = id;
    try {
      const response = await api.put(`/admin/users/reports/${id}/resolve`);
      toastState.addToast(response.data.message || "User report resolved successfully", "success");
      
      if (status === 'pending') {
        reports = reports.filter((r: any) => r.id !== id);
      } else {
        // Refresh if in resolved tab (though unlikely to resolve a resolved one)
        loadReports(status);
      }
    } catch (err: any) {
      console.error("Error resolving report:", err);
      const msg = err.response?.data?.message || "Failed to resolve report.";
      toastState.addToast(msg, "error");
    } finally {
      loadingResolve = null;
    }
  }
</script>

<svelte:head>
  <title>User Reports | Admin</title>
</svelte:head>

<div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      User Reports
    </h1>
    <p class="text-gray-400">
      Manage reports against community members.
    </p>
  </div>
</div>

<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden min-h-[400px]">
  <div class="p-4 border-b border-white/5 flex gap-2 overflow-x-auto custom-scrollbar">
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'pending' ? 'bg-white/10 text-white' : 'hover:bg-white/5 text-gray-400'}"
      onclick={() => loadReports('pending')}
    >
      Pending
    </button>
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'resolved' ? 'bg-white/10 text-white' : 'hover:bg-white/5 text-gray-400'}"
      onclick={() => loadReports('resolved')}
    >
      Resolved
    </button>
  </div>

  <table class="w-full text-left text-sm text-gray-300">
    <thead class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5">
      <tr>
        <th class="px-6 py-4 font-semibold">ID / Status</th>
        <th class="px-6 py-4 font-semibold">Reported User</th>
        <th class="px-6 py-4 font-semibold w-1/3">Reason & Content</th>
        <th class="px-6 py-4 font-semibold">Reporter</th>
        <th class="px-6 py-4 font-semibold text-right">Actions</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-white/5 relative">
      {#if isLoading}
        <tr>
          <td colspan="5" class="px-6 py-24 text-center text-gray-500">
            <div class="flex flex-col items-center justify-center gap-3">
              <div class="w-8 h-8 border-2 border-white/10 border-t-primary rounded-full animate-spin"></div>
              <span class="text-sm font-medium">Fetching {status} reports...</span>
            </div>
          </td>
        </tr>
      {:else}
        {#each reports as rpt (rpt.id)}
          <tr class="hover:bg-white/2 transition-colors" in:fade>
            <td class="px-6 py-4">
              <div class="font-medium text-white mb-1">#{rpt.id}</div>
              <div class="text-[10px] font-bold px-2 py-0.5 rounded-full w-fit {rpt.status ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'} capitalize">
                {rpt.status ? 'Resolved' : 'Pending'}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="size-10 rounded-full overflow-hidden bg-white/5 border border-white/10">
                  {#if rpt.reported_user?.avatar_url}
                    <img src={rpt.reported_user.avatar_url} alt="" class="size-full object-cover" />
                  {:else}
                    <div class="size-full flex items-center justify-center text-gray-600">
                      <User size={20} />
                    </div>
                  {/if}
                </div>
                <div>
                  <a href="/users/{rpt.reported_user?.slug}" target="_blank" class="font-bold text-white hover:text-primary transition-colors flex items-center gap-1">
                    {rpt.reported_user?.name || 'Unknown'}
                    <ExternalLink size={12} class="opacity-40" />
                  </a>
                  <p class="text-[10px] text-gray-500">ID: {rpt.reported_user_id}</p>
                </div>
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-2 text-primary font-bold text-xs mb-1">
                <Flag size={12} />
                {rpt.reason}
              </div>
              <p class="text-xs text-gray-400 line-clamp-2" title={rpt.content}>
                {rpt.content}
              </p>
            </td>
            <td class="px-6 py-4">
              <a href="/users/{rpt.reporter_user?.slug}" target="_blank" class="text-white hover:text-primary font-medium transition-colors">
                {rpt.reporter_user?.name || 'Unknown'}
              </a>
              <div class="text-[10px] text-gray-500 mt-1">
                {new Date(rpt.created_at).toLocaleDateString()}
              </div>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">
                {#if !rpt.status}
                  <button
                    class="p-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 rounded-lg transition-colors border border-emerald-500/20 disabled:opacity-50"
                    title="Mark as resolved"
                    onclick={() => resolveReport(rpt.id)}
                    disabled={loadingResolve === rpt.id}
                  >
                    {#if loadingResolve === rpt.id}
                      <div class="size-4 border-2 border-emerald-400/20 border-t-emerald-400 rounded-full animate-spin"></div>
                    {:else}
                      <CheckCircle size={18} />
                    {/if}
                  </button>
                {/if}
                <button
                  class="p-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 rounded-lg transition-colors border border-rose-500/20 disabled:opacity-50"
                  title="Delete report"
                  onclick={() => deleteReport(rpt.id)}
                  disabled={loadingDelete === rpt.id}
                >
                  {#if loadingDelete === rpt.id}
                    <div class="size-4 border-2 border-rose-400/20 border-t-rose-400 rounded-full animate-spin"></div>
                  {:else}
                    <Trash2 size={18} />
                  {/if}
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="5" class="px-6 py-24 text-center text-gray-500">
              <div class="flex flex-col items-center justify-center gap-2">
                <span class="material-symbols-outlined text-4xl opacity-20">inventory_2</span>
                <p>No {status} user reports found.</p>
              </div>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
