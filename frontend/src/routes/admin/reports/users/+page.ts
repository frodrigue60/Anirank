import api from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url }) => {
    const status = url.searchParams.get("status") || "pending";
    const statusParam = status === "resolved" ? "fixed" : "pending";
    
    try {
        const response = await api.get(`/admin/users/reports?status=${statusParam}`);
        return {
            reports: response.data.data
        };
    } catch (error) {
        console.error("Error loading user reports:", error);
        return {
            reports: []
        };
    }
};
