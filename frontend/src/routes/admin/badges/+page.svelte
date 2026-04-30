<script lang="ts">
  import api from "$lib/api";
  import BadgeModal from "./BadgeModal.svelte";

  let { data } = $props();
  // svelte-ignore state_referenced_locally
  let badges = $state(data.badges || []);

  $effect(() => {
    badges = data.badges || [];
  });

  let loadingDelete = $state<number | null>(null);

  // Modal State
  let showModal = $state(false);
  let editingBadge = $state<any>(null);

  function openCreateModal() {
    editingBadge = null;
    showModal = true;
  }

  function openEditModal(badge: any) {
    editingBadge = badge;
    showModal = true;
  }

  async function deleteBadge(id: number) {
    if (!confirm("Are you sure you want to delete this badge?")) return;

    loadingDelete = id;
    try {
      await api.delete(`/admin/badges/${id}`);
      badges = badges.filter((b: any) => b.id !== id);
    } catch (err) {
      console.error("Error deleting badge:", err);
      alert("Failed to delete badge.");
    } finally {
      loadingDelete = null;
    }
  }

  function handleSave(updatedBadge: any) {
    if (editingBadge) {
      badges = badges.map((b: any) =>
        b.id === updatedBadge.id ? updatedBadge : b,
      );
    } else {
      badges = [...badges, updatedBadge];
    }
    showModal = false;
  }
</script>

<div class="space-y-6 animate-fade-in">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-on-surface">Badges</h1>
      <p class="text-on-surface-variant/70 mt-1">Manage user badges and achievements</p>
    </div>
    <button
      onclick={openCreateModal}
      class="px-4 py-2 bg-primary hover:bg-primary/90 text-on-surface font-medium rounded-xl transition-colors flex items-center gap-2"
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
          d="M12 4v16m8-8H4"
        />
      </svg>
      Add Badge
    </button>
  </div>

  <div
    class="bg-surface-container border border-outline-variant rounded-2xl overflow-hidden"
  >
    <div class="overflow-x-auto">
      <table class="w-full text-left text-sm text-on-surface-variant">
        <thead
          class="bg-surface-highest text-on-surface-variant/70 text-xs uppercase tracking-wider"
        >
          <tr>
            <th class="px-6 py-4 font-medium">Icon</th>
            <th class="px-6 py-4 font-medium">Name</th>
            <th class="px-6 py-4 font-medium hidden md:table-cell"
              >Description</th
            >
            <th class="px-6 py-4 font-medium">Automation</th>
            <th class="px-6 py-4 font-medium">Status</th>
            <th class="px-6 py-4 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/5">
          {#each badges as badge (badge.id)}
            <tr class="hover:bg-white/2 transition-colors group">
              <td class="px-6 py-4">
                <img
                  src={badge?.icon_url || "/images/placeholders/default.svg"}
                  alt={badge.name}
                  class="w-10 h-10 object-contain rounded-md bg-black/20 p-1"
                />
              </td>
              <td class="px-6 py-4 font-medium text-on-surface">
                {badge.name}
              </td>
              <td
                class="px-6 py-4 hidden md:table-cell max-w-[200px] truncate text-on-surface-variant/70"
              >
                {badge.description || "-"}
              </td>
              <td class="px-6 py-4">
                {#if badge.is_automatic}
                  <div class="flex flex-col gap-0.5">
                    <span
                      class="inline-flex items-center w-fit px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-indigo-500/10 text-indigo-400 border border-indigo-500/20"
                    >
                      Auto
                    </span>
                    <span class="text-[10px] text-on-surface-variant/40 font-medium">
                      {#if badge.requirement_type === "level"}
                        Level {badge.requirement_value}
                      {:else if badge.requirement_type === "ratings"}
                        {badge.requirement_value} Ratings
                      {:else if badge.requirement_type === "anilist"}
                        AniList Link
                      {:else if badge.requirement_type === "comments"}
                        {badge.requirement_value} Comments
                      {/if}
                    </span>
                  </div>
                {:else}
                  <span
                    class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-surface-highest text-on-surface-variant/40 border border-outline-variant"
                  >
                    Manual
                  </span>
                {/if}
              </td>
              <td class="px-6 py-4">
                <span
                  class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium {badge.is_active
                    ? 'bg-green-500/10 text-green-400'
                    : 'bg-red-500/10 text-red-400'}"
                >
                  {badge.is_active ? "Active" : "Inactive"}
                </span>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    onclick={() => openEditModal(badge)}
                    class="p-2 hover:bg-surface-highest text-on-surface-variant/70 hover:text-on-surface rounded-lg transition-colors"
                    title="Edit"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                      ></path></svg
                    >
                  </button>
                  <button
                    onclick={() => deleteBadge(badge.admin_id)}
                    disabled={loadingDelete === badge.id}
                    class="p-2 hover:bg-red-500/20 text-on-surface-variant/70 hover:text-red-400 rounded-lg transition-colors disabled:opacity-50"
                    title="Delete"
                  >
                    {#if loadingDelete === badge.id}
                      <svg
                        class="w-4 h-4 animate-spin"
                        viewBox="0 0 24 24"
                        fill="none"
                      >
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
                        ></path></svg
                      >
                    {/if}
                  </button>
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="5" class="px-6 py-12 text-center text-on-surface-variant/40">
                No badges found. Click "Add Badge" to create one.
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

{#if showModal}
  <BadgeModal
    badge={editingBadge}
    onclose={() => (showModal = false)}
    onsave={handleSave}
  />
{/if}
