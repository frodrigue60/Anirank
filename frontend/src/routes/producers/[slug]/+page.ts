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

        const response = await api.get(`/producers/${params.slug}`, {
            params: cleanParams
        });
        return {
            producer: response.data.producer,
            animes: response.data.animes,
            params: queryParams
        };
    } catch (err: any) {
        console.error('Error loading producer:', err);
        throw error(404, 'Producer not found');
    }
};
