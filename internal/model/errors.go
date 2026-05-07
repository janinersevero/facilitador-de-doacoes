package model

import "errors"

var ErrNotFound = errors.New("not found")
var ErrEmailAlreadyInUse = errors.New("email already in use")
var ErrCPFAlreadyInUse = errors.New("cpf already in use")
var ErrCNPJAlreadyInUse = errors.New("cnpj already in use")
var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalidDonationTarget = errors.New("exactly one of institution_id or campaign_id must be provided")
var ErrCampaignNotActive = errors.New("campaign is not active")
