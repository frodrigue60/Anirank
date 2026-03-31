import api from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    try {
        const response = await api.get(`/admin/users/reports/${params.id}`);
        return {
            report: response.data.data
        };
    } catch (error) {
        console.error("Error loading user report detail:", error);
        return {
            report: null
        };
    }
};
