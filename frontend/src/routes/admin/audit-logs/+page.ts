import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
    const page = url.searchParams.get('page') || '1';
    const limit = url.searchParams.get('limit') || '20';
    const event = url.searchParams.get('event') || '';
    const resource = url.searchParams.get('resource') || url.searchParams.get('auditable_type') || '';
    const user = url.searchParams.get('user') || url.searchParams.get('user_id') || '';

    let query = `page=${page}&limit=${limit}`;
    if (event) query += `&event=${event}`;
    if (resource) query += `&resource=${resource}`;
    if (user) query += `&user=${user}`;

    try {
        const res = await api.get(`/admin/audit-logs?${query}`);
        const data = res.data;

        return {
            logs: data.data || [],
            pagination: data.pagination || {
                total: 0,
                current_page: 1,
                per_page: 20,
                last_page: 0
            }
        };
    } catch (err: any) {
        console.error("Error loading audit logs:", err);
        return {
            logs: [],
            pagination: {
                total: 0,
                current_page: 1,
                per_page: 20,
                last_page: 0
            },
            error: err.response?.data?.message || "Error al cargar los logs de auditoría"
        };
    }
};
