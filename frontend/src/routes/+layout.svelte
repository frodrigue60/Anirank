<script lang="ts">
  import "../app.css";
  import { onMount } from "svelte";
  import {
    authState,
    setUser,
    getAuthToken,
    removeAuthToken,
  } from "$lib/state/auth.svelte";
  import api from "$lib/api";
  import NavbarMaster from "$lib/components/NavbarMaster.svelte";
  import FooterMaster from "$lib/components/FooterMaster.svelte";
  import ToastContainer from "$lib/components/ToastContainer.svelte";

  let { children } = $props();

  onMount(async () => {
    const token = getAuthToken();
    if (!token) {
      setUser(null);
      return;
    }

    try {
      const response = await api.get("/profile");
      setUser(response.data.data);
    } catch (error) {
      console.error("Failed to hydrate user session", error);
      removeAuthToken();
      setUser(null);
    }
  });
</script>

<svelte:head>
  <link href="https://fonts.googleapis.com" rel="preconnect" />
  <link
    crossorigin="anonymous"
    href="https://fonts.gstatic.com"
    rel="preconnect"
  />
  <link
    href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<!-- Render Modal Global -->
<ToastContainer />

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
