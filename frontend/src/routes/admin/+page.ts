import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async () => {
    try {
        const res = await api.get('/admin/dashboard');
        
        return {
            stats: res.data.stats,
            metrics: res.data.metrics,
            error: null
        };
    } catch (err: any) {
        console.error('Dashboard load error:', err);
        return {
            stats: null,
            metrics: [],
            error: err.response?.data?.message || 'Failed to load dashboard data'
        };
    }
};
