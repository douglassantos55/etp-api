package company

import (
	"api/auth"
	"api/server"
	"context"
	"errors"
	"time"
)

type (
	Credentials struct {
		Email string `form:"email" json:"email" validate:"required,email"`
		Pass  string `form:"password" json:"password" validate:"required"`
	}

	Registration struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
		Confirm  string `json:"confirm_password" validate:"required,eqfield=Password"`
	}

	Company struct {
		Id                uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name              string     `db:"name" json:"name"`
		Email             string     `db:"email" json:"email"`
		Pass              string     `db:"password" json:"-"`
		Admin             bool       `db:"is_admin" json:"-"`
		LastLogin         *time.Time `db:"last_login" json:"last_login"`
		CreatedAt         time.Time  `db:"created_at" json:"created_at"`
		AvailableCash     int        `db:"cash" json:"available_cash"`
		AvailableTerrains int8       `db:"available_terrains" json:"available_terrains"`
	}

	Service interface {
		GetById(ctx context.Context, id uint64) (*Company, error)
		GetByEmail(ctx context.Context, email string) (*Company, error)
		Login(ctx context.Context, credentials Credentials) (string, error)
		Register(ctx context.Context, registration *Registration) (*Company, error)
		PurchaseTerrain(ctx context.Context, companyId uint64, position int) error
	}

	service struct {
		repository Repository
	}
)

const (
	TERRAIN_BASE_VALUE     = 1_000_000_00
	TERRAIN_UNIT_VALUE     = 500_000_00
	TERRAIN_POSITION_VALUE = 100_000_00
)

func (c *Company) IsAdmin() bool {
	return c.Admin
}

func (c *Company) GetCreditScore() int64 {
	return c.TerrainValue(c.AvailableTerrains)
}

func (c *Company) TerrainValue(position int8) int64 {
	return int64(TERRAIN_BASE_VALUE + TERRAIN_UNIT_VALUE*((int(position)-1)/5) + (TERRAIN_POSITION_VALUE * int(position)))
}

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetById(ctx context.Context, id uint64) (*Company, error) {
	return s.repository.GetById(ctx, id)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*Company, error) {
	return s.repository.GetByEmail(ctx, email)
}

func (s *service) PurchaseTerrain(ctx context.Context, companyId uint64, position int) error {
	company, err := s.repository.GetById(ctx, companyId)
	if err != nil {
		return err
	}

	if company == nil {
		return server.NewBusinessRuleError("company not found")
	}

	total := (TERRAIN_BASE_VALUE + (TERRAIN_UNIT_VALUE * (int(company.AvailableTerrains) / 5)) + (TERRAIN_POSITION_VALUE * position))
	if company.AvailableCash < total {
		return server.NewBusinessRuleError("not enough cash")
	}

	return s.repository.PurchaseTerrain(ctx, total, companyId)
}

func (s *service) Login(ctx context.Context, credentials Credentials) (string, error) {
	company, err := s.GetByEmail(ctx, credentials.Email)
	if err != nil || company == nil {
		return "", errors.New("invalid credentials")
	}

	if err := auth.ComparePassword(company.Pass, credentials.Pass); err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(company.Id, server.GetJwtSecret())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) Register(ctx context.Context, registration *Registration) (*Company, error) {
	hashedPassword, err := auth.HashPassword(registration.Password)
	if err != nil {
		return nil, err
	}

	registration.Password = hashedPassword

	return s.repository.Register(ctx, registration)
}
