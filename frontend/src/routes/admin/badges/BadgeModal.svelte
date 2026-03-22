<script lang="ts">
  import api from "$lib/api";

  let { badge = null, onclose, onsave } = $props();

  let loading = $state(false);
  let error = $state("");

  // svelte-ignore state_referenced_locally
  let name = $state(badge?.name || "");
  // svelte-ignore state_referenced_locally
  let description = $state(badge?.description || "");
  // svelte-ignore state_referenced_locally
  let isActive = $state(badge ? badge.is_active : true);

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

      if (iconFile) {
        formData.append("icon", iconFile);
      }

      let res;
      if (badge) {
        // Update
        res = await api.put(`/admin/badges/${badge.id}`, formData, {
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
      error =
        err.response?.data?.error || err.message || "Failed to save badge";
    } finally {
      loading = false;
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
  <!-- Backdrop -->
  <button
    type="button"
    class="absolute inset-0 bg-black/60 backdrop-blur-sm w-full h-full border-none cursor-default"
    onclick={onclose}
    aria-label="Close modal"
  ></button>

  <!-- Modal content -->
  <div
    class="relative w-full max-w-md bg-[#1a1a24] border border-white/10 rounded-2xl shadow-2xl p-6"
  >
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl font-bold text-white">
        {badge ? "Edit Badge" : "Create Badge"}
      </h2>
      <button
        onclick={onclose}
        title="Close"
        aria-label="Close modal"
        class="text-gray-400 hover:text-white transition-colors"
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
        <label for="icon" class="block text-sm font-medium text-gray-300 mb-2"
          >Badge Icon (PNG/JPEG)</label
        >
        <div class="flex items-center gap-4">
          <div
            class="w-16 h-16 rounded-xl border border-white/10 bg-black/20 flex items-center justify-center overflow-hidden flex-shrink-0"
          >
            {#if iconPreview}
              <img
                src={iconPreview}
                alt="Preview"
                class="w-full h-full object-contain p-1"
              />
            {:else}
              <svg
                class="w-6 h-6 text-gray-500"
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
              class="block w-full text-sm text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-anirank-primary/10 file:text-anirank-primary hover:file:bg-anirank-primary/20"
            />
          </div>
        </div>
      </div>

      <!-- Name -->
      <div>
        <label
          for="name"
          class="block text-sm font-medium text-gray-300 mb-1.5"
        >
          Name <span class="text-red-500">*</span>
        </label>
        <input
          type="text"
          id="name"
          bind:value={name}
          required
          class="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-anirank-primary/50 transition-all"
          placeholder="e.g., Early Bird"
        />
      </div>

      <!-- Description -->
      <div>
        <label
          for="description"
          class="block text-sm font-medium text-gray-300 mb-1.5"
        >
          Description
        </label>
        <textarea
          id="description"
          bind:value={description}
          rows="3"
          class="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-anirank-primary/50 transition-all resize-none"
          placeholder="What is this badge for?"
        ></textarea>
      </div>

      <!-- Status Toggle -->
      <div class="flex items-center gap-3">
        <label class="relative inline-flex items-center cursor-pointer">
          <input type="checkbox" bind:checked={isActive} class="sr-only peer" />
          <div
            class="w-11 h-6 bg-white/10 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-gray-300 peer-checked:after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-anirank-primary"
          ></div>
          <span class="ml-3 text-sm font-medium text-white">Active</span>
        </label>
      </div>

      <!-- Actions -->
      <div
        class="flex items-center justify-end gap-3 pt-4 border-t border-white/10"
      >
        <button
          type="button"
          onclick={onclose}
          class="px-5 py-2 rounded-xl text-sm font-medium text-gray-400 hover:text-white hover:bg-white/5 transition-colors"
          disabled={loading}
        >
          Cancel
        </button>
        <button
          type="submit"
          class="px-5 py-2 bg-anirank-primary hover:bg-anirank-primary/90 text-white text-sm font-medium rounded-xl transition-colors disabled:opacity-50 flex items-center justify-center min-w-[100px]"
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
