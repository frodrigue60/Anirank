<script lang="ts">
  import { authState, setUser } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { toastState } from "$lib/state/toast.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import KeyRound from "lucide-svelte/icons/key-round";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Mail from "lucide-svelte/icons/mail";
  import BadgeCheck from "lucide-svelte/icons/badge-check";
  import ShieldAlert from "lucide-svelte/icons/shield-alert";
  import User from "lucide-svelte/icons/user";

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
    const state = page.url.searchParams.get("state");

    if (code) {
      try {
        if (state === "discord_link") {
          await api.post("/auth/discord/callback", { code });
          toastState.addToast("Discord account linked successfully!", "success");
        } else if (page.url.searchParams.has("scope")) {
          // Google fallback
          await api.post("/auth/google/callback", { code });
          toastState.addToast("Google account linked successfully!", "success");
        } else {
          // Anilist fallback
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

  async function handleDiscordLink() {
    try {
      const response = await api.get("/auth/discord/link");
      const url = response.data.data?.url || response.data.url;
      if (url) {
        window.location.href = url;
      }
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to get Discord link."),
        "error",
      );
    }
  }

  async function handleUnlink(provider: string) {
    if (!confirm(`Are you sure you want to unlink your ${provider} account?`))
      return;

    try {
      await api.delete(`/auth/${provider}/unlink`);
      toastState.addToast(
        `${provider.charAt(0).toUpperCase() + provider.slice(1)} account unlinked successfully.`,
        "success",
      );

      // Update local state
      if (authState.user) {
        authState.user.social_identities =
          authState.user.social_identities?.filter(
            (s) => s.provider !== provider,
          );
      }
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, `Failed to unlink ${provider} account.`),
        "error",
      );
    }
  }

  async function handleResendVerification() {
    try {
      await api.post("/users/resend-verification");
      toastState.addToast("Verification instructions sent to your email!", "success");
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to send verification email."),
        "error",
      );
    }
  }

  async function handleResetPassword() {
    if (!confirm("Are you sure you want to request a password reset link?")) return;
    
    try {
      await api.post("/forgot-password", { email: authState.user?.email });
      toastState.addToast("Password reset link sent to your email!", "success");
    } catch (err: unknown) {
      toastState.addToast(
        getApiErrorMessage(err, "Failed to send reset link."),
        "error",
      );
    }
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

  let username = $state(authState.user?.name || "");
  let isUpdatingUsername = $state(false);

  $effect(() => {
    if (authState.user?.name && !username) {
      username = authState.user.name;
    }
  });

  async function handleUpdateUsername() {
    const trimmed = username.trim();
    if (!trimmed) {
      toastState.addToast("El nombre de usuario es obligatorio.", "error");
      return;
    }
    if (trimmed.length < 3 || trimmed.length > 25) {
      toastState.addToast("El nombre de usuario debe tener entre 3 y 25 caracteres.", "error");
      return;
    }

    isUpdatingUsername = true;
    try {
      const response = await api.put("/users/username", { username: trimmed });
      toastState.addToast("Username updated successfully!", "success");

      if (response.data.data) {
        setUser(response.data.data);
      } else if (authState.user) {
        authState.user.name = trimmed;
      }
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to update username."), "error");
    } finally {
      isUpdatingUsername = false;
    }
  }

  let email = $state(authState.user?.email || "");
  let isUpdatingEmail = $state(false);

  $effect(() => {
    if (authState.user?.email && !email) {
      email = authState.user.email;
    }
  });

  async function handleUpdateEmail() {
    if (!email || !email.includes("@")) {
      toastState.addToast("Please enter a valid email address.", "error");
      return;
    }
    
    isUpdatingEmail = true;
    try {
      await api.put("/users/email", { email });
      toastState.addToast("Email updated successfully!", "success");
      
      if (authState.user) {
        authState.user.email = email;
        authState.user.email_verified_at = null;
      }
    } catch (err: any) {
      toastState.addToast(getApiErrorMessage(err, "Failed to update email."), "error");
    } finally {
      isUpdatingEmail = false;
    }
  }

  // Derived social identity helpers
  const getIdentity = (provider: string) => 
    authState.user?.social_identities?.find(s => s.provider === provider);
  
  const isLinked = (provider: string) => !!getIdentity(provider);
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
  <!-- Email Verification -->
  {#if !authState.user?.email_verified_at}
    <section
      class="bg-amber-500/5 border border-amber-500/20 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
    >
      <div class="p-6 flex items-center justify-between gap-6">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center border border-amber-500/20">
            <ShieldAlert class="w-6 h-6 text-amber-500" />
          </div>
          <div>
            <h2 class="text-sm font-bold text-amber-500 uppercase tracking-widest">Verify your account</h2>
            <p class="text-xs text-on-surface-variant mt-1 max-w-md">
              {#if !authState.user?.email}
                You need to associate an email address with your account before you can verify it.
              {:else}
                Your account is not verified. Please check your email for the verification link or request a new one below.
              {/if}
            </p>
          </div>
        </div>
        <button
          onclick={handleResendVerification}
          disabled={!authState.user?.email || !authState.user.email.includes("@")}
          class="px-5 py-2 bg-amber-500 text-black font-bold text-xs uppercase tracking-widest rounded-sm hover:scale-105 active:scale-95 transition-all shadow-lg shadow-amber-500/20 whitespace-nowrap disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:scale-100 disabled:grayscale"
        >
          Resend Link
        </button>
      </div>
    </section>
  {:else}
    <section
      class="bg-green-500/5 border border-green-500/10 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
    >
      <div class="p-6 flex items-center gap-4">
        <div class="w-10 h-10 rounded-full bg-green-500/10 flex items-center justify-center border border-green-500/20">
          <BadgeCheck class="w-5 h-5 text-green-500" />
        </div>
        <div>
          <h2 class="text-sm font-bold text-green-500 uppercase tracking-widest">Verified Account</h2>
          <p class="text-xs text-on-surface-variant mt-0.5">
            Your email address <span class="text-on-surface font-medium">{authState.user?.email}</span> has been confirmed.
          </p>
        </div>
      </div>
    </section>
  {/if}

  <!-- Username -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Username
      </h2>
    </div>
    <div class="p-8">
      <div class="grid gap-4">
        <label for="username" class="text-xs font-bold text-on-surface-variant uppercase tracking-widest">Display Username</label>
        <div class="flex gap-4">
          <div class="relative flex-1">
            <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <User class="w-4 h-4 text-on-surface-variant/40" />
            </div>
            <input
              id="username"
              type="text"
              bind:value={username}
              placeholder="Username"
              disabled={!authState.user?.email_verified_at}
              class="w-full bg-surface-lowest border border-on-surface-variant/10 rounded-sm py-3 pl-11 pr-4 text-sm text-on-surface placeholder:text-on-surface-variant/20 focus:outline-none focus:border-primary/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            />
          </div>
          <button
            onclick={handleUpdateUsername}
            disabled={isUpdatingUsername || username.trim() === authState.user?.name || !authState.user?.email_verified_at}
            class="px-8 py-3 bg-primary text-on-primary font-bold text-xs uppercase tracking-widest rounded-sm hover:scale-105 active:scale-95 transition-all shadow-lg shadow-primary/20 disabled:opacity-50 disabled:scale-100 disabled:shadow-none disabled:cursor-not-allowed"
          >
            {isUpdatingUsername ? "Updating..." : "Update"}
          </button>
        </div>
        {#if !authState.user?.email_verified_at}
          <p class="text-[10px] text-amber-500 font-bold leading-relaxed flex items-center gap-1.5 animate-pulse">
            <ShieldAlert size={12} class="shrink-0" /> Requires email verification to change your username.
          </p>
        {:else}
          <p class="text-[10px] text-on-surface-variant/60 leading-relaxed">
            Must be between 3 and 25 characters. Changing your username will automatically regenerate your profile URL slug.
          </p>
        {/if}
      </div>
    </div>
  </section>

  <!-- Email Address -->
  <section
    class="bg-surface-container border border-white/5 rounded-md overflow-hidden shadow-sm transition-all duration-500 animate-in fade-in slide-in-from-bottom-4"
  >
    <div class="px-8 py-6 border-b border-white/5 bg-surface-highest">
      <h2 class="text-lg font-bold text-on-surface tracking-tight">
        Email Address
      </h2>
    </div>
    <div class="p-8">
      <div class="grid gap-4">
        <label for="email" class="text-xs font-bold text-on-surface-variant uppercase tracking-widest">Contact Email</label>
        <div class="flex gap-4">
          <div class="relative flex-1">
            <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Mail class="w-4 h-4 text-on-surface-variant/40" />
            </div>
            <input
              id="email"
              type="email"
              bind:value={email}
              placeholder="your@email.com"
              class="w-full bg-surface-lowest border border-on-surface-variant/10 rounded-sm py-3 pl-11 pr-4 text-sm text-on-surface placeholder:text-on-surface-variant/20 focus:outline-none focus:border-primary/50 transition-colors"
            />
          </div>
          <button
            onclick={handleUpdateEmail}
            disabled={isUpdatingEmail || email === authState.user?.email}
            class="px-8 py-3 bg-primary text-on-primary font-bold text-xs uppercase tracking-widest rounded-sm hover:scale-105 active:scale-95 transition-all shadow-lg shadow-primary/20 disabled:opacity-50 disabled:scale-100 disabled:shadow-none"
          >
            {isUpdatingEmail ? "Updating..." : "Update"}
          </button>
        </div>
        <p class="text-[10px] text-on-surface-variant/60 leading-relaxed">
          {#if authState.user?.email_verified_at}
            Your email is currently verified. Changing it will require a new verification.
          {:else}
            Provide an email address to receive notifications and recover your account if you lose access.
          {/if}
        </p>
      </div>
    </div>
  </section>

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
            {#if isLinked("anilist")}
              <p class="text-xs text-on-surface-variant">
                Linked as <span class="text-primary font-bold"
                  >{getIdentity("anilist")?.provider_username}</span
                >
              </p>
            {:else}
              <p class="text-xs text-on-surface-variant">Not linked</p>
            {/if}
          </div>
        </div>
        <button
          onclick={() => isLinked("anilist") ? handleUnlink("anilist") : handleAnilistLink()}
          class="px-5 py-2 rounded-sm font-bold text-xs uppercase tracking-widest transition-all {isLinked("anilist")
            ? 'bg-surface-highest text-on-surface-variant/40 border border-on-surface-variant/10 hover:text-red-400 hover:border-red-400/30'
            : 'bg-[#02a9ff] text-white shadow-sm shadow-[#02a9ff]/20 hover:scale-105 active:scale-95'}"
        >
          {isLinked("anilist") ? "Synced" : "Sync account"}
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
            {#if isLinked("google")}
              <p class="text-xs text-on-surface-variant">
                Linked as <span class="text-primary font-bold"
                  >{getIdentity("google")?.provider_username}</span
                >
              </p>
            {:else}
              <p class="text-xs text-on-surface-variant">Not linked</p>
            {/if}
          </div>
        </div>
        <button
          onclick={() => isLinked("google") ? handleUnlink("google") : handleGoogleSync()}
          class="px-5 py-2 rounded-sm font-bold text-xs uppercase tracking-widest transition-all {isLinked("google")
            ? 'bg-surface-highest text-on-surface-variant/40 border border-on-surface-variant/10 hover:text-red-400 hover:border-red-400/30'
            : 'bg-on-surface text-surface hover:scale-105 active:scale-95 shadow-sm shadow-black/20'}"
        >
          {isLinked("google") ? "Synced" : "Sync account"}
        </button>
      </div>

      <!-- Discord -->
      <div
        class="flex items-center justify-between p-4 rounded-md bg-surface-container border border-on-surface-variant/10"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-12 h-12 rounded-md bg-[#5865F2]/10 flex items-center justify-center border border-[#5865F2]/20"
          >
            <svg class="w-6 h-6 text-[#5865F2]" viewBox="0 0 24 24">
              <path
                fill="currentColor"
                d="M19.27 4.57c-1.51-.7-3.14-1.21-4.85-1.5-.21.38-.45.8-.62 1.2-1.84-.28-3.66-.28-5.47 0-.17-.4-.41-.82-.63-1.2-1.71.29-3.34.8-4.85 1.5C.3 9.08-.32 13.5.12 17.87c2.03 1.5 3.99 2.41 5.91 3.01.48-.65.9-1.35 1.26-2.09-.69-.26-1.35-.58-1.97-.96.16-.12.33-.25.48-.38 3.73 1.73 7.78 1.73 11.41 0 .16.13.32.26.48.38-.62.38-1.28.7-1.97.96.36.74.78 1.44 1.26 2.09 1.92-.6 3.88-1.51 5.91-3.01.52-5.1-.86-9.5-2.88-13.3zM8.02 15.33c-1.18 0-2.15-1.08-2.15-2.42 0-1.33.95-2.42 2.15-2.42 1.21 0 2.17 1.08 2.15 2.42 0 1.33-.94 2.42-2.15 2.42zm7.97 0c-1.18 0-2.15-1.08-2.15-2.42 0-1.33.95-2.42 2.15-2.42 1.21 0 2.17 1.08 2.15 2.42 0 1.33-.94 2.42-2.15 2.42z"
              />
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-on-surface">Discord</h3>
            {#if isLinked("discord")}
              <p class="text-xs text-on-surface-variant">
                Linked as <span class="text-primary font-bold"
                  >{getIdentity("discord")?.provider_username}</span
                >
              </p>
            {:else}
              <p class="text-xs text-on-surface-variant">Not linked</p>
            {/if}
          </div>
        </div>
        <button
          onclick={() => isLinked("discord") ? handleUnlink("discord") : handleDiscordLink()}
          class="px-5 py-2 rounded-sm font-bold text-xs uppercase tracking-widest transition-all {isLinked("discord")
            ? 'bg-surface-highest text-on-surface-variant/40 border border-on-surface-variant/10 hover:text-red-400 hover:border-red-400/30'
            : 'bg-[#5865F2] text-white hover:scale-105 active:scale-95 shadow-sm shadow-[#5865F2]/20'}"
        >
          {isLinked("discord") ? "Synced" : "Sync account"}
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
        <KeyRound
          class="text-on-surface-variant group-hover:text-primary transition-colors"
          size={24}
        />
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
        <Trash2
          class="text-red-500/20 group-hover:text-red-500 transition-colors"
          size={24}
        />
      </button>
    </div>
  </section>
</div>
