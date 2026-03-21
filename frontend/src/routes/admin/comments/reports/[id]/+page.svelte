<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let report = $state(data.report);

  async function resolveReport() {
    if (!report) return;
    try {
      await api.put(`/admin/comments/reports/${report.id}/resolve`);
      report.status = "fixed";
    } catch (e) {
      console.error("Failed to resolve report:", e);
      alert("Error resolving report.");
    }
  }

  async function deleteReport() {
    if (!report || !confirm("Are you sure you want to delete this report?")) return;
    try {
      await api.delete(`/admin/comments/reports/${report.id}`);
      window.location.href = "/admin/comments/reports";
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
      href="/admin/comments/reports"
      class="flex items-center justify-center w-10 h-10 rounded-xl bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-all"
    >
      <span class="material-symbols-outlined">arrow_back</span>
    </a>
    <div>
      <div class="flex items-center gap-3">
        <h1 class="text-3xl font-bold tracking-tight text-white/90">
          Report #{report?.id}
        </h1>
        {#if report?.status === "pending"}
          <span
            class="px-3 py-1 bg-yellow-500/20 text-yellow-400 text-xs font-bold uppercase tracking-wider rounded-lg"
          >
            Pending
          </span>
        {:else}
          <span
            class="px-3 py-1 bg-green-500/20 text-green-400 text-xs font-bold uppercase tracking-wider rounded-lg"
          >
            Resolved
          </span>
        {/if}
      </div>
      <p class="text-white/50 text-sm mt-1">
        Reported on {new Date(report?.created_at).toLocaleString()}
      </p>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Main Info (Left Column) -->
    <div class="lg:col-span-2 space-y-6">
      <div class="bg-white/5 border border-white/5 rounded-2xl p-6 space-y-4">
        <div class="flex items-center gap-3 border-b border-white/5 pb-4">
          <span class="material-symbols-outlined text-primary text-2xl">forum</span>
          <h2 class="text-xl font-bold">Reported Comment</h2>
        </div>
        
        <div class="bg-black/20 rounded-xl p-4 text-white/80 whitespace-pre-wrap border-l-4 border-primary/50">
          {report?.comment?.content || "Comment content not found"}
        </div>
        
        <div class="pt-4 border-t border-white/5 space-y-2">
          <h3 class="font-semibold text-white/60 uppercase text-xs tracking-wider">Reporter's Issue</h3>
          <p class="text-lg font-bold text-white/90">{report?.title}</p>
          {#if report?.content}
            <p class="text-white/70 bg-white/5 rounded-lg p-3 text-sm">{report.content}</p>
          {/if}
        </div>
      </div>
    </div>

    <!-- Sidebar Info (Right Column) -->
    <div class="space-y-6">
      <!-- Reporter Info -->
      <div class="bg-white/5 border border-white/5 rounded-2xl p-6 space-y-4">
        <h3 class="text-sm font-bold text-white/60 uppercase tracking-wider border-b border-white/5 pb-3">
          Reporter Details
        </h3>
        <div class="flex items-center gap-3">
          <div class="w-12 h-12 rounded-full bg-white/10 flex items-center justify-center shrink-0">
            <span class="material-symbols-outlined text-white/40">person</span>
          </div>
          <div>
            <p class="font-bold text-white/90">{report?.user?.name}</p>
            <p class="text-xs text-white/50">ID: {report?.user?.id}</p>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="bg-white/5 border border-white/5 rounded-2xl p-6 space-y-4">
        <h3 class="text-sm font-bold text-white/60 uppercase tracking-wider border-b border-white/5 pb-3">
          Actions
        </h3>
        
        <div class="space-y-3">
          {#if report?.status === "pending"}
            <button
              onclick={resolveReport}
              class="w-full flex items-center justify-center gap-2 bg-green-500/20 hover:bg-green-500/30 text-green-400 py-3 rounded-xl font-bold text-sm transition-all"
            >
              <span class="material-symbols-outlined">check_circle</span>
              Mark as Resolved
            </button>
          {/if}

          <button
            onclick={deleteReport}
            class="w-full flex items-center justify-center gap-2 bg-red-500/20 hover:bg-red-500/30 text-red-400 py-3 rounded-xl font-bold text-sm transition-all"
          >
            <span class="material-symbols-outlined">delete</span>
            Delete Report
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
