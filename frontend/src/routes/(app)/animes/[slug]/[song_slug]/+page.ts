import type { PageLoad } from "./$types";
import api from "$lib/api";
import { error } from "@sveltejs/kit";

export const load: PageLoad = async ({ params }) => {
    const { slug, song_slug } = params;
    
    try {
        const response = await api.get(`/animes/${slug}/songs/${song_slug}`);
        const result = response.data;

        return {
            song: result.data,
            related: result.related
        };
    } catch (err: any) {
        console.error("Error loading song detail:", err);
        throw error(404, "Song not found");
    }
};
