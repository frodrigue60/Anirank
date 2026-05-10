<script lang="ts">
  import { authState, setUser } from "$lib/state/auth.svelte";
  import { themeState, type Theme } from "$lib/state/theme.svelte";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { toastState } from "$lib/state/toast.svelte";
  import Camera from "lucide-svelte/icons/camera";
import ImageIcon from "lucide-svelte/icons/image";
import Save from "lucide-svelte/icons/save";
import Loader2 from "lucide-svelte/icons/loader-2";
import Lock from "lucide-svelte/icons/lock";
import ChevronDown from "lucide-svelte/icons/chevron-down";

  let scoreFormat = $state(authState.user?.score_format || "POINT_10_DECIMAL");
  let profileColor = $state(authState.user?.profile_color || "#3db4f2");
  let about = $state(authState.user?.about || "");
  let isUploadingAvatar = $state(false);
  let isUploadingBanner = $state(false);
  let isSavingSettings = $state(false);
  let isSavingProfile = $state(false);

  const scoreFormats = [
    { id: 1, slug: "POINT_100", name: "100 Point Scale (0-100)" },
    { id: 2, slug: "POINT_10_DECIMAL", name: "10 Point Decimal (0.0-10.0)" },
    { id: 3, slug: "POINT_10", name: "10 Point Scale (0-10)" },
    { id: 4, slug: "POINT_5", name: "5 Star Scale (0.5-5.0)" },
  ];

  $effect(() => {
    if (authState.user) {
      scoreFormat = authState.user.score_format || "POINT_10_DECIMAL";
      profileColor = authState.user.profile_color || "#3db4f2";
      about = authState.user.about || "";
    }
  });

  function handleThemeChange(newTheme: Theme) {
    themeState.set(newTheme);
    toastState.addToast(`Theme changed to ${newTheme}`, "success");
  }

  async function updateProfile() {
    isSavingProfile = true;
    try {
      const response = await api.patch("/users/profile", {
        about: about,
        profile_color: profileColor,
      });
      if (response.data.success) {
        if (authState.user) {
          setUser({
            ...authState.user,
            about: about,
            profile_color: profileColor,
          });
        }
        toastState.addToast("Profile updated successfully!", "success");
      }
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to update profile."),
        "error",
      );
    } finally {
      isSavingProfile = false;
    }
  }

  async function handleColorSelect(color: string) {
    profileColor = color;
    await updateProfile();
  }

  const validateImageFile = (file: File, maxSizeMB: number): Promise<boolean> => {
    return new Promise((resolve) => {
      if (file.size > maxSizeMB * 1024 * 1024) {
        toastState.addToast(`El archivo excede el límite de ${maxSizeMB}MB`, "error");
        resolve(false);
        return;
      }

      const img = new Image();
      img.src = URL.createObjectURL(file);
      img.onload = () => {
        URL.revokeObjectURL(img.src);
        if (img.width > 4000 || img.height > 4000) {
          toastState.addToast("La resolución de la imagen excede el máximo permitido (4000x4000px)", "error");
          resolve(false);
          return;
        }
        resolve(true);
      };
      img.onerror = () => {
        URL.revokeObjectURL(img.src);
        toastState.addToast("Archivo de imagen inválido", "error");
        resolve(false);
      };
    });
  };

  async function handleAvatarUpload(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      const file = target.files[0];
      const isValid = await validateImageFile(file, 2);
      if (!isValid) {
        target.value = ""; // Reset input
        return;
      }

      isUploadingAvatar = true;
      const formData = new FormData();
      formData.append("image", file);

      try {
        const response = await api.post("/users/avatar", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        if (response.data.success) {
          if (authState.user) {
            setUser({
              ...authState.user,
              avatar_url: `${response.data.data.avatar_url}?t=${Date.now()}`,
            });
          }
          toastState.addToast("Avatar updated successfully!", "success");
        }
      } catch (err: any) {
        toastState.addToast(
          getApiErrorMessage(err, "Failed to upload avatar."),
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
      const file = target.files[0];
      const isValid = await validateImageFile(file, 4);
      if (!isValid) {
        target.value = ""; // Reset input
        return;
      }

      isUploadingBanner = true;
      const formData = new FormData();
      formData.append("banner", file);

      try {
        const response = await api.post("/users/banner", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        if (response.data.success) {
          if (authState.user) {
            setUser({
              ...authState.user,
              banner_url: `${response.data.data.banner_url}?t=${Date.now()}`,
            });
          }
          toastState.addToast("Banner updated successfully!", "success");
        }
      } catch (err: any) {
        toastState.addToast(
          getApiErrorMessage(err, "Failed to upload banner."),
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
          setUser({
            ...authState.user,
            score_format: scoreFormat,
          });
        }
        toastState.addToast("Settings saved successfully!", "success");
      }
    } catch (err: any) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to save settings."),
        "error",
      );
    } finally {
      isSavingSettings = false;
    }
  }
</script>

<div class="mb-10">
  <h1
    class="text-3xl font-black text-on-surface tracking-tighter transition-all duration-500 animate-in fade-in slide-in-from-left-4"
  >
    Profile Settings
  </h1>
  <p class="text-on-surface-variant text-sm mt-1">
    Customize your public appearance and site preferences.
  </p>
</div>

<div class="grid gap-8">
  <!-- Profile Color Section -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
  >
    <div
      class="px-8 py-6 border-b border-white/5 bg-surface-highest flex justify-between items-center"
    >
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Profile Color
      </h2>
      {#if isSavingProfile}
        <div
          class="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-primary animate-pulse"
        >
          <Loader2 size={12} class="animate-spin" />
          Syncing...
        </div>
      {/if}
    </div>
    <div class="p-8">
      <div class="flex flex-wrap gap-4">
        {#each ["#3db4f2", "#c063ff", "#4cca51", "#ef881a", "#e13333", "#fc9dd4", "#677b94"] as color}
          <button
            class="w-12 h-12 rounded-sm border-2 transition-all hover:scale-110 shadow-sm {profileColor ===
            color
              ? 'border-secondary scale-110'
              : 'border-transparent hover:border-secondary'}"
            style="background-color: {color}"
            aria-label="Select profile color {color}"
            onclick={() => handleColorSelect(color)}
          ></button>
        {/each}
        <button
          class="w-12 h-12 rounded-sm bg-white/5 border border-primary/20 flex items-center justify-center text-on-surface hover:text-on-surface/40 group transition-all"
          aria-label="Custom color locked"
          title="Custom colors available for supporters soon"
        >
          <Lock class="group-hover:scale-110 transition-transform" size={24} />

        </button>
      </div>
      <p class="text-[10px] text-on-surface/20 mt-4 px-1 font-medium italic">
        This color will be used as your accent color on your public profile.
      </p>
    </div>
  </section>

  <!-- Site Theme Section -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-5"
  >
    <div
      class="px-8 py-6 border-b border-white/5 bg-surface-highest flex justify-between items-center"
    >
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Site Theme
      </h2>
    </div>
    <div class="p-8">
      <div class="flex gap-6">
        <!-- light theme -->
        <button
          class="group flex flex-col items-center gap-3"
          onclick={() => handleThemeChange("light")}
        >
          <div
            class="w-16 h-16 rounded-sm bg-[#fff7ff] border-2 flex items-center justify-center text-[#1e1924] font-black text-2xl shadow-sm transition-all group-hover:scale-105 {themeState.current ===
            'light'
              ? 'border-primary ring-4 ring-primary/20'
              : 'border-transparent'}"
            aria-label="Select Light Theme"
          >
            A
          </div>
          <span
            class="text-[10px] font-black uppercase tracking-widest {themeState.current ===
            'light'
              ? 'text-on-surface'
              : 'text-on-surface-variant'}">Light</span
          >
        </button>

        <!-- dark theme -->
        <button
          class="group flex flex-col items-center gap-3"
          onclick={() => handleThemeChange("dark")}
        >
          <div
            class="w-16 h-16 rounded-sm bg-[#261A38] border-2 flex items-center justify-center text-[#ede6f2] font-black text-2xl shadow-sm transition-all group-hover:scale-105 {themeState.current ===
            'dark'
              ? 'border-primary ring-4 ring-primary/20'
              : 'border-white/10'}"
            aria-label="Select Dark Theme"
          >
            A
          </div>
          <span
            class="text-[10px] font-black uppercase tracking-widest {themeState.current ===
            'dark'
              ? 'text-on-surface'
              : 'text-on-surface-variant'}">Dark</span
          >
        </button>

        <!-- contrast theme -->
        <button
          class="group flex flex-col items-center gap-3"
          onclick={() => handleThemeChange("contrast")}
        >
          <div
            class="w-16 h-16 rounded-sm bg-black border-2 flex flex-col items-center justify-center text-white font-black text-2xl shadow-sm transition-all group-hover:scale-105 relative overflow-hidden {themeState.current ===
            'contrast'
              ? 'border-primary ring-4 ring-primary/20'
              : 'border-white/10'}"
            aria-label="Select Contrast Theme"
          >
            <div class="absolute inset-0 bg-black"></div>
            <div
              class="absolute bottom-0 right-0 w-8 h-8 bg-white rotate-45 translate-x-1/2 translate-y-1/2"
            ></div>
            <span class="relative z-10 text-white">A</span>
          </div>
          <span
            class="text-[10px] font-black uppercase tracking-widest {themeState.current ===
            'contrast'
              ? 'text-on-surface'
              : 'text-on-surface-variant'}">Contrast</span
          >
        </button>
      </div>
    </div>
  </section>

  <!-- About Section -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-6"
  >
    <div
      class="px-8 py-6 border-b border-white/5 bg-surface-highest flex justify-between items-center"
    >
      <h2 class="text-lg font-bold text-on-surface tracking-tight">About</h2>
      {#if isSavingProfile}
        <div
          class="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-primary animate-pulse"
        >
          <Loader2 size={12} class="animate-spin" />
          Saving...
        </div>
      {/if}
    </div>
    <!-- about markdown editor -->
    <div class="p-8 space-y-4">
      <div
        class="w-full rounded-md bg-surface-low border border-white/5 p-4 min-h-[200px] flex flex-col focus-within:border-primary/30 transition-all shadow-inner"
      >
        <div
          class="flex gap-4 mb-4 pb-4 border-b border-on-surface-variant/10 text-on-surface-variant"
        >
          <span class="text-[10px] font-black uppercase tracking-widest"
            >Controls (WIP)</span
          >
        </div>
        <textarea
          placeholder="A little about yourself..."
          class="bg-transparent border-none outline-none text-on-surface text-sm resize-none flex-1 font-medium placeholder:text-on-surface-variant/40"
          bind:value={about}
          onblur={updateProfile}
        ></textarea>
      </div>
      <div class="flex justify-between items-center px-2">
        <p class="text-[10px] text-on-surface-variant font-medium italic">
          Supports Markdown formatting. Changes are saved automatically.
        </p>
        <span
          class="text-[10px] text-on-surface-variant/30 font-bold uppercase tracking-widest"
        >
          {about.length} characters
        </span>
      </div>
    </div>
  </section>

  <!-- Appearance Section -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-7"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Appearance
      </h2>
    </div>

    <div class="p-8 space-y-10">
      <!-- Banner Upload -->
      <div>
        <span
          class="block text-[11px] font-black uppercase tracking-[0.2em] text-on-surface-variant mb-4 px-1"
          >Profile Banner</span
        >
        <div
          class="relative group rounded-md overflow-hidden h-48 bg-surface-low border border-white/5"
        >
          {#if authState.user?.banner_url}
            <img
              src={authState.user.banner_url}
              alt="Banner"
              class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
            />
          {:else}
            <div
              class="w-full h-full flex flex-col items-center justify-center bg-primary text-on-surface-variant/80 gap-2"
            >
              <ImageIcon size={48} strokeWidth={1} />
              <span
                class="text-xs font-bold uppercase tracking-widest text-on-surface-variant/80"
                >No banner</span
              >
            </div>
          {/if}
          <div
            class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
          >
            <label
              class="cursor-pointer bg-on-surface text-surface px-8 py-3 rounded-sm font-black text-xs uppercase tracking-widest shadow-sm hover:scale-105 transition-transform flex items-center gap-2"
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
        <p
          class="text-[10px] text-on-surface-variant mt-3 px-1 font-medium italic"
        >
          Optimal: 1700x330px. Max 4MB.
        </p>
      </div>

      <!-- Avatar Upload -->
      <div class="flex flex-col sm:flex-row items-center gap-8">
        <div class="relative group">
          <div
            class="w-32 h-32 rounded-md overflow-hidden border-4 border-surface-container bg-surface-low shadow-sm relative"
          >
            {#if authState.user?.avatar_url}
              <img
                src={authState.user.avatar_url}
                alt="Avatar"
                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
              />
            {:else}
              <img
                src="/images/placeholders/default.svg"
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
              class="absolute inset-0 rounded-md bg-surface-low/80 flex items-center justify-center"
            >
              <Loader2 class="animate-spin text-primary" size={24} />
            </div>
          {/if}
        </div>
        <div class="text-center sm:text-left">
          <h3 class="text-lg font-bold text-on-surface tracking-tight">
            Profile Picture
          </h3>
          <p class="text-sm text-on-surface-variant max-w-xs mt-1">
            Update your avatar to make your profile unique.
          </p>
          <p
            class="text-[10px] text-on-surface-variant mt-2 font-medium italic"
          >
          Optimal: 230x230px. Max 2MB.
          <br/><span class="text-[9px] opacity-70">Max resolution 4000x4000px</span>
        </p>
        </div>
      </div>
    </div>
  </section>

  <!-- Preferences Section -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-8"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Preferences
      </h2>
    </div>

    <div class="p-8 space-y-6">
      <div>
        <label
          for="score-format"
          class="block text-[11px] font-black uppercase tracking-[0.2em] text-on-surface-variant mb-3 px-1"
          >Score Format</label
        >
        <div class="relative max-w-md">
          <select
            id="score-format"
            bind:value={scoreFormat}
            class="w-full bg-surface-low border border-on-surface-variant/10 rounded-sm px-4 py-4 text-sm text-on-surface font-medium focus:outline-none focus:border-primary/50 transition-all appearance-none cursor-pointer"
          >
            {#each scoreFormats as format}
              <option value={format.slug}>{format.name}</option>
            {/each}
          </select>
          <div
            class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-on-surface-variant"
          >
            <ChevronDown size={20} />

          </div>
        </div>
        <p
          class="text-xs text-on-surface-variant mt-3 leading-relaxed font-medium"
        >
          Choose how scores and ratings are displayed throughout the site. This
          affects both your votes and the average ratings you see.
        </p>
      </div>

      <div class="pt-6 border-t border-on-surface-variant/10 flex justify-end">
        <button
          onclick={saveSettings}
          disabled={isSavingSettings}
          class="bg-primary hover:opacity-90 disabled:opacity-50 text-white px-10 py-4 rounded-sm font-black text-xs uppercase tracking-[0.15em] transition-all shadow-sm shadow-primary/20 flex items-center gap-3 active:scale-95"
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
