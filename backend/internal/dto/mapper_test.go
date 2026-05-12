package dto

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"testing"
)

func TestToUserMinimalDTO(t *testing.T) {
	forbiddenID := uint64(99999)
	uuid := "user-uuid-123"

	user := &domain.User{
		ID:   forbiddenID,
		UUID: uuid,
		Name: "Test User",
	}

	dto := ToUserMinimalDTO(user)

	if dto.ID != uuid {
		t.Errorf("Expected ID %s, got %s", uuid, dto.ID)
	}

	// Security check: ensure no leakage of forbiddenID
	testutil.AssertNoInternalIDs(t, dto, forbiddenID)
}

func TestToUserDTO(t *testing.T) {
	forbiddenID := uint64(88888)
	badgeForbiddenID := uint64(77777)
	userUUID := "user-uuid"
	badgeUUID := "badge-uuid"

	user := &domain.User{
		ID:   forbiddenID,
		UUID: userUUID,
		Name: "Test User",
		Badges: []domain.Badge{
			{
				ID:   badgeForbiddenID,
				UUID: badgeUUID,
				Name: "Collector",
			},
		},
		Roles: []domain.Role{
			{
				Slug: "admin",
			},
		},
		SocialIdentities: []domain.UserSocialIdentity{
			{
				Provider:         "google",
				ProviderUsername: pointer("google_user"),
			},
		},
	}

	dto := ToUserDTO(user)

	if dto.ID != userUUID {
		t.Errorf("Expected UserDTO.ID %s, got %s", userUUID, dto.ID)
	}

	if len(dto.Badges) != 1 {
		t.Fatalf("Expected 1 badge, got %d", len(dto.Badges))
	}

	if dto.Badges[0].ID != badgeUUID {
		t.Errorf("Expected Badge DTO ID %s, got %s", badgeUUID, dto.Badges[0].ID)
	}

	if len(dto.Roles) != 1 || dto.Roles[0] != "admin" {
		t.Errorf("Expected role 'admin', got %v", dto.Roles)
	}

	// Security checks for both user and child relations
	testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	testutil.AssertNoInternalIDs(t, dto, badgeForbiddenID)
}

func TestToSongMappers(t *testing.T) {
	forbiddenID := uint64(11111)
	animeID := uint64(22222)
	songUUID := "song-uuid-1"
	animeUUID := "anime-uuid-1"

	song := &domain.Song{
		ID:         forbiddenID,
		UUID:       songUUID,
		SongRomaji: pointer("Gurenge"),
		Slug:       "gurenge-test",
		AnimeID:    animeID,
		Anime: &domain.Anime{
			ID:    animeID,
			UUID:  animeUUID,
			Title: "Demon Slayer",
			Year: &domain.Year{
				ID:   55555,
				UUID: "year-uuid-1",
			},
			Season: &domain.Season{
				ID:   66666,
				UUID: "season-uuid-1",
			},
		},
		Artists: []domain.Artist{
			{
				ID:   33333,
				UUID: "artist-uuid-1",
				Name: "LiSA",
			},
		},
		SongType: &domain.SongType{
			ID:   pointer(uint64(44444)),
			UUID: pointer("type-uuid-1"),
			Name: pointer("Opening"),
		},
	}

	t.Run("ToSongMinimalDTO", func(t *testing.T) {
		dto := ToSongMinimalDTO(song)
		if dto.ID != songUUID {
			t.Errorf("Expected Song ID %s, got %s", songUUID, dto.ID)
		}
		if dto.AnimeID != animeUUID {
			t.Errorf("Expected AnimeID UUID %s, got %s", animeUUID, dto.AnimeID)
		}
		if dto.YearID != "year-uuid-1" {
			t.Errorf("Expected YearID year-uuid-1, got %s", dto.YearID)
		}
		if dto.TypeID != "type-uuid-1" {
			t.Errorf("Expected TypeID type-uuid-1, got %s", dto.TypeID)
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
		testutil.AssertNoInternalIDs(t, dto, animeID)
	})

	t.Run("ToSongDTO", func(t *testing.T) {
		dto := ToSongDTO(song)
		if dto.ID != songUUID {
			t.Errorf("Expected Song ID %s, got %s", songUUID, dto.ID)
		}
		if len(dto.Artists) != 1 {
			t.Errorf("Artists not mapped")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})
}

func TestToArtistMappers(t *testing.T) {
	forbiddenID := uint64(55555)
	artistUUID := "artist-uuid-2"

	artist := &domain.Artist{
		ID:           forbiddenID,
		UUID:         artistUUID,
		Name:         "Aimer",
		FavoritesCount: 100,
	}

	t.Run("ToArtistMinimalDTO", func(t *testing.T) {
		dto := ToArtistMinimalDTO(artist)
		if dto.ID != artistUUID {
			t.Errorf("Expected Artist ID %s, got %s", artistUUID, dto.ID)
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})

	t.Run("ToArtistDTO", func(t *testing.T) {
		dto := ToArtistDTO(artist)
		if dto.FavoritesCount != 100 {
			t.Errorf("Expected 100 favorites, got %d", dto.FavoritesCount)
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})
}

func TestToAnimeMappers(t *testing.T) {
	forbiddenID := uint64(66666)
	animeUUID := "anime-uuid-2"
	genreUUID := "genre-uuid-1"
	studioUUID := "studio-uuid-1"

	anime := &domain.Anime{
		ID:        forbiddenID,
		UUID:      animeUUID,
		Title:     "Attack on Titan",
		Slug:      "aot",
		AnilistID: pointer(int64(16498)),
		Year: &domain.Year{
			ID:   111,
			UUID: "year-uuid-anime",
			Name: "2013",
		},
		Season: &domain.Season{
			ID:   222,
			UUID: "season-uuid-anime",
			Name: "Spring",
		},
		Genres: []domain.Genre{
			{ID: 333, UUID: genreUUID, Name: "Action", Slug: "action"},
		},
		Studios: []domain.Studio{
			{ID: 77777, UUID: studioUUID, Name: "WIT Studio", Slug: "wit-studio"},
		},
		ExternalLinks: []domain.ExternalLink{
			{UUID: "link-uuid-1", Name: "Official Site", URL: "https://example.com"},
		},
	}

	t.Run("ToAnimeMinimalDTO", func(t *testing.T) {
		dto := ToAnimeMinimalDTO(anime)
		if dto.Year == nil || dto.Year.ID != "year-uuid-anime" {
			t.Errorf("Year ID not mapped correctly, got %v", dto.Year.ID)
		}
		if dto.Season == nil || dto.Season.ID != "season-uuid-anime" {
			t.Errorf("Season ID not mapped correctly, got %v", dto.Season.ID)
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})

	t.Run("ToAnimeDTO", func(t *testing.T) {
		dto := ToAnimeDTO(anime)
		if len(dto.Genres) != 1 || dto.Genres[0].ID != genreUUID {
			t.Errorf("Genre ID not mapped correctly")
		}
		if len(dto.Studios) != 1 || dto.Studios[0].ID != studioUUID {
			t.Errorf("Studio ID not mapped correctly")
		}
		if len(dto.ExternalLinks) != 1 || dto.ExternalLinks[0].ID != "link-uuid-1" {
			t.Errorf("ExternalLink UUID not mapped correctly")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})
}

func TestToPlaylistMappers(t *testing.T) {
	forbiddenID := uint64(99001)
	userForbiddenID := uint64(99002)
	playlistUUID := "playlist-uuid-1"

	playlist := &domain.Playlist{
		ID:        forbiddenID,
		UUID:      playlistUUID,
		Name:      "My Favorites",
		SongCount: 5,
		User:      &domain.User{ID: userForbiddenID, UUID: "user-uuid-playlist"},
	}

	t.Run("ToPlaylistMinimalDTO", func(t *testing.T) {
		dto := ToPlaylistMinimalDTO(playlist)
		if dto.ID != playlistUUID {
			t.Errorf("ID mismatch")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})

	t.Run("ToPlaylistDTO", func(t *testing.T) {
		dto := ToPlaylistDTO(playlist)
		if dto.User.ID == "" {
			t.Errorf("User minimal DTO not mapped")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
		testutil.AssertNoInternalIDs(t, dto, userForbiddenID)
	})
}

func TestToCommentMappers(t *testing.T) {
	forbiddenID := uint64(1212)
	comment := &domain.Comment{
		ID:      forbiddenID,
		UUID:    "comment-uuid",
		Content: "Nice song!",
		User:    &domain.User{ID: 1213, UUID: "user-uuid-comment"},
	}

	dto := ToCommentDTO(comment)
	if dto.Content != "Nice song!" {
		t.Errorf("Content mismatch")
	}
	testutil.AssertNoInternalIDs(t, dto, forbiddenID)
}

func TestToTournamentMappers(t *testing.T) {
	forbiddenID := uint64(3434)
	tournament := &domain.Tournament{
		ID:   forbiddenID,
		UUID: "tournament-uuid",
		Name: "Best OP 2024",
	}

	dto := ToTournamentMinimalDTO(tournament)
	if dto.Name != "Best OP 2024" {
		t.Errorf("Name mismatch")
	}
	testutil.AssertNoInternalIDs(t, dto, forbiddenID)
}

func TestToAnnouncementMappers(t *testing.T) {
	forbiddenID := uint64(5656)
	announcement := domain.Announcement{
		ID:    forbiddenID,
		UUID:  "ann-uuid",
		Title: "New Update",
	}

	t.Run("ToAnnouncementDTO", func(t *testing.T) {
		dto := ToAnnouncementDTO(announcement)
		// ToAnnouncementDTO uses UUID as ID (interface{})
		if dto.ID != "ann-uuid" {
			t.Errorf("Expected UUID as ID")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})

	t.Run("ToAdminAnnouncementDTO", func(t *testing.T) {
		dto := ToAdminAnnouncementDTO(announcement)
		// Admin version DOES export numeric ID by design (mapper.go:717)
		if dto.ID != forbiddenID {
			t.Errorf("Admin DTO should have numeric ID %d, got %v", forbiddenID, dto.ID)
		}
	})
}

func TestToNotificationDTO(t *testing.T) {
	notifID := "notif-uuid-1"
	notif := domain.Notification{
		ID:   notifID,
		Type: "new_song",
	}

	dto := ToNotificationDTO(notif)
	// NotificationDTO maps ID directly as string
	if dto.ID != notifID {
		t.Errorf("Notification ID mismatch")
	}
}

func TestToActivityMappers(t *testing.T) {
	forbiddenID := uint64(9999)
	song := &domain.Song{ID: forbiddenID, UUID: "activity-song-uuid", SongRomaji: pointer("Identity")}

	t.Run("ToActivityItemDTO", func(t *testing.T) {
		item := domain.ActivityItem{
			Type:       "like",
			Target:     song,
			TargetType: "song",
			User:       domain.User{ID: 8888, UUID: "user-uuid"},
		}
		dto := ToActivityItemDTO(item)
		if dto.TargetID != "activity-song-uuid" {
			t.Errorf("Target ID not mapped")
		}
		testutil.AssertNoInternalIDs(t, dto, forbiddenID)
	})
}

func pointer[T any](v T) *T {
	return &v
}
