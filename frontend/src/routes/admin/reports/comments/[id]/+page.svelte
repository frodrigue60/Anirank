<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import User from "lucide-svelte/icons/user";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import Trash2 from "lucide-svelte/icons/trash-2";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let report = $state(data.report);

  async function resolveReport(isAccepted: boolean) {
    if (!report) return;
    try {
      await api.put(`/admin/comments/reports/${report.id}/resolve`, { is_accepted: isAccepted });
      report.status = true;
      report.is_accepted = isAccepted;
    } catch (e) {
      console.error("Failed to resolve report:", e);
      alert("Error resolving report.");
    }
  }

  async function deleteReport() {
    if (!report || !confirm("Are you sure you want to delete this report?")) return;
    try {
      await api.delete(`/admin/comments/reports/${report.id}`);
      window.location.href = "/admin/reports/comments";
    } catch (e) {
      console.error("Failed to delete report:", e);
      alert("Error deleting report.");
    }
  }
</script>

<div class="space-y-6">
  <!-- Header & Navigation -->
  <div class="flex items-center gap-4">
    <a
      href="/admin/reports/comments"
      class="flex items-center justify-center w-10 h-10 rounded-xl bg-surface-highest hover:bg-surface-highest text-on-surface/60 hover:text-on-surface transition-all font-bold"
    >
      <ArrowLeft size={20} />
    </a>
    <div>
      <div class="flex items-center gap-3">
        <h1 class="text-3xl font-bold tracking-tight text-on-surface/90">
          Report #{report?.id}
        </h1>
        {#if !report?.status}
          <span
            class="px-3 py-1 bg-yellow-500/20 text-yellow-400 text-xs font-bold uppercase tracking-wider rounded-lg"
          >
            Pending
          </span>
        {:else}
          <div class="flex items-center gap-2">
            <span
              class="px-3 py-1 bg-green-500/20 text-green-400 text-xs font-bold uppercase tracking-wider rounded-lg"
            >
              Resolved
            </span>
            <span
              class="px-3 py-1 {report.is_accepted ? 'bg-blue-500/20 text-blue-400' : 'bg-red-500/20 text-red-400'} text-xs font-bold uppercase tracking-wider rounded-lg"
            >
              {report.is_accepted ? 'Accepted' : 'Rejected'}
            </span>
          </div>
        {/if}
      </div>
      <p class="text-on-surface/50 text-sm mt-1">
        Reported on {new Date(report?.created_at).toLocaleString()}
      </p>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Main Info (Left Column) -->
    <div class="lg:col-span-2 space-y-6">
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 space-y-4">
        <div class="flex items-center gap-3 border-b border-outline-variant pb-4">
          <MessageSquare size={24} class="text-blue-400" />
          <h2 class="text-xl font-bold text-on-surface">Reported Comment</h2>
        </div>
        
        <div class="bg-black/20 rounded-xl p-4 text-on-surface/80 whitespace-pre-wrap border-l-4 border-blue-600/50">
          {report?.comment?.content || "Comment content not found"}
        </div>
        
        <div class="pt-4 border-t border-outline-variant space-y-2">
          <h3 class="font-semibold text-on-surface/60 uppercase text-xs tracking-wider">Reporter's Issue</h3>
          <p class="text-lg font-bold text-on-surface/90">{report?.title}</p>
          {#if report?.content}
            <p class="text-on-surface/70 bg-surface-highest rounded-lg p-3 text-sm">{report.content}</p>
          {/if}
        </div>
      </div>
    </div>

    <!-- Sidebar Info (Right Column) -->
    <div class="space-y-6">
      <!-- Reporter Info -->
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 space-y-4">
        <h3 class="text-sm font-bold text-on-surface/60 uppercase tracking-wider border-b border-outline-variant pb-3">
          Reporter Details
        </h3>
        <div class="flex items-center gap-3">
          <div class="w-12 h-12 rounded-full bg-surface-highest flex items-center justify-center shrink-0">
            <User size={24} class="text-on-surface/40" />
          </div>
          <div>
            <p class="font-bold text-on-surface/90">{report?.user?.name}</p>
            <div class="flex items-center gap-2 mt-0.5">
              <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-sm bg-blue-500/10 text-blue-400 border border-blue-500/20">
                Score: {report?.user?.truth_score}
              </span>
              {#if report?.user?.is_shadowbanned}
                <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-sm bg-red-500/10 text-red-400 border border-red-500/20">
                  Shadowbanned
                </span>
              {/if}
            </div>
            <p class="text-xs text-on-surface/50 mt-1">ID: {report?.user?.id}</p>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="bg-surface-container border border-outline-variant rounded-2xl p-6 space-y-4">
        <h3 class="text-sm font-bold text-on-surface/60 uppercase tracking-wider border-b border-outline-variant pb-3">
          Actions
        </h3>
        
        <div class="space-y-3">
          {#if !report?.status}
            <div class="grid grid-cols-2 gap-2">
              <button
                onclick={() => resolveReport(true)}
                class="flex items-center justify-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 py-3 rounded-xl font-bold text-xs transition-all border border-blue-600/30"
              >
                <CheckCircle2 size={16} />
                Accept
              </button>
              <button
                onclick={() => resolveReport(false)}
                class="flex items-center justify-center gap-2 bg-amber-600/20 hover:bg-amber-600/30 text-amber-400 py-3 rounded-xl font-bold text-xs transition-all border border-amber-600/30"
              >
                <Trash2 size={16} />
                Reject
              </button>
            </div>
          {/if}

          <button
            onclick={deleteReport}
            class="w-full flex items-center justify-center gap-2 bg-red-500/20 hover:bg-red-500/30 text-red-400 py-3 rounded-xl font-bold text-sm transition-all"
          >
            <Trash2 size={18} />
            Delete Report
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
