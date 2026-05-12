import api from '$lib/api';
import { setConfig } from '$lib/state/config.svelte';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = async () => {
    try {
        const res = await api.get('/admin/init');
        setConfig(res.data);
    } catch (e) {
        console.error('Failed to load admin config', e);
    }
    return {};
};
