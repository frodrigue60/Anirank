import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const res = await api.get('/admin/taxonomies/seasons');
        return {
            seasons: res.data?.data || []
        };
    } catch (error) {
        console.error("Error loading admin seasons:", error);
        return { seasons: [] };
    }
};
