<script lang="ts">
  import NavbarMaster from "$lib/components/NavbarMaster.svelte";
  import FooterMaster from "$lib/components/FooterMaster.svelte";
  import { authState } from "$lib/state/auth.svelte";
  import { notificationState } from "$lib/state/notifications.svelte";

  let { children } = $props();

  $effect(() => {
    // Only connect the SSE tunnel if the user is authenticated
    if (authState.isAuthenticated) {
      notificationState.init();
    } else {
      notificationState.disconnect();
    }
  });
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
