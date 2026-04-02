<script lang="ts">
  import { setAuthToken, setUser, authState } from "$lib/state/auth.svelte";
  import { Mail, Lock, Loader2, ArrowLeft } from "lucide-svelte";
  import api from "$lib/api";
  import { getApiErrorMessage } from "$lib/api-errors";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import SEO from "$lib/components/SEO.svelte";

  let loading = $state(false);
  let errorMessage = $state("");
  let email = $state("");
  let password = $state("");

  const redirectTo = $derived(page.url.searchParams.get("redirect") || "/");

  onMount(async () => {
    if (authState.isAuthenticated) {
      goto(redirectTo);
      return;
    }

    const code = page.url.searchParams.get("code");
    const oauthState = page.url.searchParams.get("state");
    const isAnilistLogin = oauthState === "anilistrank_login";
    if (code) {
      loading = true;
      try {
        const response = isAnilistLogin
          ? await api.post("/auth/anilist/login-callback", { code })
          : await api.post("/auth/google/login-callback", { code });
        const payload = response.data.data;
        if (payload?.token) {
          setAuthToken(payload.token);
          setUser(payload.user);

          const url = new URL(window.location.href);
          url.search = "";
          window.history.replaceState({}, "", url.toString());

          goto(redirectTo);
        }
      } catch (error: unknown) {
        errorMessage = getApiErrorMessage(
          error,
          isAnilistLogin ? "AniList login failed." : "Google login failed.",
        );
        const url = new URL(window.location.href);
        url.search = "";
        window.history.replaceState({}, "", url.toString());
      } finally {
        loading = false;
      }
    }
  });

  async function handleLogin(e: Event) {
    e.preventDefault();
    loading = true;
    errorMessage = "";

    try {
      const response = await api.post("/login", { email, password });
      const payload = response.data.data;
      if (payload?.token) {
        setAuthToken(payload.token);
        setUser(payload.user);
        goto(redirectTo);
      }
    } catch (error: any) {
      errorMessage =
        error.response?.data?.message ||
        "Login failed. Please check your credentials.";
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
      errorMessage = getApiErrorMessage(
        error,
        "Failed to initiate Google login.",
      );
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
      errorMessage = getApiErrorMessage(
        error,
        "Failed to initiate AniList login.",
      );
    }
  }
</script>

<SEO 
  title="Login" 
  description="Sign in to your AniRank account to sync your favorite anime theme songs and participate in community rankings." 
/>

<div class="min-h-[80vh] flex items-center justify-center p-4 py-20">
  <div
    class="w-full max-w-md overflow-hidden rounded-2xl border border-white/10 bg-dark-900 shadow-2xl transition-all"
  >
    <div class="p-8">
      <div class="mb-8 text-center">
        <div class="flex justify-start mb-6">
          <a
            href="/"
            class="flex items-center gap-2 text-sm text-white/40 hover:text-white transition-colors"
          >
            <ArrowLeft size={16} />
            Back to home
          </a>
        </div>
        <h2 class="text-3xl font-black text-white">Welcome Back</h2>
        <p class="mt-2 text-sm text-white/50">
          Sign in to sync your favorite anime and lists.
        </p>
      </div>

      {#if errorMessage}
        <div
          class="mb-6 rounded-lg bg-red-500/10 p-3 text-sm text-red-500 border border-red-500/20"
        >
          {errorMessage}
        </div>
      {/if}

      <form onsubmit={handleLogin} class="flex flex-col gap-4">
        <div class="relative">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40"
          >
            <Mail size={18} />
          </div>
          <input
            type="email"
            bind:value={email}
            required
            placeholder="Email Address"
            class="w-full rounded-xl border border-white/5 bg-dark-800 py-3.5 pl-11 pr-4 text-white placeholder-white/40 transition-colors focus:border-primary-500 focus:bg-dark-700 focus:outline-none"
          />
        </div>
        <div class="relative">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40"
          >
            <Lock size={18} />
          </div>
          <input
            type="password"
            bind:value={password}
            required
            placeholder="Password"
            class="w-full rounded-xl border border-white/5 bg-dark-800 py-3.5 pl-11 pr-4 text-white placeholder-white/40 transition-colors focus:border-primary-500 focus:bg-dark-700 focus:outline-none"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          class="mt-2 flex w-full items-center justify-center gap-2 rounded-xl bg-primary-600 py-4 font-bold text-white transition-all hover:bg-primary-500 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50"
        >
          {#if loading}
            <Loader2 size={18} class="animate-spin" />
          {/if}
          Sign In
        </button>
      </form>
      <div class="mt-4 mb-4 text-center text-sm text-white/40">
        Don't have an account?
        <a
          href="/register?redirect={encodeURIComponent(redirectTo)}"
          class="font-bold text-primary-500 hover:text-primary-400 ml-1"
        >
          Create one now
        </a>
      </div>

      <div class="flex flex-col sm:flex-row gap-4">
        <button
          type="button"
          onclick={handleAnilistLogin}
          disabled={loading}
          class="mt-2 flex w-full items-center justify-center gap-2 rounded-sm bg-[#02a9ff] py-3 font-bold text-white transition-all hover:bg-[#0290d9] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50"
        >
          Login with AniList
        </button>
        <button
          type="button"
          onclick={handleGoogleLogin}
          disabled={loading}
          class="mt-2 flex w-full items-center justify-center gap-2 rounded-sm bg-white py-3 font-bold text-black transition-all hover:bg-gray-200 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50"
        >
          Login with Google
        </button>
      </div>
    </div>
  </div>
</div>
