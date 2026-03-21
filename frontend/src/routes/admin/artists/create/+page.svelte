<script lang="ts">
  import { goto } from "$app/navigation";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";

  let name = $state("");
  let name_jp = $state("");
  let status = $state(true);
  let isSubmitting = $state(false);

  async function handleSubmit() {
    name = name.trim().replace(/\s+/g, " ");
    if (name_jp) {
      name_jp = name_jp.trim().replace(/\s+/g, " ");
    }

    if (!name) {
      toastState.addToast("Name is required", "error");
      return;
    }

    isSubmitting = true;
    try {
      const res = await api.post("/admin/artists", {
        name,
        name_jp: name_jp || null,
        status,
      });

      // The backend triggers avatar generation in the background
      toastState.addToast(
        "Artist created successfully! Redirecting...",
        "success",
      );
      goto("/admin/artists");
    } catch (error: any) {
      console.error("Error creating artist:", error);
      toastState.addToast(
        error.response?.data?.message || "Failed to create artist",
        "error",
      );
    } finally {
      isSubmitting = false;
    }
  }
</script>

<svelte:head>
  <title>New Artist | Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
  <div class="mb-8 flex items-center gap-4">
    <button
      onclick={() => history.back()}
      class="p-2 hover:bg-white/5 rounded-xl text-gray-400 transition-colors"
    >
      <span class="material-symbols-outlined">arrow_back</span>
    </button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
        New Artist
      </h1>
      <p class="text-gray-400">Add a new musical entity to the catalog.</p>
    </div>
  </div>

  <div
    class="bg-anirank-card border border-white/5 rounded-3xl overflow-hidden"
  >
    <div class="p-8">
      <form
        onsubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
        class="space-y-6"
      >
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label
              for="name"
              class="block text-sm font-medium text-gray-400 px-1"
            >
              Primary Name
            </label>
            <input
              id="name"
              type="text"
              bind:value={name}
              placeholder="e.g. LiSA"
              class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
              required
            />
          </div>

          <div class="space-y-2">
            <label
              for="name_jp"
              class="block text-sm font-medium text-gray-400 px-1"
            >
              Japanese Name (Optional)
            </label>
            <input
              id="name_jp"
              type="text"
              bind:value={name_jp}
              placeholder="e.g. リサ"
              class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
            />
          </div>
        </div>

        <StatusControl bind:status={status} />

        <div class="pt-4 flex items-center justify-end gap-4">
          <button
            type="button"
            onclick={() => history.back()}
            class="px-6 py-3 bg-white/5 hover:bg-white/10 text-white font-medium rounded-2xl transition-colors h-14"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            class="px-8 py-3 bg-anirank-primary hover:bg-blue-600 disabled:bg-blue-600/50 disabled:cursor-not-allowed text-white font-bold rounded-2xl transition-all shadow-lg shadow-anirank-primary/20 flex items-center justify-center gap-2 h-14 min-w-[160px]"
          >
            {#if isSubmitting}
              <svg
                class="animate-spin h-5 w-5 text-white"
                fill="none"
                viewBox="0 0 24 24"
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
              Creating...
            {:else}
              Create Artist
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>

  <div
    class="mt-8 bg-blue-500/10 border border-blue-500/20 rounded-2xl p-4 flex gap-4"
  >
    <div class="p-2 bg-blue-500/20 rounded-xl h-fit">
      <svg
        class="w-5 h-5 text-blue-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    </div>
    <div>
      <h4 class="text-blue-200 font-medium">Automatic Avatar Generation</h4>
      <p class="text-blue-300/70 text-sm mt-1">
        Once created, an avatar will be automatically generated using the
        DiceBear API and stored in S3. You can change it later in the edit view.
      </p>
    </div>
  </div>
</div>
