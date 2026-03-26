package dto

import "time"

type TournamentMinimalDTO struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Size         int       `json:"size"`
	Status       string    `json:"status"`
	CurrentRound *int      `json:"current_round,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TournamentDTO struct {
	TournamentMinimalDTO
	Description  *string                   `json:"description,omitempty"`
	Winner       *SongMinimalDTO           `json:"winner,omitempty"`
	Matchups     []TournamentMatchupDTO    `json:"matchups,omitempty"`
}

type TournamentMatchupDTO struct {
	ID              string    `json:"id"`
	Round           int       `json:"round"`
	Position        int       `json:"position"`
	Song1           *SongMinimalDTO `json:"song1,omitempty"`
	Song2           *SongMinimalDTO `json:"song2,omitempty"`
	Song1Votes      uint32    `json:"song1_votes"`
	Song2Votes      uint32    `json:"song2_votes"`
	Winner          *SongMinimalDTO `json:"winner,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	UserVotedSongID *string    `json:"user_voted_song_id,omitempty"`
}
