package domain

import (
	"context"
	"time"
)

type Tournament struct {
	ID           uint64     `db:"id" json:"id"`
	UUID         string     `db:"uuid" json:"uuid"`
	Name         string     `db:"name" json:"name"`
	Slug         string     `db:"slug" json:"slug"`
	Description  *string    `db:"description" json:"description"`
	Size         int        `db:"size" json:"size"`
	TypeFilter   *string    `db:"type_filter" json:"type_filter"`
	Status       string     `db:"status" json:"status"` // draft, active, completed
	CurrentRound *int       `db:"current_round" json:"current_round"`
	WinnerSongID *uint64    `db:"winner_song_id" json:"winner_song_id"`
	StartedAt    *time.Time `db:"started_at" json:"started_at"`
	CompletedAt  *time.Time `db:"completed_at" json:"completed_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`

	// Tree of matchups
	Matchups []TournamentMatchup `db:"-" json:"matchups,omitempty"`
	Winner   *Song               `db:"-" json:"winner,omitempty"`
}

type TournamentMatchup struct {
	ID           uint64     `db:"id" json:"id"`
	UUID         string     `db:"uuid" json:"uuid"`
	TournamentID uint64     `db:"tournament_id" json:"tournament_id"`
	Round        int        `db:"round" json:"round"`
	Position     int        `db:"position" json:"position"`
	Song1ID      *uint64    `db:"song1_id" json:"song1_id"`
	Song2ID      *uint64    `db:"song2_id" json:"song2_id"` // Nullable if waiting or got a bye
	Song1Votes   uint32     `db:"song1_votes" json:"song1_votes"`
	Song2Votes   uint32     `db:"song2_votes" json:"song2_votes"`
	WinnerSongID *uint64    `db:"winner_song_id" json:"winner_song_id"`
	EndsAt       *time.Time `db:"ends_at" json:"ends_at"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	UserVotedSongID *uint64 `db:"-" json:"user_voted_song_id,omitempty"`

	// Relations
	Song1  *Song `db:"-" json:"song1,omitempty"`
	Song2  *Song `db:"-" json:"song2,omitempty"`
	Winner *Song `db:"-" json:"winner,omitempty"`
}

type TournamentVote struct {
	ID                  uint64    `db:"id" json:"id"`
	TournamentMatchupID uint64    `db:"tournament_matchup_id" json:"tournament_matchup_id"`
	UserID              uint64    `db:"user_id" json:"user_id"`
	SongID              uint64    `db:"song_id" json:"song_id"`
	IPAddress           *string   `db:"ip_address" json:"ip_address"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

type SeedRequest struct {
	Method      string   `json:"method"` // random, top, manual, filtered
	YearID      *uint64  `json:"year_id,omitempty"`
	SeasonID    *uint64  `json:"season_id,omitempty"`
	GenreID     *uint64  `json:"genre_id,omitempty"`
	SongType    string   `json:"song_type,omitempty"` // OP, ED, INS, OTH
	ManualSongs []uint64 `json:"manual_songs,omitempty"`
	Sort        string   `json:"sort,omitempty"` // random, rating
}

// Repositories
type TournamentRepository interface {
	Create(ctx context.Context, t *Tournament) error
	Update(ctx context.Context, t *Tournament) error
	GetActive(ctx context.Context) (*Tournament, error)
	GetBySlug(ctx context.Context, slug string) (*Tournament, error)
	GetByID(ctx context.Context, id uint64) (*Tournament, error)
	List(ctx context.Context) ([]Tournament, error)
	ListPublic(ctx context.Context) ([]Tournament, error)
	Delete(ctx context.Context, id uint64) error

	// Matchups
	GetMatchupsByTournament(ctx context.Context, tournamentID uint64) ([]TournamentMatchup, error)
	GetMatchupsByRound(ctx context.Context, tournamentID uint64, round int) ([]TournamentMatchup, error)
	CreateMatchup(ctx context.Context, m *TournamentMatchup) error
	UpdateMatchup(ctx context.Context, m *TournamentMatchup) error
	GetExpiredMatchups(ctx context.Context) ([]TournamentMatchup, error)
	GetMatchupByID(ctx context.Context, id uint64) (*TournamentMatchup, error)
	FindMatchup(ctx context.Context, tournamentID uint64, round int, position int) (*TournamentMatchup, error)

	// Voting
	SubmitVote(ctx context.Context, vote *TournamentVote) error
	HasUserVoted(ctx context.Context, userID uint64, matchupID uint64) (bool, error)
	GetUserVotesInTournament(ctx context.Context, userID uint64, tournamentID uint64) (map[uint64]uint64, error)

	// Transactional Helper
	WithTransaction(ctx context.Context, fn func(repo TournamentRepository) error) error
}
