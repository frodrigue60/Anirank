<script lang="ts">
  type ArtistItem = {
    name?: string;
    slug?: string;
    status?: boolean;
  };

  let {
    artists = [],
    popoverId,
    compactAfter = 2,
    class: className = "",
  }: {
    artists?: ArtistItem[];
    popoverId: string;
    compactAfter?: number;
    class?: string;
  } = $props();

  function displayName(artist: ArtistItem): string {
    if (artist?.status === false) return "N/A";
    return artist?.name || "N/A";
  }

  let visibleArtists = $derived(
    (artists ?? []).filter((artist) => artist != null),
  );

  let usePopover = $derived(visibleArtists.length > compactAfter);

  let compactLabel = $derived.by(() => {
    if (visibleArtists.length === 0) return "N/A";
    const names = visibleArtists.map(displayName);
    if (names.length <= compactAfter) return names.join(", ");
    const shown = names.slice(0, compactAfter).join(", ");
    const remaining = names.length - compactAfter;
    return `${shown} +${remaining} more`;
  });

  let fullLabel = $derived(
    visibleArtists.length === 0
      ? "N/A"
      : visibleArtists.map(displayName).join(", "),
  );

  let triggerEl = $state<HTMLButtonElement | null>(null);
  let popoverEl = $state<HTMLDivElement | null>(null);

  function positionPopover() {
    if (!triggerEl || !popoverEl) return;

    const gap = 6;
    const padding = 8;
    const triggerRect = triggerEl.getBoundingClientRect();

    popoverEl.style.position = "fixed";
    popoverEl.style.inset = "auto";
    popoverEl.style.margin = "0";

    let top = triggerRect.bottom + gap;
    let left = triggerRect.left;

    popoverEl.style.top = `${top}px`;
    popoverEl.style.left = `${left}px`;

    const popoverRect = popoverEl.getBoundingClientRect();

    if (popoverRect.right > window.innerWidth - padding) {
      left = window.innerWidth - popoverRect.width - padding;
    }
    if (left < padding) left = padding;

    if (popoverRect.bottom > window.innerHeight - padding) {
      top = triggerRect.top - popoverRect.height - gap;
    }
    if (top < padding) top = padding;

    popoverEl.style.top = `${top}px`;
    popoverEl.style.left = `${left}px`;
  }

  function handleBeforeToggle(event: ToggleEvent) {
    if (event.newState !== "open" || !popoverEl) return;
    popoverEl.dataset.positioning = "true";
  }

  function handlePopoverToggle(event: ToggleEvent) {
    if (!popoverEl) return;

    if (event.newState !== "open") {
      delete popoverEl.dataset.positioning;
      return;
    }

    positionPopover();
    requestAnimationFrame(() => {
      if (!popoverEl?.matches(":popover-open")) return;
      positionPopover();
      delete popoverEl.dataset.positioning;
    });
  }
</script>

{#if visibleArtists.length === 0}
  <span class="text-on-surface-variant/80 {className}">N/A</span>
{:else if !usePopover}
  <span class="text-on-surface-variant/80 text-xs leading-snug {className}">
    {#each visibleArtists as artist, index}
      {#if index > 0}<span>, </span>{/if}
      {#if artist.status === false}
        <span>N/A</span>
      {:else if artist.slug}
        <a
          href="/artists/{artist.slug}"
          class="hover:text-primary transition-colors"
          title="View artist profile: {artist.name}"
        >
          {artist.name}
        </a>
      {:else}
        <span>{artist.name || "N/A"}</span>
      {/if}
    {/each}
  </span>
{:else}
  <span class="block min-w-0 {className}">
    <button
      type="button"
      bind:this={triggerEl}
      popovertarget={popoverId}
      class="max-w-full truncate text-left text-xs text-on-surface-variant/80 leading-snug hover:text-primary transition-colors cursor-pointer underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      aria-label="Show all artists: {fullLabel}"
    >
      {compactLabel}
    </button>

    <div
      popover="auto"
      id={popoverId}
      bind:this={popoverEl}
      onbeforetoggle={handleBeforeToggle}
      ontoggle={handlePopoverToggle}
      class="artist-list-popover m-0 w-72 max-w-[calc(100vw-2rem)] rounded-sm border border-outline-variant bg-surface-highest p-3 text-on-surface shadow-none"
    >
      <p
        class="text-[10px] font-bold uppercase tracking-widest text-on-surface-variant/80 mb-2"
      >
        Artists ({visibleArtists.length})
      </p>
      <ul class="flex flex-col gap-1.5 max-h-56 overflow-y-auto">
        {#each visibleArtists as artist}
          <li class="text-xs leading-snug">
            {#if artist.status === false}
              <span class="text-on-surface-variant/70">N/A</span>
            {:else if artist.slug}
              <a
                href="/artists/{artist.slug}"
                class="font-medium text-on-surface hover:text-primary transition-colors"
                title="View artist profile: {artist.name}"
              >
                {artist.name}
              </a>
            {:else}
              <span class="text-on-surface-variant/80">{artist.name || "N/A"}</span>
            {/if}
          </li>
        {/each}
      </ul>
    </div>
  </span>
{/if}

<style>
  :global(.artist-list-popover[data-positioning="true"]) {
    opacity: 0;
    pointer-events: none;
  }

  :global(.artist-list-popover:popover-open) {
    position: fixed;
    inset: auto;
    margin: 0;
    transition: opacity 120ms ease;
  }
</style>
