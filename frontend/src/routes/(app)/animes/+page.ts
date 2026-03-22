import api from '$lib/api';

export const load = async ({ url }) => {
    const params = {
        name: url.searchParams.get('name') || '',
        year_id: url.searchParams.get('year_id') || '',
        season_id: url.searchParams.get('season_id') || '',
        type: url.searchParams.get('type') || '',
        sort: url.searchParams.get('sort') || '',
        page: url.searchParams.get('page') || 1
    };

    const cleanParams = Object.fromEntries(
        Object.entries(params).filter(([_, v]) => v !== '')
    );

    try {
        const response = await api.get('/animes', { params: cleanParams });
        return {
            animes: response.data,
            params
        };
    } catch (e) {
        console.error("Failed to load animes", e);
        return {
            animes: null,
            params
        };
    }
};
