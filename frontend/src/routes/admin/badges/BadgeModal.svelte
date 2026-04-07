<script lang="ts">
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";

  let { badge = null, onclose, onsave } = $props();

  let loading = $state(false);
  let error = $state("");

  // svelte-ignore state_referenced_locally
  let name = $state(badge?.name || "");
  // svelte-ignore state_referenced_locally
  let description = $state(badge?.description || "");
  // svelte-ignore state_referenced_locally
  let isActive = $state(badge ? badge.is_active : true);
  // svelte-ignore state_referenced_locally
  let isAutomatic = $state(badge?.is_automatic || false);
  // svelte-ignore state_referenced_locally
  let requirementType = $state(badge?.requirement_type || "level");
  // svelte-ignore state_referenced_locally
  let requirementValue = $state(badge?.requirement_value || 0);

  let iconFile: File | null = $state(null);
  // svelte-ignore state_referenced_locally
  let iconPreview = $state(badge?.icon_url || "");

  function handleFileChange(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      iconFile = target.files[0];
      iconPreview = URL.createObjectURL(iconFile);
    }
  }

  async function handleSubmit(event: Event) {
    event.preventDefault();
    loading = true;
    error = "";

    try {
      const formData = new FormData();
      formData.append("name", name);
      formData.append("description", description);
      formData.append("is_active", isActive ? "true" : "false");
      formData.append("is_automatic", isAutomatic ? "true" : "false");
      if (isAutomatic) {
        formData.append("requirement_type", requirementType);
        formData.append("requirement_value", requirementValue.toString());
      }

      if (iconFile) {
        formData.append("icon", iconFile);
      }

      let res;
      if (badge) {
        // Update
        res = await api.put(`/admin/badges/${badge.admin_id}`, formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        });
      } else {
        // Create
        res = await api.post("/admin/badges", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        });
      }

      onsave(res.data.data);
    } catch (err: any) {
      console.error("Error saving badge:", err);
      error = getApiErrorMessage(err, "Failed to save badge");
    } finally {
      loading = false;
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
  <!-- Backdrop -->
  <button
    type="button"
    class="absolute inset-0 bg-black/60 w-full h-full border-none cursor-default"
    onclick={onclose}
    aria-label="Close modal"
  ></button>

  <!-- Modal content -->
  <div
    class="relative w-full max-w-md bg-[#1a1a24] border border-outline-variant rounded-2xl shadow-2xl p-6"
  >
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl font-bold text-on-surface">
        {badge ? "Edit Badge" : "Create Badge"}
      </h2>
      <button
        onclick={onclose}
        title="Close"
        aria-label="Close modal"
        class="text-on-surface-variant/70 hover:text-on-surface transition-colors"
      >
        <svg
          class="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          /></svg
        >
      </button>
    </div>

    {#if error}
      <div class="mb-4 p-3 rounded-xl bg-red-500/10 text-red-500 text-sm">
        {error}
      </div>
    {/if}

    <form onsubmit={handleSubmit} class="space-y-5">
      <!-- Icon Upload -->
      <div>
        <label for="icon" class="block text-sm font-medium text-on-surface-variant mb-2"
          >Badge Icon (PNG/JPEG)</label
        >
        <div class="flex items-center gap-4">
          <div
            class="w-16 h-16 rounded-xl border border-outline-variant bg-black/20 flex items-center justify-center overflow-hidden shrink-0"
          >
            {#if iconPreview}
              <img
                src={iconPreview}
                alt="Preview"
                class="w-full h-full object-contain p-1"
              />
            {:else}
              <svg
                class="w-6 h-6 text-on-surface-variant/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                ></path></svg
              >
            {/if}
          </div>
          <div class="flex-1">
            <input
              id="icon"
              type="file"
              accept="image/png, image/jpeg, image/gif, image/webp"
              onchange={handleFileChange}
              class="block w-full text-sm text-on-surface-variant/70 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20"
            />
          </div>
        </div>
      </div>

      <!-- Name -->
      <div>
        <label
          for="name"
          class="block text-sm font-medium text-on-surface-variant mb-1.5"
        >
          Name <span class="text-red-500">*</span>
        </label>
        <input
          type="text"
          id="name"
          bind:value={name}
          required
          class="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-outline-variant text-on-surface placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
          placeholder="e.g., Early Bird"
        />
      </div>

      <!-- Description -->
      <div>
        <label
          for="description"
          class="block text-sm font-medium text-on-surface-variant mb-1.5"
        >
          Description
        </label>
        <textarea
          id="description"
          bind:value={description}
          rows="3"
          class="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-outline-variant text-on-surface placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all resize-none"
          placeholder="What is this badge for?"
        ></textarea>
      </div>

      <!-- Status & Automation -->
      <div class="grid grid-cols-2 gap-4">
        <div class="flex items-center gap-3">
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              bind:checked={isActive}
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-surface-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-gray-300 peer-checked:after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"
            ></div>
            <span class="ml-3 text-sm font-medium text-on-surface">Active</span>
          </label>
        </div>

        <div class="flex items-center gap-3">
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              bind:checked={isAutomatic}
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-surface-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-gray-300 peer-checked:after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-500"
            ></div>
            <span class="ml-3 text-sm font-medium text-on-surface">Automatic</span>
          </label>
        </div>
      </div>

      {#if isAutomatic}
        <div class="space-y-4 p-4 rounded-xl bg-indigo-500/5 border border-indigo-500/10 animate-scale-in">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="reqType" class="block text-sm font-medium text-on-surface-variant mb-1.5">
                Type
              </label>
              <select
                id="reqType"
                bind:value={requirementType}
                class="w-full px-3 py-2 rounded-lg bg-black/40 border border-outline-variant text-on-surface text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/50 transition-all [&>option]:bg-[#1a1a24]"
              >
                <option value="level">Level</option>
                <option value="ratings">Ratings</option>
                <option value="anilist">AniList</option>
                <option value="comments">Comments</option>
              </select>
            </div>
            <div>
              <label for="reqValue" class="block text-sm font-medium text-on-surface-variant mb-1.5">
                Value
              </label>
              <input
                type="number"
                id="reqValue"
                bind:value={requirementValue}
                min="0"
                disabled={requirementType === 'anilist'}
                class="w-full px-3 py-2 rounded-lg bg-black/40 border border-outline-variant text-on-surface text-sm placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 transition-all disabled:opacity-30"
              />
            </div>
          </div>
          <p class="text-[10px] text-indigo-400/60 leading-tight">
            {#if requirementType === 'level'}
              Reward users when they reach level <strong>{requirementValue}</strong>.
            {:else if requirementType === 'ratings'}
              Reward users after they submit <strong>{requirementValue}</strong> ratings.
            {:else if requirementType === 'anilist'}
              Reward users when they connect their AniList account.
            {:else if requirementType === 'comments'}
              Reward users after they post <strong>{requirementValue}</strong> comments.
            {/if}
          </p>
        </div>
      {/if}

      <!-- Actions -->
      <div
        class="flex items-center justify-end gap-3 pt-4 border-t border-outline-variant"
      >
        <button
          type="button"
          onclick={onclose}
          class="px-5 py-2 rounded-xl text-sm font-medium text-on-surface-variant/70 hover:text-on-surface hover:bg-surface-highest transition-colors"
          disabled={loading}
        >
          Cancel
        </button>
        <button
          type="submit"
          class="px-5 py-2 bg-primary hover:bg-primary/90 text-on-surface text-sm font-medium rounded-xl transition-colors disabled:opacity-50 flex items-center justify-center min-w-[100px]"
          disabled={loading || !name}
        >
          {#if loading}
            <svg class="w-5 h-5 animate-spin" viewBox="0 0 24 24" fill="none">
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
            Save Badge
          {/if}
        </button>
      </div>
    </form>
  </div>
</div>
