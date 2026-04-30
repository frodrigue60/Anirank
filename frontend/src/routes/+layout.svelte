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
  import ToastContainer from "$lib/components/ToastContainer.svelte";
  import { themeState } from "$lib/state/theme.svelte";

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

  // Apply theme class to document element
  $effect(() => {
    themeState.apply();
  });
</script>

<!-- Render Modal Global -->
<ToastContainer />

{@render children()}
