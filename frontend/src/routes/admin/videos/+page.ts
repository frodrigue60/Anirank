import api from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url }) => {
    const page = Number(url.searchParams.get('page')) || 1;
    const search = url.searchParams.get('search') || '';
    const anime = url.searchParams.get('anime') || url.searchParams.get('anime_id') || '';
    const status = url.searchParams.get('status') || '';
    
    try {
        let query = `/admin/videos?page=${page}&search=${search}`;
        if (anime) query += `&anime=${anime}`;
        if (status !== '') query += `&status=${status}`;

        const res = await api.get(query);
        if (res.status === 200) {
            return {
                data: res.data.data,
                pagination: res.data.pagination || { current_page: 1, last_page: 1 },
                filters: {
                    anime,
                    status,
                    search
                }
            };
        }
    } catch (err) {
        console.error("Failed to load videos", err);
    }
    
    return {
        data: [],
        pagination: { current_page: 1, last_page: 1, total: 0 },
        filters: { anime: anime, status: status, search: search }
    };
};
