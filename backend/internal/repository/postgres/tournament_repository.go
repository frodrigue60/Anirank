package postgres

import (
	"context"
	"database/sql"
	"errors"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type tournamentRepository struct {
	db sqlx.ExtContext
}

func NewTournamentRepository(db *sqlx.DB) domain.TournamentRepository {
	return &tournamentRepository{db: db}
}

func (r *tournamentRepository) WithTransaction(ctx context.Context, fn func(repo domain.TournamentRepository) error) error {
	db, ok := r.db.(*sqlx.DB)
	if !ok {
		return errors.New("cannot start transaction from a transaction")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	repo := &tournamentRepository{db: tx}
	err = fn(repo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *tournamentRepository) Create(ctx context.Context, t *domain.Tournament) error {
	query := `
		INSERT INTO tournaments (uuid, name, slug, description, size, type_filter, status, current_round, winner_song_id, started_at, completed_at, matchup_duration_hours, created_at, updated_at)
		VALUES (:uuid, :name, :slug, :description, :size, :type_filter, :status, :current_round, :winner_song_id, :started_at, :completed_at, :matchup_duration_hours, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, t)
	if err != nil {
		return err
	}
	err = sqlx.GetContext(ctx, r.db, &t.ID, boundQuery, args...)
	return err
}

func (r *tournamentRepository) Update(ctx context.Context, t *domain.Tournament) error {
	query := `
		UPDATE tournaments SET
			name = :name, slug = :slug, description = :description, size = :size,
			type_filter = :type_filter, status = :status, current_round = :current_round,
			winner_song_id = :winner_song_id, started_at = :started_at, completed_at = :completed_at,
			matchup_duration_hours = :matchup_duration_hours,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	_, err := sqlx.NamedExecContext(ctx, r.db, query, t)
	return err
}

func (r *tournamentRepository) GetActive(ctx context.Context) (*domain.Tournament, error) {
	var t domain.Tournament
	query := "SELECT * FROM tournaments WHERE status = 'active' LIMIT 1"
	err := sqlx.GetContext(ctx, r.db, &t, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tournamentRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tournament, error) {
	var t domain.Tournament
	query := "SELECT * FROM tournaments WHERE slug = $1"
	err := sqlx.GetContext(ctx, r.db, &t, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tournamentRepository) GetByID(ctx context.Context, id uint64) (*domain.Tournament, error) {
	var t domain.Tournament
	query := "SELECT * FROM tournaments WHERE id = $1"
	err := sqlx.GetContext(ctx, r.db, &t, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tournamentRepository) List(ctx context.Context) ([]domain.Tournament, error) {
	var tt []domain.Tournament
	query := "SELECT * FROM tournaments ORDER BY created_at DESC"
	err := sqlx.SelectContext(ctx, r.db, &tt, query)
	if tt == nil {
		tt = []domain.Tournament{}
	}
	return tt, err
}

func (r *tournamentRepository) ListPublic(ctx context.Context) ([]domain.Tournament, error) {
	var tt []domain.Tournament
	query := "SELECT * FROM tournaments WHERE status IN ('active', 'completed') ORDER BY created_at DESC"
	err := sqlx.SelectContext(ctx, r.db, &tt, query)
	if tt == nil {
		tt = []domain.Tournament{}
	}
	return tt, err
}

func (r *tournamentRepository) GetMatchupsByTournament(ctx context.Context, tournamentID uint64) ([]domain.TournamentMatchup, error) {
	return r.fetchMatchups(ctx, "WHERE tm.tournament_id = $1", tournamentID)
}

func (r *tournamentRepository) GetMatchupsByRound(ctx context.Context, tournamentID uint64, round int) ([]domain.TournamentMatchup, error) {
	return r.fetchMatchups(ctx, "WHERE tm.tournament_id = $1 AND tm.round = $2", tournamentID, round)
}

func (r *tournamentRepository) CreateMatchup(ctx context.Context, m *domain.TournamentMatchup) error {
	query := `
		INSERT INTO tournament_matchups (tournament_id, round, position, song1_id, song2_id, song1_votes, song2_votes, winner_song_id, ends_at, is_active, created_at, updated_at)
		VALUES (:tournament_id, :round, :position, :song1_id, :song2_id, :song1_votes, :song2_votes, :winner_song_id, :ends_at, :is_active, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, m)
	if err != nil {
		return err
	}
	err = sqlx.GetContext(ctx, r.db, &m.ID, boundQuery, args...)
	return err
}

func (r *tournamentRepository) UpdateMatchup(ctx context.Context, m *domain.TournamentMatchup) error {
	query := `
		UPDATE tournament_matchups SET 
			song1_id = :song1_id, song2_id = :song2_id, song1_votes = :song1_votes, 
			song2_votes = :song2_votes, winner_song_id = :winner_song_id, 
			ends_at = :ends_at, is_active = :is_active, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	_, err := sqlx.NamedExecContext(ctx, r.db, query, m)
	return err
}

func (r *tournamentRepository) GetExpiredMatchups(ctx context.Context) ([]domain.TournamentMatchup, error) {
	return r.fetchMatchups(ctx, "WHERE tm.is_active = true AND tm.ends_at <= CURRENT_TIMESTAMP")
}

func (r *tournamentRepository) GetMatchupByID(ctx context.Context, id uint64) (*domain.TournamentMatchup, error) {
	mm, err := r.fetchMatchups(ctx, "WHERE tm.id = $1", id)
	if err != nil {
		return nil, err
	}
	if len(mm) == 0 {
		return nil, domain.ErrNotFound
	}
	return &mm[0], nil
}

func (r *tournamentRepository) GetMatchupByUUID(ctx context.Context, uuid string) (*domain.TournamentMatchup, error) {
	mm, err := r.fetchMatchups(ctx, "WHERE tm.uuid = $1", uuid)
	if err != nil {
		return nil, err
	}
	if len(mm) == 0 {
		return nil, domain.ErrNotFound
	}
	return &mm[0], nil
}

func (r *tournamentRepository) FindMatchup(ctx context.Context, tournamentID uint64, round int, position int) (*domain.TournamentMatchup, error) {
	mm, err := r.fetchMatchups(ctx, "WHERE tm.tournament_id = $1 AND tm.round = $2 AND tm.position = $3", tournamentID, round, position)
	if err != nil {
		return nil, err
	}
	if len(mm) == 0 {
		return nil, nil // Not found but not error
	}
	return &mm[0], nil
}

func (r *tournamentRepository) HasUserVoted(ctx context.Context, userID uint64, matchupID uint64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM tournament_votes WHERE user_id = $1 AND tournament_matchup_id = $2"
	err := sqlx.GetContext(ctx, r.db, &count, query, userID, matchupID)
	return count > 0, err
}

func (r *tournamentRepository) GetUserVotesInTournament(ctx context.Context, userID uint64, tournamentID uint64) (map[uint64]uint64, error) {
	query := `
		SELECT tournament_matchup_id, song_id 
		FROM tournament_votes 
		WHERE user_id = $1 AND tournament_matchup_id IN (
			SELECT id FROM tournament_matchups WHERE tournament_id = $2
		)
	`
	rows, err := r.db.QueryContext(ctx, query, userID, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	votes := make(map[uint64]uint64)
	for rows.Next() {
		var matchupID, songID uint64
		if err := rows.Scan(&matchupID, &songID); err != nil {
			return nil, err
		}
		votes[matchupID] = songID
	}
	return votes, nil
}

func (r *tournamentRepository) SubmitVote(ctx context.Context, vote *domain.TournamentVote) error {
	// First, insert the vote
	queryInsert := `
		INSERT INTO tournament_votes (tournament_matchup_id, user_id, song_id, ip_address, created_at, updated_at) 
		VALUES (:tournament_matchup_id, :user_id, :song_id, :ip_address, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := sqlx.NamedExecContext(ctx, r.db, queryInsert, vote)
	if err != nil {
		return err
	}

	// Then, increment the vote count in the matchup
	// Determine which column to increment
	m, err := r.GetMatchupByID(ctx, vote.TournamentMatchupID)
	if err != nil {
		return err
	}

	column := ""
	if m.Song1ID != nil && *m.Song1ID == vote.SongID {
		column = "song1_votes"
	} else if m.Song2ID != nil && *m.Song2ID == vote.SongID {
		column = "song2_votes"
	} else {
		return errors.New("voted song is not in this matchup")
	}

	queryUpdate := "UPDATE tournament_matchups SET " + column + " = " + column + " + 1 WHERE id = $1"
	_, err = r.db.ExecContext(ctx, queryUpdate, vote.TournamentMatchupID)
	return err
}

func (r *tournamentRepository) Delete(ctx context.Context, id uint64) error {
	// Cascading delete: votes -> matchups -> tournament
	
	// 1. Delete votes
	_, err := r.db.ExecContext(ctx, "DELETE FROM tournament_votes WHERE tournament_matchup_id IN (SELECT id FROM tournament_matchups WHERE tournament_id = $1)", id)
	if err != nil {
		return err
	}

	// 2. Delete matchups
	_, err = r.db.ExecContext(ctx, "DELETE FROM tournament_matchups WHERE tournament_id = $1", id)
	if err != nil {
		return err
	}

	// 3. Delete tournament
	_, err = r.db.ExecContext(ctx, "DELETE FROM tournaments WHERE id = $1", id)
	return err
}

// Internal helper for fetching matchups with song details
func (r *tournamentRepository) fetchMatchups(ctx context.Context, whereClause string, args ...interface{}) ([]domain.TournamentMatchup, error) {
	query := `
		SELECT 
			tm.*,
			s1.id as "song1.id", s1.uuid as "song1.uuid", s1.song_romaji as "song1.song_romaji", s1.song_jp as "song1.song_jp", s1.song_en as "song1.song_en",
			s1.anime_id as "song1.anime_id", st1.slug as "song1.type", (st1.slug || s1.theme_num) as "song1.slug",
			s2.id as "song2.id", s2.uuid as "song2.uuid", s2.song_romaji as "song2.song_romaji", s2.song_jp as "song2.song_jp", s2.song_en as "song2.song_en",
			s2.anime_id as "song2.anime_id", st2.slug as "song2.type", (st2.slug || s2.theme_num) as "song2.slug",
			w.id as "winner.id", w.uuid as "winner.uuid", w.song_romaji as "winner.song_romaji", w.song_jp as "winner.song_jp", w.song_en as "winner.song_en",
			w.anime_id as "winner.anime_id", stw.slug as "winner.type", (stw.slug || w.theme_num) as "winner.slug"
		FROM tournament_matchups tm
		LEFT JOIN songs s1 ON tm.song1_id = s1.id
		LEFT JOIN song_types st1 ON s1.type_id = st1.id
		LEFT JOIN songs s2 ON tm.song2_id = s2.id
		LEFT JOIN song_types st2 ON s2.type_id = st2.id
		LEFT JOIN songs w ON tm.winner_song_id = w.id
		LEFT JOIN song_types stw ON w.type_id = stw.id
		` + whereClause + `
		ORDER BY tm.position ASC
	`

	type MatchupRow struct {
		domain.TournamentMatchup
		S1ID      *uint64 `db:"song1.id"`
		S1UUID    *string `db:"song1.uuid"`
		S1Romaji  *string `db:"song1.song_romaji"`
		S1JP      *string `db:"song1.song_jp"`
		S1EN      *string `db:"song1.song_en"`
		S1AnimeID *uint64 `db:"song1.anime_id"`
		S1Type    *string `db:"song1.type"`
		S1Slug    *string `db:"song1.slug"`

		S2ID      *uint64 `db:"song2.id"`
		S2UUID    *string `db:"song2.uuid"`
		S2Romaji  *string `db:"song2.song_romaji"`
		S2JP      *string `db:"song2.song_jp"`
		S2EN      *string `db:"song2.song_en"`
		S2AnimeID *uint64 `db:"song2.anime_id"`
		S2Type    *string `db:"song2.type"`
		S2Slug    *string `db:"song2.slug"`

		WID       *uint64 `db:"winner.id"`
		WUUID     *string `db:"winner.uuid"`
		WRomaji   *string `db:"winner.song_romaji"`
		WJP       *string `db:"winner.song_jp"`
		WEN       *string `db:"winner.song_en"`
		WAnimeID  *uint64 `db:"winner.anime_id"`
		WType     *string `db:"winner.type"`
		WSlug     *string `db:"winner.slug"`
	}

	var rows []MatchupRow
	err := sqlx.SelectContext(ctx, r.db, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	matchups := make([]domain.TournamentMatchup, len(rows))
	for i, row := range rows {
		m := row.TournamentMatchup
		if row.S1ID != nil {
			m.Song1 = &domain.Song{
				ID: *row.S1ID, UUID: *row.S1UUID, SongRomaji: row.S1Romaji,
				SongJP: row.S1JP, SongEN: row.S1EN, AnimeID: *row.S1AnimeID,
				Type: *row.S1Type, Slug: *row.S1Slug,
			}
		}
		if row.S2ID != nil {
			m.Song2 = &domain.Song{
				ID: *row.S2ID, UUID: *row.S2UUID, SongRomaji: row.S2Romaji,
				SongJP: row.S2JP, SongEN: row.S2EN, AnimeID: *row.S2AnimeID,
				Type: *row.S2Type, Slug: *row.S2Slug,
			}
		}
		if row.WID != nil {
			m.Winner = &domain.Song{
				ID: *row.WID, UUID: *row.WUUID, SongRomaji: row.WRomaji,
				SongJP: row.WJP, SongEN: row.WEN, AnimeID: *row.WAnimeID,
				Type: *row.WType, Slug: *row.WSlug,
			}
		}
		matchups[i] = m
	}

	return matchups, nil
}
