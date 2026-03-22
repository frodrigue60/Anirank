<script lang="ts">
  import { authState, setUser } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import {
    Camera,
    Image as ImageIcon,
    Save,
    Loader2,
  } from "lucide-svelte";

  let scoreFormat = $state(authState.user?.score_format_id || 1);
  let isUploadingAvatar = $state(false);
  let isUploadingBanner = $state(false);
  let isSavingSettings = $state(false);

  const scoreFormats = [
    { id: 1, slug: "POINT_100", name: "100 Point Scale (0-100)" },
    { id: 2, slug: "POINT_10_DECIMAL", name: "10 Point Decimal (0.0-10.0)" },
    { id: 3, slug: "POINT_10", name: "10 Point Scale (0-10)" },
    { id: 4, slug: "POINT_5", name: "5 Star Scale (0.5-5.0)" },
  ];

  $effect(() => {
    if (authState.user) {
      scoreFormat = authState.user.score_format_id || 1;
    }
  });

  async function handleAvatarUpload(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      isUploadingAvatar = true;
      const formData = new FormData();
      formData.append("image", target.files[0]);

      try {
        const response = await api.post("/users/avatar", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        if (response.data.success) {
          if (authState.user) {
            setUser({
              ...authState.user,
              avatar_url: `${response.data.avatar_url}?t=${Date.now()}`,
            });
          }
          toastState.addToast("Avatar updated successfully!", "success");
        }
      } catch (err: any) {
        toastState.addToast(
          err.response?.data?.message || "Failed to upload avatar.",
          "error",
        );
      } finally {
        isUploadingAvatar = false;
      }
    }
  }

  async function handleBannerUpload(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      isUploadingBanner = true;
      const formData = new FormData();
      formData.append("banner", target.files[0]);

      try {
        const response = await api.post("/users/banner", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        if (response.data.success) {
          if (authState.user) {
            setUser({
              ...authState.user,
              banner_url: `${response.data.banner_url}?t=${Date.now()}`,
            });
          }
          toastState.addToast("Banner updated successfully!", "success");
        }
      } catch (err: any) {
        toastState.addToast(
          err.response?.data?.message || "Failed to upload banner.",
          "error",
        );
      } finally {
        isUploadingBanner = false;
      }
    }
  }

  async function saveSettings() {
    isSavingSettings = true;
    try {
      const response = await api.post("/users/score-format", {
        score_format: scoreFormat,
      });
      if (response.data.success) {
        if (authState.user) {
          const selectedFormat = scoreFormats.find((f) => f.id === scoreFormat);
          setUser({
            ...authState.user,
            score_format_id: scoreFormat,
            score_format: selectedFormat?.slug || "",
          });
        }
        toastState.addToast("Settings saved successfully!", "success");
      }
    } catch (err: any) {
      toastState.addToast(
        err.response?.data?.message || "Failed to save settings.",
        "error",
      );
    } finally {
      isSavingSettings = false;
    }
  }
</script>

<div class="mb-10">
  <h1
    class="text-3xl font-black text-white tracking-tighter transition-all duration-500 animate-in fade-in slide-in-from-left-4"
  >
    Profile Settings
  </h1>
  <p class="text-white/40 text-sm mt-1">
    Customize your public appearance and site preferences.
  </p>
</div>

<div class="grid gap-8">
  <!-- Profile Color Section -->
  <section
    class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
  >
    <div
      class="px-8 py-6 border-b border-white/5 bg-white/2 flex justify-between items-center"
    >
      <h2 class="text-lg font-bold text-white tracking-tight">Profile Color</h2>
      <div
        class="px-3 py-1 bg-white/5 rounded-full text-[10px] font-black uppercase tracking-widest text-white/20 border border-white/5"
      >
        Coming Soon
      </div>
    </div>
    <div class="p-8">
      <div class="flex flex-wrap gap-4">
        {#each ["#3db4f2", "#c063ff", "#4cca51", "#ef881a", "#e13333", "#fc9dd4", "#677b94"] as color}
          <button
            class="w-12 h-12 rounded-xl border-2 border-transparent hover:border-white/20 transition-all hover:scale-110 shadow-lg"
            style="background-color: {color}"
            aria-label="Select profile color {color}"
          ></button>
        {/each}
        <button
          class="w-12 h-12 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center text-white/20 hover:text-white/40 group transition-all"
          aria-label="Custom color locked"
        >
          <span
            class="material-symbols-outlined text-2xl group-hover:scale-110 transition-transform"
            >lock</span
          >
        </button>
      </div>
    </div>
  </section>

  <!-- Site Theme Section -->
  <section
    class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 animate-in fade-in slide-in-from-bottom-5"
  >
    <div
      class="px-8 py-6 border-b border-white/5 bg-white/2 flex justify-between items-center"
    >
      <h2 class="text-lg font-bold text-white tracking-tight">Site Theme</h2>
      <div
        class="px-3 py-1 bg-white/5 rounded-full text-[10px] font-black uppercase tracking-widest text-white/20 border border-white/5"
      >
        WIP
      </div>
    </div>
    <div class="p-8">
      <div class="flex gap-4">
        <button
          class="w-12 h-12 rounded-xl bg-white border-2 border-primary flex items-center justify-center text-background-dark font-black text-xl shadow-xl"
          aria-label="Select Light Theme">A</button
        >
        <button
          class="w-12 h-12 rounded-xl bg-background-dark border border-white/10 flex items-center justify-center text-white/40 font-black text-xl hover:bg-white/5 transition-colors cursor-pointer"
          aria-label="Select Dark Theme">A</button
        >
        <button
          class="w-12 h-12 rounded-xl bg-white border border-white/10 flex flex-col items-center justify-center text-background-dark font-black text-xl hover:bg-white/90 transition-colors cursor-pointer relative overflow-hidden"
          aria-label="Select Contrast Theme"
        >
          <div class="absolute inset-0 bg-white"></div>
          <div
            class="absolute bottom-0 right-0 w-0 h-0 border-t-12 border-t-transparent border-r-12 border-r-background-dark"
          ></div>
          <span class="relative z-10">A</span>
        </button>
      </div>
    </div>
  </section>

  <!-- About Section -->
  <section
    class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 animate-in fade-in slide-in-from-bottom-6"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-white/2">
      <h2 class="text-lg font-bold text-white tracking-tight">About</h2>
    </div>
    <div class="p-8 space-y-4">
      <div
        class="w-full rounded-2xl bg-background-dark border border-white/5 p-4 min-h-[160px] flex flex-col"
      >
        <div
          class="flex gap-4 mb-4 pb-4 border-b border-white/5 text-white/20"
        >
          <span
            class="material-symbols-outlined text-xl hover:text-white cursor-pointer transition-colors"
            >format_bold</span
          >
          <span
            class="material-symbols-outlined text-xl hover:text-white cursor-pointer transition-colors"
            >format_italic</span
          >
          <span
            class="material-symbols-outlined text-xl hover:text-white cursor-pointer transition-colors"
            >link</span
          >
          <span
            class="material-symbols-outlined text-xl hover:text-white cursor-pointer transition-colors"
            >image</span
          >
          <span
            class="material-symbols-outlined text-xl hover:text-white cursor-pointer transition-colors"
            >format_list_bulleted</span
          >
        </div>
        <textarea
          placeholder="A little about yourself..."
          class="bg-transparent border-none outline-none text-white/80 text-sm resize-none flex-1 font-medium placeholder:text-white/10"
        ></textarea>
      </div>
      <p class="text-[10px] text-white/20 px-2 font-medium italic">
        Supports Markdown formatting.
      </p>
    </div>
  </section>

  <!-- Appearance Section -->
  <section
    class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 animate-in fade-in slide-in-from-bottom-7"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-white/2">
      <h2 class="text-lg font-bold text-white tracking-tight">Appearance</h2>
    </div>

    <div class="p-8 space-y-10">
      <!-- Banner Upload -->
      <div>
        <span
          class="block text-[11px] font-black uppercase tracking-[0.2em] text-white/20 mb-4 px-1"
          >Profile Banner</span
        >
        <div
          class="relative group rounded-2xl overflow-hidden h-48 bg-background-dark border border-white/5"
        >
          {#if authState.user?.banner_url}
            <img
              src={authState.user.banner_url}
              alt="Banner"
              class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
            />
          {:else}
            <div
              class="w-full h-full flex flex-col items-center justify-center text-white/10 gap-2"
            >
              <ImageIcon size={48} strokeWidth={1} />
              <span
                class="text-xs font-bold uppercase tracking-widest text-white/20"
                >No banner</span
              >
            </div>
          {/if}
          <div
            class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
          >
            <label
              class="cursor-pointer bg-white text-background-dark px-8 py-3 rounded-full font-black text-xs uppercase tracking-widest shadow-2xl hover:scale-105 transition-transform flex items-center gap-2"
            >
              {#if isUploadingBanner}
                <Loader2 class="animate-spin" size={16} />
                Uploading...
              {:else}
                <Camera size={16} />
                Update Banner
              {/if}
              <input
                type="file"
                class="hidden"
                accept="image/*"
                onchange={handleBannerUpload}
                disabled={isUploadingBanner}
              />
            </label>
          </div>
        </div>
        <p class="text-[10px] text-white/20 mt-3 px-1 font-medium italic">
          Optimal: 1700x330px. Max 6MB.
        </p>
      </div>

      <!-- Avatar Upload -->
      <div class="flex flex-col sm:flex-row items-center gap-8">
        <div class="relative group">
          <div
            class="w-32 h-32 rounded-full overflow-hidden border-4 border-background-dark bg-background-dark shadow-2xl relative"
          >
            {#if authState.user?.avatar_url}
              <img
                src={authState.user.avatar_url}
                alt="Avatar"
                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
              />
            {:else}
              <img
                src="https://ui-avatars.com/api/?name={authState.user
                  ?.name}&background=7f13ec&color=fff&size=128"
                alt="Default Avatar"
                class="w-full h-full object-cover"
              />
            {/if}
            <div
              class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
            >
              <label
                class="cursor-pointer text-white flex flex-col items-center gap-1"
              >
                <Camera size={24} />
                <span class="text-[10px] font-black uppercase tracking-widest"
                  >Edit</span
                >
                <input
                  type="file"
                  class="hidden"
                  accept="image/*"
                  onchange={handleAvatarUpload}
                  disabled={isUploadingAvatar}
                />
              </label>
            </div>
          </div>
          {#if isUploadingAvatar}
            <div
              class="absolute inset-0 rounded-full bg-background-dark/80 flex items-center justify-center"
            >
              <Loader2 class="animate-spin text-primary" size={24} />
            </div>
          {/if}
        </div>
        <div class="text-center sm:text-left">
          <h3 class="text-lg font-bold text-white tracking-tight">
            Profile Picture
          </h3>
          <p class="text-sm text-white/40 max-w-xs mt-1">
            Update your avatar to make your profile unique.
          </p>
          <p class="text-[10px] text-white/20 mt-2 font-medium italic">
            Optimal: 230x230px. Max 3MB.
          </p>
        </div>
      </div>
    </div>
  </section>

  <!-- Preferences Section -->
  <section
    class="bg-surface-dark border border-white/5 rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 animate-in fade-in slide-in-from-bottom-8"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-white/2">
      <h2 class="text-lg font-bold text-white tracking-tight">Preferences</h2>
    </div>

    <div class="p-8 space-y-6">
      <div>
        <label
          for="score-format"
          class="block text-[11px] font-black uppercase tracking-[0.2em] text-white/20 mb-3 px-1"
          >Score Format</label
        >
        <div class="relative max-w-md">
          <select
            id="score-format"
            bind:value={scoreFormat}
            class="w-full bg-background-dark border border-white/5 rounded-xl px-4 py-4 text-sm text-white font-medium focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer"
          >
            {#each scoreFormats as format}
              <option value={format.id}>{format.name}</option>
            {/each}
          </select>
          <div
            class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-white/20"
          >
            <span class="material-symbols-outlined text-[20px]"
              >expand_more</span
            >
          </div>
        </div>
        <p class="text-xs text-white/30 mt-3 leading-relaxed font-medium">
          Choose how scores and ratings are displayed throughout the site. This
          affects both your votes and the average ratings you see.
        </p>
      </div>

      <div class="pt-6 border-t border-white/5 flex justify-end">
        <button
          onclick={saveSettings}
          disabled={isSavingSettings}
          class="bg-primary hover:opacity-90 disabled:opacity-50 text-white px-10 py-4 rounded-xl font-black text-xs uppercase tracking-[0.15em] transition-all shadow-[0_0_30px_rgba(127,19,236,0.2)] flex items-center gap-3 active:scale-95"
        >
          {#if isSavingSettings}
            <Loader2 class="animate-spin" size={18} />
            Saving...
          {:else}
            <Save size={18} />
            Save Preferences
          {/if}
        </button>
      </div>
    </div>
  </section>
</div>

<style lang="postcss">
  select option {
    background: var(--color-background-dark);
    color: white;
  }
</style>
