import api from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url }) => {
    const page = url.searchParams.get("page") || "1";
    const type = url.searchParams.get("type") || "";
    
    try {
        const response = await api.get(`/notifications?page=${page}&type=${type}`);
        return {
            notifications: response.data.data,
            total: response.data.total,
            currentPage: response.data.current_page,
            lastPage: response.data.last_page,
            unreadCount: response.data.unread_count,
            filterType: type
        };
    } catch (error) {
        console.error("Error loading notifications:", error);
        return {
            notifications: [],
            total: 0,
            currentPage: 1,
            lastPage: 1,
            unreadCount: 0
        };
    }
};
