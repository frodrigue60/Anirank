import type { PageLoad } from './$types';
import api from '$lib/api';

export const load: PageLoad = async ({ params }) => {
    try {
        const response = await api.get(`/admin/songs/reports/${params.id}`);
        return {
            report: response.data.data
        };
    } catch (error) {
        console.error('Error loading report:', error);
        return {
            report: null,
            error: 'Failed to load report'
        };
    }
};
