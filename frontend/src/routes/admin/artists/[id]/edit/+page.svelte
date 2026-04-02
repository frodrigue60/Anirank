<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  let { data } = $props();

  // svelte-ignore state_referenced_locally
  let artist = $state(
    data.artist || { name: "", name_jp: "", slug: "", avatar_url: "" },
  );
  let isSubmitting = $state(false);
  let isGenerating = $state(false);
  let avatarFile = $state<File | null>(null);
  // svelte-ignore state_referenced_locally
  let previewUrl = $state(artist.avatar_url);

  async function handleFileChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      avatarFile = target.files[0];
      previewUrl = URL.createObjectURL(avatarFile);
    }
  }

  async function generateAvatar() {
    if (!artist.id) return;
    isGenerating = true;
    try {
      await api.post(`/admin/artists/${artist.id}/avatar/generate`);
      // Reload artist data to get new URL
      const res = await api.get(`/admin/artists/${artist.id}`);
      artist = res.data.data;
      previewUrl = artist.avatar_url;
      toastState.addToast("Avatar generated successfully!", "success");
    } catch (error: any) {
      toastState.addToast(
        "Failed to generate avatar: " +
          (error.response?.data?.message || error.message),
        "error",
      );
    } finally {
      isGenerating = false;
    }
  }

  async function getAvatarFromAniList() {
    if (!artist.id) return;
    isGenerating = true;
    try {
      const res = await api.post(`/admin/artists/${artist.id}/sync-avatar`);
      artist = res.data.data;
      previewUrl = artist.avatar_url;
      toastState.addToast("Avatar synced successfully!", "success");
    } catch (error: any) {
      toastState.addToast(
        "Failed to sync avatar: " +
          (error.response?.data?.message || error.message),
        "error",
      );
    } finally {
      isGenerating = false;
    }
  }

  async function handleSubmit() {
    artist.name = artist.name.trim().replace(/\s+/g, " ");
    if (artist.name_jp) {
      artist.name_jp = artist.name_jp.trim().replace(/\s+/g, " ");
    }

    if (!artist.name) {
      toastState.addToast("Name is required", "error");
      return;
    }

    isSubmitting = true;
    try {
      if (avatarFile) {
        // Use Multipart for file upload + data
        const formData = new FormData();
        formData.append("name", artist.name);
        formData.append("name_jp", artist.name_jp || "");
        formData.append("slug", artist.slug);
        formData.append("status", String(artist.status));
        if (artist.anilist_id)
          formData.append("anilist_id", String(artist.anilist_id));
        if (artist.animethemes_id)
          formData.append("animethemes_id", String(artist.animethemes_id));
        formData.append("avatar", avatarFile);

        await api.put(`/admin/artists/${artist.id}`, formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
      } else {
        // Regular JSON update
        await api.put(`/admin/artists/${artist.id}`, {
          name: artist.name,
          name_jp: artist.name_jp || null,
          slug: artist.slug,
          status: artist.status,
          anilist_id: artist.anilist_id ? Number(artist.anilist_id) : null,
          animethemes_id: artist.animethemes_id
            ? Number(artist.animethemes_id)
            : null,
        });
      }

      toastState.addToast("Artist updated successfully!", "success");
      goto("/admin/artists");
    } catch (error: any) {
      console.error("Error updating artist:", error);
      toastState.addToast(
        error.response?.data?.message || "Failed to update artist",
        "error",
      );
    } finally {
      isSubmitting = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Artist | Admin</title>
</svelte:head>

<div class="max-w-5xl mx-auto pb-20">
  <div class="mb-8 flex items-center gap-4">
    <button
      onclick={() => history.back()}
      class="p-2 hover:bg-white/5 rounded-xl text-gray-400 transition-colors"
      aria-label="Back"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 19l-7-7m0 0l7-7m-7 7h18"
        /></svg
      >
    </button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-white mb-1">
        Edit Artist
      </h1>
      <p class="text-gray-400">Modify artist details and visual identity.</p>
    </div>
  </div>

  {#if !artist.id}
    <div
      class="bg-anirank-card border border-white/5 rounded-3xl p-12 text-center"
    >
      <p class="text-gray-400">Artist not found or error loading data.</p>
      <button
        onclick={() => history.back()}
        class="mt-4 text-anirank-primary hover:underline">Go back</button
      >
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left: Form -->
      <div class="lg:col-span-2 space-y-8">
        <div
          class="bg-anirank-card border border-white/5 rounded-3xl overflow-hidden p-8"
        >
          <h2 class="text-xl font-semibold text-white mb-6">
            General Information
          </h2>
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
                  bind:value={artist.name}
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
                  bind:value={artist.name_jp}
                  class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
                />
              </div>
              <div class="space-y-2">
                <label
                  for="anilist_id"
                  class="block text-sm font-medium text-gray-400 px-1"
                >
                  AniList ID (Optional)
                </label>
                <input
                  id="anilist_id"
                  type="text"
                  bind:value={artist.anilist_id}
                  class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
                />
              </div>
              <div class="space-y-2">
                <label
                  for="animethemes_id"
                  class="block text-sm font-medium text-gray-400 px-1"
                >
                  AnimeThemes ID (Optional)
                </label>
                <input
                  id="animethemes_id"
                  type="text"
                  bind:value={artist.animethemes_id}
                  class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
                />
              </div>
              <div class="space-y-2">
                <label
                  for="status"
                  class="block text-sm font-medium text-gray-400 px-1"
                >
                  Status
                </label>
                <select
                  id="status"
                  bind:value={artist.status}
                  class="w-full bg-white/5 border border-white/10 rounded-2xl py-3 px-4 text-white focus:outline-none focus:border-anirank-primary transition-colors h-14"
                >
                  <option value={true}>Active</option>
                  <option value={false}>Inactive</option>
                </select>
              </div>
            </div>

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
                class="px-8 py-3 bg-anirank-primary hover:bg-blue-600 disabled:bg-blue-600/50 disabled:cursor-not-allowed text-white font-bold rounded-2xl transition-all shadow-lg shadow-anirank-primary/20 flex items-center justify-center gap-2 h-14 min-w-[180px]"
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
                  Updating...
                {:else}
                  Save Changes
                {/if}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Right: Avatar Management -->
      <div class="space-y-8">
        <div class="bg-anirank-card border border-white/5 rounded-3xl p-8">
          <h2 class="text-xl font-semibold text-white mb-6">Artist Avatar</h2>

          <div class="flex flex-col items-center">
            <div class="relative group mb-6">
              <div
                class="w-48 h-48 rounded-full overflow-hidden border-4 border-white/5 bg-white/5 relative"
              >
                {#if previewUrl}
                  <img
                    src={previewUrl}
                    alt="Avatar Preview"
                    class="w-full h-full object-cover"
                  />
                {:else}
                  <div
                    class="w-full h-full flex items-center justify-center text-gray-600"
                  >
                    <svg
                      class="w-16 h-16"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                      />
                    </svg>
                  </div>
                {/if}

                {#if isGenerating}
                  <div
                    class="absolute inset-0 bg-black/60 flex items-center justify-center"
                  >
                    <svg
                      class="animate-spin h-8 w-8 text-anirank-primary"
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
                  </div>
                {/if}
              </div>

              <label
                for="avatar-input"
                class="absolute bottom-2 right-2 p-3 bg-anirank-primary text-white rounded-full shadow-lg hover:bg-blue-600 cursor-pointer transition-colors border-4 border-[#0a0a0b]"
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
                    d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                </svg>
                <input
                  id="avatar-input"
                  type="file"
                  accept="image/*"
                  onchange={handleFileChange}
                  class="hidden"
                />
              </label>
            </div>

            <div class="w-full space-y-3">
              <!-- generate avatares for all artists-->
              <button
                onclick={generateAvatar}
                disabled={isGenerating || isSubmitting}
                class="w-full py-3 bg-white/5 hover:bg-white/10 text-white rounded-2xl border border-white/10 transition-colors flex items-center justify-center gap-2"
              >
                Generate with Magic
              </button>
              <button
                onclick={getAvatarFromAniList}
                disabled={isGenerating || isSubmitting}
                class="w-full py-3 bg-white/5 hover:bg-white/10 text-white rounded-2xl border border-white/10 transition-colors flex items-center justify-center gap-2"
              >
                Avatar from AT/AL
              </button>
              <p class="text-gray-500 text-xs text-center">
                Tries to fetch the official avatar from AniList. If not found,
                it keeps the current one or uses a placeholder.
              </p>
            </div>
          </div>
        </div>

        <div class="bg-rose-500/10 border border-rose-500/20 rounded-3xl p-6">
          <h3 class="text-rose-400 font-semibold mb-2">Danger Zone</h3>
          <p class="text-rose-300/60 text-sm mb-4">
            Deleting this artist is permanent and will remove all song
            associations.
          </p>
          <button
            class="w-full py-3 bg-rose-500/20 hover:bg-rose-500/40 text-rose-400 rounded-2xl border border-rose-500/20 transition-colors"
          >
            Delete Artist
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
