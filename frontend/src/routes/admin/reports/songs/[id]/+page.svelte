<script lang="ts">
  import api from "$lib/api";
  import { getSongName } from "$lib/song-utils";
  import { goto } from "$app/navigation";
  import { toastState } from "$lib/state/toast.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import Inbox from "lucide-svelte/icons/inbox";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import Crosshair from "lucide-svelte/icons/crosshair";
  import Music from "lucide-svelte/icons/music";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import User from "lucide-svelte/icons/user";
  import Gavel from "lucide-svelte/icons/gavel";
  import History from "lucide-svelte/icons/history";
  import ChevronDown from "lucide-svelte/icons/chevron-down";
  import ChevronUp from "lucide-svelte/icons/chevron-up";
  import ShieldAlert from "lucide-svelte/icons/shield-alert";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let report = $state(data.report);

  $effect(() => {
    report = data.report;
  });

  let isResolving = $state(false);
  let isDeleting = $state(false);
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
    isResolving = true;
    try {
      await api.put(`/admin/songs/reports/${report.id}/resolve`, { is_accepted: isAccepted });
      toastState.addToast(`Report marked as ${isAccepted ? 'Accepted' : 'Rejected'}`, "success");
      report.status = true;
      report.is_accepted = isAccepted;
      goto("/admin/reports/songs");
    } catch (err: any) {
      console.error("Error resolving report:", err);
      toastState.addToast(
        err.response?.data?.message || "Failed to resolve report",
        "error",
      );
    } finally {
      isResolving = false;
    }
  }

  async function deleteReport() {
    if (!report) return;
    if (!confirm("Are you sure you want to delete this report?")) return;

    isDeleting = true;
    try {
      await api.delete(`/admin/songs/reports/${report.id}`);
      toastState.addToast("Report deleted", "success");
      goto("/admin/reports/songs");
    } catch (err: any) {
      console.error("Error deleting report:", err);
      toastState.addToast("Failed to delete report", "error");
    } finally {
      isDeleting = false;
    }
  }

  const statusColors: Record<string, string> = {
    pending: "bg-amber-500/10 text-amber-500 border-amber-500/20",
    fixed: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20",
  };
</script>

<svelte:head>
  <title>Report #{report?.id || "..."} | Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
  <!-- Header -->
  <div
    class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
  >
    <div>
      <div class="flex items-center gap-3 mb-2">
        <a
          href="/admin/reports/songs"
          class="p-2 bg-surface-highest hover:bg-surface-highest rounded-full text-on-surface-variant/70 font-bold transition-colors"
        >
          <ArrowLeft size={20} />
        </a>
        <h1 class="text-3xl font-bold tracking-tight text-on-surface">
          Report <span class="text-on-surface/40">#{report?.id}</span>
        </h1>
      </div>
      <p class="text-on-surface-variant/70">Reviewing violation report submitted by user.</p>
    </div>

    {#if !report?.status}
      <div class="flex gap-3 w-full sm:w-auto">
        <button
          onclick={deleteReport}
          disabled={isDeleting || isResolving}
          class="flex-1 sm:flex-none px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 rounded-xl font-medium transition-all disabled:opacity-50"
        >
          {isDeleting ? "Deleting..." : "Delete Report"}
        </button>
        <button
          onclick={() => resolveReport(true)}
          disabled={isResolving || isDeleting}
          class="flex-1 sm:flex-none px-6 py-2 bg-blue-600 text-on-surface hover:bg-blue-700 rounded-xl font-medium transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50"
        >
          {isResolving ? "Accepting..." : "Accept (Valid)"}
        </button>
        <button
          onclick={() => resolveReport(false)}
          disabled={isResolving || isDeleting}
          class="flex-1 sm:flex-none px-6 py-2 bg-amber-600 text-on-surface hover:bg-amber-700 rounded-xl font-medium transition-all shadow-lg shadow-amber-500/20 disabled:opacity-50"
        >
          {isResolving ? "Rejecting..." : "Reject (False)"}
        </button>
      </div>
    {/if}
  </div>

  {#if !report}
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl p-12 text-center text-on-surface-variant/40"
    >
      <Inbox size={48} class="mb-4 opacity-20" />

      <p>Report not found or failed to load.</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Content -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Reason/Title -->
        <div
          class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden"
        >
          <div
            class="p-4 border-b border-outline-variant bg-surface-highest flex justify-between items-center"
          >
            <h2 class="font-bold text-on-surface flex items-center gap-2">
              <AlertCircle size={18} class="text-rose-400" />

              Report Content
            </h2>
            <div class="flex items-center gap-2">
              <span
                class="px-2 py-0.5 rounded-full text-xs font-bold border capitalize {report.status ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-amber-500/10 text-amber-500 border-amber-500/20'}"
              >
                {report.status ? 'Resolved' : 'Pending'}
              </span>
              {#if report.status}
                <span
                  class="px-2 py-0.5 rounded-full text-xs font-bold border capitalize {report.is_accepted ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20'}"
                >
                  {report.is_accepted ? 'Accepted' : 'Rejected'}
                </span>
              {/if}
            </div>
          </div>
          <div class="p-6">
            <h3 class="text-xl font-bold text-on-surface mb-4">{report.title}</h3>
            <div
              class="bg-black/20 rounded-xl p-5 text-on-surface-variant leading-relaxed border border-outline-variant whitespace-pre-wrap"
            >
              {report.content}
            </div>
          </div>
        </div>

        <!-- Target Context -->
        <div
          class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden"
        >
          <div class="p-4 border-b border-outline-variant bg-surface-highest">
            <h2 class="font-bold text-on-surface flex items-center gap-2">
              <Crosshair size={18} class="text-blue-400" />

              Target Entity
            </h2>
          </div>
          <div class="p-6">
            {#if report.song}
              <div
                class="flex items-center gap-4 bg-surface-highest p-4 rounded-2xl border border-outline-variant"
              >
                <div
                  class="w-16 h-16 bg-blue-500/20 rounded-xl flex items-center justify-center text-blue-400"
                >
                  <Music size={24} />

                </div>
                <div>
                  <div
                    class="text-xs text-blue-400 font-bold uppercase tracking-wider mb-1"
                  >
                    <a
                      href="/animes/{report.song?.anime?.slug}/{report.song?.slug}"
                      target="_blank"
                      class="text-lg font-bold text-on-surface hover:text-blue-400 transition-colors"
                    >
                      {getSongName(report.song)}
                    </a>
                  </div>

                  <div class="text-sm text-on-surface-variant/40">
                    ID: {report.song?.id} • Slug: {report.song?.slug}
                  </div>
                </div>
                <a
                  href="/animes/{report.song?.anime?.slug}/{report.song?.slug}"
                  target="_blank"
                  class="ml-auto p-2 hover:bg-surface-highest rounded-lg text-on-surface-variant/70"
                >
                  <ExternalLink size={18} />
                </a>
              </div>
            {:else}
              <div
                class="p-4 bg-surface-highest rounded-xl text-on-surface-variant/40 text-center italic border border-outline-variant"
              >
                Target context information is unavailable or could not be
                loaded.
              </div>
            {/if}
          </div>
        </div>

        <!-- Snapshot Evidence (Immutable) -->
        {#if report.snapshot}
          <div
            class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden shadow-sm"
          >
            <div
              class="p-4 border-b border-outline-variant bg-white/2 flex justify-between items-center"
            >
              <h2 class="font-bold text-on-surface flex items-center gap-2">
                <History size={18} class="text-amber-400" />
                Snapshot Evidence
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
                <div>
                  <h4 class="text-sm font-bold text-amber-200">Point-in-time Evidence</h4>
                  <p class="text-xs text-amber-500/70 mt-1">
                    This data represents the exact state of the entity when it was reported. 
                    It is immutable and serves as primary evidence even if the current entity is deleted.
                  </p>
                </div>
              </div>

              {#if snapshotData()}
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="space-y-1">
                    <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">Song Name (Captured)</span>
                    <p class="text-on-surface font-medium truncate">{snapshotData().song_romaji || snapshotData().song_en || 'Unknown'}</p>
                  </div>
                  <div class="space-y-1">
                    <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">Type (Captured)</span>
                    <p class="text-on-surface font-medium capitalize">{snapshotData().type || 'N/A'}</p>
                  </div>
                  <div class="space-y-1">
                    <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">Slug (Captured)</span>
                    <p class="text-on-surface font-mono text-xs">{snapshotData().slug}</p>
                  </div>
                  <div class="space-y-1">
                    <span class="text-[10px] font-bold text-on-surface-variant/40 uppercase tracking-widest">UUID (Captured)</span>
                    <p class="text-on-surface font-mono text-xs truncate">{snapshotData().uuid}</p>
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

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Reporter Info -->
        <div
          class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden"
        >
          <div class="p-4 border-b border-outline-variant bg-surface-highest">
            <h2 class="font-bold text-on-surface flex items-center gap-2">
              <User size={18} class="text-on-surface-variant/70" />

              Reporter
            </h2>
          </div>
          <div class="p-6">
            <div class="flex items-center gap-3 mb-4">
              <div
                class="w-12 h-12 rounded-full bg-linear-to-br from-blue-500 to-purple-500 flex items-center justify-center text-on-surface font-bold text-lg"
              >
                {report.user?.name?.[0].toUpperCase() || "?"}
              </div>
              <div>
                <a
                  href="/users/{report.user?.slug}"
                  target="_blank"
                  class="font-bold text-on-surface hover:text-blue-400 hover:underline transition-all"
                >
                  {report.user?.name}
                </a>
                <div class="flex items-center gap-2 mt-0.5">
                  <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-sm bg-blue-500/10 text-blue-400 border border-blue-500/20">
                    Score: {report.user?.truth_score}
                  </span>
                  {#if report.user?.is_shadowbanned}
                    <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-sm bg-red-500/10 text-red-400 border border-red-500/20">
                      Shadowbanned
                    </span>
                  {/if}
                </div>
                <div class="text-xs text-on-surface-variant/40 mt-1">#{report.user?.id}</div>
              </div>
            </div>
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-on-surface-variant/40">Submitted</span>
                <span class="text-on-surface-variant"
                  >{new Date(report.created_at).toLocaleDateString()}</span
                >
              </div>
              <div class="flex justify-between">
                <span class="text-on-surface-variant/40">Source</span>
                <span class="text-on-surface-variant capitalize">{report.source}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Guidelines Check -->
        <div class="bg-amber-500/5 border border-amber-500/10 rounded-2xl p-6">
          <div class="flex items-center gap-2 text-amber-500 font-bold mb-3">
            <Gavel size={18} />
            <span>Moderator Tip</span>
          </div>
          <p class="text-sm text-amber-500/80 leading-relaxed italic">
            "Review the reported content carefully against our community
            guidelines. If the report is valid, proceed with taking action on
            the target content."
          </p>
        </div>
      </div>
    </div>
  {/if}
</div>
