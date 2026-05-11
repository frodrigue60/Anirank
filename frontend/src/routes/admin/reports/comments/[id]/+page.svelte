<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import MessageSquare from "lucide-svelte/icons/message-square";
  import User from "lucide-svelte/icons/user";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import ShieldAlert from "lucide-svelte/icons/shield-alert";
  import History from "lucide-svelte/icons/history";
  import ChevronDown from "lucide-svelte/icons/chevron-down";
  import ChevronUp from "lucide-svelte/icons/chevron-up";
  import { fade, scale } from "svelte/transition";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let report = $state(data.report);
  let showSnapshot = $state(false);

  const snapshotData = $derived(() => {
    if (!report?.snapshot) return null;
    try {
      return JSON.parse(report.snapshot);
    } catch (e) {
      console.error("Failed to parse snapshot:", e);
      return null;
    }
  });

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

<div class="space-y-6 max-w-5xl mx-auto px-4 py-6">
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
          Report <span class="text-on-surface/30">#{report?.id}</span>
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
          {report?.comment?.content || "Comment content not found (possibly deleted)"}
        </div>
        
        <div class="pt-4 border-t border-outline-variant space-y-2">
          <h3 class="font-semibold text-on-surface/60 uppercase text-xs tracking-wider">Reporter's Issue</h3>
          <p class="text-lg font-bold text-on-surface/90">{report?.title}</p>
          {#if report?.content}
            <p class="text-on-surface/70 bg-surface-highest rounded-lg p-3 text-sm">{report.content}</p>
          {/if}
        </div>
      </div>

      <!-- Snapshot Evidence (Immutable) -->
      {#if report.snapshot}
        <div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-sm">
          <div class="p-4 border-b border-outline-variant bg-white/2 flex justify-between items-center">
            <h2 class="font-bold text-on-surface flex items-center gap-2">
              <History size={18} class="text-amber-400" />
              Evidence Snapshot
            </h2>
            <button
              onclick={() => (showSnapshot = !showSnapshot)}
              class="text-xs font-bold text-primary hover:underline flex items-center gap-1"
            >
              {showSnapshot ? "Hide JSON" : "View Raw Source"}
              {#if showSnapshot}
                <ChevronUp size={14} />
              {:else}
                <ChevronDown size={14} />
              {/if}
            </button>
          </div>
          <div class="p-6">
            <div class="flex items-start gap-4 bg-amber-500/5 p-4 rounded-xl border border-amber-500/10 mb-4">
              <div class="p-2 bg-amber-500/20 rounded-lg text-amber-400 shrink-0">
                <ShieldAlert size={20} />
              </div>
              <div class="flex-1">
                <h4 class="text-sm font-bold text-amber-200">Point-in-time Evidence</h4>
                <p class="text-xs text-amber-500/70 mt-1 leading-relaxed">
                  Captured exact state when reported. This content is <b>immutable</b> and remains available even if the current comment is edited or purged.
                </p>
              </div>
            </div>

            {#if snapshotData()}
              <div class="space-y-4">
                <div class="space-y-1">
                  <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">Comment Content (Captured)</span>
                  <div class="bg-black/40 rounded-xl p-4 text-sm text-on-surface border border-outline-variant italic">
                    "{snapshotData().content}"
                  </div>
                </div>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                   <div class="space-y-1">
                      <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">UUID (Captured)</span>
                      <p class="text-[10px] font-mono text-on-surface-variant truncate">{snapshotData().uuid}</p>
                    </div>
                    <div class="space-y-1">
                      <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">Last Updated (Captured)</span>
                      <p class="text-[10px] font-mono text-on-surface-variant">{new Date(snapshotData().updated_at).toLocaleString()}</p>
                    </div>
                </div>
              </div>
            {/if}

            {#if showSnapshot}
              <div class="mt-6 border-t border-outline-variant pt-4" transition:fade>
                <pre class="bg-black/40 p-4 rounded-xl text-[10px] font-mono text-emerald-400/80 overflow-x-auto border border-outline-variant max-h-60 custom-scrollbar">{JSON.stringify(snapshotData(), null, 2)}</pre>
              </div>
            {/if}
          </div>
        </div>
      {/if}
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
                class="flex items-center justify-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 py-3 rounded-xl font-bold text-xs transition-all border border-blue-600/30 active:scale-95"
              >
                <CheckCircle2 size={16} />
                Accept
              </button>
              <button
                onclick={() => resolveReport(false)}
                class="flex items-center justify-center gap-2 bg-amber-600/20 hover:bg-amber-600/30 text-amber-400 py-3 rounded-xl font-bold text-xs transition-all border border-amber-600/30 active:scale-95"
              >
                <Trash2 size={16} />
                Reject
              </button>
            </div>
          {/if}

          <button
            onclick={deleteReport}
            class="w-full flex items-center justify-center gap-2 bg-red-500/20 hover:bg-red-500/30 text-red-400 py-3 rounded-xl font-bold text-sm transition-all active:scale-95"
          >
            <Trash2 size={18} />
            Delete Report
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
