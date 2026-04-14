import { readable, writable } from 'svelte/store';
import { vi } from 'vitest';

export const goto = vi.fn();
export const pushState = vi.fn();
export const replaceState = vi.fn();

export const page = {
    subscribe: readable({ 
        url: new URL('http://localhost:5173'), 
        params: {}, 
        status: 200, 
        error: null, 
        data: {}, 
        form: null 
    }).subscribe,
    url: new URL('http://localhost:5173')
};
