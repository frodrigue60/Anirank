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
        banner_url: string;
        slug: string;
    };
    year?: {
        id: number;
        name: string;
    };
    season?: {
        id: number;
        name: string;
    };
    song_variants?: SongVariant[];
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
}

export interface Artist {
    id: number;
    name: string;
    slug: string;
    status?: boolean;
    name_jp?: string;
    avatar_url?: string;
    image_url?: string;
    is_favorited?: boolean;
}

export interface SongVariant {
    id: number;
    version_number: number;
    song_id: number;
    slug: string;
    views: number;
    video?: SongVariantVideo;
}

export interface SongVariantVideo {
    type: 'file' | 'embed';
    embed_url?: string;
    local_url?: string;
}

export interface SongType {
    id: string;
    name: string;
    slug: string;
    description?: string;
}
