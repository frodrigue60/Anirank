import type { ImageSource } from './media';

export interface Badge {
    id: string;
    name: string;
    description?: string;
    icon_url?: string;
    image_url?: string; // Legacy field
    icon_sources?: ImageSource[];
}

export interface User {
    id: number;
    name: string;
    slug: string;
    avatar_url?: string;
    avatar_sources?: ImageSource[];
    banner_url?: string;
    banner_sources?: ImageSource[];
    xp: number;
    level: number;
    created_at: string;
    updated_at: string;
    followers_count?: number;
    following_count?: number;
    is_following?: boolean;

    // OAuth Connections
    anilist_id?: number | null;
    anilist_username?: string | null;
    google_id?: string | null;
    google_email?: string | null;
    profile_color?: string;
    about?: string;
    score_format_id?: number;
    score_format?: string;
    truth_score?: number;
    is_shadowbanned?: boolean;
    is_softbanned?: boolean;
    badges?: Badge[];
}

export interface RankingUser extends User {
    ratings_count: number;
    comments_count: number;
}
