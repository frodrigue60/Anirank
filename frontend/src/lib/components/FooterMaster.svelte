<script lang="ts">
  const currentYear = new Date().getFullYear();
</script>

<footer
  class="mt-20 border-t border-white/5 bg-surface-darker/50 backdrop-blur-xl relative overflow-hidden"
>
  <!-- Decorative glow effect -->
  <div
    class="absolute -top-24 -left-24 w-48 h-48 bg-primary/10 rounded-full blur-3xl pointer-events-none"
  ></div>
  <div
    class="absolute -bottom-24 -right-24 w-48 h-48 bg-primary/10 rounded-full blur-3xl pointer-events-none"
  ></div>

  <div class="max-w-[1440px] mx-auto px-6 py-16 relative">
    <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-12">
      <!-- Brand Section -->
      <div class="col-span-1 lg:col-span-2">
        <a class="group mb-6 flex items-center gap-2" href="/" title="AniRank Home" aria-label="Go to AniRank Home">
          <div
            class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/40 group-hover:scale-105 transition-transform"
          >
            <span class="material-symbols-outlined text-[24px]">music_note</span
            >
          </div>
          <span class="text-2xl font-black tracking-tighter text-white"
            >Anirank</span
          >
        </a>
        <p class="max-w-xs text-sm leading-relaxed text-white/40">
          Discover, rate, and explore the best anime openings and endings from
          all your favorite series. Join our community of anime music
          enthusiasts.
        </p>
        <div class="mt-8 flex items-center gap-4">
          <a
            href="https://x.com"
            target="_blank"
            class="flex h-10 w-10 items-center justify-center rounded-full bg-white/5 text-white/40 transition-all hover:bg-primary hover:text-white"
            title="Follow us on X (Twitter)"
            aria-label="Follow us on X (Twitter)"
          >
            <svg class="w-4 h-4 fill-current" viewBox="0 0 24 24"
              ><path
                d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.008 5.925H5.06l12.023 13.845z"
              ></path></svg
            >
          </a>
          <a
            href="https://discord.gg/anirank"
            target="_blank"
            class="flex h-10 w-10 items-center justify-center rounded-full bg-white/5 text-white/40 transition-all hover:bg-primary hover:text-white"
            title="Join our Discord community"
            aria-label="Join our Discord community"
          >
            <svg class="h-5 w-5 fill-current" viewBox="0 0 24 24"
              ><path
                d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189Z"
              ></path></svg
            >
          </a>
        </div>
      </div>

      <!-- Links Sections -->
      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-white"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >info</span
          >
          About
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Learn more about AniRank"
              >About Us</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Read our Cookie Policy"
              >Cookie Policy</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Read our Privacy Policy"
              >Privacy Policy</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Read our Terms and Conditions"
              >Terms and Conditions</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Read our DMCA Disclaimer"
              >DMCA Disclaimer</a
            >
          </li>
        </ul>
      </div>

      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-white"
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
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Browse Anime Series"
              >Series</a
            >
          </li>
          <li>
            <a
              href="/artists"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Browse Artists"
              >Artists</a
            >
          </li>
          <li>
            <a
              href="/songs"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Browse Themes and Songs"
              >Themes</a
            >
          </li>
          <li>
            <a
              href="/studios"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Browse Animation Studios"
              >Studios</a
            >
          </li>
          <li>
            <a
              href="/producers"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Browse Music Producers"
              >Producers</a
            >
          </li>
        </ul>
      </div>

      <div>
        <h4
          class="mb-6 flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-white"
        >
          <span class="material-symbols-outlined text-primary text-[18px]"
            >help</span
          >
          Help
        </h4>
        <ul class="space-y-4">
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Report an error"
              >Report Error</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Report abuse"
              >Abuse</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Contact us"
              >Contact Us</a
            >
          </li>
          <li>
            <a
              href="/"
              class="text-sm font-medium text-white/40 transition-colors hover:text-primary"
              title="Get support"
              >Support</a
            >
          </li>
        </ul>
      </div>
    </div>

    <div
      class="mt-16 flex flex-col items-center justify-between gap-4 border-t border-white/5 pt-8 md:flex-row"
    >
      <p class="text-xs font-medium text-white/20">
        &copy; {currentYear} Anirank. All rights reserved.
      </p>
      <div class="flex items-center gap-6">
        <span class="flex items-center gap-1 text-xs font-medium text-white/20">
          Made with <span
            class="material-symbols-outlined text-primary text-[14px] filled"
            >favorite</span
          > by fans
        </span>
      </div>
    </div>
  </div>
</footer>
