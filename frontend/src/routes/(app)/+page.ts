import api from '$lib/api';

export const load = async () => {
    try {
        const response = await api.get('/home');
        return {
            homeData: response.data.data
        };
    } catch (e) {
        console.error("Failed to load home page data", e);
        return {
            homeData: null
        };
    }
};
