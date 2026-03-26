import api from '$lib/api';

export const load = async ({ params }) => {
    const { slug } = params;

    try {
        const response = await api.get(`/users/${slug}/following`);
        return {
            following: response.data.data,
            pagination: response.data.pagination
        };
    } catch (e) {
        console.error("Failed to load following users", e);
        return {
            following: [],
            pagination: { current_page: 1, last_page: 1, total: 0 }
        };
    }
};
