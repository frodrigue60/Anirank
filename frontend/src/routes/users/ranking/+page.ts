import axios from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url }) => {
  const sort = url.searchParams.get("sort") || "xp";
  const page = url.searchParams.get("page") || "1";

  try {
    const response = await axios.get("/users/ranking", {
      params: { sort, page }
    });
    return {
      ranking: response.data,
      sort
    };
  } catch (error) {
    console.error("Error loading user ranking:", error);
    return {
      ranking: { data: [], total: 0 },
      sort
    };
  }
};
