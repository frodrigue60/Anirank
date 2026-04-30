<script lang="ts">
  import api from "$lib/api";
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
      await api.put(`/admin/songs/reports/${report.id}/resolve`);
      toastState.addToast("Report marked as resolved", "success");
      report.status = "fixed";
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

    {#if report?.status === "pending"}
      <div class="flex gap-3 w-full sm:w-auto">
        <button
          onclick={deleteReport}
          disabled={isDeleting || isResolving}
          class="flex-1 sm:flex-none px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 rounded-xl font-medium transition-all disabled:opacity-50"
        >
          {isDeleting ? "Deleting..." : "Delete Report"}
        </button>
        <button
          onclick={resolveReport}
          disabled={isResolving || isDeleting}
          class="flex-1 sm:flex-none px-6 py-2 bg-emerald-500 text-on-surface hover:bg-emerald-600 rounded-xl font-medium transition-all shadow-lg shadow-emerald-500/20 disabled:opacity-50"
        >
          {isResolving ? "Resolving..." : "Mark as Fixed"}
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
            <span
              class="px-2 py-0.5 rounded-full text-xs font-bold border capitalize {statusColors[
                report.status
              ] || ''}"
            >
              {report.status}
            </span>
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
                      href="/songs/{report.song?.anime?.slug}/{report.song?.slug}"
                      target="_blank"
                      class="text-lg font-bold text-on-surface hover:text-blue-400 transition-colors"
                    >
                      {report.song?.song_romaji ||
                        report.song?.song_en ||
                        report.song?.song_jp ||
                        "Unknown Song"}
                    </a>
                  </div>

                  <div class="text-sm text-on-surface-variant/40">
                    ID: {report.song?.id} • Slug: {report.song?.slug}
                  </div>
                </div>
                <a
                  href="/songs/{report.song?.anime?.slug}/{report.song?.slug}"
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
                <div class="text-xs text-on-surface-variant/40">#{report.user?.id}</div>
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
