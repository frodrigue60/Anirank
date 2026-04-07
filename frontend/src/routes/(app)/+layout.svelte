<script lang="ts">
  import NavbarMaster from "$lib/components/NavbarMaster.svelte";
  import FooterMaster from "$lib/components/FooterMaster.svelte";

  let { children } = $props();
</script>

<div
  class="flex min-h-screen flex-col bg-surface font-sans selection:bg-primary selection:text-white"
>
  <NavbarMaster />
  <main class="flex-1">
    {@render children()}
  </main>
  <FooterMaster />
</div>
