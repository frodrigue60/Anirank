package dto

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/public"
	"time"
)

// ─── User Mappers ───

func ToUserMinimalDTO(u *domain.User) UserMinimalDTO {
	if u == nil {
		return UserMinimalDTO{}
	}
	return UserMinimalDTO{
		ID:            u.UUID,
		UUID:          u.UUID,
		Name:          u.Name,
		Slug:          u.Slug,
		AvatarUrl:     u.AvatarUrl,
		AvatarSources: u.AvatarSources,
		XP:            u.XP,
		Level:         u.Level,
		RatingsCount:  u.RatingsCount,
		CommentsCount: u.CommentsCount,
		BannerUrl:     u.BannerUrl,
		BannerSources: u.BannerSources,
		CreatedAt:     u.CreatedAt,
	}
}

func ToUserDTO(u *domain.User) UserDTO {
	if u == nil {
		return UserDTO{}
	}

	badges := make([]BadgeDTO, 0)
	for _, b := range u.Badges {
		badges = append(badges, BadgeDTO{
			ID:          b.UUID,
			Name:        b.Name,
			Description: b.Description,
			IconUrl:     b.IconUrl,
		})
	}

	roles := make([]string, 0)
	for _, r := range u.Roles {
		roles = append(roles, r.Slug)
	}

	return UserDTO{
		UserMinimalDTO:  ToUserMinimalDTO(u),
		About:           u.About,
		ProfileColor:    u.ProfileColor,
		FollowersCount:  u.FollowersCount,
		FollowingCount:  u.FollowingCount,
		IsFollowing:     u.IsFollowing,
		Roles:           roles,
		Badges:          badges,
		AnilistID:       u.AnilistID,
		AnilistUsername: u.AnilistUsername,
		GoogleID:        u.GoogleID,
		GoogleEmail:     u.GoogleEmail,
	}
}

func ToAuthUserDTO(u *domain.User) AuthUserDTO {
	if u == nil {
		return AuthUserDTO{}
	}

	return AuthUserDTO{
		UserDTO:       ToUserDTO(u),
		Email:         u.Email,
		ScoreFormat:   u.ScoreFormat,
	}
}

func ToAuthResponseDTO(res *auth.AuthTokenResponse) AuthResponseDTO {
	if res == nil {
		return AuthResponseDTO{}
	}
	return AuthResponseDTO{
		Token: res.Token,
		User:  ToAuthUserDTO(res.User),
	}
}

// ─── Song Mappers ───

func ToSongMinimalDTO(s *domain.Song) SongMinimalDTO {
	if s == nil {
		return SongMinimalDTO{}
	}

	var anime *SongAnimeDTO
	if s.Anime != nil {
		coverUrl := ""
		if s.Anime.CoverUrl != nil {
			coverUrl = *s.Anime.CoverUrl
		}
		anime = &SongAnimeDTO{
			Title:         s.Anime.Title,
			Slug:          s.Anime.Slug,
			CoverUrl:      coverUrl,
			CoverSources:  s.Anime.CoverSources,
			BannerUrl:     s.Anime.BannerUrl,
			BannerSources: s.Anime.BannerSources,
			AnilistID:     s.Anime.AnilistID,
		}
	}

	artists := make([]ArtistMinimalDTO, 0)
	for _, a := range s.Artists {
		artists = append(artists, ToArtistMinimalDTO(&a))
	}

	res := SongMinimalDTO{
		ID:             s.UUID,
		SongRomaji:     s.SongRomaji,
		SongEN:         s.SongEN,
		SongJP:         s.SongJP,
		Slug:           s.Slug,
		Type:           s.Type,
		AverageRating:  s.AverageRating,
		Artists:        artists,
		Anime:          anime,
		FavoritesCount: s.FavoritesCount,
		Views:          s.Views,
		UserRating:     s.UserRating,
	}

	if s.PrevMainRank != nil {
		val := uint64(*s.PrevMainRank)
		res.PrevMainRank = &val
	}
	if s.PrevSeasonalRank != nil {
		val := uint64(*s.PrevSeasonalRank)
		res.PrevSeasonalRank = &val
	}

	return res
}

func ToSongDTO(s *domain.Song) SongDTO {
	if s == nil {
		return SongDTO{}
	}

	variants := make([]SongVariantDTO, 0)
	for _, v := range s.Variants {
		var videoUrl *string
		if v.Video != nil {
			if v.Video.LocalUrl != nil {
				videoUrl = v.Video.LocalUrl
			} else if v.Video.EmbedUrl != nil {
				videoUrl = v.Video.EmbedUrl
			}
		}
		variants = append(variants, SongVariantDTO{
			ID:            v.UUID,
			VersionNumber: v.VersionNumber,
			Slug:          v.Slug,
			VideoUrl:      videoUrl,
			Spoiler:       v.Spoiler,
		})
	}

	return SongDTO{
		SongMinimalDTO: ToSongMinimalDTO(s),
		LikesCount:     s.LikesCount,
		DislikesCount:  s.DislikesCount,
		IsFavorited:    s.IsFavorited,
		IsLiked:        s.IsLiked,
		IsDisliked:     s.IsDisliked,
		Variants:       variants,
	}
}

func ToArtistMinimalDTO(a *domain.Artist) ArtistMinimalDTO {
	if a == nil {
		return ArtistMinimalDTO{}
	}

	return ArtistMinimalDTO{
		ID:            a.UUID,
		Name:          a.Name,
		NameJP:        a.NameJP,
		Slug:          a.Slug,
		AvatarUrl:     a.AvatarUrl,
		AvatarSources: a.AvatarSources,
		BannerUrl:     a.LatestBannerUrl,
		BannerSources: a.BannerSources,
		EnabledSongs:  a.EnabledSongs,
		DisabledSongs: a.DisabledSongs,
	}
}

func ToArtistDTO(a *domain.Artist) ArtistDTO {
	if a == nil {
		return ArtistDTO{}
	}
	return ArtistDTO{
		ArtistMinimalDTO: ToArtistMinimalDTO(a),
		FavoritesCount:   a.FavoritesCount,
		IsFavorited:      a.IsFavorited,
	}
}

func ToHomeDTO(data *public.HomeData) HomeDTO {
	if data == nil {
		return HomeDTO{}
	}

	recent := make([]SongMinimalDTO, len(data.RecentlyAdded))
	for i, s := range data.RecentlyAdded {
		recent[i] = ToSongMinimalDTO(&s)
	}

	popular := make([]SongMinimalDTO, len(data.MostPopular))
	for i, s := range data.MostPopular {
		popular[i] = ToSongMinimalDTO(&s)
	}

	viewed := make([]SongMinimalDTO, len(data.MostViewed))
	for i, s := range data.MostViewed {
		viewed[i] = ToSongMinimalDTO(&s)
	}

	artists := make([]ArtistMinimalDTO, len(data.FeaturedArtists))
	for i, a := range data.FeaturedArtists {
		artists[i] = ToArtistMinimalDTO(&a)
	}

	op := make([]SongMinimalDTO, len(data.WeaklyRanking.OP))
	for i, s := range data.WeaklyRanking.OP {
		op[i] = ToSongMinimalDTO(&s)
	}

	ed := make([]SongMinimalDTO, len(data.WeaklyRanking.ED))
	for i, s := range data.WeaklyRanking.ED {
		ed[i] = ToSongMinimalDTO(&s)
	}

	var featured *SongDTO
	if data.FeaturedSong != nil {
		f := ToSongDTO(data.FeaturedSong)
		featured = &f
	}

	return HomeDTO{
		FeaturedSong:    featured,
		RecentlyAdded:   recent,
		MostPopular:     popular,
		MostViewed:      viewed,
		FeaturedArtists: artists,
		WeaklyRanking: WeaklyRankingDTO{
			OP: op,
			ED: ed,
		},
	}
}

// ─── Anime Mappers ───

func ToAnimeMinimalDTO(a *domain.Anime) AnimeMinimalDTO {
	if a == nil {
		return AnimeMinimalDTO{}
	}

	var season *SeasonDTO
	if a.Season != nil {
		sDTO := ToSeasonDTO(a.Season)
		season = &sDTO
	}
	var year *YearDTO
	if a.Year != nil {
		yDTO := ToYearDTO(a.Year)
		year = &yDTO
	}
	var format *FormatDTO
	if a.Format != nil {
		fDTO := ToFormatDTO(a.Format)
		format = &fDTO
	}

	return AnimeMinimalDTO{
		AnilistID:     a.AnilistID,
		Title:         a.Title,
		Slug:          a.Slug,
		CoverUrl:      a.CoverUrl,
		CoverSources:  a.CoverSources,
		BannerUrl:     a.BannerUrl,
		BannerSources: a.BannerSources,
		SongsCount:    a.EnabledSongs, // Abstract as only enabled songs for public
		EnabledSongs:  a.EnabledSongs,
		DisabledSongs: a.DisabledSongs,
		Season:        season,
		Year:          year,
		Format:        format,
	}
}

func ToAnimeDTO(a *domain.Anime) AnimeDTO {
	if a == nil {
		return AnimeDTO{}
	}

	studios := make([]StudioDTO, 0)
	for _, s := range a.Studios {
		studios = append(studios, ToStudioDTO(&s))
	}

	producers := make([]ProducerDTO, 0)
	for _, p := range a.Producers {
		producers = append(producers, ToProducerDTO(&p))
	}

	genres := make([]GenreDTO, 0)
	for _, g := range a.Genres {
		genres = append(genres, ToGenreDTO(&g))
	}

	songs := make([]SongMinimalDTO, 0)
	for _, s := range a.Songs {
		songs = append(songs, ToSongMinimalDTO(&s))
	}

	links := make([]ExternalLinkDTO, 0)
	for _, l := range a.ExternalLinks {
		links = append(links, ExternalLinkDTO{Name: l.Name, Type: l.Type, URL: l.URL, Icon: l.Icon})
	}

	return AnimeDTO{
		AnimeMinimalDTO: ToAnimeMinimalDTO(a),
		Description:     a.Description,
		Studios:         studios,
		Producers:       producers,
		Genres:          genres,
		Songs:           songs,
		ExternalLinks:   links,
	}
}

func ToStudioDTO(s *domain.Studio) StudioDTO {
	if s == nil {
		return StudioDTO{}
	}
	return StudioDTO{
		Name:       s.Name,
		Slug:       s.Slug,
		LogoUrl:    s.LogoUrl,
		BannerUrl:  s.BannerUrl,
		AnimeCount: s.AnimeCount,
	}
}

func ToProducerDTO(p *domain.Producer) ProducerDTO {
	if p == nil {
		return ProducerDTO{}
	}
	return ProducerDTO{
		Name:       p.Name,
		Slug:       p.Slug,
		LogoUrl:    p.LogoUrl,
		BannerUrl:  p.BannerUrl,
		AnimeCount: p.AnimeCount,
	}
}

func ToGenreDTO(g *domain.Genre) GenreDTO {
	if g == nil {
		return GenreDTO{}
	}
	return GenreDTO{
		Name: g.Name,
		Slug: g.Slug,
	}
}

func ToYearDTO(y *domain.Year) YearDTO {
	if y == nil {
		return YearDTO{}
	}
	return YearDTO{
		Name: y.Name,
		Slug: y.Slug,
	}
}

func ToSeasonDTO(s *domain.Season) SeasonDTO {
	if s == nil {
		return SeasonDTO{}
	}
	return SeasonDTO{
		Name: s.Name,
		Slug: s.Slug,
	}
}

func ToFormatDTO(f *domain.Format) FormatDTO {
	if f == nil {
		return FormatDTO{}
	}
	return FormatDTO{
		Name: f.Name,
		Slug: f.Slug,
	}
}

// ─── Playlist Mappers ───

func ToPlaylistMinimalDTO(p *domain.Playlist) PlaylistMinimalDTO {
	if p == nil {
		return PlaylistMinimalDTO{}
	}
	return PlaylistMinimalDTO{
		ID:           p.UUID,
		Name:         p.Name,
		Slug:         p.Name,
		BannerUrl:    p.BannerUrl,
		SongCount:    p.SongCount,
		IsPublic:     p.IsPublic,
		ContainsSong: p.ContainsSong,
	}
}

func ToPlaylistDTO(p *domain.Playlist) PlaylistDTO {
	if p == nil {
		return PlaylistDTO{}
	}
	return PlaylistDTO{
		PlaylistMinimalDTO: ToPlaylistMinimalDTO(p),
		Description:        p.Description,
		User:               ToUserMinimalDTO(p.User),
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func ToPlaylistSongDTO(ps *domain.PlaylistSong) SongMinimalDTO {
	if ps == nil || ps.Song == nil {
		return SongMinimalDTO{}
	}
	return ToSongMinimalDTO(ps.Song)
}

// ─── Comment Mappers ───

func ToCommentDTO(c *domain.Comment) CommentDTO {
	if c == nil {
		return CommentDTO{}
	}

	replies := make([]CommentDTO, 0)
	for _, r := range c.Replies {
		replies = append(replies, ToCommentDTO(&r))
	}

	return CommentDTO{
		ID:            c.UUID,
		Content:       c.Content,
		LikesCount:    c.LikesCount,
		DislikesCount: c.DislikesCount,
		RepliesCount:  c.RepliesCount,
		CreatedAt:     c.Created_At,
		UpdatedAt:     c.Updated_At,
		IsLiked:       c.IsLiked,
		IsDisliked:    c.IsDisliked,
		User:          ToUserMinimalDTO(c.User),
		Replies:       replies,
	}
}

func ToActivityItemDTO(item domain.ActivityItem) ActivityItemDTO {
	targetID := ""
	var target interface{}

	// Handle polymorphism to extract UUID or map to DTO
	switch t := item.Target.(type) {
	case *domain.Song:
		if t != nil {
			targetID = t.UUID
			target = ToSongMinimalDTO(t)
		}
	case domain.Song:
		targetID = t.UUID
		target = ToSongMinimalDTO(&t)
	case *domain.Anime:
		if t != nil {
			targetID = t.UUID
			target = ToAnimeMinimalDTO(t)
		}
	case domain.Anime:
		targetID = t.UUID
		target = ToAnimeMinimalDTO(&t)
	case *domain.Comment:
		if t != nil {
			targetID = t.UUID
			target = ToCommentDTO(t)
		}
	case domain.Comment:
		targetID = t.UUID
		target = ToCommentDTO(&t)
	case *domain.Artist:
		if t != nil {
			targetID = t.UUID
			target = ToArtistMinimalDTO(t)
		}
	case domain.Artist:
		targetID = t.UUID
		target = ToArtistMinimalDTO(&t)
	default:
		// Fallback to original target if not handled
		target = item.Target
	}

	return ActivityItemDTO{
		Type:      item.Type,
		User:       ToUserMinimalDTO(&item.User),
		TargetID:   targetID,
		TargetType: item.TargetType,
		Target:     target,
		Value:      item.Value,
		CreatedAt: item.CreatedAt,
	}
}

func ToActivityDTO(item domain.Activity) ActivityItemDTO {
	targetID := ""
	var target interface{}

	// Handle polymorphism from Activity struct
	if item.Song != nil {
		targetID = item.Song.UUID
		target = ToSongMinimalDTO(item.Song)
	} else if item.Artist != nil {
		targetID = item.Artist.UUID
		target = ToArtistMinimalDTO(item.Artist)
	} else if item.UserTarget != nil {
		targetID = item.UserTarget.UUID
		target = ToUserMinimalDTO(item.UserTarget)
	} else if item.Target != nil {
		// Fallback for generic targets
		switch t := item.Target.(type) {
		case *domain.Song:
			targetID = t.UUID
			target = ToSongMinimalDTO(t)
		case *domain.Anime:
			targetID = t.UUID
			target = ToAnimeMinimalDTO(t)
		case *domain.Comment:
			targetID = t.UUID
			target = ToCommentDTO(t)
		case *domain.Artist:
			targetID = t.UUID
			target = ToArtistMinimalDTO(t)
		}
	}

	userDto := UserMinimalDTO{}
	if item.User != nil {
		userDto = ToUserMinimalDTO(item.User)
	}

	return ActivityItemDTO{
		Type:       item.ActionType,
		User:       userDto,
		TargetID:   targetID,
		TargetType: item.TargetType,
		Target:     target,
		Value:      item.ActionValue,
		CreatedAt:  item.CreatedAt.Format(time.RFC3339),
	}
}

// ─── Tournament Mappers ───

func ToTournamentMinimalDTO(t *domain.Tournament) TournamentMinimalDTO {
	if t == nil {
		return TournamentMinimalDTO{}
	}
	return TournamentMinimalDTO{
		ID:        t.UUID,
		Name:      t.Name,
		Slug:      t.Slug,
		Size:      t.Size,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
	}
}

func ToTournamentDTO(t *domain.Tournament) TournamentDTO {
	if t == nil {
		return TournamentDTO{}
	}

	matchups := make([]TournamentMatchupDTO, 0)
	for _, m := range t.Matchups {
		matchups = append(matchups, ToTournamentMatchupDTO(&m))
	}

	return TournamentDTO{
		TournamentMinimalDTO: ToTournamentMinimalDTO(t),
		Description:          t.Description,
		Winner:               nil, // Map if winner exists
		Matchups:             matchups,
	}
}

func ToTournamentMatchupDTO(m *domain.TournamentMatchup) TournamentMatchupDTO {
	if m == nil {
		return TournamentMatchupDTO{}
	}

	var s1, s2, winner *SongMinimalDTO
	if m.Song1 != nil {
		val := ToSongMinimalDTO(m.Song1)
		s1 = &val
	}
	if m.Song2 != nil {
		val := ToSongMinimalDTO(m.Song2)
		s2 = &val
	}
	if m.Winner != nil {
		val := ToSongMinimalDTO(m.Winner)
		winner = &val
	}

	return TournamentMatchupDTO{
		ID:         m.UUID,
		Round:      m.Round,
		Position:   m.Position,
		Song1:      s1,
		Song2:      s2,
		Song1Votes: m.Song1Votes,
		Song2Votes: m.Song2Votes,
		Winner:     winner,
		EndsAt:     m.EndsAt,
		IsActive:   m.IsActive,
	}
}

func ToAnnouncementDTO(a domain.Announcement) AnnouncementDTO {
	return AnnouncementDTO{
		ID:        a.UUID,
		UUID:      a.UUID,
		Title:     a.Title,
		Content:   a.Content,
		Type:      a.Type,
		Icon:      a.Icon,
		URL:       a.URL,
		ImageUrl:  a.ImageUrl,
		ImageSources: a.ImageSources,
		Priority:  a.Priority,
		IsActive:  a.IsActive,
		StartsAt:  a.StartsAt,
		EndsAt:    a.EndsAt,
	}
}

func ToAdminAnnouncementDTO(a domain.Announcement) AnnouncementDTO {
	dto := ToAnnouncementDTO(a)
	dto.ID = a.ID // Sequential numeric ID for Admin Panel
	return dto
}

func ToNotificationDTO(n domain.Notification) NotificationDTO {
	var subjectID *string
	if n.SubjectUUID != nil {
		subjectID = n.SubjectUUID
	}

	return NotificationDTO{
		ID:          n.ID,
		Type:        n.Type,
		SubjectID:   subjectID,
		SubjectType: n.SubjectType,
		Data:        n.Data,
		ReadAt:      n.ReadAt,
		CreatedAt:   n.CreatedAt,
	}
}
