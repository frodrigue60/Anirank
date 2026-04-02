<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";

  let { data } = $props();

  let name = $state("");
  let email = $state("");

  let selectedRoleIds = $state<number[]>([]);
  let selectedBadgeIds = $state<number[]>([]);

  $effect(() => {
    name = data.user.name;
    email = data.user.email;
    selectedRoleIds = data.user.roles?.map((r: any) => r.id) || [];
    selectedBadgeIds = data.user.badges?.map((b: any) => b.id) || [];
  });

  let isSaving = $state(false);
  let isResetting = $state(false);
  let newPassword = $state<string | null>(null);

  function toggleRole(roleId: number) {
    if (selectedRoleIds.includes(roleId)) {
      selectedRoleIds = selectedRoleIds.filter((id) => id !== roleId);
    } else {
      selectedRoleIds = [...selectedRoleIds, roleId];
    }
  }

  function toggleBadge(badgeId: number) {
    if (selectedBadgeIds.includes(badgeId)) {
      selectedBadgeIds = selectedBadgeIds.filter((id) => id !== badgeId);
    } else {
      selectedBadgeIds = [...selectedBadgeIds, badgeId];
    }
  }

  async function copyPassword() {
    if (newPassword) {
      await navigator.clipboard.writeText(newPassword);
      toastState.addToast("Password copied to clipboard!", "success");
    }
  }

  async function handleResetPassword() {
    if (
      !confirm(
        `Are you sure you want to reset the password for ${data.user.name}?`,
      )
    )
      return;

    isResetting = true;
    newPassword = null;
    try {
      const res = await api.post(`/admin/users/${data.user.uuid}/reset-password`);
      newPassword = res.data.password;
      toastState.addToast("Password reset successfully", "success");
    } catch (err: any) {
      console.error("Failed to reset password", err);
      const msg = err.response?.data?.message || "Failed to reset password.";
      toastState.addToast(msg, "error");
    } finally {
      isResetting = false;
    }
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    isSaving = true;

    try {
      await api.put(`/admin/users/${data.user.uuid}`, {
        name,
        email,
        role_ids: selectedRoleIds,
        badge_ids: selectedBadgeIds,
      });
      toastState.addToast("User updated successfully", "success");
      goto("/admin/users");
    } catch (err: any) {
      console.error("Failed to update user", err);
      const msg = err.response?.data?.message || "Uh oh! Failed to update the user details.";
      toastState.addToast(msg, "error");
    } finally {
      isSaving = false;
    }
  }
</script>

<svelte:head>
  <title>Edit User | Admin</title>
</svelte:head>

<div class="mb-8">
  <div class="flex items-center gap-4 mb-2">
    <a
      href="/admin/users"
      class="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors border border-transparent hover:border-white/10"
      title="Back to Users"
    >
      <span class="material-symbols-outlined text-xl">arrow_back</span>
    </a>
    <h1 class="text-3xl font-bold tracking-tight text-white">Edit User</h1>
  </div>
  <p class="text-gray-400 pl-14">
    Modify <span class="text-anirank-primary font-medium">{data.user.name}</span
    >'s profile details and account access privileges.
  </p>
</div>

<form onsubmit={handleSubmit} class="space-y-6 max-w-4xl">
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Left Column: Basic Info -->
    <div class="space-y-6">
      <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
        <h2
          class="text-lg font-semibold text-white border-b border-white/10 pb-3 mb-4"
        >
          Profile Details
        </h2>

        <div class="space-y-4">
          <!-- Name -->
          <div>
            <label
              for="name"
              class="block text-sm font-medium text-gray-300 mb-1"
            >
              Username
            </label>
            <input
              type="text"
              id="name"
              bind:value={name}
              required
              class="w-full bg-white/5 border border-white/10 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors"
            />
          </div>

          <!-- Email -->
          <div>
            <label
              for="email"
              class="block text-sm font-medium text-gray-300 mb-1"
            >
              Email Address
            </label>
            <input
              type="email"
              id="email"
              bind:value={email}
              required
              class="w-full bg-white/5 border border-white/10 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors focus:ring-1 focus:ring-anirank-primary/50"
            />
          </div>
        </div>
      </div>

      <!-- Password Reset Section -->
      <div class="bg-anirank-card border border-rose-500/20 rounded-2xl p-6">
        <h2
          class="text-lg font-semibold text-rose-400 border-b border-rose-500/10 pb-3 mb-4 flex items-center gap-2"
        >
          <span class="material-symbols-outlined">lock_reset</span>
          Security
        </h2>
        <p class="text-sm text-gray-400 mb-4">
          If the user lost access to their account, you can generate a new
          temporary secure password for them.
        </p>

        {#if newPassword}
          <div
            class="mb-4 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl"
          >
            <p class="text-emerald-400 text-sm font-medium mb-1">
              New password generated successfully:
            </p>
            <div
              class="flex items-center gap-3 mt-2"
            >
              <div class="text-white font-mono bg-black/30 px-3 py-2 rounded-lg inline-block text-lg tracking-wider">
                 {newPassword}
              </div>
              <button
                type="button"
                onclick={copyPassword}
                class="flex items-center gap-1 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-500/10 px-3 py-2 rounded-lg transition-colors text-sm font-medium border border-transparent hover:border-emerald-500/20"
              >
                <span class="material-symbols-outlined text-[18px]">content_copy</span>
                Copy
              </button>
            </div>
            <p class="text-xs text-gray-500 mt-2">
              Make sure to copy this password now, it won't be shown again.
            </p>
          </div>
        {/if}

        <button
          type="button"
          onclick={handleResetPassword}
          disabled={isResetting}
          class="px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 rounded-xl transition-colors border border-rose-500/20 text-sm font-medium flex items-center gap-2 disabled:opacity-50"
        >
          {#if isResetting}
            <span class="material-symbols-outlined animate-spin text-[18px]"
              >progress_activity</span
            >
            Generating...
          {:else}
            <span class="material-symbols-outlined text-[18px]">key</span>
            Generate Random Password
          {/if}
        </button>
      </div>
    </div>

    <!-- Right Column: RBAC & Rewards -->
    <div class="space-y-6">
      <!-- Roles Selection -->
      <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
        <h2
          class="text-lg font-semibold text-white border-b border-white/10 pb-3 mb-4 flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-rose-400"
            >shield_person</span
          >
          Access Roles
        </h2>

        <div class="space-y-3 max-h-48 overflow-y-auto pr-2 custom-scrollbar">
          {#each data.allRoles as role}
            <label
              class="flex items-center gap-3 p-3 rounded-xl border border-white/5 bg-white/5 hover:bg-white/10 transition-colors cursor-pointer {selectedRoleIds.includes(
                role.id,
              )
                ? 'border-anirank-primary/50 bg-anirank-primary/10'
                : ''}"
            >
              <div
                class="relative flex items-center border border-white/20 rounded-md w-5 h-5 bg-black/20 overflow-hidden shrink-0 {selectedRoleIds.includes(
                  role.id,
                )
                  ? 'border-anirank-primary bg-anirank-primary'
                  : ''}"
              >
                <input
                  type="checkbox"
                  class="absolute opacity-0 cursor-pointer"
                  checked={selectedRoleIds.includes(role.id)}
                  onchange={() => toggleRole(role.id)}
                />
                {#if selectedRoleIds.includes(role.id)}
                  <span class="material-symbols-outlined text-[16px] text-white"
                    >check</span
                  >
                {/if}
              </div>
              <div>
                <div class="text-sm font-medium text-white">{role.name}</div>
                <div class="text-xs text-gray-500">{role.description}</div>
              </div>
            </label>
          {/each}
        </div>
      </div>

      <!-- Badges Selection -->
      <div class="bg-anirank-card border border-white/5 rounded-2xl p-6">
        <h2
          class="text-lg font-semibold text-white border-b border-white/10 pb-3 mb-4 flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-yellow-400"
            >military_tech</span
          >
          Earned Badges
        </h2>

        <div
          class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-60 overflow-y-auto pr-2 custom-scrollbar"
        >
          {#each data.allBadges as badge}
            <label
              class="flex items-center gap-3 p-2 rounded-xl border border-white/5 bg-white/5 hover:bg-white/10 transition-colors cursor-pointer {selectedBadgeIds.includes(
                badge.id,
              )
                ? 'border-anirank-primary/50 bg-anirank-primary/10'
                : ''}"
            >
              <div
                class="relative flex items-center border border-white/20 rounded-md w-5 h-5 bg-black/20 overflow-hidden shrink-0 {selectedBadgeIds.includes(
                  badge.id,
                )
                  ? 'border-anirank-primary bg-anirank-primary'
                  : ''}"
              >
                <input
                  type="checkbox"
                  class="absolute opacity-0 cursor-pointer"
                  checked={selectedBadgeIds.includes(badge.id)}
                  onchange={() => toggleBadge(badge.id)}
                />
                {#if selectedBadgeIds.includes(badge.id)}
                  <span class="material-symbols-outlined text-[16px] text-white"
                    >check</span
                  >
                {/if}
              </div>
              <div class="flex items-center gap-2 overflow-hidden">
                {#if badge.icon_url}
                  <img
                    src={badge.icon_url}
                    alt={badge.name}
                    class="w-6 h-6 object-contain shrink-0"
                  />
                {/if}
                <div class="text-sm font-medium text-white truncate">
                  {badge.name}
                </div>
              </div>
            </label>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <!-- Form Actions -->
  <div
    class="flex items-center justify-end gap-3 pt-6 border-t border-white/10"
  >
    <a
      href="/admin/users"
      class="px-5 py-2.5 rounded-xl border border-white/10 text-white hover:bg-white/5 transition-colors font-medium"
      onclick={(e) => {
        if (isSaving) e.preventDefault();
      }}
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={isSaving}
      class="px-6 py-2.5 bg-anirank-primary hover:bg-anirank-secondary text-white rounded-xl transition-colors font-medium shadow-lg shadow-anirank-primary/20 disabled:opacity-70 flex items-center gap-2"
    >
      {#if isSaving}
        <span
          class="material-symbols-outlined animate-spin align-middle shrink-0"
          >progress_activity</span
        >
        Saving Changes...
      {:else}
        <span class="material-symbols-outlined align-middle shrink-0">save</span
        >
        Save User Details
      {/if}
    </button>
  </div>
</form>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.02);
    border-radius: 8px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 8px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
