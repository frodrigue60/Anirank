import { defineConfig } from "vite";
import laravel from "laravel-vite-plugin";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
    plugins: [
        tailwindcss(),
        laravel({
            input: [
                "resources/js/app.js",
                "resources/js/api/config.js",
                "resources/js/api/endpoints.js",
                "resources/js/api/index.js",
                "resources/js/ajaxSearch.js",
                "resources/js/api_get_video.js",
                "resources/js/filter_animes.js",
                "resources/js/filter_favorites.js",
                "resources/js/filter_artist_themes.js",
                "resources/js/filter_studios.js",
                "resources/js/filter_studio_themes.js",
                "resources/js/filter_studio_animes.js",
                "resources/js/ranking_songs.js",
                "resources/js/seasonal.js",
                "resources/js/api_get_video.js",
                "resources/js/modules/songs/delete_comment.js",
                "resources/js/modules/songs/make_comment.js",
                "resources/js/modules/songs/like.js",
                "resources/js/modules/songs/dislike.js",
                "resources/js/modules/songs/toggle_favorite.js",
                "resources/js/modules/songs/rate.js",
                "resources/js/modules/songs/report.js",
                "resources/js/modules/comments/like.js",
                "resources/js/modules/comments/dislike.js",
                "resources/js/modules/users/upload_avatar.js",
                "resources/js/modules/users/upload_banner.js",
                "resources/js/modules/users/set_score_format.js",
                "resources/js/modules/comments/reply.js",
                "resources/js/modules/songs/get_comments.js",
                "resources/js/filter_user_list",
                "resources/js/make_request.js",
                "resources/js/theme_switch.js",
                "resources/css/app.css",
                /* "resources/sass/app.scss", */
                "resources/css/userProfile.css",
                "resources/css/post.css",
                "resources/css/ranking.css",
                "resources/css/fivestars.css",
                "resources/js/playlists.js",
                "resources/css/fonts.css",
            ],
            refresh: true,
        }),
    ],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./resources/js"),
            "@api": path.resolve(__dirname, "./resources/js/api"),
            "~": path.resolve(__dirname, "./resources"),
            "@modules": path.resolve(__dirname, "./resources/js/modules"),
        },
    },
    build: {
        emptyOutDir: true,
        rollupOptions: {
            output: {
                entryFileNames: `js/[name]-[hash].js`,
                chunkFileNames: `js/[name]-[hash].js`,
                assetFileNames: (assetInfo) => {
                    const extType = assetInfo.name.split(".").at(1);
                    if (/png|jpe?g|svg|gif|tiff|bmp|ico/i.test(extType)) {
                        return `images/[name]-[hash][extname]`;
                    }
                    if (/css/i.test(extType)) {
                        return `css/[name]-[hash][extname]`;
                    }
                    if (/woff|woff2|eot|ttf|otf/i.test(extType)) {
                        return `fonts/[name]-[hash][extname]`;
                    }
                    return `assets/[name]-[hash][extname]`;
                },
            },
        },
        outDir: "public/build",
        assetsDir: "",
        manifest: "manifest.json",
    },
});
