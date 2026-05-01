import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/svelte';
import { afterEach, beforeAll, afterAll, vi } from 'vitest';
import { setupServer } from 'msw/node';

// Setup MSW
export const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterAll(() => server.close());

afterEach(() => {
    cleanup();
    server.resetHandlers();
    vi.clearAllMocks();
    localStorage.clear();
});

// Mock LocalStorage
const localStorageMock = (() => {
    let store: Record<string, string> = {};
    return {
        getItem: (key: string) => store[key] || null,
        setItem: (key: string, value: string) => { store[key] = value.toString(); },
        clear: () => { store = {}; },
        removeItem: (key: string) => { delete store[key]; }
    };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock SvelteKit $app/navigation
vi.mock('$app/navigation', () => ({
    goto: vi.fn(),
    invalidate: vi.fn(),
    prefetch: vi.fn(),
    prefetchRoutes: vi.fn()
}));

// Mock SvelteKit $app/state
vi.mock('$app/state', () => ({
    page: {
        url: new URL('http://localhost:5173'),
        params: {},
        data: {}
    }
}));

// Mock SvelteKit $app/stores
vi.mock('$app/stores', () => {
    const { readable } = require('svelte/store');
    return {
        page: readable({
            url: new URL('http://localhost:5173'),
            params: {},
            data: {}
        }),
        navigating: readable(null),
        updated: readable(false)
    };
});

// Mock Web Animations API (JSDOM lacks this, needed for Svelte 5 transitions)
if (typeof window !== 'undefined') {
    if (!HTMLDivElement.prototype.animate) {
        const animateMock = vi.fn().mockReturnValue({
            finished: Promise.resolve(),
            cancel: vi.fn(),
            pause: vi.fn(),
            play: vi.fn(),
            reverse: vi.fn(),
            onfinish: null,
            oncancel: null,
        });
        HTMLDivElement.prototype.animate = animateMock;
        HTMLElement.prototype.animate = animateMock;
        (window as any).Element.prototype.animate = animateMock;
    }
}
