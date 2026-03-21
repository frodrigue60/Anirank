package usecase

import (
	"anirank/api/internal/domain"
	"context"
)

type activityUsecase struct {
	activityRepo domain.ActivityRepository
	userRepo     domain.UserRepository
	songRepo     domain.SongRepository
	artistRepo   domain.ArtistRepository
}

func NewActivityUsecase(
	a domain.ActivityRepository,
	u domain.UserRepository,
	s domain.SongRepository,
	ar domain.ArtistRepository,
) domain.ActivityUsecase {
	return &activityUsecase{
		activityRepo: a,
		userRepo:     u,
		songRepo:     s,
		artistRepo:   ar,
	}
}

func (u *activityUsecase) GetFeed(ctx context.Context, limit, offset int) ([]domain.Activity, error) {
	activities, err := u.activityRepo.GetPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range activities {
		// Populate compatibility Action field
		activities[i].Action = activities[i].ActionType

		// Hydrate User
		user, err := u.userRepo.GetByID(ctx, activities[i].UserID)
		if err == nil {
			activities[i].User = user
		}

		// Hydrate Target based on type
		switch activities[i].TargetType {
		case "song":
			song, err := u.songRepo.GetByID(ctx, activities[i].TargetID)
			if err == nil {
				activities[i].Target = song
				activities[i].Song = song // Virtual field
			}
		case "artist":
			artist, err := u.artistRepo.GetByID(ctx, activities[i].TargetID)
			if err == nil {
				activities[i].Target = artist
				activities[i].Artist = artist // Virtual field
			}
		}
	}

	return activities, nil
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
