package building

import (
	"api/resource"
	"context"
)

type (
	Building struct {
		Id              uint64              `db:"id" json:"id"`
		Name            string              `db:"name" json:"name"`
		WagesHour       uint64              `db:"wages_per_hour" json:"wages_per_hour"`
		AdminHour       uint64              `db:"admin_per_hour" json:"admin_per_hour"`
		MaintenanceHour uint64              `db:"maintenance_per_hour" json:"maintenance_per_hour"`
		Downtime        *uint8              `db:"downtime" json:"downtime"`
		Requirements    []*resource.Item    `json:"requirements"`
		Resources       []*BuildingResource `json:"resources"`
	}

	BuildingResource struct {
		*resource.Resource `db:"resource" json:"resource"`
		QtyPerHours        uint64 `db:"qty_per_hour" json:"qty_per_hour"`
	}

	Service interface {
		// List all buildings
		GetAll(ctx context.Context) ([]*Building, error)

		// Get a building by ID
		GetById(ctx context.Context, id uint64) (*Building, error)
	}

	service struct {
		repository Repository
	}
)

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetAll(ctx context.Context) ([]*Building, error) {
	return s.repository.GetAll(ctx)
}

func (s *service) GetById(ctx context.Context, id uint64) (*Building, error) {
	return s.repository.GetById(ctx, id)
}
