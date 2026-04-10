<script lang="ts">
  import { page } from "$app/state";
  import SEO from "$lib/components/SEO.svelte";
  import { ArrowLeft, ShieldAlert, FileQuestion } from "lucide-svelte";

  let status = $derived(page.status);
  let error = $derived(page.error);

  let title = $derived(status === 404 ? "Page Not Found" : "System Error");
  let description = $derived(
    status === 404 
      ? "It looks like you've wandered off the path. The page you're searching for is unavailable."
      : error?.message || "An unexpected server error occurred. Our team has been notified."
  );
</script>

<SEO {title} {description} />

<div class="min-h-[70vh] flex items-center justify-center p-4 py-24">
  <div class="w-full max-w-lg overflow-hidden flex flex-col items-center text-center">
    <div class="mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-surface-container border border-outline-variant/10 shadow-xl shadow-primary/5">
      {#if status === 404}
        <FileQuestion size={32} class="text-primary opacity-80" />
      {:else}
        <ShieldAlert size={32} class="text-primary opacity-80" />
      {/if}
    </div>
    
    <div class="text-8xl font-black text-primary mb-2 tracking-tighter opacity-10 leading-none">
      {status}
    </div>
    
    <h1 class="text-3xl sm:text-4xl font-black text-on-surface mb-3 tracking-tighter">
      {title}
    </h1>
    
    <p class="text-on-surface-variant/70 font-medium max-w-md mx-auto mb-10 leading-relaxed text-sm">
      {description}
    </p>
    
    <a href="/" class="group flex items-center gap-2 rounded-sm bg-primary px-8 py-3.5 font-black text-xs sm:text-sm uppercase tracking-widest text-white transition-all hover:bg-primary/90 hover:scale-[1.02] shadow-xl shadow-primary/20 active:scale-95">
      <ArrowLeft size={16} class="transition-transform group-hover:-translate-x-1" />
      Return to Home
    </a>
  </div>
</div>
