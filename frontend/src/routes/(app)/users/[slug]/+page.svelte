<script lang="ts">
  import SongCard from "$lib/components/SongCard.svelte";
  import SEO from "$lib/components/SEO.svelte";
  import { page } from "$app/state";
  const PUBLIC_API_URL =
    import.meta.env.VITE_API_URL || "http://localhost:8080/api";

  let { data } = $props();

  const accentColor = $derived(data.profile?.profile_color || "#3db4f2");

  function renderMarkdown(text: string) {
    if (!text) return "";
    // Very simple markdown replacement
    return text
      .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.*?)\*/g, "<em>$1</em>")
      .replace(
        /\[(.*?)\]\((.*?)\)/g,
        '<a href="$2" class="text-primary hover:underline" target="_blank">$1</a>',
      )
      .replace(/\n/g, "<br>");
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
      class="bg-surface-dark border border-white/5 rounded-3xl p-8 shadow-2xl relative overflow-hidden group"
    >
      <div
        class="absolute top-0 left-0 w-1 h-full"
        style="background-color: {accentColor}"
      ></div>
      <h3
        class="text-sm font-black uppercase tracking-widest mb-4 flex items-center gap-2"
        style="color: {accentColor}"
      >
        <span class="material-symbols-outlined text-sm">person</span>
        About Me
      </h3>
      <div class="text-white/70 leading-relaxed text-sm font-medium">
        {@html renderMarkdown(data.profile.about)}
      </div>
    </div>
  </section>
{/if}

{#if data.initialSongs && data.initialSongs.length > 0}
  <!-- Favorite Themes Section (Horizontal Cards) -->
  <section class="mb-12">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-2xl font-bold text-slate-100 flex items-center gap-2">
        <span class="material-symbols-outlined" style="color: {accentColor}"
          >favorite</span
        >
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
      <h3 class="text-2xl font-bold text-slate-100 flex items-center gap-2">
        <span class="material-symbols-outlined" style="color: {accentColor}"
          >artist</span
        >
        Top Artists
      </h3>
      <a
        href={`/users/${data.profile.slug}/artists`}
        class="font-semibold hover:underline text-sm"
        style="color: {accentColor}">View All</a
      >
    </div>
    <div class="flex-wrap gap-8 px-4 flex">
      {#each data.artists.slice(0, 6) as artist}
        <a
          href={`/artists/${artist.slug}`}
          class="flex flex-col items-center group cursor-pointer"
        >
          <div
            class="size-24 sm:size-28 rounded-full border-2 p-1 transition-all mb-3 bg-surface-darker/50"
            style="border-color: {accentColor}4D; --hover-border: {accentColor}"
            role="presentation"
            onmouseenter={(e) =>
              (e.currentTarget.style.borderColor = accentColor)}
            onmouseleave={(e) =>
              (e.currentTarget.style.borderColor = accentColor + "4D")}
          >
            <!-- svelte-ignore a11y_missing_attribute -->
            <img
              class="h-full w-full rounded-full object-cover grayscale group-hover:grayscale-0 transition-all duration-300"
              data-alt="Portrait of a musician"
              src={artist.avatar_url || "/images/placeholders/default.jpg"}
            />
          </div>
          <h5 class="text-slate-100 font-bold text-center text-sm">
            {artist.name || "Anonymous"}
          </h5>
          <p class="text-slate-500 text-xs mt-1">
            {artist.enabled_songs || 0} Themes
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
