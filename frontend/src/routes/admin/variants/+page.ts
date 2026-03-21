import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
    const page = Number(url.searchParams.get('page')) || 1;
    const search = url.searchParams.get('search') || '';
    const anime = url.searchParams.get('anime') || url.searchParams.get('anime_id') || '';
    const status = url.searchParams.get('status') || '';
    
    try {
        let query = `/admin/variants?page=${page}&search=${search}`;
        if (anime) query += `&anime=${anime}`;
        if (status !== '') query += `&status=${status}`;

        const res = await api.get(query);
        if (res.status === 200) {
            return {
                data: res.data.data,
                meta: {
                    ...res.data.meta,
                    anime: anime,
                    status: status
                }
            };
        }
    } catch (err) {
        console.error("Failed to load variants", err);
    }
    
    return {
        data: [],
        meta: { current_page: 1, total_pages: 1, total_items: 0, search: '', anime: '', status: '' }
    };
};
