<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import StatusControl from "$lib/components/admin/StatusControl.svelte";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";

  let name = $state("");
  let name_jp = $state("");
  let status = $state(true);
  let anilist_id = $state("");
  let animethemes_id = $state("");
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
      await api.post("/admin/artists", {
        name,
        name_jp: name_jp || null,
        status,
        anilist_id: anilist_id ? Number(anilist_id) : null,
        animethemes_id: animethemes_id ? Number(animethemes_id) : null,
      });

      // The backend triggers avatar generation synchronously
      toastState.addToast(
        "Artist created and avatar generated successfully!",
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
      class="p-2 hover:bg-surface-highest rounded-xl text-on-surface-variant/70 transition-colors"
    >
      <ArrowLeft size={20} />
    </button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-on-surface mb-1">
        New Artist
      </h1>
      <p class="text-on-surface-variant/70">Add a new musical entity to the catalog.</p>
    </div>
  </div>

  <div
    class="bg-surface-container border border-outline-variant rounded-3xl overflow-hidden"
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
              class="block text-sm font-medium text-on-surface-variant/70 px-1"
            >
              Primary Name
            </label>
            <input
              id="name"
              type="text"
              bind:value={name}
              placeholder="e.g. LiSA"
              class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-3 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors h-14"
              required
            />
          </div>

          <div class="space-y-2">
            <label
              for="name_jp"
              class="block text-sm font-medium text-on-surface-variant/70 px-1"
            >
              Japanese Name (Optional)
            </label>
            <input
              id="name_jp"
              type="text"
              bind:value={name_jp}
              placeholder="e.g. リサ"
              class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-3 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors h-14"
            />
          </div>
          <div class="space-y-2">
            <label
              for="anilist_id"
              class="block text-sm font-medium text-on-surface-variant/70 px-1"
            >
              AniList ID (Optional)
            </label>
            <input
              id="anilist_id"
              type="text"
              bind:value={anilist_id}
              placeholder="e.g. 12345"
              class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-3 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors h-14"
            />
          </div>
          <div class="space-y-2">
            <label
              for="animethemes_id"
              class="block text-sm font-medium text-on-surface-variant/70 px-1"
            >
              AnimeThemes ID (Optional)
            </label>
            <input
              id="animethemes_id"
              type="text"
              bind:value={animethemes_id}
              placeholder="e.g. 12345"
              class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-3 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors h-14"
            />
          </div>
        </div>

        <div class="space-y-2">
          <label
            for="status"
            class="block text-sm font-medium text-on-surface-variant/70 px-1"
          >
            Status
          </label>
          <select
            id="status"
            bind:value={status}
            class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-3 px-4 text-on-surface focus:outline-none focus:border-primary/30 focus:bg-surface-highest transition-colors h-14"
          >
            <option value={true}>Active</option>
            <option value={false}>Inactive</option>
          </select>
        </div>

        <div class="pt-4 flex items-center justify-end gap-4">
          <button
            type="button"
            onclick={() => history.back()}
            class="px-6 py-3 bg-surface-highest hover:bg-surface-highest text-on-surface font-medium rounded-2xl transition-colors h-14"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            class="px-8 py-3 bg-primary hover:bg-primary-container disabled:bg-blue-600/50 disabled:cursor-not-allowed text-on-surface font-bold rounded-2xl transition-all shadow-lg shadow-anirank-primary/20 flex items-center justify-center gap-2 h-14 min-w-[160px]"
          >
            {#if isSubmitting}
              <svg
                class="animate-spin h-5 w-5 text-on-surface"
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
        AniList API (staff search) or UI-Avatars as fallback.
      </p>
    </div>
  </div>
</div>
