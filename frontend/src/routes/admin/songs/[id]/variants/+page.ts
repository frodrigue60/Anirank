import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import api from '$lib/api';

export const ssr = false;

export const load: PageLoad = async ({ params }) => {
    const id = params.id;

    try {
        const response = await api.get(`/admin/songs/${id}`);
        return {
            song: response.data.data
        };
    } catch (err: any) {
        console.error('Error loading variants:', err);
        throw error(err.response?.status || 500, err.response?.data?.message || 'Internal Server Error');
    }
};
