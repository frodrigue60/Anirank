import api from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async () => {
  try {
    const [rolesRes, badgesRes] = await Promise.all([
      api.get("/admin/roles"),
      api.get("/admin/badges"),
    ]);

    return {
      allRoles: rolesRes.data.data || [],
      allBadges: badgesRes.data.data || [],
    };
  } catch (err: any) {
    console.error("Error loading create user data", err);
    throw error(err.response?.status || 500, "Failed to load creation data");
  }
};
