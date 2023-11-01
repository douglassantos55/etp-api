package building

import "api/resource"

type (
	Building struct {
		Id              uint64              `db:"id" json:"id"`
		Name            string              `db:"name" json:"name"`
		WagesHour       uint64              `db:"wages_per_hour" json:"wages_per_hour"`
		AdminHour       uint64              `db:"admin_per_hour" json:"admin_per_hour"`
		MaintenanceHour uint64              `db:"maintenance_per_hour" json:"maintenance_per_hour"`
		Requirements    []*resource.Item    `json:"requirements"`
		Resources       []*BuildingResource `json:"resources"`
	}

	BuildingResource struct {
		Resource    *resource.Resource `db:"resource" json:"resource"`
		QtyPerHours uint64             `db:"qty_per_hour" json:"qty_per_hour"`
	}

	Service interface {
		GetAll() ([]*Building, error)
		GetById(id uint64) (*Building, error)
	}

	service struct {
		repository Repository
	}
)

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetAll() ([]*Building, error) {
	return s.repository.GetAll()
}

func (s *service) GetById(id uint64) (*Building, error) {
	return s.repository.GetById(id)
}
