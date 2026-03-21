import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params, url }) => {
    try {
        const queryParams = {
            page: url.searchParams.get('page') || 1,
            name: url.searchParams.get('name') || '',
            year_id: url.searchParams.get('year_id') || '',
            season_id: url.searchParams.get('season_id') || '',
            type: url.searchParams.get('type') || '',
            sort: url.searchParams.get('sort') || 'recent'
        };

        const cleanParams = Object.fromEntries(
            Object.entries(queryParams).filter(([_, v]) => v !== '')
        );

        const response = await api.get(`/artists/${params.slug}/songs`, {
            params: cleanParams
        });

        return {
            artist: response.data.artist,
            songs: response.data.songs,
            params: queryParams
        };
    } catch (e) {
        console.error('Failed to load artist songs:', e);
        return {
            artist: null,
            songs: null,
            params: {}
        };
    }
};
