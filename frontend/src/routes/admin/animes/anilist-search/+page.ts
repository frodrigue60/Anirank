import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ url }) => {
    const q = url.searchParams.get('q') ?? '';
    const format = url.searchParams.get('format') ?? '';

    if (!q.trim()) {
        return { results: [], q, format };
    }

    try {
        const response = await api.get('/admin/animes/anilist-search', { 
            params: { q, format } 
        });
        return {
            results: response.data.data ?? [],
            q,
            format,
        };
    } catch (err: any) {
        console.error('Anilist search error:', err);
        throw error(500, 'Failed to search Anilist');
    }
};
