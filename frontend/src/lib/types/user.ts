export interface User {
    id: number;
    name: string;
    slug: string;
    avatar_url?: string;
    banner_url?: string;
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
}

export interface RankingUser extends User {
    ratings_count: number;
    comments_count: number;
}
