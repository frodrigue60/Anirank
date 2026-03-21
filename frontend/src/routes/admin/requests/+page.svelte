<script lang="ts">
  import api from "$lib/api";

  let { data } = $props();
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
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      User Requests
    </h1>
    <p class="text-gray-400">
      Manage catalog suggestions and general support tickets.
    </p>
  </div>
</div>

<div class="bg-anirank-card border border-white/5 rounded-2xl overflow-hidden">
  <div
    class="p-4 border-b border-white/5 flex gap-2 overflow-x-auto custom-scrollbar"
  >
    <button
      onclick={() => loadRequests("pending")}
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {currentStatus ===
      'pending'
        ? 'bg-white/10 text-white'
        : 'hover:bg-white/5 text-gray-400'}">Pending</button
    >
    <button
      onclick={() => loadRequests("attended")}
      class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors {currentStatus ===
      'attended'
        ? 'bg-white/10 text-white'
        : 'hover:bg-white/5 text-gray-400'}">Attended</button
    >
    {#if loadingStatus}
      <div class="flex items-center ml-2 text-gray-400">
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
      </div>
    {/if}
  </div>

  <table class="w-full text-left text-sm text-gray-300">
    <thead
      class="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/5"
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
        <tr class="hover:bg-white/[0.02] transition-colors">
          <td class="px-6 py-4 font-medium text-gray-500">#{req.id}</td>
          <td class="px-6 py-4">
            <div class="flex items-center gap-2">
              <div
                class="w-6 h-6 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0 uppercase text-xs font-bold"
              >
                {req.user?.name.charAt(0)}
              </div>
              <span class="font-medium text-white">{req.user?.name}</span>
            </div>
          </td>
          <td class="px-6 py-4">
            <div class="font-medium text-white mb-1">
              <a href="/admin/requests/{req.id}" class="hover:text-white"
                >{req.title || "General Request"}</a
              >
            </div>
            <div class="text-gray-400 text-xs line-clamp-2">{req.content}</div>
          </td>
          <td class="px-6 py-4 text-xs text-gray-500">
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
          <td colspan="5" class="px-6 py-12 text-center text-gray-500">
            No {currentStatus} requests.
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
