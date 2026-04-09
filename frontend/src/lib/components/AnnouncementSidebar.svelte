<script lang="ts">
  import { onMount } from "svelte";
  import api from "$lib/api";

  interface Announcement {
    id: number;
    title: string;
    content?: string;
    type: string;
    icon?: string;
    url?: string;
    image_url?: string;
    priority: number;
  }

  let announcements = $state<Announcement[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const response = await api.get("/announcements");
      announcements = response.data.data;
    } catch (error) {
      console.error("Failed to load announcements:", error);
    } finally {
      loading = false;
    }
  });

  const typeConfig: Record<
    string,
    { bg: string; text: string; border: string; glow: string }
  > = {
    info: {
      bg: "bg-blue-500/10",
      text: "text-blue-400",
      border: "border-blue-500/10",
      glow: "shadow-sm",
    },
    success: {
      bg: "bg-green-500/10",
      text: "text-green-400",
      border: "border-green-500/10",
      glow: "shadow-sm",
    },
    warning: {
      bg: "bg-yellow-500/10",
      text: "text-yellow-400",
      border: "border-yellow-500/10",
      glow: "shadow-sm",
    },
    danger: {
      bg: "bg-red-500/10",
      text: "text-red-400",
      border: "border-red-500/10",
      glow: "shadow-sm",
    },
    event: {
      bg: "bg-primary/10",
      text: "text-primary",
      border: "border-primary/10",
      glow: "shadow-sm",
    },
  };
</script>

{#if !loading && announcements.length > 0}
  <div class="flex flex-col gap-4">
    {#each announcements as item}
      <svelte:element
        this={item.url ? "a" : "div"}
        href={item.url}
        target={item.url?.startsWith("http") ? "_blank" : undefined}
        title={item.content}
        class="group relative flex min-h-[150px] flex-col justify-end overflow-hidden rounded-md p-5 transition-all hover:scale-[1.02] active:scale-[0.98] border {typeConfig[
          item.type
        ]?.border || 'border-outline-variant/10'} {item.url
          ? 'cursor-pointer'
          : ''}"
      >
        <!-- Background Image -->
        {#if item.image_url}
          <div
            class="absolute inset-0 bg-cover bg-center transition-transform duration-700 group-hover:scale-110"
            style="background-image: url('{item.image_url}');"
          ></div>
          <div
            class="absolute inset-0 bg-linear-to-t from-black/90 via-black/40 to-transparent"
          ></div>
        {:else}
          <div class="absolute inset-0 bg-surface-low"></div>
          <div class="absolute inset-0 {typeConfig[item.type]?.bg || ''}"></div>
        {/if}

        <!-- Content -->
        <div class="relative z-10 flex flex-col gap-1">
          <!-- <div class="flex items-center gap-2 mb-1">
            {#if item.icon}
              <span
                class="material-symbols-outlined text-[18px] {typeConfig[
                  item.type
                ]?.text || 'text-white'}"
              >
                {item.icon}
              </span>
            {/if}
            <span
              class="text-[10px] font-bold uppercase tracking-widest {typeConfig[
                item.type
              ]?.text || 'text-white/50'}"
            >
              {item.type === "event" ? "Special Event" : item.type}
            </span>
          </div> -->

          <h3
            class="line-clamp-2 text-lg font-bold leading-tight {item.image_url
              ? 'text-white'
              : 'text-on-surface'}"
          >
            {item.title}
          </h3>

          {#if item.content}
            <p
              class="line-clamp-2 mt-1 text-xs {item.image_url
                ? 'text-white/70'
                : 'text-on-surface-variant'}"
            >
              {item.content}
            </p>
          {/if}
        </div>

        <!-- Hover Glow -->
        <div
          class="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none {typeConfig[
            item.type
          ]?.glow || ''}"
        ></div>
      </svelte:element>
    {/each}
  </div>
{/if}

<style>
  /* Custom scroll behavior hidden if needed */
</style>
