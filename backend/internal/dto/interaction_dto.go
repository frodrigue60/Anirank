package dto

import "time"

type CommentDTO struct {
	ID            string       `json:"id"`
	ParentID      *string      `json:"parent_id,omitempty"`
	Content       string       `json:"content"`
	LikesCount    int          `json:"likes_count"`
	DislikesCount int          `json:"dislikes_count"`
	RepliesCount  int          `json:"replies_count"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	IsLiked       bool         `json:"is_liked"`
	IsDisliked    bool         `json:"is_disliked"`
	User          UserMinimalDTO `json:"user"`
	Replies       []CommentDTO `json:"replies,omitempty"`
}

type ActivityItemDTO struct {
	Type      string      `json:"type"`
	User      UserMinimalDTO `json:"user"`
	TargetID  string      `json:"target_id"`
	Target    interface{} `json:"target"`
	Value     interface{} `json:"value,omitempty"`
	CreatedAt string      `json:"created_at"`
}
