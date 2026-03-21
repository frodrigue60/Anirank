import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
    try {
        const response = await api.get(`/animes/${params.slug}`);
        return {
            anime: response.data
        };
    } catch (err: any) {
        console.error('Error loading anime:', err);
        throw error(404, 'Anime not found');
    }
};
