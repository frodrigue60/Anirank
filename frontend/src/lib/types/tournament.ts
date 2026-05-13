import type { Song } from './song';

export interface Tournament {
    id: number;
    name: string;
    slug: string;
    description: string | null;
    size: number;
    type_filter: string | null;
    status: 'draft' | 'active' | 'completed';
    current_round: number | null;
    winner_song_id: number | null;
    started_at: string | null;
    completed_at: string | null;
    created_at: string;
    updated_at: string;
    matchups?: TournamentMatchup[];
    winner?: Song;
}

export interface TournamentMatchup {
    id: number | string;
    tournament_id: number;
    round: number;
    position: number;
    song1_id: number | string | null;
    song2_id: number | string | null;
    song1_votes: number;
    song2_votes: number;
    winner_song_id: number | string | null;
    ends_at: string | null;
    is_active: boolean;
    song1?: Song;
    song2?: Song;
    winner?: Song;
    user_voted_song_id?: number | string;
}
