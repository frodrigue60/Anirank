import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params, url }) => {
    try {
        const queryParams = {
            page: url.searchParams.get('page') || 1,
            name: url.searchParams.get('name') || '',
            sort: url.searchParams.get('sort') || 'title'
        };
        const cleanParams = Object.fromEntries(
            Object.entries(queryParams).filter(([_, v]) => v !== '')
        );

        const response = await api.get(`/studios/${params.slug}`, {
            params: cleanParams
        });
        return {
            studio: response.data.studio,
            animes: response.data.animes,
            params: queryParams
        };
    } catch (err: any) {
        console.error('Error loading studio:', err);
        throw error(404, 'Studio not found');
    }
};
