<script lang="ts">
  import api from "$lib/api";
  import { goto } from "$app/navigation";
  import { toastState } from "$lib/state/toast.svelte";
  import { User, Flag, CheckCircle, Trash2, ArrowLeft, ExternalLink, ShieldAlert } from "lucide-svelte";
  import { fade, scale } from "svelte/transition";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let report = $state(data.report);

  $effect(() => {
    report = data.report;
  });

  let isResolving = $state(false);
  let isDeleting = $state(false);

  async function resolveReport() {
    if (!report) return;
    isResolving = true;
    try {
      const response = await api.put(`/admin/users/reports/${report.id}/resolve`);
      toastState.addToast(response.data.message || "User report marked as resolved", "success");
      report.status = true;
      goto("/admin/reports/users");
    } catch (err: any) {
      console.error("Error resolving report:", err);
      const msg = err.response?.data?.message || "Failed to resolve report";
      toastState.addToast(msg, "error");
    } finally {
      isResolving = false;
    }
  }

  async function deleteReport() {
    if (!report) return;
    if (!confirm("Are you sure you want to delete this report?")) return;

    isDeleting = true;
    try {
      const response = await api.delete(`/admin/users/reports/${report.id}`);
      toastState.addToast(response.data.message || "Report deleted", "success");
      goto("/admin/reports/users");
    } catch (err: any) {
      console.error("Error deleting report:", err);
      const msg = err.response?.data?.message || "Failed to delete report";
      toastState.addToast(msg, "error");
    } finally {
      isDeleting = false;
    }
  }
</script>

<svelte:head>
  <title>Report #{report?.id || "..."} | Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 py-8">
  <!-- Header -->
  <div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
    <div class="flex items-center gap-4">
      <button
        onclick={() => goto("/admin/reports/users")}
        class="p-2 bg-white/5 hover:bg-white/10 rounded-xl text-gray-400 transition-all active:scale-95"
      >
        <ArrowLeft size={20} />
      </button>
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-white">
          User Report <span class="text-white/30">#{report?.id}</span>
        </h1>
        <p class="text-gray-400 text-sm">Reviewing community conduct violation.</p>
      </div>
    </div>

    {#if report && !report.status}
      <div class="flex gap-3 w-full sm:w-auto">
        <button
          onclick={deleteReport}
          disabled={isDeleting || isResolving}
          class="flex-1 sm:flex-none px-5 py-2.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 rounded-xl font-bold text-sm transition-all disabled:opacity-50 flex items-center justify-center gap-2"
        >
          <Trash2 size={16} />
          {isDeleting ? "Deleting..." : "Delete Report"}
        </button>
        <button
          onclick={resolveReport}
          disabled={isResolving || isDeleting}
          class="flex-1 sm:flex-none px-6 py-2.5 bg-primary text-white hover:opacity-90 rounded-xl font-bold text-sm transition-all shadow-lg shadow-primary/20 disabled:opacity-50 flex items-center justify-center gap-2"
        >
          <CheckCircle size={16} />
          {isResolving ? "Resolving..." : "Mark as Resolved"}
        </button>
      </div>
    {/if}
  </div>

  {#if !report}
    <div class="bg-anirank-card border border-white/5 rounded-3xl p-12 text-center text-gray-500">
      <ShieldAlert size={64} class="mx-auto mb-4 opacity-10" />
      <p>Report not found or failed to load.</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Content -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Reason & Content -->
        <div class="bg-anirank-card border border-white/5 rounded-3xl overflow-hidden shadow-xl">
          <div class="p-5 border-b border-white/5 bg-white/2 flex justify-between items-center">
            <h2 class="font-bold text-white flex items-center gap-2">
              <Flag size={18} class="text-primary" />
              Report Details
            </h2>
            <div class="text-[10px] font-bold px-3 py-1 rounded-full border {report.status ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border-amber-500/20'} uppercase tracking-wider">
              {report.status ? 'Resolved' : 'Pending'}
            </div>
          </div>
          <div class="p-8">
            <div class="mb-6">
              <span class="text-[10px] font-bold text-gray-500 uppercase tracking-widest block mb-1">Reason Category</span>
              <h3 class="text-2xl font-black text-white">{report.reason}</h3>
            </div>
            
            <span class="text-[10px] font-bold text-gray-500 uppercase tracking-widest block mb-2">Description</span>
            <div class="bg-black/40 rounded-2xl p-6 text-gray-300 leading-relaxed border border-white/5 whitespace-pre-wrap text-sm italic">
              "{report.content || 'No detailed description provided.'}"
            </div>
          </div>
        </div>

        <!-- Reported User Info -->
        <div class="bg-anirank-card border border-white/5 rounded-3xl overflow-hidden shadow-xl">
          <div class="p-5 border-b border-white/5 bg-white/2">
            <h2 class="font-bold text-white flex items-center gap-2">
              <User size={18} class="text-rose-400" />
              Reported User
            </h2>
          </div>
          <div class="p-8">
            <div class="flex items-center gap-6 bg-white/5 p-6 rounded-2xl border border-white/5">
              <div class="size-20 rounded-full overflow-hidden border-2 border-white/10 shadow-2xl bg-black/50">
                {#if report.reported_user?.avatar_url}
                  <img src={report.reported_user.avatar_url} alt="" class="size-full object-cover" />
                {:else}
                  <div class="size-full flex items-center justify-center text-gray-700">
                    <User size={40} />
                  </div>
                {/if}
              </div>
              <div class="flex-1">
                <div class="flex items-center justify-between">
                  <h4 class="text-xl font-black text-white">{report.reported_user?.name || 'Unknown User'}</h4>
                  <a href="/users/{report.reported_user?.slug}" target="_blank" class="p-2 bg-white/5 hover:bg-white/10 rounded-lg text-primary transition-all">
                    <ExternalLink size={18} />
                  </a>
                </div>
                <p class="text-sm text-gray-500 mt-1 font-mono">UUID: {report.reported_user?.uuid || 'N/A'}</p>
                <div class="mt-4 flex gap-2">
                  <span class="px-3 py-1 bg-white/5 rounded-full text-[10px] font-bold text-gray-400">LEVEL {report.reported_user?.level || 1}</span>
                  <span class="px-3 py-1 bg-white/5 rounded-full text-[10px] font-bold text-gray-400">XP {report.reported_user?.xp || 0}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Reporter Info -->
        <div class="bg-anirank-card border border-white/5 rounded-3xl overflow-hidden shadow-xl">
          <div class="p-5 border-b border-white/5 bg-white/2">
            <h2 class="font-bold text-white flex items-center gap-2">
              <ShieldAlert size={18} class="text-blue-400" />
              Reporter
            </h2>
          </div>
          <div class="p-6">
            <div class="flex items-center gap-3 mb-6">
              <div class="size-12 rounded-full overflow-hidden bg-white/5 border border-white/10">
                {#if report.reporter_user?.avatar_url}
                  <img src={report.reporter_user.avatar_url} alt="" class="size-full object-cover" />
                {:else}
                  <div class="size-full flex items-center justify-center text-gray-700">
                    <User size={20} />
                  </div>
                {/if}
              </div>
              <div>
                <a href="/users/{report.reporter_user?.slug}" target="_blank" class="font-black text-white hover:text-primary transition-colors">
                  {report.reporter_user?.name}
                </a>
                <p class="text-[10px] text-gray-500">ID: {report.reporter_user_id}</p>
              </div>
            </div>

            <div class="space-y-4 pt-4 border-t border-white/5">
              <div class="flex justify-between items-center text-xs">
                <span class="text-gray-500 font-bold uppercase tracking-wider">Submitted</span>
                <span class="text-gray-300 font-medium">{new Date(report.created_at).toLocaleDateString()}</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="text-gray-500 font-bold uppercase tracking-wider">Source</span>
                <span class="text-primary font-black uppercase tracking-widest">{report.source || 'WEB'}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Guidelines -->
        <div class="bg-primary/5 border border-primary/10 rounded-3xl p-6 shadow-xl">
          <div class="flex items-center gap-2 text-primary font-black uppercase tracking-widest text-xs mb-4">
            <ShieldAlert size={14} />
            Admin Guidance
          </div>
          <p class="text-xs text-white/50 leading-relaxed italic">
            "Before taking action, verify if the reported behavior violates current community guidelines. Punitive actions should be proportional to the offense."
          </p>
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  :global(.bg-anirank-card) {
    background: rgba(13, 8, 18, 0.6);
    backdrop-filter: blur(20px);
  }
</style>
