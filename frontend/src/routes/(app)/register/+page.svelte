<script lang="ts">
  import { setAuthToken, setUser, authState } from "$lib/state/auth.svelte";
  import Mail from "lucide-svelte/icons/mail";
import Lock from "lucide-svelte/icons/lock";
import UserIcon from "lucide-svelte/icons/user";
  import Loader2 from "lucide-svelte/icons/loader-2";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import SEO from "$lib/components/SEO.svelte";

  let loading = $state(false);
  let errorMessage = $state("");
  let name = $state("");
  let email = $state("");
  let password = $state("");
  let password_confirmation = $state("");

  const redirectTo = $derived(page.url.searchParams.get("redirect") || "/");

  onMount(() => {
    if (authState.isAuthenticated) {
      goto(redirectTo);
    }
  });

  async function handleRegister(e: Event) {
    e.preventDefault();
    loading = true;
    errorMessage = "";

    if (password !== password_confirmation) {
      errorMessage = "Passwords do not match.";
      loading = false;
      return;
    }

    try {
      const response = await api.post("/register", {
        name,
        email,
        password,
      });
      const payload = response.data.data;
      if (payload?.token) {
        setAuthToken(payload.token);
        setUser(payload.user);
        goto(redirectTo);
      }
    } catch (error: any) {
      errorMessage =
        error.response?.data?.message ||
        "Registration failed. Please check your data.";
    } finally {
      loading = false;
    }
  }

  async function handleGoogleLogin() {
    try {
      const response = await api.get("/auth/google/login");
      const url = response.data.data?.url || response.data.url;
      if (url) {
        window.location.href = url;
      }
    } catch (error: unknown) {
      errorMessage = getApiErrorMessage(error, "Failed to initiate Google login.");
    }
  }

  async function handleAnilistLogin() {
    try {
      const response = await api.get("/auth/anilist/login");
      const url = response.data.data?.url || response.data.url;
      if (url) {
        window.location.href = url;
      }
    } catch (error: unknown) {
      errorMessage = getApiErrorMessage(error, "Failed to initiate AniList login.");
    }
  }
</script>

<SEO
  title="Create Account"
  description="Join AniRank to track your favorite anime opening and ending themes, rate songs, and share your lists."
/>

<div class="min-h-[80vh] flex items-center justify-center p-4 py-24">
  <div
    class="w-full max-w-md overflow-hidden rounded-md border border-outline-variant/10 bg-surface-container shadow-2xl transition-all"
  >
    <div class="p-10">
      <div class="mb-10 text-center">
        <div class="flex justify-start mb-8">
          <a
            href="/"
            class="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant/40 hover:text-primary transition-all group/back"
          >
            <ArrowLeft
              size={14}
              class="transition-transform group-hover/back:-translate-x-1"
            />
            Back to home
          </a>
        </div>
        <h2 class="text-3xl font-black text-on-surface tracking-tighter">
          Join Anirank
        </h2>
        <p class="mt-2 text-sm text-on-surface-variant/60 font-medium">
          Create an account to track your favorites and vote.
        </p>
      </div>

      {#if errorMessage}
        <div
          class="mb-8 rounded-md bg-red-500/5 p-4 text-[11px] font-black uppercase tracking-wider text-red-500 border border-red-500/10 text-center leading-tight flex items-center justify-center gap-2"
        >
          <AlertCircle size={16} />
          {errorMessage}
        </div>
      {/if}

      <form onsubmit={handleRegister} class="flex flex-col gap-5">
        <div class="relative group">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-5 text-on-surface-variant/30 group-focus-within:text-primary transition-colors"
          >
            <UserIcon size={18} />
          </div>
          <input
            type="text"
            bind:value={name}
            required
            placeholder="Username"
            class="w-full rounded-sm border border-outline-variant/10 bg-surface-highest/30 py-4 pl-12 pr-5 text-on-surface placeholder-on-surface-variant/30 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
          />
        </div>
        <div class="relative group">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-5 text-on-surface-variant/30 group-focus-within:text-primary transition-colors"
          >
            <Mail size={18} />
          </div>
          <input
            type="email"
            bind:value={email}
            required
            placeholder="Email Address"
            class="w-full rounded-sm border border-outline-variant/10 bg-surface-highest/30 py-4 pl-12 pr-5 text-on-surface placeholder-on-surface-variant/30 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
          />
        </div>
        <div class="relative group">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-5 text-on-surface-variant/30 group-focus-within:text-primary transition-colors"
          >
            <Lock size={18} />
          </div>
          <input
            type="password"
            bind:value={password}
            required
            placeholder="Password"
            class="w-full rounded-sm border border-outline-variant/10 bg-surface-highest/30 py-4 pl-12 pr-5 text-on-surface placeholder-on-surface-variant/30 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
          />
        </div>
        <div class="relative group">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-5 text-on-surface-variant/30 group-focus-within:text-primary transition-colors"
          >
            <Lock size={18} />
          </div>
          <input
            type="password"
            bind:value={password_confirmation}
            required
            placeholder="Confirm Password"
            class="w-full rounded-sm border border-outline-variant/10 bg-surface-highest/30 py-4 pl-12 pr-5 text-on-surface placeholder-on-surface-variant/30 font-medium transition-all focus:bg-surface-highest focus:ring-4 focus:ring-primary/5 focus:outline-none"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          class="mt-2 flex w-full items-center justify-center gap-2 rounded-sm bg-primary py-4 font-black text-sm uppercase tracking-widest text-white transition-all hover:bg-primary/90 hover:scale-[1.02] shadow-xl shadow-primary/20 active:scale-95 disabled:opacity-50"
        >
          {#if loading}
            <Loader2 size={18} class="animate-spin" />
          {/if}
          Create Account
        </button>
      </form>

      <div
        class="mt-10 text-center text-xs font-medium text-on-surface-variant/40"
      >
        Already have an account?
        <a
          href="/login?redirect={encodeURIComponent(redirectTo)}"
          class="font-black text-primary hover:text-primary/80 ml-1 underline underline-offset-4 decoration-primary/20 hover:decoration-primary/50 transition-all"
        >
          Sign in instead
        </a>
      </div>

      <div class="relative mb-8">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-outline-variant/10"></div>
        </div>
        <div class="relative flex justify-center text-[10px] font-black uppercase tracking-widest leading-none">
          <span class="bg-surface-container px-4 text-on-surface-variant/30">
            Or join with
          </span>
        </div>
      </div>

      <div class="flex flex-col sm:flex-row gap-4">
        <button
          type="button"
          onclick={handleAnilistLogin}
          disabled={loading}
          class="flex w-full items-center justify-center gap-2 rounded-sm shadow-sm bg-[#02a9ff] hover:bg-[#0290d9] py-3.5 font-black text-[10px] uppercase tracking-widest text-white transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-50"
        >
          AniList
        </button>
        <button
          type="button"
          onclick={handleGoogleLogin}
          disabled={loading}
          class="flex w-full items-center justify-center gap-2 rounded-sm shadow-sm bg-white hover:bg-white/80 text-black border border-outline-variant py-3.5 font-black text-[10px] uppercase tracking-widest text-black transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-50"
        >
          Google
        </button>
      </div>
    </div>
  </div>
</div>
