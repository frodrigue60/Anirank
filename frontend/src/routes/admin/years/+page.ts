import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const res = await api.get('/init');
        return {
            years: res.data?.data?.years || []
        };
    } catch (error) {
        console.error("Error loading admin years:", error);
        return { years: [] };
    }
};
