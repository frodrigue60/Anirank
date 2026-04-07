<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import type { PageData } from "./$types";

  let { data } = $props<{ data: PageData }>();
  // svelte-ignore state_referenced_locally
  let req = $state(data.request);

  let loadingDelete = $state(false);
  let loadingStatus = $state(false);

  async function updateStatus(newStatus: string) {
    if (!req) return;

    const oldStatus = req.status;

    if (!confirm(`Mark this request as ${newStatus}?`)) {
      req.status = oldStatus;
      return;
    }

    loadingStatus = true;
    try {
      await api.patch(`/admin/user-requests/${req.id}/status`, {
        status: newStatus,
      });
      req.status = newStatus;
    } catch (err) {
      console.error("Error updating status:", err);
      alert("Failed to update status.");
      req.status = oldStatus;
    } finally {
      loadingStatus = false;
    }
  }

  async function deleteRequest() {
    if (!req) return;
    if (!confirm("Are you sure you want to delete this user request?")) return;

    loadingDelete = true;
    try {
      await api.delete(`/admin/user-requests/${req.id}`);
      goto("/admin/requests");
    } catch (err) {
      console.error("Error deleting request:", err);
      alert("Failed to delete request.");
    } finally {
      loadingDelete = false;
    }
  }

  function goBack() {
    window.history.back();
  }
</script>

<svelte:head>
  <title>View Request | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <!-- svelte-ignore a11y_consider_explicit_label -->
    <button
      onclick={goBack}
      class="text-on-surface-variant/70 hover:text-on-surface transition-colors p-2 -ml-2 rounded-lg hover:bg-surface-highest"
    >
      <svg
        class="w-5 h-5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 19l-7-7m0 0l7-7m-7 7h18"
        />
      </svg>
    </button>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface flex-1">
      {req ? req.title : "User Request"}
    </h1>
  </div>
</div>

{#if !req}
  <div
    class="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl flex items-center justify-between"
  >
    <p>Request not found.</p>
    <a
      href="/admin/requests"
      class="text-on-surface hover:underline text-sm font-medium">Back to List</a
    >
  </div>
{:else}
  <div class="max-w-4xl space-y-6">
    <div
      class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden"
    >
      <!-- Header -->
      <div
        class="p-6 border-b border-outline-variant flex flex-col md:flex-row gap-4 items-start justify-between"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-12 h-12 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center text-xl font-bold uppercase shrink-0"
          >
            {req.user?.name.charAt(0)}
          </div>
          <div>
            <div class="font-medium text-on-surface text-lg">{req.user?.name}</div>
            <div class="text-sm text-on-surface-variant/40">
              Submitted on {new Date(req.created_at).toLocaleString()}
            </div>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <select
            class="px-3 py-1 pr-8 rounded-full text-xs font-semibold uppercase tracking-wide border cursor-pointer focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-[#1a1b1e] {req.status ===
            'pending'
              ? 'bg-amber-500/10 text-amber-500 border-amber-500/20 focus:ring-amber-500/50'
              : 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20 focus:ring-emerald-500/50'}"
            value={req.status}
            onchange={(e) => updateStatus(e.currentTarget.value)}
            disabled={loadingStatus}
            aria-label="Request Status"
          >
            <option
              value="pending"
              class="bg-gray-800 text-on-surface font-sans uppercase">pending</option
            >
            <option
              value="attended"
              class="bg-gray-800 text-on-surface font-sans uppercase"
              >attended</option
            >
          </select>
        </div>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-6">
        <div>
          <h2
            class="text-xl font-medium text-on-surface-variant/70 uppercase tracking-widest mb-2"
          >
            {req.title}
          </h2>
        </div>
        <div>
          <h3
            class="text-sm font-medium text-on-surface-variant/70 uppercase tracking-widest mb-2"
          >
            Message
          </h3>
          <div
            class="bg-surface-highest border border-outline-variant rounded-xl p-5 text-on-surface-variant leading-relaxed whitespace-pre-wrap"
          >
            {req.content}
          </div>
        </div>
      </div>

      <!-- Actions Footer -->
      <div
        class="p-6 bg-black/20 border-t border-outline-variant flex items-center justify-end gap-3"
      >
        <button
          onclick={deleteRequest}
          disabled={loadingDelete}
          class="px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 font-medium rounded-xl transition-all border border-rose-500/20 disabled:opacity-50 flex items-center gap-2"
        >
          {#if loadingDelete}
            <svg class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Deleting...
          {:else}
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              /></svg
            >
            Delete Request
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
