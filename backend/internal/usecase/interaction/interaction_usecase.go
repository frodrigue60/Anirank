package interaction

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/security"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type InteractionUsecase struct {
	interactionRepo  domain.InteractionRepository
	commentRepo      domain.CommentRepository
	userRepo         domain.UserRepository
	notificationRepo domain.NotificationRepository
	songRepo         domain.SongRepository
	animeRepo        domain.AnimeRepository
	mediaService     infrastructure.MediaService
	xpUsecase        domain.XPUsecase
	activityUsecase  domain.ActivityUsecase
	badgeUsecase     domain.BadgeUsecase
	artistRepo       domain.ArtistRepository
}

func NewInteractionUsecase(
	ir domain.InteractionRepository,
	cr domain.CommentRepository,
	ur domain.UserRepository,
	nr domain.NotificationRepository,
	sr domain.SongRepository,
	ar domain.AnimeRepository,
	atr domain.ArtistRepository,
	ms infrastructure.MediaService,
	xu domain.XPUsecase,
	au domain.ActivityUsecase,
	bu domain.BadgeUsecase,
) *InteractionUsecase {
	return &InteractionUsecase{
		interactionRepo:  ir,
		commentRepo:      cr,
		userRepo:         ur,
		notificationRepo: nr,
		songRepo:         sr,
		animeRepo:        ar,
		artistRepo:       atr,
		mediaService:     ms,
		xpUsecase:        xu,
		activityUsecase:  au,
		badgeUsecase:     bu,
	}
}

// ---- RATINGS (Song-only) ----

const ratingSongType = "App\\Models\\Song"

func (u *InteractionUsecase) RateSong(ctx context.Context, userID, songID uint64, score float64) (float64, float64, error) {
	if songID == 0 {
		return 0, 0, domain.NewAppError(400, "song_id is required and cannot be 0", nil)
	}

	if score < 0 || score > 100 {
		return 0, 0, domain.NewAppError(400, "Rating score must be between 0 and 100", nil)
	}

	// The frontend always sends the raw 0-100 score, so no normalization is needed.
	rating := &domain.Rating{
		Rating: score,
		SongID: songID,
		UserID: userID,
	}

	if err := u.interactionRepo.UpsertRating(ctx, rating); err != nil {
		return 0, 0, err
	}

	avg, err := u.interactionRepo.GetAverageRating(ctx, songID)
	if err == nil {
		_ = u.xpUsecase.AwardXP(ctx, userID, "rate_song", map[string]interface{}{"song_id": songID})
		
		// Log Activity
		scoreStr := strconv.FormatFloat(score, 'f', 1, 64)
		_ = u.activityUsecase.LogActivity(ctx, userID, "rate", songID, "song", &scoreStr)

		// 5. Automatic Badge Check
		_ = u.badgeUsecase.CheckAndAwardBadges(ctx, userID, "ratings")
	}
	return avg, score, err
}

func (u *InteractionUsecase) GetUserSongRating(ctx context.Context, userID, songID uint64) (*domain.Rating, error) {
	return u.interactionRepo.GetRatingByUser(ctx, userID, songID)
}

func (u *InteractionUsecase) GetSongAverageRating(ctx context.Context, songID uint64) (float64, error) {
	return u.interactionRepo.GetAverageRating(ctx, songID)
}

// ---- REACTIONS ----

func (u *InteractionUsecase) ReactToEntity(ctx context.Context, userID, entityID uint64, entityType string, reactionType int8) error {
	if reactionType != 1 && reactionType != -1 && reactionType != 0 {
		return domain.NewAppError(400, "Invalid reaction type. Dislike = -1, Like = 1, None = 0", nil)
	}

	reaction := &domain.Reaction{
		UserID:        userID,
		ReactableID:   entityID,
		ReactableType: entityType,
		Type:          reactionType,
	}

	return u.interactionRepo.ToggleReaction(ctx, reaction)
}

func (u *InteractionUsecase) ReactToSong(ctx context.Context, userID, songID uint64, reactionType int8) (int, int, error) {
	if songID == 0 {
		return 0, 0, domain.NewAppError(400, "song_id is required", nil)
	}
	return u.interactionRepo.UpsertSongReaction(ctx, userID, songID, reactionType)
}

func (u *InteractionUsecase) ReactToComment(ctx context.Context, userID, commentID uint64, reactionType int8) (int, int, error) {
	if commentID == 0 {
		return 0, 0, domain.NewAppError(400, "comment_id is required", nil)
	}
	return u.interactionRepo.UpsertCommentReaction(ctx, userID, commentID, reactionType)
}

func (u *InteractionUsecase) GetEntityCounters(ctx context.Context, entityID uint64, entityType string) (*domain.ReactionCounter, error) {
	return u.interactionRepo.GetCounters(ctx, entityID, entityType)
}

func (u *InteractionUsecase) ToggleFavorite(ctx context.Context, userID, entityID uint64, entityType string) (bool, error) {
	fav := &domain.Favorite{
		UserID:          userID,
		FavoritableID:   entityID,
		FavoritableType: entityType,
	}
	wasFavorited, err := u.interactionRepo.ToggleFavorite(ctx, fav)
	if err == nil && wasFavorited {
		_ = u.xpUsecase.AwardXP(ctx, userID, "add_favorite", map[string]interface{}{"song_id": entityID})
		
		// Log Activity
		_ = u.activityUsecase.LogActivity(ctx, userID, "favorite", entityID, entityType, nil)
	}
	return wasFavorited, err
}

func (u *InteractionUsecase) CheckIsFavorited(ctx context.Context, userID, entityID uint64, entityType string) (bool, error) {
	return u.interactionRepo.IsFavoritedByUser(ctx, userID, entityID, entityType)
}

// ---- COMMENTS ----

func (u *InteractionUsecase) GetCommentsByEntity(ctx context.Context, userID *uint64, entityID uint64, entityType string, limit, offset int) ([]domain.Comment, int, error) {
	comments, err := u.commentRepo.GetByEntity(ctx, userID, entityID, entityType, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, _ := u.commentRepo.GetCountByEntity(ctx, entityID, entityType)

	// Fetch replies for each root comment
	for i := range comments {
		replies, err := u.commentRepo.GetReplies(ctx, userID, comments[i].ID, 50, 0)
		if err == nil {
			comments[i].Replies = replies
		}
	}

	u.enrichComments(ctx, comments)
	return comments, total, nil
}

func (u *InteractionUsecase) GetCommentReplies(ctx context.Context, userID *uint64, parentID uint64, limit, offset int) ([]domain.Comment, int, error) {
	comments, err := u.commentRepo.GetReplies(ctx, userID, parentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, _ := u.commentRepo.GetRepliesCount(ctx, parentID)
	u.enrichComments(ctx, comments)
	return comments, total, nil
}

func (u *InteractionUsecase) SongComment(ctx context.Context, userID, entityID uint64, entityType, content string, parentID *uint64) (*domain.Comment, error) {
	if len(content) < 2 {
		return nil, domain.NewAppError(400, "Comment content is too short", nil)
	}

	comment := &domain.Comment{
		ParentID: parentID,
		UserID:   userID,
		Content:  security.SanitizeStrict(content),
	}

	if entityType == "song" || entityType == "App\\Models\\Song" {
		comment.SongID = &entityID
	}

	err := u.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, domain.NewAppError(500, "Could not post comment", err)
	}

	// Enrich with user data for immediate frontend display
	user, _ := u.userRepo.GetByID(ctx, userID)
	if user != nil {
		comment.User = user
	}
	
	enrichedComments := []domain.Comment{*comment}
	u.enrichComments(ctx, enrichedComments)
	*comment = enrichedComments[0]

	// Trigger Notification for replies
	if parentID != nil {
		parentComment, err := u.commentRepo.GetByID(ctx, *parentID)
		if err == nil && parentComment != nil && parentComment.UserID != userID {
			subjectType := "comment"
			publicURL := os.Getenv("S3_PUBLIC_URL")
			var avatarUrl, bannerUrl *string
			if user.Avatar != nil {
				if strings.HasPrefix(*user.Avatar, "http") {
					avatarUrl = user.Avatar
				} else {
					u := publicURL + "/" + *user.Avatar
					avatarUrl = &u
				}
			}
			if user.Banner != nil {
				if strings.HasPrefix(*user.Banner, "http") {
					bannerUrl = user.Banner
				} else {
					u := publicURL + "/" + *user.Banner
					bannerUrl = &u
				}
			}

			dataObj := map[string]interface{}{
				"replied_by_name":   user.Name,
				"replied_by_avatar": avatarUrl,
				"replied_by_banner": bannerUrl,
				"comment_id":        comment.UUID,
			}

			// Add Entity Data (Anime/Song)
			if parentComment.SongID != nil {
				song, _ := u.songRepo.GetByID(ctx, *parentComment.SongID)
				if song != nil {
					anime, _ := u.animeRepo.GetByID(ctx, song.AnimeID)
					if anime != nil {
						dataObj["anime_name"] = anime.Title
						dataObj["anime_slug"] = anime.Slug
						dataObj["anime_cover"] = u.mediaService.Resolve(anime.Cover)
					}
					dataObj["song_name"] = song.SongRomaji
					dataObj["song_slug"] = song.Slug
				}
			}
			dataJSON, _ := json.Marshal(dataObj)

			notif := &domain.Notification{
				UserID:      parentComment.UserID,
				Type:        "comment_reply",
				SubjectID:   &comment.ID,
				SubjectUUID: &comment.UUID,
				SubjectType: &subjectType,
				Data:        dataJSON,
			}
			_ = u.notificationRepo.Create(ctx, notif)
		}
	}

	activityKey := "comment"
	if parentID != nil {
		activityKey = "reply"
	}
	_ = u.xpUsecase.AwardXP(ctx, userID, activityKey, map[string]interface{}{"comment_id": comment.ID})

	targetID := entityID
	targetType := entityType
	_ = u.activityUsecase.LogActivity(ctx, userID, activityKey, targetID, targetType, nil)

	// Automatic Badge Check
	_ = u.badgeUsecase.CheckAndAwardBadges(ctx, userID, "comments")

	return comment, nil
}

func (u *InteractionUsecase) enrichComments(ctx context.Context, comments []domain.Comment) {
	publicURL := os.Getenv("S3_PUBLIC_URL")
	for i := range comments {
		if comments[i].User != nil {
			if comments[i].User.Avatar != nil {
				if strings.HasPrefix(*comments[i].User.Avatar, "http") {
					comments[i].User.AvatarUrl = comments[i].User.Avatar
				} else {
					url := publicURL + "/" + *comments[i].User.Avatar
					comments[i].User.AvatarUrl = &url
				}
			}
			if comments[i].User.Banner != nil {
				if strings.HasPrefix(*comments[i].User.Banner, "http") {
					comments[i].User.BannerUrl = comments[i].User.Banner
				} else {
					url := publicURL + "/" + *comments[i].User.Banner
					comments[i].User.BannerUrl = &url
				}
			}

			// Load Badges
			badges, err := u.userRepo.GetBadgesByUserID(ctx, comments[i].User.ID)
			if err == nil {
				for j := range badges {
					if badges[j].Icon != nil {
						if strings.HasPrefix(*badges[j].Icon, "http") {
							badges[j].IconUrl = badges[j].Icon
						} else {
							badgeURL := publicURL + "/" + *badges[j].Icon
							badges[j].IconUrl = &badgeURL
						}
					}
				}
				comments[i].User.Badges = badges
			}
		}
		if len(comments[i].Replies) > 0 {
			u.enrichComments(ctx, comments[i].Replies)
		}
	}
}

func (u *InteractionUsecase) DeleteComment(ctx context.Context, commentID, userID uint64) error {
	err := u.commentRepo.Delete(ctx, commentID, userID)
	if err != nil {
		return domain.NewAppError(403, "You cannot delete this comment", err)
	}
	return nil
}

// ---- ACTIVITY FEED ----

func (u *InteractionUsecase) GetGlobalActivity(ctx context.Context, limit int) ([]domain.ActivityItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	activities, err := u.interactionRepo.GetRecentActivities(ctx, limit)
	if err != nil {
		return nil, err
	}

	u.enrichActivities(ctx, activities)
	return activities, nil
}

func (u *InteractionUsecase) enrichActivities(ctx context.Context, items []domain.ActivityItem) {
	// 1. Collect IDs by type to batch fetch
	songIDs := make([]uint64, 0)
	artistIDs := make([]uint64, 0)
	userIDs := make([]uint64, 0)

	for _, item := range items {
		userIDs = append(userIDs, item.UserID)
		switch item.TargetType {
		case "song":
			songIDs = append(songIDs, item.TargetID)
		case "artist":
			artistIDs = append(artistIDs, item.TargetID)
		}
	}

	// 2. Fetch data (batch)
	songs := make(map[uint64]domain.Song)
	if len(songIDs) > 0 {
		if s, err := u.songRepo.GetMany(ctx, songIDs); err == nil {
			for _, song := range s {
				songs[song.ID] = song
			}
		}
	}

	artists := make(map[uint64]domain.Artist)
	if len(artistIDs) > 0 {
		if a, err := u.artistRepo.GetMany(ctx, artistIDs); err == nil {
			for _, artist := range a {
				artists[artist.ID] = artist
			}
		}
	}

	userData := make(map[uint64]domain.User)
	if len(userIDs) > 0 {
		// Unique user IDs for query
		uniqueIDs := make([]uint64, 0)
		seen := make(map[uint64]bool)
		for _, id := range userIDs {
			if !seen[id] {
				uniqueIDs = append(uniqueIDs, id)
				seen[id] = true
			}
		}
		if users, err := u.userRepo.GetMany(ctx, uniqueIDs); err == nil {
			for _, user := range users {
				userData[user.ID] = user
			}
		}
	}

	// 3. Populate targets and user details (avatars/banners)
	for i := range items {
		// Populate User
		if user, ok := userData[items[i].UserID]; ok {
			// Populate image URLs for frontend
			if user.Avatar != nil {
				user.AvatarUrl = u.mediaService.Resolve(user.Avatar)
			}
			items[i].User = user
		}

		// Populate Target
		switch items[i].TargetType {
		case "song":
			if song, ok := songs[items[i].TargetID]; ok {
				items[i].Target = &song
			}
		case "artist":
			if artist, ok := artists[items[i].TargetID]; ok {
				// Populate image URLs for artist
				if artist.Avatar != nil {
					artist.AvatarUrl = u.mediaService.Resolve(artist.Avatar)
				}
				items[i].Target = &artist
			}
		}
	}
}

// ---- FOLLOWS ----

func (u *InteractionUsecase) FollowUser(ctx context.Context, followerID uint64, followedTarget string) error {
	// Resolve target (UUID or Slug) to ID
	followedUser, err := u.userRepo.GetByUUID(ctx, followedTarget)
	if err != nil {
		followedUser, err = u.userRepo.GetBySlug(ctx, followedTarget)
		if err != nil {
			return domain.NewAppError(404, "User not found", err)
		}
	}

	followedID := followedUser.ID

	if followerID == followedID {
		return domain.NewAppError(400, "You cannot follow yourself", nil)
	}
	if err := u.userRepo.Follow(ctx, followerID, followedID); err != nil {
		return err
	}

	// Trigger Notification
	follower, _ := u.userRepo.GetByID(ctx, followerID)
	if follower != nil {
		subjectType := "user"
		
		followerSlug := ""
		if follower.Slug != nil {
			followerSlug = *follower.Slug
		}
		
		// Prepare notification data
		dataObj := map[string]interface{}{
			"follower_name":   follower.Name,
			"follower_slug":   followerSlug,
			"follower_avatar": u.mediaService.Resolve(follower.Avatar),
			"follower_banner": u.mediaService.Resolve(follower.Banner),
		}
		dataJSON, _ := json.Marshal(dataObj)

		notif := &domain.Notification{
			UserID:      followedID,
			Type:        "follow",
			SubjectID:   &followerID,
			SubjectUUID: &follower.UUID,
			SubjectType: &subjectType,
			Data:        dataJSON,
		}
		err := u.notificationRepo.Create(ctx, notif)
		
		// Log Activity
		val := fmt.Sprintf("%d", followedID)
		_ = u.activityUsecase.LogActivity(ctx, followerID, "follow", followedID, "user", &val)
		
		return err
	}
	return nil
}

func (u *InteractionUsecase) UnfollowUser(ctx context.Context, followerID uint64, followedTarget string) error {
	followedUser, err := u.userRepo.GetByUUID(ctx, followedTarget)
	if err != nil {
		followedUser, err = u.userRepo.GetBySlug(ctx, followedTarget)
		if err != nil {
			return domain.NewAppError(404, "User not found", err)
		}
	}
	return u.userRepo.Unfollow(ctx, followerID, followedUser.ID)
}

func (u *InteractionUsecase) IsFollowing(ctx context.Context, followerID uint64, followedTarget string) (bool, error) {
	followedUser, err := u.userRepo.GetByUUID(ctx, followedTarget)
	if err != nil {
		followedUser, err = u.userRepo.GetBySlug(ctx, followedTarget)
		if err != nil {
			return false, nil // Consider not found as not following
		}
	}
	return u.userRepo.IsFollowing(ctx, followerID, followedUser.ID)
}
