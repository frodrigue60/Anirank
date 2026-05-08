<script lang="ts">
    import { adminNav, type NavCrumb } from "$lib/state/admin-nav.svelte";
    import ChevronRight from "lucide-svelte/icons/chevron-right";
    import Home from "lucide-svelte/icons/home";

    let { crumbs = [] }: { crumbs?: NavCrumb[] } = $props();

    // Use provided crumbs if available, otherwise read from adminNav state
    let displayCrumbs = $derived(crumbs.length > 0 ? crumbs : adminNav.stack);

    // On mobile, only show the last 2 crumbs with an ellipsis
    let collapsedCrumbs = $derived(
        displayCrumbs.length > 3
            ? [displayCrumbs[0], ...displayCrumbs.slice(-2)]
            : displayCrumbs
    );
    let hasCollapsed = $derived(displayCrumbs.length > 3);
</script>

{#if displayCrumbs.length > 0}
    <nav aria-label="Breadcrumb" class="flex items-center gap-1.5 text-sm mb-6 min-w-0 overflow-x-auto scrollbar-none">
        <!-- Admin Home -->
        <a
            href="/admin"
            class="flex items-center gap-1.5 text-on-surface-variant/40 hover:text-on-surface transition-colors shrink-0 p-1 rounded-lg hover:bg-surface-highest"
            title="Admin Dashboard"
        >
            <Home size={14} />
        </a>

        <!-- Desktop: Show all crumbs -->
        <div class="hidden md:flex items-center gap-1.5 min-w-0">
            {#each displayCrumbs as crumb, i}
                <ChevronRight size={12} class="text-on-surface-variant/30 shrink-0" />
                {#if i === displayCrumbs.length - 1}
                    <!-- Current (last) segment — not clickable -->
                    <span
                        class="font-semibold text-on-surface truncate max-w-[200px]"
                        title={crumb.label}
                    >
                        {crumb.label}
                    </span>
                {:else}
                    <a
                        href={crumb.href}
                        class="text-on-surface-variant/70 hover:text-on-surface transition-colors truncate max-w-[180px] hover:underline underline-offset-4"
                        title={crumb.label}
                    >
                        {crumb.label}
                    </a>
                {/if}
            {/each}
        </div>

        <!-- Mobile: Collapsed view -->
        <div class="flex md:hidden items-center gap-1.5 min-w-0">
            {#each collapsedCrumbs as crumb, i}
                <ChevronRight size={12} class="text-on-surface-variant/30 shrink-0" />
                {#if hasCollapsed && i === 0}
                    <!-- First crumb + ellipsis -->
                    <a
                        href={crumb.href}
                        class="text-on-surface-variant/70 hover:text-on-surface transition-colors truncate max-w-[120px]"
                        title={crumb.label}
                    >
                        {crumb.label}
                    </a>
                    <ChevronRight size={12} class="text-on-surface-variant/30 shrink-0" />
                    <span class="text-on-surface-variant/30">…</span>
                {:else if i === collapsedCrumbs.length - 1}
                    <span
                        class="font-semibold text-on-surface truncate max-w-[160px]"
                        title={crumb.label}
                    >
                        {crumb.label}
                    </span>
                {:else}
                    <a
                        href={crumb.href}
                        class="text-on-surface-variant/70 hover:text-on-surface transition-colors truncate max-w-[120px]"
                        title={crumb.label}
                    >
                        {crumb.label}
                    </a>
                {/if}
            {/each}
        </div>
    </nav>
{/if}
