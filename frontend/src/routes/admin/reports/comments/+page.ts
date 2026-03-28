import type { PageLoad } from "./$types";
import api from "$lib/api";

export const load: PageLoad = async () => {
  try {
    const res = await api.get("/admin/comments/reports");
    return { reports: res.data.data || [] };
  } catch (error) {
    console.error("Error loading admin reports:", error);
    return { reports: [] };
  }
};
