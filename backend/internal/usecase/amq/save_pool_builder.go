package amq

import (
	"context"
	"fmt"
	"math/rand"

	"anirank/api/internal/domain"
)

const (
	saveRoundMaxAttempts      = 12
	saveMinCandidates         = 2
	saveAnimeMinDistinctNames = 4
)

var saveThemeKinds = []string{"artist", "year", "season", "anime", "genre"}

func IsSaveGameType(gameType string) bool {
	return gameType == "save-4" || gameType == "save-6"
}

func SaveOptionCount(gameType string) int {
	if gameType == "save-6" {
		return 6
	}
	return 4
}

// resolveRoundThemeType picks OP or ED for a single save round.
// Lobby "both" means each round is either all-OP or all-ED, never mixed.
func resolveRoundThemeType(themeTypeConfig string) string {
	if themeTypeConfig == "OP" || themeTypeConfig == "ED" {
		return themeTypeConfig
	}
	if rand.Intn(2) == 0 {
		return "OP"
	}
	return "ED"
}

func roundThemeTypesFilter(roundThemeType string) []string {
	return []string{roundThemeType}
}

func buildSaveThemeLabel(anchor domain.AMQSaveThemeAnchor, roundThemeType string) string {
	switch anchor.Kind {
	case "artist":
		return fmt.Sprintf("Best %s by %s", roundThemeType, anchor.ArtistName)
	case "year":
		return fmt.Sprintf("Best %s of %s", roundThemeType, anchor.YearName)
	case "season":
		return fmt.Sprintf("Best %s of %s %s", roundThemeType, anchor.SeasonName, anchor.YearName)
	case "anime":
		return fmt.Sprintf("Best %s of %s", roundThemeType, anchor.AnimeTitle)
	case "genre":
		return fmt.Sprintf("Best %s — %s genre", roundThemeType, anchor.GenreName)
	case "fallback":
		return fmt.Sprintf("Open %s Selection", roundThemeType)
	default:
		return "Save Round"
	}
}

func buildSaveThemeKey(anchor domain.AMQSaveThemeAnchor, roundThemeType string) string {
	switch anchor.Kind {
	case "artist":
		return fmt.Sprintf("artist:%s:%s", anchor.ArtistUUID, roundThemeType)
	case "year":
		if anchor.YearID == nil {
			return ""
		}
		return fmt.Sprintf("year:%d:%s", *anchor.YearID, roundThemeType)
	case "season":
		if anchor.SeasonID == nil || anchor.YearID == nil {
			return ""
		}
		return fmt.Sprintf("season:%d:%d:%s", *anchor.SeasonID, *anchor.YearID, roundThemeType)
	case "anime":
		return fmt.Sprintf("anime:%s:%s", anchor.AnimeUUID, roundThemeType)
	case "genre":
		if anchor.GenreID == nil {
			return ""
		}
		return fmt.Sprintf("genre:%d:%s", *anchor.GenreID, roundThemeType)
	case "fallback":
		return fmt.Sprintf("fallback:%d", rand.Int())
	default:
		return ""
	}
}

func themeKindsForRound(roundIndex int, distribution string) []string {
	if distribution != "balanced" {
		kinds := append([]string(nil), saveThemeKinds...)
		rand.Shuffle(len(kinds), func(i, j int) { kinds[i], kinds[j] = kinds[j], kinds[i] })
		return kinds
	}

	primary := saveThemeKinds[roundIndex%len(saveThemeKinds)]
	rest := make([]string, 0, len(saveThemeKinds)-1)
	for _, k := range saveThemeKinds {
		if k != primary {
			rest = append(rest, k)
		}
	}
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	return append([]string{primary}, rest...)
}

func (r *LobbyRoom) buildSaveRoundPool(ctx context.Context, maxRounds int, gameType, themeTypeConfig string) ([]domain.AMQSaveRound, error) {
	optionCount := SaveOptionCount(gameType)
	distribution := r.Config.ThemeDistribution
	if distribution != "balanced" {
		distribution = "random"
	}

	usedKeys := make(map[string]bool)
	rounds := make([]domain.AMQSaveRound, 0, maxRounds)

	for roundIndex := 0; len(rounds) < maxRounds; roundIndex++ {
		round, ok := r.tryBuildSaveRound(ctx, themeTypeConfig, optionCount, usedKeys, roundIndex, distribution)
		if !ok {
			round = r.buildFallbackSaveRound(ctx, themeTypeConfig, optionCount)
		}
		if len(round.Candidates) < saveMinCandidates {
			return nil, fmt.Errorf("could not build enough save rounds (got %d)", len(rounds))
		}
		if round.ThemeKey != "" {
			usedKeys[round.ThemeKey] = true
		}
		rounds = append(rounds, round)
	}

	return rounds, nil
}

func (r *LobbyRoom) tryBuildSaveRound(ctx context.Context, themeTypeConfig string, optionCount int, usedKeys map[string]bool, roundIndex int, distribution string) (domain.AMQSaveRound, bool) {
	roundThemeType := resolveRoundThemeType(themeTypeConfig)
	themeTypes := roundThemeTypesFilter(roundThemeType)
	kinds := themeKindsForRound(roundIndex, distribution)

	for attempt := 0; attempt < saveRoundMaxAttempts; attempt++ {
		kind := kinds[attempt%len(kinds)]
		anchor, err := r.findSaveAnchor(ctx, kind, themeTypes, optionCount)
		if err != nil || anchor == nil {
			continue
		}

		themeKey := buildSaveThemeKey(*anchor, roundThemeType)
		if themeKey == "" || usedKeys[themeKey] {
			continue
		}

		songIDs, err := r.SongRepo.GetRandomSongIDsForAMQSave(ctx, *anchor, themeTypes, optionCount)
		if err != nil || len(songIDs) < saveMinCandidates {
			continue
		}

		songs, err := r.hydrateSaveSongs(ctx, songIDs)
		if err != nil || len(songs) < saveMinCandidates {
			continue
		}

		return domain.AMQSaveRound{
			ThemeKind:      anchor.Kind,
			ThemeKey:       themeKey,
			ThemeLabel:     buildSaveThemeLabel(*anchor, roundThemeType),
			RoundThemeType: roundThemeType,
			IsFallback:     false,
			Candidates:     songs,
		}, true
	}

	return domain.AMQSaveRound{}, false
}

func (r *LobbyRoom) findSaveAnchor(ctx context.Context, kind string, themeTypes []string, optionCount int) (*domain.AMQSaveThemeAnchor, error) {
	minPool := saveMinCandidates
	if kind != "anime" {
		minPool = optionCount
	}

	switch kind {
	case "artist":
		return r.SongRepo.FindRandomArtistForAMQSave(ctx, themeTypes, minPool)
	case "year":
		return r.SongRepo.FindRandomYearForAMQSave(ctx, themeTypes, minPool)
	case "season":
		return r.SongRepo.FindRandomSeasonYearForAMQSave(ctx, themeTypes, minPool)
	case "anime":
		minDistinctNames := saveAnimeMinDistinctNames
		if optionCount > minDistinctNames {
			minDistinctNames = optionCount
		}
		return r.SongRepo.FindRandomAnimeForAMQSave(ctx, themeTypes, minDistinctNames)
	case "genre":
		return r.SongRepo.FindRandomGenreForAMQSave(ctx, themeTypes, minPool)
	default:
		return nil, nil
	}
}

func (r *LobbyRoom) buildFallbackSaveRound(ctx context.Context, themeTypeConfig string, optionCount int) domain.AMQSaveRound {
	roundThemeType := resolveRoundThemeType(themeTypeConfig)
	themeTypes := roundThemeTypesFilter(roundThemeType)
	anchor := domain.AMQSaveThemeAnchor{Kind: "fallback"}

	songIDs, err := r.SongRepo.GetRandomSongIDsForAMQSave(ctx, anchor, themeTypes, optionCount)
	if err != nil || len(songIDs) < saveMinCandidates {
		return domain.AMQSaveRound{
			ThemeKind:      "fallback",
			ThemeKey:       fmt.Sprintf("fallback:%d", rand.Int()),
			ThemeLabel:     buildSaveThemeLabel(anchor, roundThemeType),
			RoundThemeType: roundThemeType,
			IsFallback:     true,
		}
	}

	songs, err := r.hydrateSaveSongs(ctx, songIDs)
	if err != nil || len(songs) < saveMinCandidates {
		return domain.AMQSaveRound{
			ThemeKind:      "fallback",
			ThemeKey:       fmt.Sprintf("fallback:%d", rand.Int()),
			ThemeLabel:     buildSaveThemeLabel(anchor, roundThemeType),
			RoundThemeType: roundThemeType,
			IsFallback:     true,
		}
	}

	return domain.AMQSaveRound{
		ThemeKind:      "fallback",
		ThemeKey:       fmt.Sprintf("fallback:%d", rand.Int()),
		ThemeLabel:     buildSaveThemeLabel(anchor, roundThemeType),
		RoundThemeType: roundThemeType,
		IsFallback:     true,
		Candidates:     songs,
	}
}

func (r *LobbyRoom) hydrateSaveSongs(ctx context.Context, songIDs []uint64) ([]domain.Song, error) {
	songs, err := r.SongRepo.GetMany(ctx, songIDs)
	if err != nil {
		return nil, err
	}

	variantsMap, err := r.SongRepo.GetVariantsBySongIDs(ctx, songIDs)
	if err != nil {
		return nil, err
	}
	artistsMap, err := r.SongRepo.GetArtistsBySongIDs(ctx, songIDs)
	if err != nil {
		return nil, err
	}

	songByID := make(map[uint64]domain.Song, len(songs))
	for i := range songs {
		s := songs[i]
		if vars, ok := variantsMap[s.ID]; ok {
			s.Variants = vars
		}
		if arts, ok := artistsMap[s.ID]; ok {
			s.Artists = arts
		}
		songByID[s.ID] = s
	}

	ordered := make([]domain.Song, 0, len(songIDs))
	for _, id := range songIDs {
		if s, ok := songByID[id]; ok {
			ordered = append(ordered, s)
		}
	}
	return ordered, nil
}

func truncateAnimeTitle(title string, maxRunes int) string {
	runes := []rune(title)
	if len(runes) <= maxRunes {
		return title
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func songThemeLabel(song *domain.Song) string {
	if song == nil {
		return ""
	}
	return fmt.Sprintf("%s%s", song.Type, song.ThemeNum)
}
