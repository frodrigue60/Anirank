import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
    const params = {
        name: url.searchParams.get('name') || '',
        sort: url.searchParams.get('sort') || 'name_asc',
        page: url.searchParams.get('page') || 1
    };

    const response = await api.get('/studios', { params });

    return {
        studios: response.data.studios,
        params
    };
};
