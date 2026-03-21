<script lang="ts">
  import {
    setAuthToken,
    setUser,
    authState,
  } from "$lib/state/auth.svelte";
  import { Mail, Lock, User as UserIcon, Loader2, ArrowLeft } from "lucide-svelte";
  import api from "$lib/api";
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
</script>

<SEO 
  title="Create Account" 
  description="Join AniRank to track your favorite anime opening and ending themes, rate songs, and share your lists." 
/>

<div class="min-h-[80vh] flex items-center justify-center p-4 py-20">
  <div class="w-full max-w-md overflow-hidden rounded-2xl border border-white/10 bg-dark-900 shadow-2xl transition-all">
    <div class="p-8">
      <div class="mb-8 text-center">
        <div class="flex justify-start mb-6">
          <a href="/" class="flex items-center gap-2 text-sm text-white/40 hover:text-white transition-colors">
            <ArrowLeft size={16} />
            Back to home
          </a>
        </div>
        <h2 class="text-3xl font-black text-white">Join Anirank</h2>
        <p class="mt-2 text-sm text-white/50">
          Create an account to track your favorites and vote.
        </p>
      </div>

      {#if errorMessage}
        <div class="mb-6 rounded-lg bg-red-500/10 p-3 text-sm text-red-500 border border-red-500/20">
          {errorMessage}
        </div>
      {/if}

      <form onsubmit={handleRegister} class="flex flex-col gap-4">
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40">
            <UserIcon size={18} />
          </div>
          <input
            type="text"
            bind:value={name}
            required
            placeholder="Username"
            class="w-full rounded-xl border border-white/5 bg-dark-800 py-3.5 pl-11 pr-4 text-white placeholder-white/40 transition-colors focus:border-primary-500 focus:bg-dark-700 focus:outline-none"
          />
        </div>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40">
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
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40">
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
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-white/40">
            <Lock size={18} />
          </div>
          <input
            type="password"
            bind:value={password_confirmation}
            required
            placeholder="Confirm Password"
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
          Create Account
        </button>
      </form>

      <div class="mt-8 text-center text-sm text-white/40">
        Already have an account? 
        <a href="/login?redirect={encodeURIComponent(redirectTo)}" class="font-bold text-primary-500 hover:text-primary-400 ml-1">
          Sign in instead
        </a>
      </div>
    </div>
  </div>
</div>
