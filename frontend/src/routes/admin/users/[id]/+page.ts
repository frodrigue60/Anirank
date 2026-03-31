import api from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  try {
    const userRes = await api.get(`/admin/users/${params.id}`);

    return {
      user: userRes.data.data,
    };
  } catch (err: any) {
    console.error("Error loading user for detail view", err);
    throw error(err.response?.status || 500, "Failed to load user details");
  }
};
