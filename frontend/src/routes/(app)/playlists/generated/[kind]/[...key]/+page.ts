import api from "$lib/api";
import { generatedPlaylistApiPath } from "$lib/generated-playlists";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const ssr = false;

export const load: PageLoad = async ({ params }) => {
  const apiPath = generatedPlaylistApiPath(params.kind, params.key);
  if (!apiPath) {
    throw error(404, "Generated playlist not found");
  }

  try {
    const response = await api.get(apiPath, { params: { page: 1, limit: 100 } });
    return {
      playlist: {
        ...response.data.playlist,
        songs: response.data.data || [],
      },
      pagination: response.data.pagination,
      generatedApiPath: apiPath,
    };
  } catch (e: any) {
    throw error(
      e.response?.status || 500,
      e.response?.data?.message || "Failed to load generated playlist",
    );
  }
};
