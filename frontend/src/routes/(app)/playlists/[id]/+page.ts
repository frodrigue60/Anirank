import api from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  try {
    const response = await api.get(`/playlists/${params.id}`);
    return {
      playlist: response.data.data,
    };
  } catch (e: any) {
    throw error(e.response?.status || 500, e.response?.data?.message || "Failed to load playlist");
  }
};
