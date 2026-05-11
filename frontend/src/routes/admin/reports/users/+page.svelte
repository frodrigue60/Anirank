<script lang="ts">
  import api from "$lib/api";
  import User from "lucide-svelte/icons/user";
import Flag from "lucide-svelte/icons/flag";
import Eye from "lucide-svelte/icons/eye";
import ExternalLink from "lucide-svelte/icons/external-link";
import Inbox from "lucide-svelte/icons/inbox";
  import { fade } from "svelte/transition";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let reports = $state(data.reports || []);
  let status = $state("pending");
  let isLoading = $state(false);

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
</script>

<svelte:head>
  <title>User Reports | Admin</title>
</svelte:head>

<div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      User Reports
    </h1>
    <p class="text-on-surface-variant/70">
      Manage reports against community members.
    </p>
  </div>
</div>

<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden min-h-[400px]">
  <div class="p-4 border-b border-outline-variant flex gap-2 overflow-x-auto custom-scrollbar">
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'pending' ? 'bg-surface-highest text-on-surface' : 'hover:bg-surface-highest text-on-surface-variant/70'}"
      onclick={() => loadReports('pending')}
    >
      Pending
    </button>
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {status === 'resolved' ? 'bg-surface-highest text-on-surface' : 'hover:bg-surface-highest text-on-surface-variant/70'}"
      onclick={() => loadReports('resolved')}
    >
      Resolved
    </button>
  </div>

  <table class="w-full text-left text-sm text-on-surface-variant">
    <thead class="text-xs text-on-surface-variant/70 uppercase bg-surface-highest border-b border-outline-variant">
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
          <td colspan="5" class="px-6 py-24 text-center text-on-surface-variant/40">
            <div class="flex flex-col items-center justify-center gap-3">
              <div class="w-8 h-8 border-2 border-outline-variant border-t-primary rounded-full animate-spin"></div>
              <span class="text-sm font-medium">Fetching {status} reports...</span>
            </div>
          </td>
        </tr>
      {:else}
        {#each reports as rpt (rpt.id)}
          <tr class="hover:bg-white/2 transition-colors" in:fade>
            <td class="px-6 py-4">
              <div class="font-medium text-on-surface mb-1">#{rpt.id}</div>
              <div class="text-[10px] font-bold px-2 py-0.5 rounded-full w-fit {rpt.status ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'} capitalize">
                {rpt.status ? 'Resolved' : 'Pending'}
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="size-10 rounded-full overflow-hidden bg-surface-highest border border-outline-variant">
                  {#if rpt.reported_user?.avatar_url}
                    <img src={rpt.reported_user.avatar_url} alt="" class="size-full object-cover" />
                  {:else}
                    <div class="size-full flex items-center justify-center text-gray-600">
                      <User size={20} />
                    </div>
                  {/if}
                </div>
                <div>
                  <a href="/users/{rpt.reported_user?.slug}" target="_blank" class="font-bold text-on-surface hover:text-primary transition-colors flex items-center gap-1">
                    {rpt.reported_user?.name || 'Unknown'}
                    <ExternalLink size={12} class="opacity-40" />
                  </a>
                  <div class="flex items-center gap-1.5 mt-0.5">
                    <span class="text-[9px] font-bold px-1 py-0.5 rounded-sm bg-blue-500/10 text-blue-400 border border-blue-500/10">
                      Score: {rpt.reported_user?.truth_score}
                    </span>
                    {#if rpt.reported_user?.is_shadowbanned}
                      <span class="text-[9px] font-bold px-1 py-0.5 rounded-sm bg-red-500/10 text-red-400 border border-red-500/10">
                        SB
                      </span>
                    {/if}
                  </div>
                  <p class="text-[10px] text-on-surface-variant/40 mt-0.5">ID: {rpt.reported_user_id}</p>
                </div>
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-2 text-primary font-bold text-xs mb-1">
                <Flag size={12} />
                {rpt.reason}
              </div>
              <p class="text-xs text-on-surface-variant/70 line-clamp-2" title={rpt.content}>
                {rpt.content}
              </p>
            </td>
            <td class="px-6 py-4">
              <a href="/users/{rpt.reporter_user?.slug}" target="_blank" class="text-on-surface hover:text-primary font-medium transition-colors">
                {rpt.reporter_user?.name || 'Unknown'}
              </a>
              <div class="flex items-center gap-1 mt-1">
                <span class="text-[9px] font-medium text-on-surface-variant/50">Score: {rpt.reporter_user?.truth_score}</span>
                {#if rpt.reporter_user?.is_shadowbanned}
                  <span class="text-[9px] text-red-400 font-bold">SB</span>
                {/if}
              </div>
              <div class="text-[10px] text-on-surface-variant/40 mt-0.5">
                {new Date(rpt.created_at).toLocaleDateString()}
              </div>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">
                <a
                  href="/admin/reports/users/{rpt.id}"
                  class="p-2 bg-surface-highest hover:bg-surface-highest text-on-surface/60 hover:text-primary rounded-lg transition-all"
                  title="View Details"
                >
                  <Eye size={18} />
                </a>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="5" class="px-6 py-24 text-center text-on-surface-variant/40">
              <div class="flex flex-col items-center justify-center gap-2">
                <Inbox size={48} class="opacity-20" />
                <p>No {status} user reports found.</p>
              </div>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
