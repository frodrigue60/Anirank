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

	for i := range activities {
		u.enrichActivity(ctx, &activities[i])
	}

	return activities, nil
}

func (u *activityUsecase) enrichActivity(ctx context.Context, activity *domain.Activity) {
	// Populate compatibility Action field
	activity.Action = activity.ActionType

	// Hydrate Author User
	user, err := u.userRepo.GetByID(ctx, activity.UserID)
	if err == nil {
		if user.Avatar != nil {
			user.AvatarUrl = u.mediaService.Resolve(user.Avatar)
		}
		activity.User = user
	}

	// Hydrate Target based on type
	switch activity.TargetType {
	case "song":
		song, err := u.songRepo.GetByID(ctx, activity.TargetID)
		if err == nil {
			if song.Anime != nil {
				song.Anime.CoverUrl = u.mediaService.Resolve(song.Anime.Cover)
			}

			// Set computed fields (Name fallback)
			if song.SongRomaji != nil && *song.SongRomaji != "" {
				song.Name = *song.SongRomaji
			} else if song.SongEN != nil && *song.SongEN != "" {
				song.Name = *song.SongEN
			} else if song.SongJP != nil && *song.SongJP != "" {
				song.Name = *song.SongJP
			} else {
				song.Name = "N/A"
			}

			activity.Target = song
			activity.Song = song
		}
	case "artist":
		artist, err := u.artistRepo.GetByID(ctx, activity.TargetID)
		if err == nil {
			if artist.Avatar != nil {
				artist.AvatarUrl = u.mediaService.Resolve(artist.Avatar)
			}
			activity.Target = artist
			activity.Artist = artist
		}
	case "user":
		targetUser, err := u.userRepo.GetByID(ctx, activity.TargetID)
		if err == nil {
			if targetUser.Avatar != nil {
				targetUser.AvatarUrl = u.mediaService.Resolve(targetUser.Avatar)
			}
			activity.Target = targetUser
			activity.UserTarget = targetUser
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
