import { getAuthToken } from '$lib/state/auth.svelte';

const apiBase =
    import.meta.env.PUBLIC_API_URL ||
    import.meta.env.VITE_API_URL ||
    'http://localhost:8080/api';

export const load: PageLoad = async ({ fetch }) => {
    try {
        const token = getAuthToken();
        const headers: Record<string, string> = {};
        
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const res = await fetch(`${apiBase}/admin/dashboard`, {
            headers
        });
        
        if (!res.ok) {
            const errData = await res.json();
            return {
                stats: null,
                metrics: [],
                error: errData.message || 'Failed to load dashboard data'
            };
        }

        const data = await res.json();
        return {
            stats: data.stats,
            metrics: data.metrics,
            error: null
        };
    } catch (err) {
        console.error('Dashboard load error:', err);
        return {
            stats: null,
            metrics: [],
            error: 'Connection error'
        };
    }
};
