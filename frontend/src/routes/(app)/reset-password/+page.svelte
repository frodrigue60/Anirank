<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import Lock from "lucide-svelte/icons/lock";
  import Loader2 from "lucide-svelte/icons/loader-2";
  import ShieldCheck from "lucide-svelte/icons/shield-check";
  import Eye from "lucide-svelte/icons/eye";
  import EyeOff from "lucide-svelte/icons/eye-off";

  let token = $state("");
  let password = $state("");
  let confirmPassword = $state("");
  let loading = $state(false);
  let success = $state(false);
  let showPassword = $state(false);

  onMount(() => {
    token = page.url.searchParams.get("token") || "";
    if (!token) {
      toastState.addToast("Invalid or missing reset token.", "error");
      goto("/forgot-password");
    }
  });

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    
    if (password !== confirmPassword) {
      toastState.addToast("Passwords do not match.", "error");
      return;
    }

    if (password.length < 8) {
      toastState.addToast("Password must be at least 8 characters.", "error");
      return;
    }

    loading = true;
    try {
      await api.post("/reset-password", { token, password });
      success = true;
      toastState.addToast("Password reset successfully!", "success");
      
      setTimeout(() => {
        goto("/login");
      }, 3000);
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to reset password.", "error");
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Reset Password - AniRank</title>
</svelte:head>

<div class="min-h-[80vh] flex items-center justify-center p-4">
  <div class="w-full max-w-md bg-surface-container border border-white/5 rounded-md p-8 shadow-2xl transition-all duration-500 animate-in fade-in zoom-in-95">
    
    {#if !success}
      <div class="mb-8">
        <h1 class="text-3xl font-black tracking-tighter text-on-surface mb-2">Set New Password</h1>
        <p class="text-on-surface-variant text-sm">Please choose a strong password that you don't use elsewhere.</p>
      </div>

      <form onsubmit={handleSubmit} class="space-y-6">
        <!-- New Password -->
        <div class="space-y-2">
          <label for="password" class="text-xs font-bold uppercase tracking-widest text-on-surface-variant ml-1">New Password</label>
          <div class="relative group">
            <div class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors">
              <Lock size={20} />
            </div>
            <input
              id="password"
              type={showPassword ? "text" : "password"}
              bind:value={password}
              required
              minlength="8"
              placeholder="••••••••"
              class="w-full bg-surface-highest border border-white/5 rounded-sm py-4 pl-12 pr-12 text-on-surface focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
            />
            <button
              type="button"
              onclick={() => showPassword = !showPassword}
              class="absolute right-4 top-1/2 -translate-y-1/2 text-on-surface-variant hover:text-on-surface transition-colors"
            >
              {#if showPassword}
                <EyeOff size={20} />
              {:else}
                <Eye size={20} />
              {/if}
            </button>
          </div>
        </div>

        <!-- Confirm Password -->
        <div class="space-y-2">
          <label for="confirm" class="text-xs font-bold uppercase tracking-widest text-on-surface-variant ml-1">Confirm Password</label>
          <div class="relative group">
            <div class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors">
              <Lock size={20} />
            </div>
            <input
              id="confirm"
              type={showPassword ? "text" : "password"}
              bind:value={confirmPassword}
              required
              placeholder="••••••••"
              class="w-full bg-surface-highest border border-white/5 rounded-sm py-4 pl-12 pr-4 text-on-surface focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
            />
          </div>
        </div>

        <button
          type="submit"
          disabled={loading}
          class="w-full py-4 bg-primary text-on-primary rounded-sm font-bold uppercase tracking-widest text-xs hover:scale-[1.02] active:scale-95 disabled:opacity-50 disabled:scale-100 transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2"
        >
          {#if loading}
            <Loader2 size={18} class="animate-spin" />
            Resetting Password...
          {:else}
            Update Password
          {/if}
        </button>
      </form>
    {:else}
      <div class="text-center py-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
        <div class="w-20 h-20 rounded-full bg-green-500/10 flex items-center justify-center border border-green-500/20 mx-auto mb-6">
          <ShieldCheck class="w-10 h-10 text-green-500" />
        </div>
        <h2 class="text-2xl font-black tracking-tight text-on-surface mb-3">Password Updated</h2>
        <p class="text-on-surface-variant text-sm mb-8">
          Your password has been reset successfully. You will be redirected to the login page shortly.
        </p>
        <button 
          onclick={() => goto("/login")}
          class="w-full py-3 bg-surface-highest text-on-surface rounded-sm font-bold uppercase tracking-widest text-xs hover:bg-white/10 transition-all border border-white/5"
        >
          Go to Login
        </button>
      </div>
    {/if}
  </div>
</div>
