import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
    const params = {
        name: url.searchParams.get('name') || '',
        year: url.searchParams.get('year') || '',
        season: url.searchParams.get('season') || '',
        genre: url.searchParams.get('genre') || '',
        type: url.searchParams.get('type') || '',
        format: url.searchParams.get('format') || '',
        sort: url.searchParams.get('sort') || '',
        page: url.searchParams.get('page') || 1
    };

    const cleanParams = Object.fromEntries(
        Object.entries(params).filter(([_, v]) => v !== '')
    );

    try {
        const response = await api.get('/songs', { params: cleanParams });
        return {
            songs: response.data,
            params
        };
    } catch (e) {
        console.error("Failed to load themes", e);
        return {
            songs: null,
            params
        };
    }
};
