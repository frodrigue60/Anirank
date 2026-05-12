package usecase

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"context"
)

type activityUsecase struct {
	activityRepo domain.ActivityRepository
	userRepo     domain.UserRepository
	songRepo     domain.SongRepository
	artistRepo   domain.ArtistRepository
	mediaService infrastructure.MediaService
}

func NewActivityUsecase(
	a domain.ActivityRepository,
	u domain.UserRepository,
	s domain.SongRepository,
	ar domain.ArtistRepository,
	media infrastructure.MediaService,
) domain.ActivityUsecase {
	return &activityUsecase{
		activityRepo: a,
		userRepo:     u,
		songRepo:     s,
		artistRepo:   ar,
		mediaService: media,
	}
}

func (u *activityUsecase) GetFeed(ctx context.Context, limit, offset int) ([]domain.Activity, error) {
	activities, err := u.activityRepo.GetPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	u.enrichActivitiesBatch(ctx, activities)

	return activities, nil
}

func (u *activityUsecase) GetCount(ctx context.Context) (int, error) {
	return u.activityRepo.Count(ctx)
}

func (u *activityUsecase) enrichActivitiesBatch(ctx context.Context, activities []domain.Activity) {
	if len(activities) == 0 {
		return
	}

	userIDs := make(map[uint64]bool)
	songIDs := make(map[uint64]bool)
	artistIDs := make(map[uint64]bool)
	targetUserIDs := make(map[uint64]bool)

	for _, a := range activities {
		userIDs[a.UserID] = true
		switch a.TargetType {
		case "song":
			songIDs[a.TargetID] = true
		case "artist":
			artistIDs[a.TargetID] = true
		case "user":
			targetUserIDs[a.TargetID] = true
		}
	}

	// Map to store fetched entities
	usersMap := make(map[uint64]*domain.User)
	if len(userIDs) > 0 {
		ids := make([]uint64, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}
		users, _ := u.userRepo.GetMany(ctx, ids)
		for i := range users {
			users[i].AvatarUrl = u.mediaService.Resolve(users[i].Avatar)
			if users[i].Avatar != nil {
				users[i].AvatarSources = u.mediaService.GetImageSources(*users[i].Avatar)
			}
			usersMap[users[i].ID] = &users[i]
		}
	}

	songsMap := make(map[uint64]*domain.Song)
	if len(songIDs) > 0 {
		ids := make([]uint64, 0, len(songIDs))
		for id := range songIDs {
			ids = append(ids, id)
		}
		songs, _ := u.songRepo.GetMany(ctx, ids)
		for i := range songs {
			if songs[i].Anime != nil {
				songs[i].Anime.CoverUrl = u.mediaService.Resolve(songs[i].Anime.Cover)
				if songs[i].Anime.Cover != nil {
					songs[i].Anime.CoverSources = u.mediaService.GetImageSources(*songs[i].Anime.Cover)
				}
				songs[i].Anime.BannerUrl = u.mediaService.Resolve(songs[i].Anime.Banner)
				if songs[i].Anime.Banner != nil {
					songs[i].Anime.BannerSources = u.mediaService.GetImageSources(*songs[i].Anime.Banner)
				}
			}
			// Set computed fields (fallback logic)
			if songs[i].SongRomaji != nil && *songs[i].SongRomaji != "" {
				songs[i].Name = *songs[i].SongRomaji
			} else if songs[i].SongEN != nil && *songs[i].SongEN != "" {
				songs[i].Name = *songs[i].SongEN
			} else if songs[i].SongJP != nil && *songs[i].SongJP != "" {
				songs[i].Name = *songs[i].SongJP
			} else {
				songs[i].Name = "N/A"
			}
			songsMap[songs[i].ID] = &songs[i]
		}
	}

	artistsMap := make(map[uint64]*domain.Artist)
	if len(artistIDs) > 0 {
		ids := make([]uint64, 0, len(artistIDs))
		for id := range artistIDs {
			ids = append(ids, id)
		}
		artists, _ := u.artistRepo.GetMany(ctx, ids)
		for i := range artists {
			artists[i].AvatarUrl = u.mediaService.Resolve(artists[i].Avatar)
			if artists[i].Avatar != nil {
				artists[i].AvatarSources = u.mediaService.GetImageSources(*artists[i].Avatar)
			}
			artistsMap[artists[i].ID] = &artists[i]
		}
	}

	targetUsersMap := make(map[uint64]*domain.User)
	if len(targetUserIDs) > 0 {
		ids := make([]uint64, 0, len(targetUserIDs))
		for id := range targetUserIDs {
			ids = append(ids, id)
		}
		tUsers, _ := u.userRepo.GetMany(ctx, ids)
		for i := range tUsers {
			tUsers[i].AvatarUrl = u.mediaService.Resolve(tUsers[i].Avatar)
			if tUsers[i].Avatar != nil {
				tUsers[i].AvatarSources = u.mediaService.GetImageSources(*tUsers[i].Avatar)
			}
			targetUsersMap[tUsers[i].ID] = &tUsers[i]
		}
	}

	// Final pass to assign relations
	for i := range activities {
		activities[i].Action = activities[i].ActionType
		activities[i].User = usersMap[activities[i].UserID]

		switch activities[i].TargetType {
		case "song":
			if song, ok := songsMap[activities[i].TargetID]; ok {
				activities[i].Target = song
				activities[i].Song = song
			}
		case "artist":
			if artist, ok := artistsMap[activities[i].TargetID]; ok {
				activities[i].Target = artist
				activities[i].Artist = artist
			}
		case "user":
			if targetUser, ok := targetUsersMap[activities[i].TargetID]; ok {
				activities[i].Target = targetUser
				activities[i].UserTarget = targetUser
			}
		}
	}
}

func (u *activityUsecase) LogActivity(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string, value *string) error {
	// Check if already exists (Uniqueness constraint: per user-target-action)
	exists, err := u.activityRepo.Exists(ctx, userID, actionType, targetID, targetType)
	if err != nil {
		return err
	}
	if exists {
		// Activity already logged for this action-target pair, skip
		return nil
	}

	activity := &domain.Activity{
		UserID:      userID,
		ActionType:  actionType,
		TargetID:    targetID,
		TargetType:  targetType,
		ActionValue: value,
	}
	return u.activityRepo.Create(ctx, activity)
}
