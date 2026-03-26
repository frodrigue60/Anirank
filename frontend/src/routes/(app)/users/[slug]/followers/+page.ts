import api from '$lib/api';

export const load = async ({ params }) => {
    const { slug } = params;

    try {
        const response = await api.get(`/users/${slug}/followers`);
        return {
            followers: response.data.data,
            pagination: response.data.pagination
        };
    } catch (e) {
        console.error("Failed to load followers", e);
        return {
            followers: [],
            pagination: { current_page: 1, last_page: 1, total: 0 }
        };
    }
};
