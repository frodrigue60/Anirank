<script lang="ts">
  import OptimizedImage from "$lib/components/OptimizedImage.svelte";
  import Globe from "lucide-svelte/icons/globe";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import SEO from "$lib/components/SEO.svelte";

  let { data } = $props();
  const partners = $derived(data.partners || []);
</script>

<SEO 
  title="Communities & Partners" 
  description="Discover the communities and partners that make AniRank possible. Explore fan groups, databases, and more."
/>

<main class="max-w-[1440px] mx-auto px-6 py-10 space-y-12">
  <header class="text-center space-y-4 max-w-2xl mx-auto">
    <h1 class="text-4xl md:text-5xl font-black tracking-tight uppercase italic text-on-surface">
      Communities & <span class="text-primary">Partners</span>
    </h1>
    <p class="text-on-surface-variant text-lg leading-relaxed">
      AniRank is supported by an amazing network of anime communities, databases, and creators. Explore our partners below.
    </p>
  </header>

  {#if partners.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
      {#each partners as partner}
        <a 
          href={partner.url} 
          target="_blank" 
          rel="noopener noreferrer"
          class="group bg-surface-container rounded-lg overflow-hidden border border-white/5 flex flex-col h-full hover:border-primary/30 transition-all hover:shadow-2xl hover:shadow-primary/5 hover:-translate-y-1"
        >
          <!-- Banner -->
          <div class="aspect-video w-full overflow-hidden relative">
            {#if partner.banner_url}
              <OptimizedImage 
                src={partner.banner_url} 
                sources={partner.banner_sources}
                alt={partner.name} 
                class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-110" 
                sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 400px"
              />
            {:else}
              <div class="flex h-full w-full flex-col items-center justify-center bg-surface-highest text-on-surface-variant/20 p-8">
                <Globe size={48} />
              </div>
            {/if}
            
            <!-- Type Badge -->
            {#if partner.type}
              <div class="absolute top-4 right-4 bg-black/60 backdrop-blur-md px-3 py-1 rounded-sm border border-white/10">
                <span class="text-[10px] uppercase font-black tracking-widest text-primary">
                  {partner.type}
                </span>
              </div>
            {/if}
          </div>

          <!-- Content -->
          <div class="p-6 flex flex-col flex-1 space-y-4">
            <div class="flex items-start justify-between gap-4">
              <h2 class="text-2xl font-bold text-on-surface group-hover:text-primary transition-colors">
                {partner.name}
              </h2>
              <ExternalLink size={20} class="text-on-surface-variant group-hover:text-primary transition-colors shrink-0 mt-1" />
            </div>

            {#if partner.description}
              <p class="text-on-surface-variant text-sm leading-relaxed line-clamp-4">
                {partner.description}
              </p>
            {/if}
          </div>
        </a>
      {/each}
    </div>
  {:else}
    <div class="text-center py-20 bg-surface-container rounded-lg border border-dashed border-white/10">
       <Globe size={48} class="mx-auto text-on-surface-variant/20 mb-4" />
       <p class="text-on-surface-variant font-bold uppercase tracking-widest">No partners found</p>
    </div>
  {/if}

  <!-- Contact Section -->
  <section class="bg-surface-highest rounded-lg p-12 text-center space-y-6 border border-white/5">
    <div class="max-w-xl mx-auto space-y-4">
      <h3 class="text-2xl font-black uppercase italic text-on-surface">Become a Partner</h3>
      <p class="text-on-surface-variant">
        Are you running an anime community, a database, or a project that would be a great fit for AniRank? We'd love to hear from you.
      </p>
      <div class="pt-4">
        <a 
          href="mailto:partnerships@anirank.net" 
          class="inline-flex items-center gap-2 text-primary font-bold uppercase tracking-widest hover:underline"
        >
          Contact Partnerships Team
        </a>
      </div>
    </div>
  </section>
</main>
