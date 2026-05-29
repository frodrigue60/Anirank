import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';
import { logLoadError } from '$lib/logger';

export const load: PageLoad = async ({ params, url }) => {
    try {
        const queryParams = {
            page: url.searchParams.get('page') || 1,
            name: url.searchParams.get('name') || '',
            sort: url.searchParams.get('sort') || ''
        };
        const cleanParams = Object.fromEntries(
            Object.entries(queryParams).filter(([_, v]) => v !== '')
        );

        const response = await api.get(`/producers/${params.slug}`, {
            params: cleanParams
        });
        return {
            producer: response.data.producer,
            animes: response.data.data,
            params: queryParams
        };
    } catch (err: any) {
        logLoadError('(app)/producers/[slug]', err);
        throw error(404, 'Producer not found');
    }
};
