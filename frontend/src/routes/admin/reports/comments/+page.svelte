<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let reports = $state(data.reports);
  let statusFilter = $state("pending");
  let isLoading = $state(false);

  async function loadReports(newStatus: string) {
    isLoading = true;
    statusFilter = newStatus;
    try {
      const resp = await api.get("/admin/comments/reports", {
        params: { status: statusFilter, limit: 50, offset: 0 },
      });
      reports = resp.data.data || [];
    } catch (e) {
      console.error("Failed to load comment reports", e);
    } finally {
      isLoading = false;
    }
  }

  async function deleteReport(id: number) {
    if (!confirm("Are you sure you want to delete this report?")) return;
    try {
      // Optimistic update
      reports = reports.filter((r: any) => r.id !== id);
      await api.delete(`/admin/comments/reports/${id}`);
    } catch (e: any) {
      console.error("Failed to delete report:", e);
      alert("Failed to delete report.");
    }
  }

  onMount(() => {
    // Initial load handled by SSR/load function
  });
</script>

<div class="space-y-6">
  <div class="flex justify-between items-end">
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-on-surface/90">
        Comment Reports
      </h1>
      <p class="text-on-surface/50 text-sm mt-1">
        Manage and review user reports for comments across the platform.
      </p>
    </div>
  </div>

  <!-- Status Tabs -->
  <div class="flex items-center gap-2 border-b border-outline-variant pb-4">
    <button
      onclick={() => loadReports("pending")}
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors {statusFilter ===
      'pending'
        ? 'bg-blue-600 text-on-surface'
        : 'text-on-surface/50 hover:bg-surface-highest hover:text-on-surface'}"
    >
      Pending
    </button>
    <button
      onclick={() => loadReports("fixed")}
      class="px-4 py-2 rounded-lg text-sm font-medium transition-colors {statusFilter ===
      'fixed'
        ? 'bg-green-500/20 text-green-400'
        : 'text-on-surface/50 hover:bg-surface-highest hover:text-on-surface'}"
    >
      Resolved
    </button>
  </div>

  <div
    class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-xl"
  >
    {#if isLoading}
      <div class="p-8 flex justify-center text-on-surface/40">
        <span class="material-symbols-outlined animate-spin text-4xl"
          >progress_activity</span
        >
      </div>
    {:else if reports.length === 0}
      <div class="p-12 text-center text-on-surface/40">
        <span class="material-symbols-outlined text-4xl mb-2">inbox</span>
        <p>No reports found.</p>
      </div>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-outline-variant bg-surface-highest text-xs text-on-surface/40 uppercase tracking-wider">
              <th class="p-4 font-medium">ID</th>
              <th class="p-4 font-medium">Title</th>
              <th class="p-4 font-medium">Comment</th>
              <th class="p-4 font-medium">Reported By</th>
              <th class="p-4 font-medium">Date</th>
              <th class="p-4 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            {#each reports as report (report.id)}
              <tr class="hover:bg-surface-highest transition-colors group">
                <td class="p-4 text-sm text-on-surface/60">#{report.id}</td>
                <td class="p-4">
                  <span class="text-sm font-bold text-on-surface/90">
                    {report.title}
                  </span>
                </td>
                <td class="p-4">
                  <div class="text-sm text-on-surface/70 max-w-[200px] truncate">
                    {report.comment?.content || 'N/A'}
                  </div>
                </td>
                <td class="p-4">
                  <div class="flex items-center gap-2">
                    <span class="material-symbols-outlined text-on-surface/40 text-[18px]">person</span>
                    <span class="text-sm text-on-surface/80">{report.user?.name}</span>
                  </div>
                </td>
                <td class="p-4 text-sm text-on-surface/50">
                  {new Date(report.created_at).toLocaleDateString()}
                </td>
                <td class="p-4">
                  <div class="flex items-center gap-2">
                    <a
                      href="/admin/reports/comments/{report.id}"
                      class="flex items-center justify-center w-8 h-8 rounded-lg bg-surface-highest hover:bg-surface-highest text-on-surface/60 hover:text-blue-400 transition-all font-bold"
                      title="View Details"
                    >
                      <span class="material-symbols-outlined text-[18px]">visibility</span>
                    </a>
                    {#if statusFilter === 'fixed'}
                      <button
                        onclick={(e) => { e.stopPropagation(); deleteReport(report.id); }}
                        class="flex items-center justify-center w-8 h-8 rounded-lg bg-surface-highest hover:bg-red-500/20 text-on-surface/60 hover:text-red-400 transition-all"
                        title="Delete Report"
                      >
                        <span class="material-symbols-outlined text-[18px]">delete</span>
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>
