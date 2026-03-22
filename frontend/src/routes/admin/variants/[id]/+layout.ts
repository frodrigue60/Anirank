import type { LayoutLoad } from './$types';
import api from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: LayoutLoad = async ({ params }) => {
    try {
        const res = await api.get(`/admin/variants/${params.id}`);
        if (res.status === 200) {
            return {
                variant: res.data.data
            };
        }
    } catch (err: any) {
        console.error("Failed to load variant details", err);
        throw error(err.response?.status || 500, 'Error loading variant');
    }
    throw error(404, 'Not found');
};
