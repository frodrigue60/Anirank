import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ url }) => {
    const params = {
        name: url.searchParams.get('name') || '',
        page: url.searchParams.get('page') || 1
    };

    const response = await api.get('/playlists', { params });

    return {
        playlists: response.data.playlists,
        params
    };
};
