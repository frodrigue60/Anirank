import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const res = await api.get('/init');
        return {
            genres: res.data?.data?.genres || []
        };
    } catch (error) {
        console.error("Error loading admin genres:", error);
        return { genres: [] };
    }
};
