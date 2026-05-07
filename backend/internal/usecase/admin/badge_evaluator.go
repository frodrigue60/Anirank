package admin

import (
	"context"

	"anirank/api/internal/domain"
)

// BadgeEvaluator defines the Strategy interface for determining if a user
// qualifies for an automatic badge based on a required threshold value.
//
// Each implementation encapsulates a single evaluation rule, making it trivial
// to add new badge types without modifying CheckAndAwardBadges.
type BadgeEvaluator interface {
	CanAward(ctx context.Context, userID uint64, requiredValue int) (bool, error)
}

// levelEvaluator awards the badge when the user's current level meets the threshold.
type levelEvaluator struct {
	userRepo domain.UserRepository
}

func (e *levelEvaluator) CanAward(ctx context.Context, userID uint64, requiredValue int) (bool, error) {
	user, err := e.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return int(user.Level) >= requiredValue, nil
}

// ratingsEvaluator awards the badge when the user has submitted at least N ratings.
type ratingsEvaluator struct {
	interactionRepo domain.InteractionRepository
}

func (e *ratingsEvaluator) CanAward(ctx context.Context, userID uint64, requiredValue int) (bool, error) {
	count, err := e.interactionRepo.CountRatingsByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return count >= requiredValue, nil
}

// commentsEvaluator awards the badge when the user has posted at least N comments.
type commentsEvaluator struct {
	commentRepo domain.CommentRepository
}

func (e *commentsEvaluator) CanAward(ctx context.Context, userID uint64, requiredValue int) (bool, error) {
	count, err := e.commentRepo.GetCountByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return count >= requiredValue, nil
}

// anilistEvaluator awards the badge when the user has linked their AniList account.
// requiredValue is unused — linking is a binary condition.
type anilistEvaluator struct {
	userRepo domain.UserRepository
}

func (e *anilistEvaluator) CanAward(ctx context.Context, userID uint64, _ int) (bool, error) {
	user, err := e.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.GetSocialID("anilist") != nil, nil
}

// buildEvaluators returns the registered map of triggerType → BadgeEvaluator.
// To add a new badge type, simply add a new entry here — no other code changes required.
func buildEvaluators(
	userRepo domain.UserRepository,
	interactionRepo domain.InteractionRepository,
	commentRepo domain.CommentRepository,
) map[string]BadgeEvaluator {
	return map[string]BadgeEvaluator{
		"level":    &levelEvaluator{userRepo: userRepo},
		"ratings":  &ratingsEvaluator{interactionRepo: interactionRepo},
		"comments": &commentsEvaluator{commentRepo: commentRepo},
		"anilist":  &anilistEvaluator{userRepo: userRepo},
	}
}
