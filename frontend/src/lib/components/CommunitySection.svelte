<script lang="ts">
  import { onMount } from "svelte";
  import { getActivePartners } from "$lib/api";
  import Globe from "lucide-svelte/icons/globe";
  import OptimizedImage from "./OptimizedImage.svelte";
  import type { ImageSource } from "$lib/types/media";

  interface Partner {
    uuid: string;
    name: string;
    url: string;
    banner_url?: string;
    banner_sources?: ImageSource[];
    description?: string;
    type?: string;
  }

  let partners = $state<Partner[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      partners = await getActivePartners();
    } catch (e) {
      console.error("Error loading partners:", e);
    } finally {
      loading = false;
    }
  });
</script>

{#if !loading && partners.length > 0}
  <section class="rounded-md bg-surface-container p-2 space-y-4 shadow-sm border border-gray-500/5">
    <div class="flex items-center justify-center border-b border-gray-500/10">
      <h3 class="text-md font-bold text-on-surface pb-1">
        Partners & Communities
      </h3>
    </div>
    
    <div class="grid grid-cols-2 gap-2">
      {#each partners.slice(0, 6) as partner}
        <a 
          href={partner.url} 
          target="_blank" 
          rel="noopener noreferrer"
          class="group block aspect-video w-full overflow-hidden rounded-md bg-surface-highest transition-all hover:scale-[1.02] active:scale-[0.98] border border-outline-variant/10"
          title={partner.name}
        >
          <div class="h-full w-full relative">
            {#if partner.banner_url}
              <OptimizedImage 
                src={partner.banner_url} 
                sources={partner.banner_sources}
                alt={partner.name} 
                class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" 
                sizes="(max-width: 640px) 50vw, 200px"
              />
            {:else}
              <div class="flex h-full w-full flex-col items-center justify-center text-on-surface-variant/20 p-4">
                <Globe size={32} />
                <span class="mt-2 text-sm font-bold text-on-surface/40 uppercase tracking-wider">{partner.name}</span>
              </div>
            {/if}
            
            <!-- Hover Overlay -->
            <div class="absolute inset-0 bg-black/80 flex flex-col items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300 p-2 text-center backdrop-blur-[2px]">
              <h4 class="text-white font-bold text-[10px] uppercase tracking-tight line-clamp-2">{partner.name}</h4>
              {#if partner.type}
                <span class="text-[8px] uppercase font-black tracking-widest text-primary mt-1">
                  {partner.type}
                </span>
              {/if}
            </div>
          </div>
        </a>
      {/each}
    </div>
    
    <div class="flex flex-col gap-2">
      <p class="text-[10px] text-center text-on-surface-variant/40 pt-1 italic">
        Want to be our partner? Contact us.
      </p>
      <a 
        href="/community" 
        class="text-[10px] text-center text-primary font-bold uppercase tracking-widest hover:underline"
      >
        View All Communities
      </a>
    </div>
  </section>
{/if}
