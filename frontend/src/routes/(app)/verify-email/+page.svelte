<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import api from "$lib/api";
  import { toastState } from "$lib/state/toast.svelte";
  import { authState, setUser } from "$lib/state/auth.svelte";
  import MailCheck from "lucide-svelte/icons/mail-check";
  import Loader2 from "lucide-svelte/icons/loader-2";
  import XCircle from "lucide-svelte/icons/x-circle";

  let status = $state<'loading' | 'success' | 'error'>('loading');
  let errorMessage = $state("");
  let processing = false;

  onMount(async () => {
    if (processing) return;
    const token = page.url.searchParams.get("token");

    if (!token) {
      status = 'error';
      errorMessage = "No verification token found in URL.";
      return;
    }

    processing = true;

    try {
      await api.get(`/verify-email?token=${token}`);
      status = 'success';
      toastState.addToast("Email verified successfully!", "success");
      
      // If user is logged in, refresh their profile to update verified status
      if (authState.user) {
        const profileRes = await api.get("/profile");
        if (profileRes.data.data) {
          setUser(profileRes.data.data);
        }
      }

      // Redirect after a short delay
      setTimeout(() => {
        goto("/settings/account");
      }, 3000);
    } catch (err: any) {
      status = 'error';
      errorMessage = err.response?.data?.message || "Failed to verify email. The link may be expired or invalid.";
    }
  });
</script>

<svelte:head>
  <title>Verify Email - AniRank</title>
</svelte:head>

<div class="min-h-[60vh] flex flex-col items-center justify-center p-4">
  <div class="w-full max-w-md bg-surface-container border border-white/5 rounded-md p-8 text-center shadow-xl transition-all duration-500 animate-in fade-in zoom-in-95">
    {#if status === 'loading'}
      <div class="flex flex-col items-center gap-4">
        <Loader2 class="w-12 h-12 text-primary animate-spin" />
        <h1 class="text-2xl font-black tracking-tight text-on-surface">Verifying your email...</h1>
        <p class="text-on-surface-variant text-sm">Please wait while we confirm your account.</p>
      </div>
    {:else}
      <div class="flex flex-col items-center gap-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
        {#if status === 'success'}
          <div class="w-20 h-20 rounded-full bg-green-500/10 flex items-center justify-center border border-green-500/20">
            <MailCheck class="w-10 h-10 text-green-500" />
          </div>
          <div>
            <h1 class="text-3xl font-black tracking-tighter text-on-surface mb-2">Success!</h1>
            <p class="text-on-surface-variant">Your email has been verified. You will be redirected to your account settings shortly.</p>
          </div>
          <button 
            onclick={() => goto("/settings/account")}
            class="w-full py-3 bg-primary text-on-primary rounded-sm font-bold uppercase tracking-widest text-xs hover:scale-[1.02] active:scale-95 transition-all shadow-lg shadow-primary/20"
          >
            Go to Settings Now
          </button>
        {:else}
          <div class="w-20 h-20 rounded-full bg-red-500/10 flex items-center justify-center border border-red-500/20">
            <XCircle class="w-10 h-10 text-red-500" />
          </div>
          <div>
            <h1 class="text-3xl font-black tracking-tighter text-on-surface mb-2">Verification Failed</h1>
            <p class="text-red-400/80 text-sm font-medium">{errorMessage}</p>
          </div>
          <div class="w-full space-y-3">
            <button 
              onclick={() => goto("/settings/account")}
              class="w-full py-3 bg-surface-highest text-on-surface rounded-sm font-bold uppercase tracking-widest text-xs hover:bg-white/10 transition-all border border-white/5"
            >
              Back to Settings
            </button>
            <p class="text-xs text-on-surface-variant">
              Need a new link? You can request one from your account settings.
            </p>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
