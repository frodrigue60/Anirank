<script lang="ts">
  import { authState, setUser } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { toastState } from "$lib/state/toast.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/state";

  onMount(async () => {
    const oauthError = page.url.searchParams.get("error");
    const oauthDesc = page.url.searchParams.get("error_description");
    if (oauthError) {
      toastState.addToast(
        oauthDesc || oauthError || "OAuth was cancelled or failed.",
        "error",
        8000,
      );
      const url = new URL(window.location.href);
      url.search = "";
      window.history.replaceState({}, "", url.toString());
      return;
    }

    const code = page.url.searchParams.get("code");
    if (code) {
      // Check if it's a Google callback (state or just try)
      // For simplicity, we can try to guess or use a state param.
      // But since we are only linking here, we can try one and then the other,
      // or check the URL for breadcrumbs.
      // Usually, Google returns 'scope' or other params.
      const isGoogle = page.url.searchParams.has("scope");

      try {
        if (isGoogle) {
          await api.post("/auth/google/callback", { code });
          toastState.addToast("Google account linked successfully!", "success");
        } else {
          await api.post("/auth/anilist/callback", { code });
          toastState.addToast(
            "Anilist account linked successfully!",
            "success",
          );
        }

        // Refresh user state to update UI buttons immediately
        const profileRes = await api.get("/profile");
        if (profileRes.data.data) {
          setUser(profileRes.data.data);
        }

        // Clean URL completely
        const url = new URL(window.location.href);
        url.search = "";
        window.history.replaceState({}, "", url.toString());
      } catch (err: any) {
        toastState.addToast(
          err.response?.data?.message || "Failed to link account.",
          "error",
        );
      }
    }
  });

  async function handleAnilistLink() {
    try {
      const response = await api.get("/auth/anilist/link");
      const url = response.data.data?.url || response.data.url;
      if (url) {
        window.location.href = url;
      }
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to get Anilist link."),
        "error",
      );
    }
  }

  async function handleGoogleSync() {
    try {
      const response = await api.get("/auth/google/link");
      const url = response.data.data?.url || response.data.url;
      if (url) {
        window.location.href = url;
      }
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to get Google link."),
        "error",
      );
    }
  }

  function handleResetPassword() {
    toastState.addToast("Password reset email sent (WIP)!", "info");
  }

  function handleDeleteAccount() {
    if (
      confirm(
        "Are you sure you want to delete your account? This action is permanent.",
      )
    ) {
      toastState.addToast("Account deletion request sent (WIP)!", "info");
    }
  }
</script>

<div class="mb-10">
  <h1
    class="text-3xl font-black text-on-surface tracking-tighter transition-all duration-500 animate-in fade-in slide-in-from-left-4"
  >
    Account Settings
  </h1>
  <p class="text-on-surface-variant text-sm mt-1">
    Manage your account connections, security, and data.
  </p>
</div>

<div class="grid gap-8">
  <!-- Account Connections -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Account Connections
      </h2>
    </div>
    <div class="p-8 space-y-6">
      <!-- Anilist -->
      <div
        class="flex items-center justify-between p-4 rounded-md bg-surface-container border border-on-surface-variant/10"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-12 h-12 rounded-md bg-[#2B2D42] flex items-center justify-center border border-[#02a9ff]/20"
          >
            <img src="/images/anilist_icon.svg" alt="Anilist" class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-sm font-bold text-on-surface">Anilist</h3>
            {#if authState.user?.anilist_username}
              <p class="text-xs text-on-surface-variant">
                Linked as <span class="text-primary font-bold"
                  >{authState.user.anilist_username}</span
                >
              </p>
            {:else}
              <p class="text-xs text-on-surface-variant">Not linked</p>
            {/if}
          </div>
        </div>
        <button
          onclick={handleAnilistLink}
          class="px-5 py-2 rounded-sm font-bold text-xs uppercase tracking-widest transition-all {authState
            .user?.anilist_id
            ? 'bg-surface-highest text-on-surface-variant/40 border border-on-surface-variant/10'
            : 'bg-[#02a9ff] text-white shadow-sm shadow-[#02a9ff]/20 hover:scale-105 active:scale-95'}"
        >
          {authState.user?.anilist_id ? "Synced" : "Sync account"}
        </button>
      </div>

      <!-- Google -->
      <div
        class="flex items-center justify-between p-4 rounded-md bg-surface-container border border-on-surface-variant/10"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-12 h-12 rounded-md bg-surface-container flex items-center justify-center border border-on-surface-variant/10"
          >
            <svg class="w-6 h-6 text-on-surface-variant" viewBox="0 0 24 24">
              <path
                fill="currentColor"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="currentColor"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="currentColor"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
              />
              <path
                fill="currentColor"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-on-surface">Google</h3>
            {#if authState.user?.google_email}
              <p class="text-xs text-on-surface-variant">
                Linked as <span class="text-primary font-bold"
                  >{authState.user.google_email}</span
                >
              </p>
            {:else}
              <p class="text-xs text-on-surface-variant">Not linked</p>
            {/if}
          </div>
        </div>
        <button
          onclick={handleGoogleSync}
          class="px-5 py-2 rounded-sm font-bold text-xs uppercase tracking-widest transition-all {authState
            .user?.google_id
            ? 'bg-surface-highest text-on-surface-variant/40 border border-on-surface-variant/10'
            : 'bg-on-surface text-surface hover:scale-105 active:scale-95 shadow-sm shadow-black/20'}"
        >
          {authState.user?.google_id ? "Synced" : "Sync account"}
        </button>
      </div>
    </div>
  </section>

  <!-- Security -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-5"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">Security</h2>
    </div>
    <div class="p-8 space-y-4">
      <button
        onclick={handleResetPassword}
        class="w-full flex items-center justify-between p-4 rounded-sm bg-surface-low border border-on-surface-variant/10 hover:bg-surface-highest transition-colors group"
      >
        <div class="text-left">
          <h3 class="text-sm font-bold text-on-surface">Reset Password</h3>
          <p class="text-xs text-on-surface-variant">
            Send a password reset link to your email.
          </p>
        </div>
        <span
          class="material-symbols-outlined text-on-surface-variant group-hover:text-primary transition-colors"
          >lock_reset</span
        >
      </button>
    </div>
  </section>

  <!-- Termination -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-6"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-red-500/80 tracking-tight">
        Danger Zone
      </h2>
    </div>
    <div class="p-8">
      <button
        onclick={handleDeleteAccount}
        class="w-full flex items-center justify-between p-4 rounded-sm bg-red-500/5 border border-red-500/10 hover:bg-red-500/10 transition-colors group"
      >
        <div class="text-left">
          <h3 class="text-sm font-bold text-red-500/80">Delete Account</h3>
          <p class="text-xs text-red-500/40">
            Permanently remove your account and all data.
          </p>
        </div>
        <span
          class="material-symbols-outlined text-red-500/20 group-hover:text-red-500 transition-colors"
          >delete_forever</span
        >
      </button>
    </div>
  </section>
</div>
