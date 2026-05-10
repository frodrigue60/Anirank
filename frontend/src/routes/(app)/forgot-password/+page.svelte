<script lang="ts">
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import Mail from "lucide-svelte/icons/mail";
  import ArrowLeft from "lucide-svelte/icons/arrow-left";
  import Loader2 from "lucide-svelte/icons/loader-2";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";

  let email = $state("");
  let loading = $state(false);
  let success = $state(false);

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!email) return;

    loading = true;
    try {
      await api.post("/forgot-password", { email });
      success = true;
      toastState.addToast("Reset link sent if account exists.", "success");
    } catch (err: any) {
      toastState.addToast(err.response?.data?.message || "Failed to send reset link.", "error");
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Forgot Password - AniRank</title>
</svelte:head>

<div class="min-h-[80vh] flex items-center justify-center p-4">
  <div class="w-full max-w-md bg-surface-container border border-white/5 rounded-md p-8 shadow-2xl relative overflow-hidden transition-all duration-500 animate-in fade-in zoom-in-95">
    
    <!-- Decorative background element -->
    <div class="absolute -top-24 -right-24 w-48 h-48 bg-primary/5 rounded-full blur-3xl"></div>
    
    <div class="relative z-10">
      <button 
        onclick={() => goto("/login")}
        class="flex items-center gap-2 text-on-surface-variant hover:text-primary transition-colors text-xs font-bold uppercase tracking-widest mb-8 group"
      >
        <ArrowLeft size={16} class="group-hover:-translate-x-1 transition-transform" />
        Back to Login
      </button>

      {#if !success}
        <div class="mb-8">
          <h1 class="text-3xl font-black tracking-tighter text-on-surface mb-2">Forgot Password?</h1>
          <p class="text-on-surface-variant text-sm">Enter your email address and we'll send you a link to reset your password.</p>
        </div>

        <form onsubmit={handleSubmit} class="space-y-6">
          <div class="space-y-2">
            <label for="email" class="text-xs font-bold uppercase tracking-widest text-on-surface-variant ml-1">Email Address</label>
            <div class="relative group">
              <div class="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors">
                <Mail size={20} />
              </div>
              <input
                id="email"
                type="email"
                bind:value={email}
                required
                placeholder="name@example.com"
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
              Sending Link...
            {:else}
              Send Reset Link
            {/if}
          </button>
        </form>
      {:else}
        <div class="text-center py-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
          <div class="w-20 h-20 rounded-full bg-green-500/10 flex items-center justify-center border border-green-500/20 mx-auto mb-6">
            <CheckCircle2 class="w-10 h-10 text-green-500" />
          </div>
          <h2 class="text-2xl font-black tracking-tight text-on-surface mb-3">Check your email</h2>
          <p class="text-on-surface-variant text-sm mb-8 leading-relaxed">
            If an account exists for <span class="text-on-surface font-bold">{email}</span>, you will receive a password reset link shortly.
          </p>
          <button 
            onclick={() => success = false}
            class="text-primary text-xs font-bold uppercase tracking-widest hover:underline"
          >
            Didn't get the email? Try again
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
