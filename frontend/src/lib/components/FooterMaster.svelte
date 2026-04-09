<script lang="ts">
  import { themeState } from "$lib/state/theme.svelte";
  const currentYear = new Date().getFullYear();
</script>

<footer
  class="relative mt-20 overflow-hidden border-t border-outline-variant/10 bg-surface-low/50"
>
  <div class="max-w-[1440px] mx-auto px-6 py-6 relative space-y-6">
    <!-- Brand Section -->
    <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-12">
      <div class="col-span-1 lg:col-span-2">
        <a
          class="group mb-6 flex items-center gap-2"
          href="/"
          title="AniRank Home"
          aria-label="Go to AniRank Home"
        >
          <span class="text-2xl font-black tracking-tighter text-on-surface"
            >AniRank</span
          >
        </a>
        <p class="max-w-xs text-sm leading-relaxed text-on-surface-variant">
          Discover, rate, and explore the best anime openings and endings from
          all your favorite series. Join our community of anime music
          enthusiasts.
        </p>

        <!-- Site Stats Snippet -->
        <div class="mt-6 flex flex-wrap gap-4 opacity-50">
          <div class="flex flex-col">
            <span class="text-xs font-bold text-on-surface">45K+</span>
            <span class="text-[10px] uppercase tracking-wider">Themes</span>
          </div>
          <div class="flex flex-col">
            <span class="text-xs font-bold text-on-surface">8K+</span>
            <span class="text-[10px] uppercase tracking-wider">Series</span>
          </div>
          <div class="flex flex-col">
            <span class="text-xs font-bold text-on-surface">12K+</span>
            <span class="text-[10px] uppercase tracking-wider">Users</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Links Sections Grid -->
    <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-12">
      <!-- Explore Section -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-on-surface"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >explore</span
          >
          Discover
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/animes"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Anime Series</a
            >
          </li>
          <li>
            <a
              href="/songs"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Themes & Songs</a
            >
          </li>
          <li>
            <a
              href="/artists"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Artists</a
            >
          </li>
          <li>
            <a
              href="/songs/ranking"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Global Rankings</a
            >
          </li>
          <li>
            <a
              href="/songs/seasonal"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Current Season</a
            >
          </li>
        </ul>
      </div>

      <!-- Resources Section -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-on-surface"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >info</span
          >
          Resources
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >About AniRank</a
            >
          </li>
          <li>
            <a
              href="/stats"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Site Statistics</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Developer API</a
            >
          </li>
          <li>
            <a
              href="/sitemap.xml"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Sitemap</a
            >
          </li>
        </ul>
      </div>

      <!-- Support Section -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-on-surface"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >help</span
          >
          Support
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Help Center</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >FAQ</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Report an Issue</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Contact Us</a
            >
          </li>
        </ul>
      </div>

      <!-- Legal Section -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-on-surface"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >gavel</span
          >
          Legal
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Privacy Policy</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Terms of Service</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Cookie Policy</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >DMCA Disclaimer</a
            >
          </li>
        </ul>
      </div>

      <!-- Social Section -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-on-surface"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >share</span
          >
          Social
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="https://discord.com/invite/HBAxfsH4WZ"
              target="_blank"
              title="Join our Discord server"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Discord</a
            >
          </li>
          <li>
            <a
              href="https://x.com"
              target="_blank"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Twitter / X</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-on-surface-variant transition-colors hover:text-primary"
              >Instagram</a
            >
          </li>
        </ul>
      </div>
    </div>

    <!-- Theme Switcher Section -->
    <div class="pt-4">
      <div
        class="flex items-center gap-1 bg-surface-highest p-1.5 rounded-md border border-outline-variant/10 w-fit"
      >
        <button
          onclick={() => themeState.set("light")}
          class="px-2 py-1 rounded-sm transition-all {themeState.current ===
          'light'
            ? 'bg-surface shadow-sm text-primary scale-110 border border-outline-variant/10'
            : 'text-on-surface-variant/40 hover:text-on-surface-variant'}"
          title="Light Mode"
        >
          <span class="material-symbols-outlined text-[18px]">light_mode</span>
        </button>
        <button
          onclick={() => themeState.set("dark")}
          class="px-2 py-1 rounded-sm transition-all {themeState.current ===
          'dark'
            ? 'bg-surface shadow-sm text-primary scale-110 border border-outline-variant/10'
            : 'text-on-surface-variant/40 hover:text-on-surface-variant'}"
          title="Dark Mode"
        >
          <span class="material-symbols-outlined text-[18px]">dark_mode</span>
        </button>
        <button
          onclick={() => themeState.set("contrast")}
          class="px-2 py-1 rounded-sm transition-all {themeState.current ===
          'contrast'
            ? 'bg-surface shadow-sm text-primary scale-110 border border-outline-variant/10'
            : 'text-on-surface-variant/40 hover:text-on-surface-variant'}"
          title="High Contrast"
        >
          <span class="material-symbols-outlined text-[18px]">contrast</span>
        </button>
      </div>
    </div>

    <!-- Bottom Section -->
    <div
      class="flex flex-col items-center justify-between gap-4 border-t border-outline-variant/10 pt-8 mt-8 md:flex-row"
    >
      <p class="text-xs font-medium text-on-surface-variant/40">
        &copy; {currentYear} AniRank. All rights reserved.
      </p>
      <div class="flex items-center gap-6">
        <span
          class="flex items-center gap-1 text-xs font-medium text-on-surface-variant/40"
        >
          Made with <span
            class="material-symbols-outlined text-primary text-[14px] filled"
            >favorite</span
          > by fans
        </span>
      </div>
    </div>
  </div>
</footer>
