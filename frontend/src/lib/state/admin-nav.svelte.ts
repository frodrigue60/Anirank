import { untrack } from 'svelte';

/**
 * Admin Navigation State
 * 
 * Tracks navigation context using sessionStorage so that back buttons
 * and breadcrumbs stay aware of the drill-down path through the entity
 * hierarchy: Anime → Song → Variant → Video.
 * 
 * Uses Svelte 5 runes for reactive state.
 */

const STORAGE_KEY = 'anirank_admin_nav';

export interface NavCrumb {
    label: string;
    href: string;
    type: 'anime' | 'song' | 'variant' | 'video' | 'list';
}

interface NavState {
    /** The breadcrumb stack representing the current drill-down path */
    stack: NavCrumb[];
}

function loadState(): NavState {
    if (typeof window === 'undefined') return { stack: [] };
    try {
        const raw = sessionStorage.getItem(STORAGE_KEY);
        if (raw) return JSON.parse(raw);
    } catch { /* ignore parse errors */ }
    return { stack: [] };
}

function saveState(state: NavState) {
    if (typeof window === 'undefined') return;
    try {
        sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch { /* ignore quota errors */ }
}

function createAdminNav() {
    let state = $state<NavState>(loadState());

    /**
     * Push a new crumb onto the navigation stack.
     * If a crumb with the same type already exists, truncate the stack
     * to that point and replace it (prevents duplicate levels).
     */
    function pushContext(crumb: NavCrumb) {
        untrack(() => {
            const existingIdx = state.stack.findIndex(c => c.type === crumb.type);
            if (existingIdx !== -1) {
                // Truncate stack to this level and replace
                state.stack = [...state.stack.slice(0, existingIdx), crumb];
            } else {
                state.stack = [...state.stack, crumb];
            }
            saveState(state);
        });
    }

    /**
     * Set the entire breadcrumb stack from entity data.
     * Call this from each [id]/+layout.svelte to rebuild
     * the breadcrumb trail based on the entity's parent relationships.
     */
    function setContext(crumbs: NavCrumb[]) {
        untrack(() => {
            state.stack = crumbs;
            saveState(state);
        });
    }

    /**
     * Clear the navigation stack (e.g. when returning to a top-level list).
     */
    function clearContext() {
        untrack(() => {
            state.stack = [];
            saveState(state);
        });
    }

    /**
     * Get the appropriate "back" URL based on the current context stack.
     * Returns the previous item in the stack, or a fallback flat list URL.
     */
    function getBackUrl(currentType: NavCrumb['type'], fallbackUrl: string): string {
        const idx = state.stack.findIndex(c => c.type === currentType);
        if (idx > 0) {
            return state.stack[idx - 1].href;
        }
        return fallbackUrl;
    }

    /**
     * Get the current breadcrumb trail.
     */
    function getBreadcrumbs(): NavCrumb[] {
        return state.stack;
    }

    return {
        pushContext,
        setContext,
        clearContext,
        getBackUrl,
        getBreadcrumbs,
        get stack() { return state.stack; }
    };
}

export const adminNav = createAdminNav();
