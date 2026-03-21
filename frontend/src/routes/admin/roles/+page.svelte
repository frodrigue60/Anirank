<script lang="ts">
  let { data } = $props();
  let roles = $state(data.roles);
</script>

<svelte:head>
  <title>Roles & Permissions | Admin</title>
</svelte:head>

<div
  class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
>
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
      Roles & Permissions
    </h1>
    <p class="text-gray-400">
      Manage system roles and their assigned capabilities.
    </p>
  </div>

  <button
    class="px-4 py-2 bg-white/5 hover:bg-white/10 text-white font-medium rounded-xl transition-colors border border-white/10 flex items-center gap-2"
    disabled
  >
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M12 4v16m8-8H4"
      /></svg
    >
    New Role (Disabled)
  </button>
</div>

<div
  class="bg-orange-500/10 border border-orange-500/20 text-orange-400 p-4 rounded-xl mb-6 flex gap-3"
>
  <svg
    class="w-6 h-6 shrink-0"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
    ><path
      stroke-linecap="round"
      stroke-linejoin="round"
      stroke-width="2"
      d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
    /></svg
  >
  <p class="text-sm">
    <strong>Fixed Architecture:</strong> The current system uses a hardcoded role
    hierarchy (`admin`, `editor`, `creator`, `user`). Creating custom roles with granular
    permissions is not supported in the current backend iteration.
  </p>
</div>

<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
  {#each roles as role}
    <div
      class="bg-anirank-card border border-white/5 rounded-2xl p-6 flex flex-col"
    >
      <div class="flex items-start justify-between mb-4">
        <div class="flex items-center gap-3">
          <div
            class="w-12 h-12 rounded-xl bg-{role.color}-500/20 text-{role.color}-400 flex items-center justify-center border border-{role.color}-500/30"
          >
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              {#if role.name === "admin"}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                />
              {:else if role.name === "editor"}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              {:else if role.name === "creator"}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                />
              {:else}
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              {/if}
            </svg>
          </div>
          <div>
            <h2 class="text-xl font-bold text-white capitalize">{role.name}</h2>
            <p class="text-sm text-gray-500">
              {role.users_count.toLocaleString()} accounts assigned
            </p>
          </div>
        </div>
        <button
          class="px-3 py-1.5 bg-white/5 hover:bg-white/10 text-white text-xs font-semibold rounded-lg transition-colors border border-white/10"
        >
          View Users
        </button>
      </div>

      <p class="text-sm text-gray-400 mb-6 flex-1">
        {role.description}
      </p>

      <div class="pt-4 border-t border-white/5 mt-auto">
        <h3
          class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3"
        >
          Key Capabilities
        </h3>
        <ul class="space-y-2 text-sm text-gray-300">
          {#if role.name === "admin"}
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Full Backend Access
            </li>
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Role Management
            </li>
          {:else if role.name === "editor"}
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Direct Catalog Publish
            </li>
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Resolve Reports
            </li>
          {:else if role.name === "creator"}
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Bypass Upload Limits
            </li>
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-amber-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                /></svg
              > Catalog submissions need review
            </li>
          {:else}
            <li class="flex items-center gap-2">
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                /></svg
              > Voting & Playlists
            </li>
            <li class="flex items-center gap-2 text-rose-400">
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                /></svg
              > No Admin Panel Access
            </li>
          {/if}
        </ul>
      </div>
    </div>
  {/each}
</div>
