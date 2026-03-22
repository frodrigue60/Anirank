<script lang="ts">
  import NavbarMaster from "$lib/components/NavbarMaster.svelte";
  import FooterMaster from "$lib/components/FooterMaster.svelte";

  let { children } = $props();
</script>

<div
  class="flex min-h-screen flex-col bg-dark-900 font-sans text-white/80 selection:bg-primary-500/30 selection:text-white"
>
  <!-- Navbar -->
  <NavbarMaster />
  <main class="flex-1">
    {@render children()}
  </main>
  <FooterMaster />
</div>
