import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const res = await api.get('/init');
        return {
            seasons: res.data?.data?.seasons || []
        };
    } catch (error) {
        console.error("Error loading admin seasons:", error);
        return { seasons: [] };
    }
};
