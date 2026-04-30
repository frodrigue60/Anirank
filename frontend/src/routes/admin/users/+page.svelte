<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import Edit2 from "lucide-svelte/icons/edit-2";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Search from "lucide-svelte/icons/search";
  import ChevronLeft from "lucide-svelte/icons/chevron-left";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  let { data } = $props();

  let searchQuery = $state("");
  let isDeleting = $state(false);

  function handleSearch() {
    goto(`/admin/users?search=${searchQuery}&page=1`, { keepFocus: true });
  }

  function changePage(newPage: number) {
    if (newPage >= 1 && newPage <= data.pagination.last_page) {
      goto(`/admin/users?search=${searchQuery}&page=${newPage}`);
    }
  }

  async function deleteUser(uuid: string) {
    if (
      !confirm(
        "Are you sure you want to delete this user? This action cannot be undone.",
      )
    )
      return;

    isDeleting = true;
    try {
      await api.delete(`/admin/users/${uuid}`);
      toastState.addToast("User deleted successfully", "success");
      // Refresh page data
      goto(
        `/admin/users?search=${searchQuery}&page=${data.pagination.current_page}`,
        { invalidateAll: true },
      );
    } catch (err: any) {
      const msg = err.response?.data?.message || "Failed to delete user.";
      toastState.addToast(msg, "error");
      console.error(err);
    } finally {
      isDeleting = false;
    }
  }
</script>

<svelte:head>
  <title>User Management | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
      Users & Access
    </h1>
    <p class="text-on-surface-variant/70">
      Manage user accounts, enforce moderation, and update roles.
    </p>
  </div>
  <div>
    <button
      onclick={() => goto("/admin/users/create")}
      class="px-4 py-2 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-xl transition-colors border border-outline-variant shrink-0"
    >
      Create User
    </button>
  </div>
</div>

<!-- Filters & Search -->
<div
  class="bg-surface-container border border-outline-variant rounded-2xl p-4 mb-6 flex flex-col sm:flex-row gap-4"
>
  <div class="relative flex-1">
    <Search size={20} class="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant/40" />
    <input
      type="text"
      bind:value={searchQuery}
      onkeydown={(e) => e.key === "Enter" && handleSearch()}
      placeholder="Search by username or email..."
      class="w-full bg-surface-highest border border-outline-variant rounded-xl py-2 pl-10 pr-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors"
    />
  </div>
  <button
    onclick={handleSearch}
    class="px-4 py-2 bg-surface-highest hover:bg-surface-highest text-on-surface rounded-xl transition-colors border border-outline-variant shrink-0"
  >
    Search Accounts
  </button>
</div>

<!-- Table -->
<div class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden">
  <div class="overflow-x-auto">
    <table class="w-full text-left text-sm text-on-surface-variant">
      <thead
        class="text-xs text-on-surface-variant/70 uppercase bg-surface-highest border-b border-outline-variant"
      >
        <tr>
          <th class="px-6 py-4 font-semibold">User</th>
          <th class="px-6 py-4 font-semibold">Joined Date</th>
          <th class="px-6 py-4 font-semibold text-center">Status</th>
          <th class="px-6 py-4 font-semibold text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-white/5">
        {#each data.users as user}
          <tr class="hover:bg-white/2 transition-colors group">
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                {#if user.avatar_url}
                  <img
                    src={user.avatar_url}
                    alt={user.name}
                    class="w-10 h-10 rounded-full object-cover shadow-sm bg-surface-highest border border-outline-variant"
                  />
                {:else}
                  <div
                    class="w-10 h-10 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center font-bold text-lg uppercase shadow-sm border border-blue-500/30"
                  >
                    {user.name.charAt(0)}
                  </div>
                {/if}
                <div>
                  <div class="font-bold text-on-surface mb-0.5">
                    <a
                      href="/admin/users/{user.id}"
                      class="hover:text-primary transition-colors"
                      >{user.name}</a
                    >
                  </div>
                  <div class="text-xs text-on-surface-variant/40 font-mono hidden sm:block">
                    ID: {user.id}
                  </div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 text-on-surface-variant/70 text-sm">
              {new Date(user.created_at).toLocaleDateString()}
            </td>
            <td class="px-6 py-4 text-center">
              <span
                class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Active
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">

                <a
                  href="/admin/users/{user.id}/edit"
                  class="p-2 text-on-surface-variant/70 hover:text-on-surface hover:bg-emerald-500/10 rounded-lg transition-colors border border-transparent hover:border-emerald-500/10"
                  title="Edit Profile"
                >
                  <Edit2 size={18} />
                </a>

                <button
                  onclick={() => deleteUser(user.id)}
                  disabled={isDeleting}
                  class="p-2 text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 rounded-lg transition-colors border border-transparent hover:border-rose-500/20 disabled:opacity-50"
                  title="Delete User"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="4" class="px-6 py-12 text-center text-on-surface-variant/40">
              No users found matching that search.
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if data.pagination?.last_page > 1}
    <div
      class="px-6 py-4 border-t border-outline-variant flex items-center justify-between"
    >
      <div class="text-sm text-on-surface-variant/70">
        Showing <span class="font-medium text-on-surface">{data.users.length}</span> items
      </div>
      <div class="flex items-center gap-2">
        <button
          disabled={data.pagination.current_page === 1}
          onclick={() => changePage(data.pagination.current_page - 1)}
          class="p-2 rounded-lg border border-outline-variant text-on-surface-variant/70 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-highest transition-colors"
          aria-label="Previous Page"
        >
          <ChevronLeft size={16} />
        </button>
        <span class="text-sm text-on-surface-variant font-medium px-2"
          >Page {data.pagination.current_page} of {data.pagination
            .last_page}</span
        >
        <button
          disabled={data.pagination.current_page === data.pagination.last_page}
          onclick={() => changePage(data.pagination.current_page + 1)}
          class="p-2 rounded-lg border border-outline-variant text-on-surface-variant/70 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-highest transition-colors"
          aria-label="Next Page"
        >
          <ChevronRight size={16} />
        </button>
      </div>
    </div>
  {/if}
</div>
