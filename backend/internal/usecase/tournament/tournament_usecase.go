package tournament

import (
	"context"
	"errors"
	"fmt"
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
	uniqueSongs := make([]*domain.Song, 0)

	// Helper to collect unique songs already partially hydrated by repo
	collect := func(s *domain.Song) {
		if s != nil && !idMap[s.ID] {
			songIDs = append(songIDs, s.ID)
			uniqueSongs = append(uniqueSongs, s)
			idMap[s.ID] = true
		}
	}

	for _, m := range matchups {
		collect(m.Song1)
		collect(m.Song2)
		collect(m.Winner)
	}

	if len(uniqueSongs) == 0 {
		return nil
	}

	// Batch fetch all relations (Artists and Variants)
	artistsMap, _ := u.songRepo.GetArtistsBySongIDs(ctx, songIDs)
	variantsMap, _ := u.songRepo.GetVariantsBySongIDs(ctx, songIDs)

	animeIDs := make([]uint64, 0)
	animeIDMap := make(map[uint64]bool)

	for _, s := range uniqueSongs {
		// Hydrate artists from batch map
		if artists, ok := artistsMap[s.ID]; ok {
			s.Artists = artists
		} else {
			s.Artists = []domain.Artist{}
		}

		// Hydrate variants from batch map
		if variants, ok := variantsMap[s.ID]; ok {
			s.Variants = variants
		} else {
			s.Variants = []domain.SongVariant{}
		}

		if _, exists := animeIDMap[s.AnimeID]; !exists {
			animeIDs = append(animeIDs, s.AnimeID)
			animeIDMap[s.AnimeID] = true
		}
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

	for _, s := range uniqueSongs {
		if anime, ok := animesMap[s.AnimeID]; ok {
			s.Anime = anime
		}
	}

	// The matchups and tournament pointers already point to the uniqueSongs 
	// because uniqueSongs was collected from them. No further mapping needed.
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
		if err != nil || len(songs) < t.Size {
			// Fallback: Try to get any active songs randomly if ranking is insufficient
			songs, err = u.songRepo.GetPaginated(ctx, t.Size, 0, domain.SongFilters{
				Type:    songType,
				Sort:    "random",
				IsAdmin: true,
			})
		}
		if err != nil {
			return err
		}
	}

	if len(songs) < t.Size {
		return fmt.Errorf("not enough songs to seed the tournament (found %d, need %d) with filters: type=%s", len(songs), t.Size, songType)
	}

	// Shuffle if requested or for auto
	if req.Sort == "random" || req.Method == "random" {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
	}

	return u.repo.WithTransaction(ctx, func(repo domain.TournamentRepository) error {
		// Create initial matchups (Round of N)
		duration := time.Duration(t.MatchupDurationHours) * time.Hour
		if duration == 0 {
			duration = 48 * time.Hour
		}
		endsAt := time.Now().Add(duration)
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
				
				t, err := repo.GetByID(ctx, m.TournamentID)
				if err != nil {
					return err
				}

				duration := time.Duration(t.MatchupDurationHours) * time.Hour
				if duration == 0 {
					duration = 48 * time.Hour
				}
				nextEndsAt := time.Now().Add(duration)
				nextMatchup.EndsAt = &nextEndsAt

				// Update Tournament CurrentRound if needed
				if t.CurrentRound == nil || *t.CurrentRound > nextRound {
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

func (u *TournamentUsecase) GetMatchupByUUID(ctx context.Context, uuid string) (*domain.TournamentMatchup, error) {
	return u.repo.GetMatchupByUUID(ctx, uuid)
}

func (u *TournamentUsecase) GetSongByUUID(ctx context.Context, uuid string) (*domain.Song, error) {
	return u.songRepo.GetByUUID(ctx, uuid)
}
