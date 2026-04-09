import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params, url }) => {
    try {
        const queryParams = {
            page: url.searchParams.get('page') || 1,
            name: url.searchParams.get('name') || '',
            year: url.searchParams.get('year') || '',
            season: url.searchParams.get('season') || '',
            type: url.searchParams.get('type') || '',
            sort: url.searchParams.get('sort') || ''
        };

        const cleanParams = Object.fromEntries(
            Object.entries(queryParams).filter(([_, v]) => v !== '')
        );

        const response = await api.get(`/artists/${params.slug}/songs`, {
            params: cleanParams
        });
        //console.log(response.data.data);
        return {
            artist: response.data.artist,
            songs: response.data.data,
            pagination: response.data.pagination,
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
