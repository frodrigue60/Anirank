import api from '$lib/api';

export const ssr = false;

export const load = async ({ url }) => {
    const type = url.searchParams.get('type') || 'all';
    const page = url.searchParams.get('page') || 1;

    try {
        const response = await api.get(`/songs/ranking/seasonal?type=${type}&page=${page}`);
        return {
            songsData: response.data,
            type
        };
    } catch (e) {
        console.error("Failed to load seasonal ranking data", e);
        return {
            ranking: null,
            type
        };
    }
};
