package dto

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/public"
)

// ─── User Mappers ───

func ToUserMinimalDTO(u *domain.User) UserMinimalDTO {
	if u == nil {
		return UserMinimalDTO{}
	}
	return UserMinimalDTO{
		ID:        u.ID,
		Name:      u.Name,
		Slug:      u.Slug,
		AvatarUrl: u.AvatarUrl,
		XP:        u.XP,
		Level:     u.Level,
	}
}

func ToUserDTO(u *domain.User) UserDTO {
	if u == nil {
		return UserDTO{}
	}

	badges := make([]BadgeDTO, 0)
	for _, b := range u.Badges {
		badges = append(badges, BadgeDTO{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			IconUrl:     b.IconUrl,
		})
	}

	return UserDTO{
		UserMinimalDTO: ToUserMinimalDTO(u),
		BannerUrl:      u.BannerUrl,
		About:          u.About,
		ProfileColor:   u.ProfileColor,
		FollowersCount: u.FollowersCount,
		FollowingCount: u.FollowingCount,
		RatingsCount:   u.RatingsCount,
		IsFollowing:    u.IsFollowing,
		Badges:         badges,
		CreatedAt:      u.CreatedAt,
	}
}

func ToAuthUserDTO(u *domain.User) AuthUserDTO {
	if u == nil {
		return AuthUserDTO{}
	}

	roles := make([]string, 0)
	for _, r := range u.Roles {
		roles = append(roles, r.Slug)
	}

	return AuthUserDTO{
		UserDTO:       ToUserDTO(u),
		Email:         u.Email,
		ScoreFormatID: u.ScoreFormatID,
		ScoreFormat:   u.ScoreFormat,
		Roles:         roles,
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
		thumbnailUrl := ""
		if s.Anime.CoverUrl != nil {
			thumbnailUrl = *s.Anime.CoverUrl
		}
		anime = &SongAnimeDTO{
			Title:        s.Anime.Title,
			Slug:         s.Anime.Slug,
			ThumbnailUrl: thumbnailUrl,
			BannerUrl:    s.Anime.BannerUrl,
		}
	}

	artists := make([]ArtistMinimalDTO, 0)
	for _, a := range s.Artists {
		artists = append(artists, ToArtistMinimalDTO(&a))
	}

	res := SongMinimalDTO{
		ID:             s.ID,
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
			ID:            v.ID,
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
		ID:            a.ID,
		Name:          a.Name,
		NameJP:        a.NameJP,
		Slug:          a.Slug,
		AvatarUrl:     a.AvatarUrl,
		SongsCount:    a.SongsCount,
		EnabledSongs:  a.EnabledSongs,
		DisabledSongs: a.DisabledSongs,
		Status:        a.Status,
	}
}

func ToArtistDTO(a *domain.Artist) ArtistDTO {
	if a == nil {
		return ArtistDTO{}
	}
	return ArtistDTO{
		ArtistMinimalDTO: ToArtistMinimalDTO(a),
		BannerUrl:        a.LatestBannerUrl,
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

	var season, year, format *string
	if a.Season != nil {
		season = &a.Season.Name
	}
	if a.Year != nil {
		year = &a.Year.Name
	}
	if a.Format != nil {
		format = &a.Format.Name
	}

	return AnimeMinimalDTO{
		ID:           a.ID,
		Title:        a.Title,
		Slug:         a.Slug,
		ThumbnailUrl: a.CoverUrl,
		BannerUrl:    a.BannerUrl,
		SongsCount:   a.SongsCount,
		Season:       season,
		Year:         year,
		Format:       format,
	}
}

func ToAnimeDTO(a *domain.Anime) AnimeDTO {
	if a == nil {
		return AnimeDTO{}
	}

	studios := make([]StudioDTO, 0)
	for _, s := range a.Studios {
		studios = append(studios, StudioDTO{ID: s.ID, Name: s.Name, Slug: s.Slug})
	}

	producers := make([]ProducerDTO, 0)
	for _, p := range a.Producers {
		producers = append(producers, ProducerDTO{ID: p.ID, Name: p.Name, Slug: p.Slug})
	}

	genres := make([]GenreDTO, 0)
	for _, g := range a.Genres {
		genres = append(genres, GenreDTO{ID: g.ID, Name: g.Name, Slug: g.Slug})
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
		ExternalLinks:   links,
	}
}
