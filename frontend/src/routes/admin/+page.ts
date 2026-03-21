import type { PageLoad } from './$types';

const apiBase =
    import.meta.env.PUBLIC_API_URL ||
    import.meta.env.VITE_API_URL ||
    'http://localhost:8080/api';

export const load: PageLoad = async ({ fetch }) => {
    try {
        const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
        const res = await fetch(`${apiBase}/admin/dashboard`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
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
