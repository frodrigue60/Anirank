import type { PageLoad } from "./$types";
import api from "$lib/api";

export const load: PageLoad = async () => {
  try {
    const response = await api.get("/admin/comments/reports", {
      params: { limit: 20, offset: 0, status: 'pending' },
    });
    return {
      reports: response.data.data,
    };
  } catch (error) {
    console.error("Error loading comment reports:", error);
    return { reports: [] };
  }
};
