package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
)

func newCampaignUC(campaignRepo *mocks.MockCampaignRepository) usecase.CampaignUseCase {
	return usecase.NewCampaignUseCase(campaignRepo)
}

func TestCampaignCreate_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	input := usecase.CreateCampaignInput{
		Title:       "Campanha Saúde",
		Description: "Ajuda médica para crianças",
		GoalAmount:  100000,
		IsUrgent:    true,
		Keywords:    []string{"saúde", "crianças"},
	}

	cr.On("Create", mock.MatchedBy(func(c *model.Campaign) bool {
		return c.Title == input.Title &&
			c.InstitutionID == instID &&
			c.GoalAmount == input.GoalAmount &&
			c.IsUrgent == true &&
			c.Status == model.CampaignStatusActive
	})).Return(nil)

	campaign, err := uc.Create(instID, input)

	assert.NoError(t, err)
	assert.Equal(t, input.Title, campaign.Title)
	assert.Equal(t, instID, campaign.InstitutionID)
	assert.Equal(t, int64(0), campaign.TotalRaised)
	assert.Equal(t, model.CampaignStatusActive, campaign.Status)
	cr.AssertExpectations(t)
}

func TestCampaignGetByID_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	id := uuid.New()
	expected := &model.Campaign{ID: id, Title: "Campanha"}
	cr.On("FindByID", id).Return(expected, nil)

	campaign, err := uc.GetByID(id)

	assert.NoError(t, err)
	assert.Equal(t, expected, campaign)
	cr.AssertExpectations(t)
}

func TestCampaignGetByID_NotFound(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	id := uuid.New()
	cr.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.GetByID(id)

	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestCampaignGetAll_NoFilter(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	campaigns := []*model.Campaign{
		{ID: uuid.New(), Title: "A"},
		{ID: uuid.New(), Title: "B"},
	}
	cr.On("FindAll", repository.CampaignFilters{}).Return(campaigns, nil)

	result, err := uc.GetAll(repository.CampaignFilters{})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	cr.AssertExpectations(t)
}

func TestCampaignGetAll_WithKeyword(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	campaigns := []*model.Campaign{
		{ID: uuid.New(), Title: "Campanha Saúde", Keywords: pq.StringArray{"saúde"}},
	}
	filters := repository.CampaignFilters{Keyword: "saúde"}
	cr.On("FindAll", filters).Return(campaigns, nil)

	result, err := uc.GetAll(filters)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	cr.AssertExpectations(t)
}

func TestCampaignGetByInstitutionID_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	campaigns := []*model.Campaign{{ID: uuid.New(), InstitutionID: instID}}
	cr.On("FindByInstitutionID", instID).Return(campaigns, nil)

	result, err := uc.GetByInstitutionID(instID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	cr.AssertExpectations(t)
}

func TestCampaignUpdate_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	id := uuid.New()
	existing := &model.Campaign{ID: id, InstitutionID: instID, Title: "Antigo"}
	input := usecase.UpdateCampaignInput{Title: "Novo"}

	cr.On("FindByID", id).Return(existing, nil)
	cr.On("Update", mock.MatchedBy(func(c *model.Campaign) bool {
		return c.Title == "Novo"
	})).Return(nil)

	campaign, err := uc.Update(id, instID, input)

	assert.NoError(t, err)
	assert.Equal(t, "Novo", campaign.Title)
	cr.AssertExpectations(t)
}

func TestCampaignUpdate_Unauthorized(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	otherInstID := uuid.New()
	id := uuid.New()

	cr.On("FindByID", id).Return(&model.Campaign{ID: id, InstitutionID: instID}, nil)

	_, err := uc.Update(id, otherInstID, usecase.UpdateCampaignInput{Title: "X"})

	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestCampaignDelete_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	id := uuid.New()

	cr.On("FindByID", id).Return(&model.Campaign{ID: id, InstitutionID: instID}, nil)
	cr.On("Delete", id).Return(nil)

	err := uc.Delete(id, instID)

	assert.NoError(t, err)
	cr.AssertExpectations(t)
}

func TestCampaignDelete_Unauthorized(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	otherInstID := uuid.New()
	id := uuid.New()

	cr.On("FindByID", id).Return(&model.Campaign{ID: id, InstitutionID: instID}, nil)

	err := uc.Delete(id, otherInstID)

	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestCampaignUpdateStatus_Success(t *testing.T) {
	cr := &mocks.MockCampaignRepository{}
	uc := newCampaignUC(cr)

	instID := uuid.New()
	id := uuid.New()

	cr.On("FindByID", id).Return(&model.Campaign{ID: id, InstitutionID: instID, Status: model.CampaignStatusActive}, nil)
	cr.On("Update", mock.MatchedBy(func(c *model.Campaign) bool {
		return c.Status == model.CampaignStatusPaused
	})).Return(nil)

	campaign, err := uc.UpdateStatus(id, instID, model.CampaignStatusPaused)

	assert.NoError(t, err)
	assert.Equal(t, model.CampaignStatusPaused, campaign.Status)
	cr.AssertExpectations(t)
}
