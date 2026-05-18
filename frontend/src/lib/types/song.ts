import type { ImageSource } from './media';

export interface Song {
    is_reported: boolean | null | undefined;
    id: number;
    name?: string;
    title?: string;
    song_romaji: string;
    song_en: string;
    song_jp: string;
    type: string;
    song_type?: SongType;
    theme_num: string;
    slug: string;
    anime_id: number;
    year_id: number;
    season_id: number;
    anime?: {
        id: number;
        title: string;
        cover_url: string;
        cover_sources?: ImageSource[];
        banner_url: string;
        banner_sources?: ImageSource[];
        slug: string;
    };
    year?: {
        id: string;
        name: string;
        current?: boolean;
    };
    season?: {
        id: string;
        name: string;
        current?: boolean;
    };
    variants?: SongVariant[];
    artists?: Artist[];
    views: number;
    likes_count: number;
    dislikes_count: number;
    average_rating?: number;
    user_rating?: number;
    is_liked?: boolean;
    is_disliked?: boolean;
    is_favorited?: boolean;
    prev_rank?: number;
    prev_main_rank?: number;
    prev_seasonal_rank?: number;
}

export interface Artist {
    id: number;
    name: string;
    slug: string;
    status?: boolean;
    name_jp?: string;
    avatar_url?: string;
    avatar_sources?: ImageSource[];
    banner_url?: string;
    banner_sources?: ImageSource[];
    image_url?: string;
    is_favorited?: boolean;
}

export interface SongVariantVideo {
    type?: string;
    video_url?: string;
    embed_url?: string;
    local_url?: string;
    embed_code?: string;
    video_src?: string;
    is_nc: boolean;
    is_bd: boolean;
    resolution: number;
    is_uncensored: boolean;
    is_subbed: boolean;
    is_lyrics: boolean;
    source: string;
    overlap: string;
}

export interface SongVariant {
    id: string;
    version_number: number;
    song_id: string;
    slug: string;
    video_url?: string;
    embed_url?: string;
    local_url?: string;
    embed_code?: string;
    video_src?: string;
    is_nc?: boolean;
    is_bd?: boolean;
    resolution?: number;
    is_uncensored?: boolean;
    is_subbed?: boolean;
    is_lyrics?: boolean;
    source?: string;
    overlap?: string;
    video?: {
        type: string;
        embed_code?: string;
        video_src?: string;
        embed_url?: string;
        local_url?: string;
        is_nc?: boolean;
        is_bd?: boolean;
        resolution?: number;
        is_uncensored?: boolean;
        is_subbed?: boolean;
        is_lyrics?: boolean;
        source?: string;
        overlap?: string;
    };
    videos?: SongVariantVideo[];
    spoiler: boolean;
    nsfw?: boolean;
    episodes?: string;
    season?: {
        id: string;
        name: string;
        current?: boolean;
    };
    year?: {
        id: string;
        name: string;
        current?: boolean;
    };
}

export interface SongType {
    id: string;
    name: string;
    slug: string;
    description?: string;
}
