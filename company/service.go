package company

import (
	"api/auth"
	"api/building"
	"api/resource"
	"api/server"
	"api/warehouse"
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

	Building struct {
		BuildingId uint64 `json:"building_id" validate:"required"`
		Position   uint8  `json:"position" validate:"required,min=0"`
	}

	Company struct {
		Id            uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name          string     `db:"name" json:"name"`
		Email         string     `db:"email" json:"email"`
		Pass          string     `db:"password" json:"-"`
		LastLogin     *time.Time `db:"last_login" json:"last_login"`
		CreatedAt     time.Time  `db:"created_at" json:"created_at"`
		AvailableCash int        `db:"cash" json:"available_cash"`
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
		BusyUntil       *time.Time `db:"busy_until" json:"busy_until"`
	}

	Production struct {
		*resource.Item
		Id             uint64           `db:"id" json:"id"`
		Building       *CompanyBuilding `db:"-" json:"-"`
		StartedAt      time.Time        `db:"created_at" json:"started_at"`
		FinishesAt     time.Time        `db:"finishes_at" json:"finishes_at"`
		CanceledAt     *time.Time       `db:"canceled_at" json:"canceled_at"`
		LastCollection *time.Time       `db:"collected_at" json:"last_collection"`
		SourcingCost   uint64           `db:"sourcing_cost" json:"sourcing_cost"`
	}

	Service interface {
		GetById(ctx context.Context, id uint64) (*Company, error)

		GetByEmail(ctx context.Context, email string) (*Company, error)

		Login(ctx context.Context, credentials Credentials) (string, error)

		Register(ctx context.Context, registration *Registration) (*Company, error)

		GetBuildings(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error)

		AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*CompanyBuilding, error)

		Produce(ctx context.Context, companyId, companyBuildingId uint64, item *resource.Item) (*Production, error)

		CancelProduction(ctx context.Context, companyId, buildingId, productionId uint64) error
	}

	service struct {
		repository Repository
		building   building.Service
		warehouse  warehouse.Service
	}
)

func (b *CompanyBuilding) GetResource(resourceId uint64) (*building.BuildingResource, error) {
	for _, resource := range b.Resources {
		if resource.Id == resourceId {
			return resource, nil
		}
	}
	return nil, server.NewBusinessRuleError("resource not found")
}

func (p *Production) ProducedUntil(t time.Time) (*warehouse.StockItem, error) {
	producedResource, err := p.Building.GetResource(p.Resource.Id)
	if err != nil {
		return nil, err
	}

	var lastCollection time.Time
	if p.LastCollection == nil {
		lastCollection = p.StartedAt
	} else {
		lastCollection = *p.LastCollection
	}

	qtyPerMinute := (float64(producedResource.QtyPerHours) / 60.0)
	qtyProduced := t.Sub(lastCollection).Minutes() * qtyPerMinute

	return &warehouse.StockItem{
		Cost: p.SourcingCost,
		Item: &resource.Item{
			Qty:        uint64(qtyProduced),
			Quality:    p.Quality,
			ResourceId: producedResource.Id,
			Resource:   producedResource.Resource,
		},
	}, nil
}

func NewService(repository Repository, building building.Service, warehouse warehouse.Service) Service {
	return &service{repository, building, warehouse}
}

func (s *service) GetById(ctx context.Context, id uint64) (*Company, error) {
	return s.repository.GetById(ctx, id)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*Company, error) {
	return s.repository.GetByEmail(ctx, email)
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

func (s *service) GetBuildings(ctx context.Context, companyId uint64) ([]*CompanyBuilding, error) {
	return s.repository.GetBuildings(ctx, companyId)
}

func (s *service) AddBuilding(ctx context.Context, companyId, buildingId uint64, position uint8) (*CompanyBuilding, error) {
	build, err := s.building.GetById(ctx, buildingId)
	if err != nil {
		return nil, err
	}

	inventory, err := s.warehouse.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if !inventory.HasResources(build.Requirements) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	return s.repository.AddBuilding(ctx, companyId, inventory, build, position)
}

func (s *service) Produce(ctx context.Context, companyId, buildingId uint64, item *resource.Item) (*Production, error) {
	building, err := s.repository.GetBuilding(ctx, buildingId, companyId)
	if err != nil {
		return nil, err
	}

	if building == nil {
		return nil, server.NewBusinessRuleError("building not found")
	}

	if building.BusyUntil != nil {
		return nil, server.NewBusinessRuleError("building is busy")
	}

	resourceToProduce, err := building.GetResource(item.ResourceId)
	if err != nil {
		return nil, err
	}

	inventory, err := s.warehouse.GetInventory(ctx, companyId)
	if err != nil {
		return nil, err
	}

	// Considers the qty
	for _, requirement := range resourceToProduce.Requirements {
		requirement.Qty *= item.Qty
	}

	if !inventory.HasResources(resourceToProduce.Requirements) {
		return nil, server.NewBusinessRuleError("not enough resources")
	}

	timeToProduce := float64(item.Qty) / (float64(resourceToProduce.QtyPerHours) / 60.0)
	adminCost := uint64(float64(building.AdminHour) * timeToProduce * 10)
	wagesCost := uint64(float64(building.WagesHour) * timeToProduce * 10)
	totalCost := int(adminCost + wagesCost)

	company, err := s.GetById(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if company.AvailableCash < totalCost {
		return nil, server.NewBusinessRuleError("not enough cash")
	}

	return s.repository.Produce(ctx, companyId, inventory, building, item, totalCost)
}

func (s *service) CancelProduction(ctx context.Context, companyId, buildingId, productionId uint64) error {
	companyBuilding, err := s.repository.GetBuilding(ctx, buildingId, companyId)
	if err != nil {
		return err
	}

	if companyBuilding == nil {
		return server.NewBusinessRuleError("building not found")
	}

	if companyBuilding.BusyUntil == nil {
		return server.NewBusinessRuleError("no production in process")
	}

	return s.repository.CancelProduction(ctx, productionId, buildingId, companyId)
}
