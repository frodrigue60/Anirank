<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";

  let { data } = $props();

  let name = $state("");
  let email = $state("");
  let password = $state("");

  let selectedRoleIds = $state<number[]>([]);
  let selectedBadgeIds = $state<number[]>([]);

  let isSaving = $state(false);

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

  function generatePassword() {
    const charset =
      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*";
    let pw = "";
    for (let i = 0; i < 12; i++) {
      pw += charset.charAt(Math.floor(Math.random() * charset.length));
    }
    password = pw;
  }

  async function copyPassword() {
    if (password) {
      await navigator.clipboard.writeText(password);
      toastState.addToast("Password copied to clipboard!", "success");
    }
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    isSaving = true;

    try {
      await api.post(`/admin/users`, {
        name,
        email,
        password,
        role_ids: selectedRoleIds,
        badge_ids: selectedBadgeIds,
      });
      goto("/admin/users");
    } catch (err: any) {
      console.error("Failed to create user", err);
      const msg =
        err.response?.data?.message || "Uh oh! Failed to create the user.";
      alert(msg);
    } finally {
      isSaving = false;
    }
  }
</script>

<svelte:head>
  <title>Create User | Admin</title>
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
    <h1 class="text-3xl font-bold tracking-tight text-white">Create User</h1>
  </div>
  <p class="text-gray-400 pl-14">
    Add a new member manually and configure their account access privileges.
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

          <!-- Password -->
          <div>
            <label
              for="password"
              class="block text-sm font-medium text-gray-300 mb-1"
            >
              Password
            </label>
            <div class="flex gap-2">
              <input
                type="text"
                id="password"
                bind:value={password}
                required
                class="flex-1 bg-white/5 border border-white/10 rounded-xl py-2 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors focus:ring-1 focus:ring-anirank-primary/50 font-mono"
              />
              <button
                type="button"
                onclick={generatePassword}
                class="px-3 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-xl transition-colors flex items-center justify-center shrink-0"
                title="Generate Random Password"
              >
                <span class="material-symbols-outlined text-[20px]">casino</span>
              </button>
              <button
                type="button"
                onclick={copyPassword}
                class="px-3 bg-white/5 hover:bg-white/10 border border-white/10 text-emerald-400 rounded-xl transition-colors flex items-center justify-center shrink-0"
                title="Copy Password"
              >
                <span class="material-symbols-outlined text-[20px]">content_copy</span>
              </button>
            </div>
            <p class="text-xs text-gray-500 mt-2">
              Save this password securely. It will be required for the user's
              first login.
            </p>
          </div>
        </div>
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
        Saving...
      {:else}
        <span class="material-symbols-outlined align-middle shrink-0">add</span>
        Create User
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
