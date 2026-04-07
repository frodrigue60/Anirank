<script lang="ts">
  import { goto } from "$app/navigation";
  import { configState as config } from "$lib/state/config.svelte";
  import TagsInput from "$lib/components/admin/TagsInput.svelte";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { getApiErrorMessage } from "$lib/api-errors";
  import type { PageData } from "./$types";
  import { authState } from "$lib/state/auth.svelte";
  import { Loader2 } from "lucide-svelte";

  let { data } = $props<{ data: PageData }>();
  // svelte-ignore state_referenced_locally
  const anime = data.anime;

  // Form State
  let title = $state(anime.title || "");
  let description = $state(anime.description || "");
  let status = $state(!!anime.status);
  let year_id = $state(anime.year_id || 0);
  let season_id = $state(anime.season_id || 0);
  let format_id = $state(anime.format_id || 0);
  let anilist_id = $state<number | null>(anime.anilist_id || null);

  // Tags State
  let studiosString = $state(
    anime.studios ? anime.studios.map((s: any) => s.name).join(", ") : "",
  );
  let producersString = $state(
    anime.producers ? anime.producers.map((s: any) => s.name).join(", ") : "",
  );
  let genresString = $state(
    anime.genres ? anime.genres.map((s: any) => s.name).join(", ") : "",
  );

  // Files
  let coverFile: File | null = $state(null);
  let bannerFile: File | null = $state(null);

  // Existing images
  let coverPreview = $state(anime.cover_url || "");
  let bannerPreview = $state(anime.banner_url || "");

  // Status Permissions
  const canEditStatus = $derived.by(() => {
    if (!authState.user || !authState.user.roles) return false;
    return (authState.user.roles as any[]).some(
      (r) =>
        ["admin", "editor"].includes(r.name?.toLowerCase()) ||
        ["admin", "editor"].includes(r.slug?.toLowerCase()),
    );
  });

  // If the user can't edit status, ensure it remains in its current state or draft if new logic applies
  // In edit mode, we don't automatically force to false unless we want to strip permissions.
  // We'll just disable the input below.

  // UI State
  let loading = $state(false);

  // Handlers for file inputs
  function onCoverChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      coverFile = target.files[0];
      coverPreview = URL.createObjectURL(coverFile);
    }
  }

  function onBannerChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      bannerFile = target.files[0];
      bannerPreview = URL.createObjectURL(bannerFile);
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    loading = true;

    try {
      const formData = new FormData();
      formData.append("title", title);
      formData.append("description", description);
      formData.append("status", status ? "1" : "0");
      if (year_id > 0) formData.append("year_id", year_id.toString());
      if (season_id > 0) formData.append("season_id", season_id.toString());
      if (format_id > 0) formData.append("format_id", format_id.toString());
      if (anilist_id) formData.append("anilist_id", anilist_id.toString());

      // Relations
      formData.append("studios", studiosString);
      formData.append("producers", producersString);
      formData.append("genres", genresString);

      if (coverFile) {
        formData.append("cover", coverFile);
      }
      if (bannerFile) {
        formData.append("banner", bannerFile);
      }

      const res = await api.put(`/admin/animes/${anime.id}`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (res.data.success || res.status === 200) {
        toastState.addToast(
          res.data.message || "Anime updated successfully",
          "success",
        );
        goto(`/admin/animes/${anime.id}`);
      } else {
        toastState.addToast(
          res.data.message || "Failed to update anime",
          "error",
        );
      }
    } catch (err: any) {
      console.error(err);
      toastState.addToast(
        getApiErrorMessage(err, "An error occurred while updating the anime"),
        "error",
      );
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Anime | Admin</title>
</svelte:head>

<div class="mb-10">
  <div class="flex items-center gap-4 mb-3">
    <a
      href="/admin/animes"
      aria-label="Back to Animes"
      class="text-on-surface-variant hover:text-primary transition-all p-2 -ml-2 rounded-xl hover:bg-primary-container group"
    >
      <span
        class="material-symbols-outlined transition-transform group-hover:-translate-x-1"
        >arrow_back</span
      >
    </a>
    <h1 class="text-3xl font-black tracking-tighter text-on-surface uppercase">
      Edit Anime
    </h1>
  </div>
  <p class="text-on-surface-variant/60 font-medium ml-10">
    Update general details, taxonomies and media assets.
  </p>
</div>

<form onsubmit={handleSubmit} class="space-y-8 max-w-5xl">
  <!-- General Info -->
  <div
    class="bg-surface-container border border-outline-variant rounded-4xl p-8 shadow-sm"
  >
    <h2
      class="text-xl font-black text-on-surface mb-8 flex items-center gap-3 uppercase tracking-tight"
    >
      <div
        class="bg-primary/10 w-10 h-10 rounded-2xl flex items-center justify-center text-primary"
      >
        <span class="material-symbols-outlined text-[20px]">info</span>
      </div>
      General Information
    </h2>

    <div class="space-y-6">
      <div class="group">
        <label
          for="title"
          class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
          >Title <span class="text-red-500/60">*</span></label
        >
        <input
          type="text"
          id="title"
          bind:value={title}
          required
          class="bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface placeholder-on-surface-variant font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
          placeholder="e.g. Shingeki no Kyojin"
        />
      </div>

      <div class="group">
        <label
          for="description"
          class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
          >Description</label
        >
        <textarea
          id="description"
          bind:value={description}
          rows="5"
          class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface placeholder-on-surface-variant/20 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none resize-none"
          placeholder="Synopsis or brief description..."
        ></textarea>
      </div>
    </div>
  </div>

  <!-- Taxonomies & Metadata -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
    <div
      class="bg-surface-container border border-outline-variant rounded-4xl p-8 shadow-sm"
    >
      <h2
        class="text-xl font-black text-on-surface mb-8 flex items-center gap-3 uppercase tracking-tight"
      >
        <div
          class="bg-primary/10 w-10 h-10 rounded-2xl flex items-center justify-center text-primary"
        >
          <span class="material-symbols-outlined text-[20px]">category</span>
        </div>
        Taxonomy
      </h2>

      <div class="space-y-8">
        <div>
          <label
            for="studios"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-3 ml-1"
            >Studios</label
          >
          <TagsInput
            endpoint="/admin/studios"
            bind:value={studiosString}
            placeholder="e.g. Mappa, Wit Studio"
            entityName="Studio"
          />
        </div>

        <div>
          <label
            for="producers"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-3 ml-1"
            >Producers</label
          >
          <TagsInput
            endpoint="/admin/producers"
            bind:value={producersString}
            placeholder="e.g. Aniplex, Kadokawa"
            entityName="Producer"
          />
        </div>

        <div>
          <label
            for="genres"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-3 ml-1"
            >Genres</label
          >
          <TagsInput
            endpoint="/admin/genres"
            bind:value={genresString}
            placeholder="e.g. Action, Romance"
            entityName="Genre"
          />
        </div>
      </div>

      <div class="mt-10 space-y-5 pt-8 border-t border-outline-variant">
        <div class="group">
          <label
            for="year"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
            >Release Year</label
          >
          <select
            id="year"
            bind:value={year_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none appearance-none"
          >
            <option value={0}>Select Year</option>
            {#each config.years as year}
              <option value={year.id}>{year.name}</option>
            {/each}
          </select>
        </div>

        <div class="group">
          <label
            for="season"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
            >Season</label
          >
          <select
            id="season"
            bind:value={season_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none appearance-none"
          >
            <option value={0}>Select Season</option>
            {#each config.seasons as season}
              <option value={season.id}>{season.name}</option>
            {/each}
          </select>
        </div>

        <div class="group">
          <label
            for="format"
            class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
            >Format</label
          >
          <select
            id="format"
            bind:value={format_id}
            class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none appearance-none"
          >
            <option value={0}>Select Format</option>
            {#each config.formats as format}
              <option value={format.id}>{format.name}</option>
            {/each}
          </select>
        </div>
      </div>
    </div>

    <!-- External & States -->
    <div class="space-y-8">
      <div
        class="bg-surface-container border border-outline-variant rounded-4xl p-8 shadow-sm"
      >
        <h2
          class="text-xl font-black text-on-surface mb-8 flex items-center gap-3 uppercase tracking-tight"
        >
          <div
            class="bg-primary/10 w-10 h-10 rounded-2xl flex items-center justify-center text-primary"
          >
            <span class="material-symbols-outlined text-[20px]">link</span>
          </div>
          Connections & Status
        </h2>

        <div class="space-y-8">
          <div class="group">
            <label
              for="anilist_id"
              class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-2 ml-1"
              >Anilist ID</label
            >
            <input
              type="number"
              id="anilist_id"
              title="Anilist ID"
              bind:value={anilist_id}
              class="w-full bg-surface-highest border border-outline-variant rounded-2xl py-4 px-5 text-on-surface placeholder-on-surface-variant/20 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
              placeholder="e.g. 16498"
            />
            <p
              class="text-[10px] font-bold text-on-surface-variant/30 mt-3 px-1"
            >
              Leave empty if it's a manual entry not tied to Anilist.
            </p>
          </div>

          <div class="pt-4 border-t border-outline-variant">
            <label
              class="flex items-center gap-4 {canEditStatus
                ? 'cursor-pointer'
                : 'cursor-not-allowed opacity-50'} group/status"
            >
              <div class="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  bind:checked={status}
                  disabled={!canEditStatus}
                  class="sr-only peer"
                />
                <div
                  class="w-14 h-7 bg-surface-highest border border-outline-variant rounded-full peer peer-checked:bg-primary/20 peer-checked:border-primary/30 transition-all duration-300 after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-on-surface-variant/20 after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-7 peer-checked:after:bg-primary shadow-inner"
                ></div>
              </div>
              <div class="flex flex-col">
                <span
                  class="text-[11px] font-black uppercase tracking-widest {status
                    ? 'text-primary'
                    : 'text-on-surface-variant/60'} transition-colors"
                >
                  {status
                    ? "Published / Global Visibility"
                    : "Draft / Private Entry"}
                </span>
                {#if !canEditStatus}
                  <span
                    class="text-[9px] font-bold text-amber-500/60 uppercase tracking-tight mt-0.5"
                  >
                    Requires moderator review before publishing
                  </span>
                {:else}
                  <span
                    class="text-[9px] font-bold text-on-surface-variant/20 uppercase tracking-tight mt-0.5"
                  >
                    Toggle visibility on the public catalog
                  </span>
                {/if}
              </div>
            </label>
          </div>
        </div>
      </div>

      <div
        class="bg-surface-container border border-outline-variant rounded-4xl p-8 shadow-sm overflow-hidden"
      >
        <h2
          class="text-xl font-black text-on-surface mb-8 flex items-center gap-3 uppercase tracking-tight"
        >
          <div
            class="bg-primary/10 w-10 h-10 rounded-2xl flex items-center justify-center text-primary"
          >
            <span class="material-symbols-outlined text-[20px]">image</span>
          </div>
          Quick Help
        </h2>
        <p
          class="text-[11px] font-bold uppercase tracking-wider text-on-surface-variant/40 leading-relaxed"
        >
          Ensure you have the <span class="text-primary">Anilist ID</span> ready for
          automatic metadata hydration if available. High quality banners and covers
          are recommended for the best display on the platform.
        </p>
      </div>
    </div>
  </div>

  <!-- Assets -->
  <div
    class="bg-surface-container border border-outline-variant rounded-4xl p-8 shadow-sm"
  >
    <h2
      class="text-xl font-black text-on-surface mb-8 flex items-center gap-3 uppercase tracking-tight"
    >
      <div
        class="bg-primary/10 w-10 h-10 rounded-2xl flex items-center justify-center text-primary"
      >
        <span class="material-symbols-outlined text-[20px]">media_output</span>
      </div>
      Media Assets
    </h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <!-- Thumbnail -->
      <div class="group">
        <label
          for="cover"
          class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-3 ml-1"
          >Thumbnail (Cover)</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="cover"
            class="flex flex-col items-center justify-center w-full h-48 border-2 border-outline-variant border-dashed rounded-3xl cursor-pointer bg-surface-highest hover:bg-surface-highest hover:border-primary/30 transition-all overflow-hidden relative group/asset"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-6 text-center z-10 transition-transform group-hover/asset:scale-105"
            >
              <div
                class="bg-on-surface-variant/5 w-12 h-12 rounded-2xl flex items-center justify-center text-on-surface-variant/30 mb-4"
              >
                <span class="material-symbols-outlined text-3xl"
                  >cloud_upload</span
                >
              </div>
              <p
                class="mb-1 text-[11px] font-black uppercase tracking-widest text-on-surface"
              >
                Click to upload
              </p>
              <p class="text-[10px] font-bold text-on-surface-variant/30">
                PNG, JPG up to 2MB
              </p>
            </div>
            <!-- PREVIEW -->
            {#if coverPreview}
              <div
                class="absolute inset-0 z-0 opacity-80 select-none bg-black/50"
              >
                <span
                  class="absolute inset-0 flex items-center justify-center text-on-surface bg-black/60 z-10 font-medium opacity-0 group-hover/asset:opacity-100 transition-opacity uppercase tracking-widest text-[10px]"
                  >Change Image</span
                >
                <img
                  src={coverPreview}
                  alt="Preview"
                  class="w-full h-full object-cover"
                />
              </div>
            {:else if coverFile}
              <div class="absolute inset-0 z-0 select-none">
                <div
                  class="absolute inset-0 flex items-center justify-center bg-primary/80 z-10"
                >
                  <span
                    class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface"
                    >Cover Selected</span
                  >
                </div>
                <img
                  src={URL.createObjectURL(coverFile)}
                  alt="Preview"
                  class="w-full h-full object-cover blur-[2px]"
                />
              </div>
            {/if}
            <input
              id="cover"
              type="file"
              accept="image/*"
              class="hidden"
              onchange={onCoverChange}
            />
          </label>
        </div>
      </div>

      <!-- Banner -->
      <div class="group">
        <label
          for="banner"
          class="block text-[10px] font-black uppercase tracking-[0.2em] text-on-surface-variant/40 mb-3 ml-1"
          >Banner</label
        >
        <div class="flex items-center justify-center w-full">
          <label
            for="banner"
            class="flex flex-col items-center justify-center w-full h-48 border-2 border-outline-variant border-dashed rounded-3xl cursor-pointer bg-surface-highest hover:bg-surface-highest hover:border-primary/30 transition-all overflow-hidden relative group/asset"
          >
            <div
              class="flex flex-col items-center justify-center pt-5 pb-6 px-6 text-center z-10 transition-transform group-hover/asset:scale-105"
            >
              <div
                class="bg-on-surface-variant/5 w-12 h-12 rounded-2xl flex items-center justify-center text-on-surface-variant/30 mb-4"
              >
                <span class="material-symbols-outlined text-3xl">landscape</span
                >
              </div>
              <p
                class="mb-1 text-[11px] font-black uppercase tracking-widest text-on-surface"
              >
                Click to upload
              </p>
              <p class="text-[10px] font-bold text-on-surface-variant/30">
                Wide aspect ratio
              </p>
            </div>
            <!-- PREVIEW -->
            {#if bannerPreview}
              <div
                class="absolute inset-0 z-0 opacity-80 select-none bg-black/50"
              >
                <span
                  class="absolute inset-0 flex items-center justify-center text-on-surface bg-black/60 z-10 font-medium opacity-0 group-hover/asset:opacity-100 transition-opacity uppercase tracking-widest text-[10px]"
                  >Change Image</span
                >
                <img
                  src={bannerPreview}
                  alt="Preview"
                  class="w-full h-full object-cover"
                />
              </div>
            {:else if bannerFile}
              <div class="absolute inset-0 z-0 select-none">
                <div
                  class="absolute inset-0 flex items-center justify-center bg-primary/80 z-10"
                >
                  <span
                    class="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface"
                    >Banner Selected</span
                  >
                </div>
                <img
                  src={URL.createObjectURL(bannerFile)}
                  alt="Preview"
                  class="w-full h-full object-cover blur-[2px]"
                />
              </div>
            {/if}
            <input
              id="banner"
              type="file"
              accept="image/*"
              class="hidden"
              onchange={onBannerChange}
            />
          </label>
        </div>
      </div>
    </div>
  </div>

  <div
    class="flex items-center justify-end gap-5 pt-10 border-t border-outline-variant"
  >
    <a
      href="/admin/animes"
      class="px-8 py-4 text-[11px] font-black uppercase tracking-widest text-on-surface-variant hover:text-on-surface transition-all"
    >
      Cancel
    </a>
    <button
      type="submit"
      disabled={loading || !title}
      class="px-10 py-4 text-[11px] font-black uppercase tracking-widest text-on-surface bg-primary hover:bg-primary-container rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-3 active:scale-95"
    >
      {#if loading}
        <Loader2 class="animate-spin w-4 h-4" />
        Saving...
      {:else}
        Update Anime
      {/if}
    </button>
  </div>
</form>
