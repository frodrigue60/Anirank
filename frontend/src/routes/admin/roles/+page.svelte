<script lang="ts">
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { onMount } from "svelte";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import User from "lucide-svelte/icons/user";
  import Users from "lucide-svelte/icons/users";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import ShieldCheck from "lucide-svelte/icons/shield-check";
  import Award from "lucide-svelte/icons/award";
  import Megaphone from "lucide-svelte/icons/megaphone";
  import Trophy from "lucide-svelte/icons/trophy";
  import Settings from "lucide-svelte/icons/settings";
  import Save from "lucide-svelte/icons/save";
  import Check from "lucide-svelte/icons/check";

  let { data } = $props();

  // Roles state initialization
  // svelte-ignore state_referenced_locally
  let roles = $state(data.roles || []);
  // svelte-ignore state_referenced_locally
  let allPermissions = $state(data.allPermissions || []);
  let selectedRoleIndex = $state(0);
  let selectedRole = $derived(roles[selectedRoleIndex] || null);

  // Buffer for currently selected permissions for the active role
  let selectedPermissionIds = $state<Set<number>>(new Set());
  let isSaving = $state(false);

  // Group permissions by resource prefix
  let groupedPermissions = $derived(() => {
    const groups: Record<string, typeof allPermissions> = {};
    allPermissions.forEach((p: any) => {
      const parts = p.slug.split(".");
      let group = "system";
      
      // Handle granular taxonomy grouping: taxonomy.genres.create -> taxonomy: genres
      if (parts[0] === "taxonomy" && parts.length >= 2) {
        group = `taxonomy: ${parts[1]}`;
      } else if (parts.length > 1) {
        group = parts[0];
      }

      if (!groups[group]) groups[group] = [];
      groups[group].push(p);
    });
    return groups;
  });

  // Sync selected permissions set when the selected role changes
  $effect(() => {
    if (selectedRole && selectedRole.permissions) {
      selectedPermissionIds = new Set(selectedRole.permissions.map((p: any) => p.id));
    } else {
      selectedPermissionIds = new Set();
    }
  });

  function togglePermission(id: number) {
    if (selectedRole?.slug === 'owner') return; // Owner bypass logic

    if (selectedPermissionIds.has(id)) {
      selectedPermissionIds.delete(id);
    } else {
      selectedPermissionIds.add(id);
    }
    // Trigger reactivity
    selectedPermissionIds = new Set(selectedPermissionIds);
  }

  async function handleSave() {
    if (!selectedRole || isSaving) return;

    isSaving = true;
    try {
      const resp = await api.post(`/admin/roles/${selectedRole.id}/permissions`, {
        permission_ids: Array.from(selectedPermissionIds),
      });

      if (resp.data.success) {
        toastState.addToast("Permissions updated successfully!", "success");
        // Update local role object to reflect new permissions (important!)
        const updatedPermissions = allPermissions.filter((p: any) =>
          selectedPermissionIds.has(p.id)
        );
        roles[selectedRoleIndex].permissions = updatedPermissions;
      }
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        err.response?.data?.message || "Failed to update permissions",
        "error"
      );
    } finally {
      isSaving = false;
    }
  }

  const getResourceIcon = (group: string) => {
    if (group.startsWith("taxonomy")) return LayoutGrid;
    switch (group) {
      case "anime": return Film;
      case "song": return Music;
      case "artist": return User;
      case "users": return Users;
      case "reports": return AlertTriangle;
      case "permissions": return ShieldCheck;
      case "badge": return Award;
      case "announcement": return Megaphone;
      case "tournament": return Trophy;
      default: return Settings;
    }
  };

  const formatGroupName = (group: string) => {
    if (group.startsWith("taxonomy: ")) {
       const entity = group.replace("taxonomy: ", "");
       return `Taxonomy: ${entity.charAt(0).toUpperCase() + entity.slice(1)}`;
    }
    return group.charAt(0).toUpperCase() + group.slice(1);
  };

  const getRoleColor = (slug: string) => {
    switch (slug) {
      case "owner": return "rose";
      case "admin": return "orange";
      case "editor": return "blue";
      case "creator": return "emerald";
      default: return "gray";
    }
  };
</script>

<svelte:head>
  <title>Roles & Permissions | Admin</title>
</svelte:head>

<div class="mb-8 overflow-hidden">
  <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-2">Roles & Permissions</h1>
  <p class="text-on-surface-variant/70">Configure what each account can do within the system.</p>
</div>

<div class="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
  <!-- ROLES LIST (Maestro) -->
  <div class="lg:col-span-1 space-y-3">
    <h2 class="text-xs font-semibold text-on-surface-variant/40 uppercase tracking-wider px-2 mb-2">System Roles</h2>
    {#each roles as role, i}
      <button
        onclick={() => (selectedRoleIndex = i)}
        class="w-full flex items-center gap-3 p-4 rounded-2xl border transition-all text-left {selectedRoleIndex === i 
          ? 'bg-surface-highest border-outline-variant ring-1 ring-white/10' 
          : 'bg-surface-container border-outline-variant hover:bg-surface-highest grayscale hover:grayscale-0'}"
      >
        <div class="w-10 h-10 rounded-xl bg-{getRoleColor(role.slug)}-500/20 text-{getRoleColor(role.slug)}-400 flex items-center justify-center shrink-0">
          <span class="font-bold text-lg">{role.name?.[0].toUpperCase()}</span>
        </div>
        <div class="overflow-hidden">
          <div class="font-bold text-on-surface text-sm truncate">{role.name}</div>
          <div class="text-xs text-on-surface-variant/70 truncate capitalize">{role.slug}</div>
        </div>
      </button>
    {/each}
  </div>

  <!-- PERMISSIONS MATRIX (Detalle) -->
  <div class="lg:col-span-3 space-y-6">
    {#if selectedRole}
      <div class="bg-surface-container border border-outline-variant rounded-3xl overflow-hidden">
        <!-- Role Info Header -->
        <div class="p-6 border-b border-outline-variant flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white/[0.02]">
          <div>
            <div class="flex items-center gap-2 mb-1">
              <h2 class="text-2xl font-bold text-on-surface capitalize">{selectedRole.name}</h2>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-{getRoleColor(selectedRole.slug)}-500/20 text-{getRoleColor(selectedRole.slug)}-400 border border-{getRoleColor(selectedRole.slug)}-500/30">
                {selectedRole.slug}
              </span>
            </div>
            <p class="text-sm text-on-surface-variant/70">{selectedRole.description || 'No description provided.'}</p>
          </div>

          {#if selectedRole.slug !== 'owner'}
            <button
              onclick={handleSave}
              disabled={isSaving}
              class="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-on-surface font-bold rounded-xl transition-all shadow-lg shadow-indigo-500/20 flex items-center gap-2 text-sm"
            >
              {#if isSaving}
                <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                Saving...
              {:else}
                <Save size={16} />
                Save Changes
              {/if}
            </button>
          {/if}
        </div>

        {#if selectedRole.slug === 'owner'}
          <div class="p-8 text-center bg-rose-500/5 border-b border-outline-variant">
            <div class="inline-flex p-3 rounded-full bg-rose-500/20 text-rose-400 mb-4 font-bold">Total Access (Locked)</div>
            <p class="text-sm text-on-surface-variant/70 max-w-lg mx-auto">
              The <strong>Owner</strong> role has a hardcoded bypass in the backend for security and redundancy.
              Permissions for this role cannot be restricted through the UI.
            </p>
          </div>
        {/if}

        <!-- Permissions Grid -->
        <div class="p-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            {#each Object.entries(groupedPermissions()) as [group, perms]}
              {@const GroupIcon = getResourceIcon(group)}
              <div class="space-y-4">
                <div class="flex items-center gap-2 pb-2 border-b border-outline-variant">
                  <div class="p-1 px-1.5 rounded-md bg-indigo-500/20 text-indigo-400">
                    <GroupIcon size={16} />
                  </div>
                  <span class="text-indigo-400 text-sm font-bold tracking-tight">{formatGroupName(group)}</span>
                </div>
                <div class="space-y-2">
                  {#each perms as perm}
                    <button
                      onclick={() => togglePermission(perm.id)}
                      disabled={selectedRole.slug === 'owner'}
                      class="w-full flex items-center justify-between p-3 rounded-xl border transition-all group {selectedPermissionIds.has(perm.id) 
                        ? 'bg-indigo-500/10 border-indigo-500/30' 
                        : 'bg-white/[0.02] border-outline-variant hover:border-outline-variant'}"
                    >
                      <div class="flex flex-col items-start gap-0.5 text-left">
                        <span class="text-sm font-semibold {selectedPermissionIds.has(perm.id) ? 'text-on-surface' : 'text-on-surface-variant'}">{perm.name}</span>
                        <span class="text-[10px] text-on-surface-variant/40 font-mono tracking-tighter uppercase">{perm.slug}</span>
                      </div>
                      
                      <div class="w-6 h-6 rounded-lg flex items-center justify-center transition-all {selectedPermissionIds.has(perm.id)
                        ? 'bg-indigo-500 text-on-surface' 
                        : 'bg-surface-highest text-gray-600 group-hover:bg-surface-highest'}">
                        {#if selectedPermissionIds.has(perm.id)}
                          <Check size={16} />
                        {:else}
                          <div class="w-2 h-2 rounded-full bg-current opacity-20"></div>
                        {/if}
                      </div>
                    </button>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {:else}
      <div class="h-64 flex items-center justify-center text-on-surface-variant/40 animate-pulse bg-surface-container rounded-3xl border border-outline-variant">
        Select a role to manage its permissions
      </div>
    {/if}
  </div>
</div>

<style>
  /* Optional transition or focus styles if needed */
</style>
