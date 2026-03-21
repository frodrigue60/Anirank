import api from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  try {
    const [userRes, rolesRes, badgesRes] = await Promise.all([
      api.get(`/admin/users/${params.id}`),
      api.get("/admin/roles"),
      api.get("/admin/badges"),
    ]);

    return {
      user: userRes.data.data,
      allRoles: rolesRes.data.data || [],
      allBadges: badgesRes.data.data || [],
    };
  } catch (err: any) {
    console.error("Error loading user for edit", err);
    throw error(err.response?.status || 500, "Failed to load user edit data");
  }
};
