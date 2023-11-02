package company

import (
	"api/auth"
	"api/building"
	"api/server"
	"api/warehouse"
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

	Building struct {
		BuildingId uint64 `json:"building_id" validate:"required"`
		Position   uint8  `json:"position" validate:"required,min=0"`
	}

	Company struct {
		Id        uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name      string     `db:"name" json:"name"`
		Email     string     `db:"email" json:"email"`
		Pass      string     `db:"password" json:"-"`
		LastLogin *time.Time `db:"last_login" json:"last_login"`
		CreatedAt time.Time  `db:"created_at" json:"created_at"`
	}

	CompanyBuilding struct {
		Id              uint64 `db:"id" json:"id"`
		Name            string `db:"name" json:"name"`
		WagesHour       uint64 `db:"wages_per_hour" json:"wages_per_hour"`
		AdminHour       uint64 `db:"admin_per_hour" json:"admin_per_hour"`
		MaintenanceHour uint64 `db:"maintenance_per_hour" json:"maintenance_per_hour"`
		Level           uint8  `db:"level" json:"level"`
		Position        *uint8 `db:"position" json:"position"`
		Resources       []*building.BuildingResource
	}

	Service interface {
		GetById(id uint64) (*Company, error)

		GetByEmail(email string) (*Company, error)

		Login(credentials Credentials) (string, error)

		Register(registration *Registration) (*Company, error)

		GetBuildings(companyId uint64) ([]*CompanyBuilding, error)

		AddBuilding(companyId, buildingId uint64, position uint8) (*CompanyBuilding, error)
	}

	service struct {
		repository Repository
		building   building.Service
		warehouse  warehouse.Service
	}
)

func NewService(repository Repository, building building.Service, warehouse warehouse.Service) Service {
	return &service{repository, building, warehouse}
}

func (s *service) GetById(id uint64) (*Company, error) {
	return s.repository.GetById(id)
}

func (s *service) GetByEmail(email string) (*Company, error) {
	return s.repository.GetByEmail(email)
}

func (s *service) Login(credentials Credentials) (string, error) {
	company, err := s.GetByEmail(credentials.Email)
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

func (s *service) Register(registration *Registration) (*Company, error) {
	hashedPassword, err := auth.HashPassword(registration.Password)
	if err != nil {
		return nil, err
	}

	registration.Password = hashedPassword

	return s.repository.Register(registration)
}

func (s *service) GetBuildings(companyId uint64) ([]*CompanyBuilding, error) {
	return s.repository.GetBuildings(companyId)
}

func (s *service) AddBuilding(companyId, buildingId uint64, position uint8) (*CompanyBuilding, error) {
	build, err := s.building.GetById(buildingId)
	if err != nil {
		return nil, err
	}

	inventory, err := s.warehouse.GetInventory(companyId)
	if err != nil {
		return nil, err
	}

	if !inventory.HasResources(build.Requirements) {
		return nil, errors.New("not enough resources")
	}

	if err := s.warehouse.ReduceStock(companyId, inventory, build.Requirements); err != nil {
		return nil, err
	}

	return s.repository.AddBuilding(companyId, build, position)
}
