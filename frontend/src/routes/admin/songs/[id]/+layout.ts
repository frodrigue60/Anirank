import type { LayoutLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: LayoutLoad = async ({ params }) => {
    try {
        const res = await api.get(`/admin/songs/${params.id}`);
        if (res.status === 200) {
            return {
                song: res.data.data
            };
        }
    } catch (err: any) {
        console.error("Failed to load song details", err);
        throw error(err.response?.status || 500, 'Error loading song');
    }
    throw error(404, 'Not found');
};
