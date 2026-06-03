package usecase

import (
	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type rankingUseCase struct {
	donationRepo repository.DonationRepository
}

func NewRankingUseCase(donationRepo repository.DonationRepository) RankingUseCase {
	return &rankingUseCase{donationRepo: donationRepo}
}

func (uc *rankingUseCase) GetRanking(limit int) ([]*model.RankingEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return uc.donationRepo.GetRanking(limit)
}
