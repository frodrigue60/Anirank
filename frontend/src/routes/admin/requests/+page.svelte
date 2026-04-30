<script lang="ts">
  import api from "$lib/api";

  import Loader2 from "lucide-svelte/icons/loader-2";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let requests = $state(data.requests || []);
  let currentStatus = $state("pending");
  let loadingStatus = $state(false);

  $effect(() => {
    // Only set initial data from SSR if it matches current status
    // Usually SSR loads 'pending'
    if (currentStatus === "pending" && !loadingStatus) {
      requests = data.requests || [];
    }
  });

  async function loadRequests(status: string) {
    if (currentStatus === status) return;
    currentStatus = status;
    loadingStatus = true;
    try {
      const res = await api.get(`/admin/user-requests?status=${status}`);
      requests = res.data.data || [];
    } catch (err) {
      console.error("Error loading requests:", err);
    } finally {
      loadingStatus = false;
    }
  }

  let loadingDelete = $state<number | null>(null);

  async function deleteRequest(id: number) {
    if (!confirm("Are you sure you want to delete this user request?")) return;

    loadingDelete = id;
    try {
      await api.delete(`/admin/user-requests/${id}`);
      requests = requests.filter((r: any) => r.id !== id);
    } catch (err) {
      console.error("Error deleting request:", err);
      alert("Failed to delete request.");
    } finally {
      loadingDelete = null;
    }
  }
</script>

<svelte:head>
  <title>User Requests | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      User Requests
    </h1>
    <p class="text-on-surface-variant/70">
      Manage catalog suggestions and general support tickets.
    </p>
  </div>
</div>

<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
  <div
    class="p-4 border-b border-outline-variant flex gap-2 overflow-x-auto custom-scrollbar"
  >
    <button
      onclick={() => loadRequests("pending")}
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {currentStatus ===
      'pending'
        ? 'bg-surface-highest text-on-surface'
        : 'hover:bg-surface-highest text-on-surface-variant/70'}">Pending</button
    >
    <button
      onclick={() => loadRequests("attended")}
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {currentStatus ===
      'attended'
        ? 'bg-surface-highest text-on-surface'
        : 'hover:bg-surface-highest text-on-surface-variant/70'}">Attended</button
    >
    {#if loadingStatus}
      <div class="flex items-center ml-2 text-on-surface-variant/70">
        <Loader2 size={16} class="animate-spin" />
      </div>
    {/if}
  </div>

  <table class="w-full text-left text-sm text-on-surface-variant">
    <thead
      class="text-xs text-on-surface-variant/70 uppercase bg-surface-highest border-b border-outline-variant"
    >
      <tr>
        <th class="px-6 py-4 font-semibold">ID</th>
        <th class="px-6 py-4 font-semibold">User</th>
        <th class="px-6 py-4 font-semibold w-1/2">Subject / Content</th>
        <th class="px-6 py-4 font-semibold">Date</th>
        <th class="px-6 py-4 font-semibold text-right">Actions</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-white/5">
      {#each requests as req}
        <tr class="hover:bg-white/2 transition-colors">
          <td class="px-6 py-4 font-medium text-on-surface-variant/40">#{req.id}</td>
          <td class="px-6 py-4">
            <div class="flex items-center gap-2">
              <div
                class="w-6 h-6 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0 uppercase text-xs font-bold"
              >
                {req.user?.name.charAt(0)}
              </div>
              <span class="font-medium text-on-surface">{req.user?.name}</span>
            </div>
          </td>
          <td class="px-6 py-4">
            <div class="font-medium text-on-surface mb-1">
              <a href="/admin/requests/{req.id}" class="hover:text-on-surface"
                >{req.title || "General Request"}</a
              >
            </div>
            <div class="text-on-surface-variant/70 text-xs line-clamp-2">{req.content}</div>
          </td>
          <td class="px-6 py-4 text-xs text-on-surface-variant/40">
            {new Date(req.created_at).toLocaleDateString()}
          </td>
          <td class="px-6 py-4 text-right">
            <button
              class="px-3 py-1.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 text-xs font-semibold rounded-lg transition-colors border border-rose-500/20 disabled:opacity-50"
              onclick={() => deleteRequest(req.id)}
              disabled={loadingDelete === req.id}
            >
              {loadingDelete === req.id ? "Deleting..." : "Delete"}
            </button>
          </td>
        </tr>
      {:else}
        <tr>
          <td colspan="5" class="px-6 py-12 text-center text-on-surface-variant/40">
            No {currentStatus} requests.
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
