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

        const response = await api.get(`/studios/${params.slug}`, {
            params: cleanParams
        });
        return {
            studio: response.data.studio,
            animes: response.data.data,
            params: queryParams
        };
    } catch (err: any) {
        logLoadError('(app)/studios/[slug]', err);
        const status = err?.response?.status;
        if (status === 404) {
            throw error(404, 'Studio not found');
        }
        throw error(status && status >= 400 ? status : 500, 'Failed to load studio');
    }
};
