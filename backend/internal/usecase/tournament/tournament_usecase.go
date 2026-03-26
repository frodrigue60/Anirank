package tournament

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

type TournamentUsecase struct {
	repo      domain.TournamentRepository
	songRepo  domain.SongRepository
	animeRepo domain.AnimeRepository
	storage   infrastructure.StorageService
}

func NewTournamentUsecase(repo domain.TournamentRepository, songRepo domain.SongRepository, animeRepo domain.AnimeRepository, storage infrastructure.StorageService) *TournamentUsecase {
	return &TournamentUsecase{repo: repo, songRepo: songRepo, animeRepo: animeRepo, storage: storage}
}

// GetActiveTournament builds the tree of matchups for the current active tournament.
func (u *TournamentUsecase) GetActiveTournament(ctx context.Context) (*domain.Tournament, error) {
	t, err := u.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	allMatchups, err := u.repo.GetMatchupsByTournament(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	if err := u.enrichMatchups(ctx, allMatchups, t); err != nil {
		return nil, err
	}

	t.Matchups = allMatchups
	return t, nil
}

// GetTournament fetches a specific tournament by its numeric ID.
func (u *TournamentUsecase) GetTournament(ctx context.Context, id uint64) (*domain.Tournament, error) {
	return u.repo.GetByID(ctx, id)
}

// GetTournamentBySlug fetches a specific tournament and its bracket.
func (u *TournamentUsecase) GetTournamentBySlug(ctx context.Context, slug string, userID *uint64) (*domain.Tournament, error) {
	t, err := u.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	allMatchups, err := u.repo.GetMatchupsByTournament(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	if userID != nil {
		votes, err := u.repo.GetUserVotesInTournament(ctx, *userID, t.ID)
		if err == nil {
			for i := range allMatchups {
				if songID, ok := votes[allMatchups[i].ID]; ok {
					allMatchups[i].UserVotedSongID = &songID
				}
			}
		}
	}

	if err := u.enrichMatchups(ctx, allMatchups, t); err != nil {
		return nil, err
	}

	t.Matchups = allMatchups
	return t, nil
}

func (u *TournamentUsecase) enrichMatchups(ctx context.Context, matchups []domain.TournamentMatchup, t *domain.Tournament) error {
	songIDs := make([]uint64, 0)
	idMap := make(map[uint64]bool)

	if t != nil && t.WinnerSongID != nil {
		songIDs = append(songIDs, *t.WinnerSongID)
		idMap[*t.WinnerSongID] = true
	}

	for _, m := range matchups {
		if m.Song1ID != nil && !idMap[*m.Song1ID] {
			songIDs = append(songIDs, *m.Song1ID)
			idMap[*m.Song1ID] = true
		}
		if m.Song2ID != nil && !idMap[*m.Song2ID] {
			songIDs = append(songIDs, *m.Song2ID)
			idMap[*m.Song2ID] = true
		}
		if m.WinnerSongID != nil && !idMap[*m.WinnerSongID] {
			songIDs = append(songIDs, *m.WinnerSongID)
			idMap[*m.WinnerSongID] = true
		}
	}

	if len(songIDs) == 0 {
		return nil
	}

	songs, err := u.songRepo.GetMany(ctx, songIDs)
	if err != nil {
		return err
	}

	songMap := make(map[uint64]*domain.Song)
	// Fetch and associate details
	animeIDs := make([]uint64, 0)
	animeIDMap := make(map[uint64]bool)

	for i := range songs {
		s := &songs[i]
		if _, exists := animeIDMap[s.AnimeID]; !exists {
			animeIDs = append(animeIDs, s.AnimeID)
			animeIDMap[s.AnimeID] = true
		}

		artists, _ := u.songRepo.GetArtistsBySongID(ctx, s.ID, false)
		s.Artists = artists
		variants, _ := u.songRepo.GetVariantsBySongID(ctx, s.ID)

		// Validate video sources (S3 check)
		for j := range variants {
			if variants[j].Video != nil && variants[j].Video.LocalUrl != nil {
				exists, _ := u.storage.FileExists(ctx, *variants[j].Video.LocalUrl)
				if !exists {
					variants[j].Video.LocalUrl = nil
				}
			}
		}
		s.Variants = variants
	}

	// Load all animes in batch
	animes, _ := u.animeRepo.GetMany(ctx, animeIDs)
	animesMap := make(map[uint64]*domain.Anime)
	for i := range animes {
		a := &animes[i]
		if a.Cover != nil {
			url := u.storage.GetURL(*a.Cover)
			a.CoverUrl = &url
		}
		if a.Banner != nil {
			url := u.storage.GetURL(*a.Banner)
			a.BannerUrl = &url
		}
		animesMap[a.ID] = a
	}

	for i := range songs {
		s := &songs[i]
		if anime, ok := animesMap[s.AnimeID]; ok {
			s.Anime = anime
		}
		songMap[s.ID] = s
	}

	for i := range matchups {
		m := &matchups[i]
		if m.Song1ID != nil {
			m.Song1 = songMap[*m.Song1ID]
		}
		if m.Song2ID != nil {
			m.Song2 = songMap[*m.Song2ID]
		}
		if m.WinnerSongID != nil {
			m.Winner = songMap[*m.WinnerSongID]
		}
	}

	if t != nil && t.WinnerSongID != nil {
		t.Winner = songMap[*t.WinnerSongID]
	}

	return nil
}

// CreateTournament creates a new tournament in draft mode.
func (u *TournamentUsecase) CreateTournament(ctx context.Context, t *domain.Tournament) error {
	if t.Name == "" {
		return domain.NewAppError(400, "Name is required", nil)
	}

	// Always start as draft
	t.Status = "draft"
	t.UUID = uuid.New().String()

	// Basic slug generation if not provided
	if t.Slug == "" {
		t.Slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		// Add regex or more robust slugification if needed
	}

	return u.repo.Create(ctx, t)
}

// ListAll returns all tournaments for administrative purposes.
func (u *TournamentUsecase) ListAll(ctx context.Context) ([]domain.Tournament, error) {
	return u.repo.List(ctx)
}

// ListPublic returns active and completed tournaments for the public index.
func (u *TournamentUsecase) ListPublic(ctx context.Context) ([]domain.Tournament, error) {
	return u.repo.ListPublic(ctx)
}

// DeleteTournament removes a tournament and its associated records if in valid state.
func (u *TournamentUsecase) DeleteTournament(ctx context.Context, id uint64) error {
	t, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Restriction: Only delete draft or completed
	if t.Status == "active" {
		return domain.NewAppError(400, "Cannot delete an active tournament", nil)
	}

	return u.repo.WithTransaction(ctx, func(txRepo domain.TournamentRepository) error {
		return txRepo.Delete(ctx, id)
	})
}

// SeedTournament initializes a tournament from draft to active using various methods.
func (u *TournamentUsecase) SeedTournament(ctx context.Context, id uint64, req domain.SeedRequest) error {
	t, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if t.Status != "draft" {
		return errors.New("only draft tournaments can be seeded")
	}

	var songs []domain.Song
	songType := "all"
	if t.TypeFilter != nil {
		songType = *t.TypeFilter
	}

	switch req.Method {
	case "manual":
		if len(req.ManualSongs) < t.Size {
			return errors.New("not enough manual songs provided")
		}
		songs, err = u.songRepo.GetMany(ctx, req.ManualSongs)
		if err != nil {
			return err
		}
	case "filtered", "random", "top":
		filters := domain.SongFilters{
			Type:    songType,
			Sort:    req.Sort,
			IsAdmin: true,
		}
		if req.SongType != "" {
			filters.Type = req.SongType
		}
		if req.YearID != nil {
			filters.YearID = *req.YearID
		}
		if req.SeasonID != nil {
			filters.SeasonID = *req.SeasonID
		}
		if req.GenreID != nil {
			filters.GenreID = *req.GenreID
		}

		// If method is random, force sort
		if req.Method == "random" {
			filters.Sort = "random"
		} else if filters.Sort == "" {
			filters.Sort = "rating" // Default to top ranked
		}

		songs, err = u.songRepo.GetPaginated(ctx, t.Size, 0, filters)
		if err != nil {
			return err
		}
	default: // Legacy behavior
		songs, err = u.songRepo.GetRanking(ctx, "global", songType, t.Size, 0)
		if err != nil {
			return err
		}
	}

	if len(songs) < t.Size {
		return errors.New("not enough songs to seed the tournament with selected filters")
	}

	// Shuffle if requested or for auto
	if req.Sort == "random" || req.Method == "random" {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
	}

	return u.repo.WithTransaction(ctx, func(repo domain.TournamentRepository) error {
		// Create initial matchups (Round of N)
		endsAt := time.Now().Add(48 * time.Hour)
		matchupCount := t.Size / 2

		for i := 0; i < matchupCount; i++ {
			m := &domain.TournamentMatchup{
				TournamentID: t.ID,
				Round:        t.Size,
				Position:     i + 1,
				Song1ID:      &songs[i*2].ID,
				Song2ID:      &songs[i*2+1].ID,
				IsActive:     true,
				EndsAt:       &endsAt,
			}
			if err := repo.CreateMatchup(ctx, m); err != nil {
				return err
			}
		}

		// Update tournament status
		currentRound := t.Size
		startedAt := time.Now()
		t.Status = "active"
		t.CurrentRound = &currentRound
		t.StartedAt = &startedAt

		return repo.Update(ctx, t)
	})
}

// SubmitVote validates and records a user's vote.
func (u *TournamentUsecase) SubmitVote(ctx context.Context, userID uint64, matchupID uint64, songID uint64, ip string) error {
	m, err := u.repo.GetMatchupByID(ctx, matchupID)
	if err != nil {
		return err
	}

	if !m.IsActive || (m.EndsAt != nil && m.EndsAt.Before(time.Now())) {
		return errors.New("matchup is not active or has already closed")
	}

	// Verify song belongs to matchup
	if (m.Song1ID == nil || *m.Song1ID != songID) && (m.Song2ID == nil || *m.Song2ID != songID) {
		return errors.New("the selected song is not part of this matchup")
	}

	// Check if user already voted
	hasVoted, err := u.repo.HasUserVoted(ctx, userID, matchupID)
	if err != nil {
		return err
	}
	if hasVoted {
		return errors.New("you have already voted in this matchup")
	}

	vote := &domain.TournamentVote{
		TournamentMatchupID: matchupID,
		UserID:              userID,
		SongID:              songID,
		IPAddress:           &ip,
	}

	return u.repo.SubmitVote(ctx, vote)
}

// AdvanceTournament manually resolves all active matchups in a tournament and advances to the next round.
func (u *TournamentUsecase) AdvanceTournament(ctx context.Context, tournamentID uint64) error {
	t, err := u.repo.GetByID(ctx, tournamentID)
	if err != nil {
		return err
	}

	if t.Status != "active" {
		return errors.New("only active tournaments can be advanced")
	}

	// Get all matchups for this tournament
	allMatchups, err := u.repo.GetMatchupsByTournament(ctx, tournamentID)
	if err != nil {
		return err
	}

	// Filter only active ones
	var activeMatchups []domain.TournamentMatchup
	for _, m := range allMatchups {
		if m.IsActive {
			activeMatchups = append(activeMatchups, m)
		}
	}

	if len(activeMatchups) == 0 {
		return errors.New("no active matchups found to advance")
	}

	// Resolve each active matchup
	for _, m := range activeMatchups {
		if err := u.resolveMatchup(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

// ProcessTournaments scans for expired matchups and advances the brackets.
func (u *TournamentUsecase) ProcessTournaments(ctx context.Context) error {
	matchups, err := u.repo.GetExpiredMatchups(ctx)
	if err != nil {
		return err
	}

	for _, m := range matchups {
		if err := u.resolveMatchup(ctx, m); err != nil {
			// Log error but continue with others
			continue
		}
	}

	return nil
}

func (u *TournamentUsecase) resolveMatchup(ctx context.Context, m domain.TournamentMatchup) error {
	return u.repo.WithTransaction(ctx, func(repo domain.TournamentRepository) error {
		// 1. Determine winner
		var winnerID uint64
		if m.Song2ID == nil {
			// Bye
			winnerID = *m.Song1ID
		} else if m.Song1Votes > m.Song2Votes {
			winnerID = *m.Song1ID
		} else if m.Song2Votes > m.Song1Votes {
			winnerID = *m.Song2ID
		} else {
			// Tie-breaker
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 0 {
				winnerID = *m.Song1ID
			} else {
				winnerID = *m.Song2ID
			}
		}

		// 2. Close current matchup
		m.IsActive = false
		m.WinnerSongID = &winnerID
		if err := repo.UpdateMatchup(ctx, &m); err != nil {
			return err
		}

		// 3. Check if it was the final (Round 2)
		if m.Round == 2 {
			t, err := repo.GetByID(ctx, m.TournamentID)
			if err != nil {
				return err
			}
			t.Status = "completed"
			t.WinnerSongID = &winnerID
			now := time.Now()
			t.CompletedAt = &now
			return repo.Update(ctx, t)
		}

		// 4. Advance to next round
		nextRound := m.Round / 2
		nextPosition := int(math.Ceil(float64(m.Position) / 2.0))
		isSong1 := m.Position%2 != 0 // Odd is Song1, Even is Song2

		nextMatchup, err := repo.FindMatchup(ctx, m.TournamentID, nextRound, nextPosition)
		if err != nil {
			return err
		}

		if nextMatchup == nil {
			// Create placeholder matchup for next round
			nextMatchup = &domain.TournamentMatchup{
				TournamentID: m.TournamentID,
				Round:        nextRound,
				Position:     nextPosition,
				IsActive:     false,
			}
			if isSong1 {
				nextMatchup.Song1ID = &winnerID
			} else {
				nextMatchup.Song2ID = &winnerID
			}
			if err := repo.CreateMatchup(ctx, nextMatchup); err != nil {
				return err
			}
		} else {
			// Update existing matchup
			if isSong1 {
				nextMatchup.Song1ID = &winnerID
			} else {
				nextMatchup.Song2ID = &winnerID
			}

			// If both songs are present, ACTIVATE the matchup
			if nextMatchup.Song1ID != nil && nextMatchup.Song2ID != nil {
				nextMatchup.IsActive = true
				nextEndsAt := time.Now().Add(48 * time.Hour)
				nextMatchup.EndsAt = &nextEndsAt

				// Update Tournament CurrentRound if needed
				t, err := repo.GetByID(ctx, m.TournamentID)
				if err == nil && (t.CurrentRound == nil || *t.CurrentRound > nextRound) {
					t.CurrentRound = &nextRound
					repo.Update(ctx, t)
				}
			}
			if err := repo.UpdateMatchup(ctx, nextMatchup); err != nil {
				return err
			}
		}

		return nil
	})
}
