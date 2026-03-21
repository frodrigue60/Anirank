import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
    try {
        const res = await api.get(`/admin/animes/${params.id}`);
        if (res.status === 200) {
            return {
                anime: res.data.data
            };
        }
    } catch (err: any) {
        console.error("Failed to load anime details", err);
        throw error(err.response?.status || 500, 'Error loading anime');
    }
    throw error(404, 'Not found');
};
