import api from '$lib/api';

export const load = async ({ params }) => {
    const { slug } = params;

    try {
        const response = await api.get(`/users/${slug}/following`);
        return {
            following: response.data.data,
            pagination: {
                current_page: response.data.current_page,
                last_page: response.data.last_page,
                total: response.data.total
            }
        };
    } catch (e) {
        console.error("Failed to load following users", e);
        return {
            following: [],
            pagination: { current_page: 1, last_page: 1, total: 0 }
        };
    }
};
