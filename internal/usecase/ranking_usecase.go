package usecase

import "facilitador-de-doacoes/internal/model"

type RankingUseCase interface {
	GetRanking(limit int) ([]*model.RankingEntry, error)
}
