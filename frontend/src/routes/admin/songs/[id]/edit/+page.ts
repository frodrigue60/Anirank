import type { PageLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
    try {
        const res = await api.get(`/admin/songs/${params.id}`);
        if (res.status === 200) {
            return {
                song: res.data.data
            };
        }
    } catch (err: any) {
        console.error("Failed to load song for edit", err);
        throw error(err.response?.status || 500, 'Error loading song');
    }
    throw error(404, 'Not found');
};
