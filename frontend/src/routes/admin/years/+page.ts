import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        // Use the admin taxonomies endpoint to get IDs and full data
        const res = await api.get('/admin/taxonomies/years');
        return {
            years: res.data?.data || []
        };
    } catch (error) {
        console.error("Error loading admin years:", error);
        return { years: [] };
    }
};
