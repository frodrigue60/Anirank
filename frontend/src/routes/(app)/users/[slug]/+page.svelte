<script lang="ts">
  import SongCard from "$lib/components/SongCard.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import { page } from "$app/state";
  import snarkdown from "snarkdown";
  import ArtistAvatarCard from "$lib/components/ArtistAvatarCard.svelte";
  import Heart from "lucide-svelte/icons/heart";
  import Mic2 from "lucide-svelte/icons/mic-2";
  import History from "lucide-svelte/icons/history";
  const PUBLIC_API_URL =
    import.meta.env.VITE_API_URL || "http://localhost:8080/api";

  let { data } = $props();

  const accentColor = $derived(data.profile?.profile_color || "#3db4f2");

  function renderMarkdown(text: string | null | undefined) {
    if (!text) return "";
    return snarkdown(text);
  }
</script>

<SEO
  title={`${data.profile.name}'s Profile`}
  description={`Check out ${data.profile.name}'s anime theme song favorites and stats on AniRank.`}
  image={`${PUBLIC_API_URL}/og/user/${data.profile.slug}`}
  type="profile"
/>

<!-- Bio / About Section -->
{#if data.profile.about}
  <section class="mb-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
    <div
      class="bg-surface-container border border-on-surface-variant/10 rounded-md p-8 shadow-2xl relative overflow-hidden group"
    >
      <div
        class="absolute top-0 left-0 w-1 h-full"
        style="background-color: {accentColor}"
      ></div>
      <h3
        class="text-sm font-black uppercase tracking-widest mb-4 flex items-center gap-2"
        style="color: {accentColor}"
      >
        About Me
      </h3>
      <div
        class="text-on-surface-variant leading-relaxed text-sm font-medium markdown-content"
      >
        {@html renderMarkdown(data.profile.about)}
      </div>
    </div>
  </section>
{/if}

{#if data.initialSongs && data.initialSongs.length > 0}
  <section class="mb-12">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-2xl font-bold text-on-surface flex items-center gap-2">
        <Heart size={20} style="color: {accentColor}" />

        Favorite Themes
      </h3>
      <a
        href={`/users/${data.profile.slug}/favorites`}
        class="font-semibold hover:underline text-sm"
        style="color: {accentColor}">View All</a
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
      <h3 class="text-2xl font-bold text-on-surface flex items-center gap-2">
        <Mic2 size={20} style="color: {accentColor}" />

        Top Artists
      </h3>
      <a
        href={`/users/${data.profile.slug}/artists`}
        class="font-semibold hover:underline text-sm"
        style="color: {accentColor}">View All</a
      >
    </div>
    <div
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-8 mb-12"
    >
      {#each data.artists.slice(0, 6) as artist}
        <ArtistAvatarCard {artist} />
      {/each}
    </div>
  </section>
{/if}

{#if (!data.initialSongs || data.initialSongs.length === 0) && (!data.artists || data.artists.length === 0)}
  <div
    class="py-20 flex flex-col items-center justify-center text-center space-y-4 rounded-md border border-on-surface-variant/10 bg-surface-container"
  >
    <div
      class="w-20 h-20 rounded-full bg-surface-highest flex items-center justify-center text-on-surface-variant/40"
    >
      <History size={40} />
    </div>
    <div>
      <h3 class="text-xl font-bold text-on-surface uppercase italic">
        Quiet Profile
      </h3>
      <p class="text-sm text-on-surface-variant max-w-xs mx-auto mt-2">
        This user hasn't added any favorites or recent activity yet.
      </p>
    </div>
  </div>
{/if}
