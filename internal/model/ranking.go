package model

import "github.com/google/uuid"

type RankingEntry struct {
	Position      int       `json:"position"`
	UserID        uuid.UUID `json:"user_id"`
	UserName      string    `json:"user_name"`
	AvatarURL     string    `json:"avatar_url"`
	TotalDonated  int64     `json:"total_donated"`
	DonationCount int       `json:"donation_count"`
}
