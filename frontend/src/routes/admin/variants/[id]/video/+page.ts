import type { PageLoad } from './$types';
import api from '$lib/api';

export const ssr = false;

export const load: PageLoad = async ({ params }) => {
    const { id } = params;
    
    try {
        const res = await api.get(`/admin/variants/${id}`);
        return {
            variant: res.data.data
        };
    } catch (err) {
        console.error('Failed to load variant:', err);
        return {
            variant: null,
            error: 'Failed to load variant data'
        };
    }
};
