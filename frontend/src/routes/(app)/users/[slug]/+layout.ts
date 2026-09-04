export const ssr = true;

import api from "$lib/api";

export const load = async ({ params }: { params: { slug: string } }) => {
  const slug = params.slug;

  try {
    const userResponse = await api.get(`/users/${slug}`);
    const user = userResponse.data.data;

    if (!user) {
      return {
        profile: null,
        playlistsCount: 0,
        favoritesCount: 0,
      };
    }

    const [playlistsRes, favoritesRes] = await Promise.all([
      api
        .get(`/users/${slug}/playlists`)
        .then((res) => res.data)
        .catch(() => null),
      api
        .post(`/users/favorites/themes`, {
          user_uuid: user.uuid,
          page: 1,
        })
        .then((res) => res.data)
        .catch(() => null),
    ]);

    const playlistsCount =
      playlistsRes?.pagination?.total ??
      (Array.isArray(playlistsRes?.data) ? playlistsRes.data.length : 0);
    const favoritesCount =
      favoritesRes?.pagination?.total ??
      (Array.isArray(favoritesRes?.data) ? favoritesRes.data.length : 0);

    return {
      profile: user,
      playlistsCount,
      favoritesCount,
    };
  } catch (e: any) {
    console.error("Failed to load user profile layout", e);
    return {
      profile: null,
      playlistsCount: 0,
      favoritesCount: 0,
    };
  }
};
