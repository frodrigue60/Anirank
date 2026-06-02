package dto

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/public"
	"fmt"
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
			IconSources: b.IconSources,
		})
	}

	roles := make([]string, 0)
	for _, r := range u.Roles {
		roles = append(roles, r.Slug)
	}

	identities := make([]UserSocialIdentityDTO, 0)
	var anilistID *int64
	var anilistUsername *string
	for _, si := range u.SocialIdentities {
		username := ""
		if si.ProviderUsername != nil {
			username = *si.ProviderUsername
		}
		identities = append(identities, UserSocialIdentityDTO{
			Provider:         si.Provider,
			ProviderUsername: username,
		})

		if si.Provider == "anilist" {
			var idVal int64
			if _, err := fmt.Sscanf(si.ProviderID, "%d", &idVal); err == nil {
				anilistID = &idVal
			}
			anilistUsername = si.ProviderUsername
		}
	}

	return UserDTO{
		UserMinimalDTO:   ToUserMinimalDTO(u),
		About:            u.About,
		ProfileColor:     u.ProfileColor,
		FollowersCount:   u.FollowersCount,
		FollowingCount:   u.FollowingCount,
		IsFollowing:      u.IsFollowing,
		TruthScore:       u.TruthScore,
		IsShadowbanned:   u.IsShadowbanned,
		IsSoftbanned:     u.IsSoftbanned,
		Roles:            roles,
		Badges:           badges,
		SocialIdentities: identities,
		AnilistID:        anilistID,
		AnilistUsername:  anilistUsername,
	}
}

func ToAuthUserDTO(u *domain.User) AuthUserDTO {
	if u == nil {
		return AuthUserDTO{}
	}

	return AuthUserDTO{
		UserDTO:         ToUserDTO(u),
		Email:           u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
		ScoreFormat:     u.ScoreFormat,
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

func ToBadgeDTO(b *domain.Badge) BadgeDTO {
	if b == nil {
		return BadgeDTO{}
	}
	return BadgeDTO{
		ID:          b.UUID,
		Name:        b.Name,
		Description: b.Description,
		IconUrl:     b.IconUrl,
		IconSources: b.IconSources,
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
		var animeSeason *SeasonDTO
		if s.Anime.Season != nil {
			sd := ToSeasonDTO(s.Anime.Season)
			animeSeason = &sd
		}
		var animeYear *YearDTO
		if s.Anime.Year != nil {
			yd := ToYearDTO(s.Anime.Year)
			animeYear = &yd
		}

		anime = &SongAnimeDTO{
			Title:         s.Anime.Title,
			TitleEnglish:  s.Anime.TitleEnglish,
			TitleNative:   s.Anime.TitleNative,
			Slug:          s.Anime.Slug,
			CoverUrl:      coverUrl,
			CoverSources:  s.Anime.CoverSources,
			BannerUrl:     s.Anime.BannerUrl,
			BannerSources: s.Anime.BannerSources,
			AnilistID:     s.Anime.AnilistID,
			Season:        animeSeason,
			Year:          animeYear,
		}
	}


	artists := make([]ArtistMinimalDTO, 0)
	for _, a := range s.Artists {
		artists = append(artists, ToArtistMinimalDTO(&a))
	}

	var typeID string
	if s.SongType != nil && s.SongType.UUID != nil {
		typeID = *s.SongType.UUID
	}

	var animeID, yearID, seasonID string
	if s.Anime != nil {
		animeID = s.Anime.UUID
		if s.Anime.Year != nil {
			yearID = s.Anime.Year.UUID
		}
		if s.Anime.Season != nil {
			seasonID = s.Anime.Season.UUID
		}
	}

	// ─── Heritage Fallback Logic (Song Level) ───
	var seasonDTO *SeasonDTO
	var yearDTO *YearDTO

	// 1. Try Song's own season/year
	if s.Season != nil {
		sd := ToSeasonDTO(s.Season)
		seasonDTO = &sd
		seasonID = sd.ID
	}
	if s.Year != nil {
		yd := ToYearDTO(s.Year)
		yearDTO = &yd
		yearID = yd.ID
	}

	// 2. Fallback to Anime's season/year if still nil
	if seasonDTO == nil && s.Anime != nil && s.Anime.Season != nil {
		sd := ToSeasonDTO(s.Anime.Season)
		seasonDTO = &sd
		seasonID = sd.ID
	}
	if yearDTO == nil && s.Anime != nil && s.Anime.Year != nil {
		yd := ToYearDTO(s.Anime.Year)
		yearDTO = &yd
		yearID = yd.ID
	}

	var songType *SongTypeDTO
	if s.SongType != nil {
		st := ToSongTypeDTO(s.SongType)
		songType = &st
	}

	res := SongMinimalDTO{
		ID:             s.UUID,
		Name:           s.Name,
		SongRomaji:     s.SongRomaji,
		SongEN:         s.SongEN,
		SongJP:         s.SongJP,
		Slug:           s.Slug,
		Type:           s.Type,
		ThemeNum:       s.ThemeNum,
		SongType:       songType,
		Season:         seasonDTO,
		Year:           yearDTO,
		AverageRating:  s.AverageRating,
		Artists:        artists,
		Anime:          anime,
		FavoritesCount: s.FavoritesCount,
		Views:          s.Views,
		UserRating:     s.UserRating,
		TypeID:         typeID,
		AnimeID:        animeID,
		YearID:         yearID,
		SeasonID:       seasonID,
	}

	if s.PrevMainRank != nil {
		val := *s.PrevMainRank
		res.PrevMainRank = &val
	}
	if s.PrevSeasonalRank != nil {
		val := *s.PrevSeasonalRank
		res.PrevSeasonalRank = &val
	}
	if s.PrevRank != nil {
		val := *s.PrevRank
		res.PrevRank = &val
	}

	return res
}

func ToSongSlimDTO(s *domain.Song) SongSlimDTO {
	if s == nil {
		return SongSlimDTO{}
	}

	artists := make([]ArtistSlimDTO, 0)
	for _, a := range s.Artists {
		artists = append(artists, ArtistSlimDTO{
			Name: a.Name,
			Slug: a.Slug,
		})
	}

	anime := AnimeSlimDTO{}
	if s.Anime != nil {
		anime.Title = s.Anime.Title
		anime.Slug = s.Anime.Slug
		if s.Anime.CoverUrl != nil {
			anime.CoverUrl = *s.Anime.CoverUrl
		}
		anime.BannerUrl = s.Anime.BannerUrl
	}

	var seasonDTO *SeasonDTO
	var yearDTO *YearDTO

	// 1. Try Song's own season/year
	if s.Season != nil {
		sd := ToSeasonDTO(s.Season)
		seasonDTO = &sd
	}
	if s.Year != nil {
		yd := ToYearDTO(s.Year)
		yearDTO = &yd
	}

	// 2. Fallback to Anime's season/year if still nil
	if seasonDTO == nil && s.Anime != nil && s.Anime.Season != nil {
		sd := ToSeasonDTO(s.Anime.Season)
		seasonDTO = &sd
	}
	if yearDTO == nil && s.Anime != nil && s.Anime.Year != nil {
		yd := ToYearDTO(s.Anime.Year)
		yearDTO = &yd
	}

	res := SongSlimDTO{
		ID:            s.UUID,
		Name:          s.Name,
		Slug:          s.Slug,
		Type:          s.Type,
		AverageRating: s.AverageRating,
		Artists:       artists,
		Anime:         anime,
		Views:         s.Views,
		UserRating:    s.UserRating,
		Season:        seasonDTO,
		Year:          yearDTO,
	}

	if s.PrevMainRank != nil {
		val := *s.PrevMainRank
		res.PrevMainRank = &val
	}
	if s.PrevSeasonalRank != nil {
		val := *s.PrevSeasonalRank
		res.PrevSeasonalRank = &val
	}
	if s.PrevRank != nil {
		val := *s.PrevRank
		res.PrevRank = &val
	}

	return res
}

func ToSongTypeDTO(st *domain.SongType) SongTypeDTO {
	if st == nil {
		return SongTypeDTO{}
	}

	id := ""
	if st.UUID != nil {
		id = *st.UUID
	}
	name := ""
	if st.Name != nil {
		name = *st.Name
	}
	slug := ""
	if st.Slug != nil {
		slug = *st.Slug
	}

	return SongTypeDTO{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: st.Description,
	}
}

func ToSongDTO(s *domain.Song) SongDTO {
	if s == nil {
		return SongDTO{}
	}

	minimalDTO := ToSongMinimalDTO(s)

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

		vDTO := SongVariantDTO{
			ID:            v.UUID,
			VersionNumber: v.VersionNumber,
			Slug:          v.Slug,
			VideoUrl:      videoUrl,
			Episodes:      v.Episodes,
			Spoiler:       v.Spoiler,
			NSFW:          v.NSFW,
			Videos:        []SongVariantVideoDTO{},
		}

		if v.Video != nil {
			vDTO.EmbedCode = v.Video.EmbedCode
			vDTO.VideoSrc = v.Video.VideoSrc
			vDTO.EmbedUrl = v.Video.EmbedUrl
			vDTO.LocalUrl = v.Video.LocalUrl
			vDTO.IsNC = v.Video.IsNC
			vDTO.IsBD = v.Video.IsBD
			vDTO.Resolution = v.Video.Resolution
			vDTO.IsUncensored = v.Video.IsUncensored
			vDTO.IsSubbed = v.Video.IsSubbed
			vDTO.IsLyrics = v.Video.IsLyrics
			vDTO.Source = v.Video.Source
			vDTO.Overlap = v.Video.Overlap
		}

		for _, vid := range v.Videos {
			var vidUrl *string
			if vid.LocalUrl != nil {
				vidUrl = vid.LocalUrl
			} else if vid.EmbedUrl != nil {
				vidUrl = vid.EmbedUrl
			}

			vDTO.Videos = append(vDTO.Videos, SongVariantVideoDTO{
				VideoUrl:     vidUrl,
				EmbedUrl:     vid.EmbedUrl,
				LocalUrl:     vid.LocalUrl,
				EmbedCode:    vid.EmbedCode,
				VideoSrc:     vid.VideoSrc,
				IsNC:         vid.IsNC,
				IsBD:         vid.IsBD,
				Resolution:   vid.Resolution,
				IsUncensored: vid.IsUncensored,
				IsSubbed:     vid.IsSubbed,
				IsLyrics:     vid.IsLyrics,
				Source:       vid.Source,
				Overlap:      vid.Overlap,
			})
		}

		// ─── Heritage Fallback Logic (Variant Level) ───
		// 1. Try Variant's own season/year
		if v.Season != nil {
			sd := ToSeasonDTO(v.Season)
			vDTO.Season = &sd
		}
		if v.Year != nil {
			yd := ToYearDTO(v.Year)
			vDTO.Year = &yd
		}

		// 2. Fallback to parent Song's resolved season/year if still nil
		if vDTO.Season == nil {
			vDTO.Season = minimalDTO.Season
		}
		if vDTO.Year == nil {
			vDTO.Year = minimalDTO.Year
		}

		variants = append(variants, vDTO)
	}

	return SongDTO{
		SongMinimalDTO: minimalDTO,
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
			ID:            a.UUID,
			AnilistID:     a.AnilistID,
			Title:         a.Title,
			TitleEnglish:  a.TitleEnglish,
			TitleNative:   a.TitleNative,
			Synonyms:      []string(a.Synonyms),
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
		links = append(links, ExternalLinkDTO{
			ID:   l.UUID,
			Name: l.Name,
			Type: l.Type,
			URL:  l.URL,
			Icon: l.Icon,
		})
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
		ID:         s.UUID,
		Name:       s.Name,
		Slug:       s.Slug,
		LogoUrl:       s.LogoUrl,
		LogoSources:   s.LogoSources,
		BannerUrl:     s.BannerUrl,
		BannerSources: s.BannerSources,
		AnimeCount:    s.AnimeCount,
	}
}

func ToProducerDTO(p *domain.Producer) ProducerDTO {
	if p == nil {
		return ProducerDTO{}
	}
	return ProducerDTO{
		ID:         p.UUID,
		Name:       p.Name,
		Slug:       p.Slug,
		LogoUrl:       p.LogoUrl,
		LogoSources:   p.LogoSources,
		BannerUrl:     p.BannerUrl,
		BannerSources: p.BannerSources,
		AnimeCount:    p.AnimeCount,
	}
}

func ToGenreDTO(g *domain.Genre) GenreDTO {
	if g == nil {
		return GenreDTO{}
	}
	return GenreDTO{
		ID:   g.UUID,
		Name: g.Name,
		Slug: g.Slug,
	}
}

func ToYearDTO(y *domain.Year) YearDTO {
	if y == nil {
		return YearDTO{}
	}
	return YearDTO{
		ID:      y.UUID,
		Name:    y.Name,
		Slug:    y.Slug,
		Current: y.Current,
	}
}

func ToSeasonDTO(s *domain.Season) SeasonDTO {
	if s == nil {
		return SeasonDTO{}
	}
	return SeasonDTO{
		ID:      s.UUID,
		Name:    s.Name,
		Slug:    s.Slug,
		Current: s.Current,
	}
}

func ToFormatDTO(f *domain.Format) FormatDTO {
	if f == nil {
		return FormatDTO{}
	}
	return FormatDTO{
		ID:   f.UUID,
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
		BannerUrl:     p.BannerUrl,
		BannerSources: p.BannerSources,
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
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
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
	case *domain.Badge:
		if t != nil {
			targetID = t.UUID
			target = ToBadgeDTO(t)
		}
	case domain.Badge:
		targetID = t.UUID
		target = ToBadgeDTO(&t)
	case uint64:
		targetID = fmt.Sprintf("%d", t)
		target = t
	default:
		// Fallback to original target if not handled
		target = item.Target
	}

	var badgeDto *BadgeDTO
	if item.Badge != nil {
		dto := ToBadgeDTO(item.Badge)
		badgeDto = &dto
		if targetID == "" {
			targetID = item.Badge.UUID
		}
	}

	return ActivityItemDTO{
		Type:       item.Type,
		User:       ToUserMinimalDTO(&item.User),
		TargetID:   targetID,
		TargetType: item.TargetType,
		Target:     target,
		Badge:      badgeDto,
		Value:      item.Value,
		CreatedAt:  item.CreatedAt,
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
		case *domain.Badge:
			targetID = t.UUID
			target = ToBadgeDTO(t)
		case uint64:
			targetID = fmt.Sprintf("%d", t)
			target = t
		}
	}

	userDto := UserMinimalDTO{}
	if item.User != nil {
		userDto = ToUserMinimalDTO(item.User)
	}

	var badgeDto *BadgeDTO
	if item.Badge != nil {
		dto := ToBadgeDTO(item.Badge)
		badgeDto = &dto
		
		// If target is empty, set it to the badge UUID for consistent TargetID
		if targetID == "" {
			targetID = item.Badge.UUID
		}
	}

	return ActivityItemDTO{
		Type:       item.ActionType,
		User:       userDto,
		TargetID:   targetID,
		TargetType: item.TargetType,
		Target:     target,
		Badge:      badgeDto,
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
		CurrentRound: t.CurrentRound,
		MatchupDurationHours: t.MatchupDurationHours,
		CreatedAt: t.CreatedAt,
	}
}

func ToAdminTournamentMinimalDTO(t *domain.Tournament) TournamentMinimalDTO {
	dto := ToTournamentMinimalDTO(t)
	dto.ID = t.ID
	return dto
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

func ToAdminTournamentDTO(t *domain.Tournament) TournamentDTO {
	dto := ToTournamentDTO(t)
	dto.ID = t.ID
	return dto
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

	var s1Id, s2Id, vId *any
	if m.Song1 != nil {
		val := any(m.Song1.UUID)
		s1Id = &val
	}
	if m.Song2 != nil {
		val := any(m.Song2.UUID)
		s2Id = &val
	}
	if m.UserVotedSongID != nil {
		if m.Song1 != nil && *m.UserVotedSongID == m.Song1.ID {
			val := any(m.Song1.UUID)
			vId = &val
		} else if m.Song2 != nil && *m.UserVotedSongID == m.Song2.ID {
			val := any(m.Song2.UUID)
			vId = &val
		}
	}

	id := m.UUID
	if id == "" {
		id = fmt.Sprintf("%d", m.ID)
	}

	return TournamentMatchupDTO{
		ID:              id,
		Round:           m.Round,
		Position:        m.Position,
		Song1:           s1,
		Song2:           s2,
		Song1ID:         s1Id,
		Song2ID:         s2Id,
		Song1Votes:      m.Song1Votes,
		Song2Votes:      m.Song2Votes,
		Winner:          winner,
		EndsAt:          m.EndsAt,
		IsActive:        m.IsActive,
		UserVotedSongID: vId,
	}
}

func ToAnnouncementDTO(a domain.Announcement) AnnouncementDTO {
	return AnnouncementDTO{
		ID:           a.UUID,
		UUID:         a.UUID,
		Title:        a.Title,
		Content:      a.Content,
		Type:         a.Type,
		Icon:         a.Icon,
		URL:          a.URL,
		ImageUrl:     a.ImageUrl,
		ImageSources: a.ImageSources,
		Priority:     a.Priority,
		IsActive:     a.IsActive,
		StartsAt:     a.StartsAt,
		EndsAt:       a.EndsAt,
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
