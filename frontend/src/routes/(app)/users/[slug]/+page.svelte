<script lang="ts">
  import SongCard from "$lib/components/SongCard.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import { page } from "$app/state";
  const PUBLIC_API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

  let { data } = $props();
</script>

<SEO 
  title={`${data.profile.name}'s Profile`}
  description={`Check out ${data.profile.name}'s anime theme song favorites and stats on AniRank.`}
  image={`${PUBLIC_API_URL}/og/user/${data.profile.slug}`}
  type="profile"
/>

{#if data.initialSongs && data.initialSongs.length > 0}
  <!-- Favorite Themes Section (Horizontal Cards) -->
  <section class="mb-12">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-2xl font-bold text-slate-100 flex items-center gap-2">
        <span class="material-symbols-outlined text-primary">favorite</span>
        Favorite Themes
      </h3>
      <a
        href={`/users/${data.profile.slug}/favorites`}
        class="text-primary font-semibold hover:underline text-sm">View All</a
      >
    </div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each data.initialSongs.slice(0, 6) as song}
        <SongCard {song} />
      {/each}
    </div>
  </section>
{/if}

{#if data.artists && data.artists.length > 0}
  <!-- Favorite Artists Section (Circular) -->
  <section class="mb-12">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-2xl font-bold text-slate-100 flex items-center gap-2">
        <span class="material-symbols-outlined text-primary">artist</span>
        Top Artists
      </h3>
      <a
        href={`/users/${data.profile.slug}/artists`}
        class="text-primary font-semibold hover:underline text-sm">View All</a
      >
    </div>
    <div class="flex-wrap gap-8 px-4 flex">
      {#each data.artists.slice(0, 6) as artist}
        <a
          href={`/artists/${artist.slug}`}
          class="flex flex-col items-center group cursor-pointer"
        >
          <div
            class="size-24 sm:size-28 rounded-full border-2 border-primary/30 p-1 group-hover:border-primary transition-all mb-3 bg-surface-darker/50"
          >
            <!-- svelte-ignore a11y_missing_attribute -->
            <img
              class="h-full w-full rounded-full object-cover grayscale group-hover:grayscale-0 transition-all duration-300"
              data-alt="Portrait of a musician"
              src={artist.avatar_url ||
                "https://placehold.co/200x200/2a2136/white?text=Artist"}
            />
          </div>
          <h5 class="text-slate-100 font-bold text-center text-sm">
            {artist.name || "Anonymous"}
          </h5>
          <p class="text-slate-500 text-xs mt-1">
            {artist.songs_count || 0} Themes
          </p>
        </a>
      {/each}
    </div>
  </section>
{/if}

{#if (!data.initialSongs || data.initialSongs.length === 0) && (!data.artists || data.artists.length === 0)}
  <div
    class="py-20 flex flex-col items-center justify-center text-center space-y-4 rounded-4xl border border-white/5 bg-white/2"
  >
    <div
      class="w-20 h-20 rounded-full bg-white/5 flex items-center justify-center text-white/20"
    >
      <span class="material-symbols-outlined text-4xl">history</span>
    </div>
    <div>
      <h3 class="text-xl font-bold text-white uppercase italic">
        Quiet Profile
      </h3>
      <p class="text-sm text-white/40 max-w-xs mx-auto mt-2">
        This user hasn't added any favorites or recent activity yet.
      </p>
    </div>
  </div>
{/if}
