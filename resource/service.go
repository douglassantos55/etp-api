package resource

import "context"

type (
	Service interface {
		GetAll(ctx context.Context) ([]*Resource, error)
		GetById(ctx context.Context, id uint64) (*Resource, error)
		CreateResource(ctx context.Context, resource *Resource) (*Resource, error)
		UpdateResource(ctx context.Context, resource *Resource) (*Resource, error)
	}

	Category struct {
		Id   uint64 `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name string `db:"name" json:"name" validate:"required"`
	}

	Item struct {
		Qty        uint64    `db:"quantity" json:"quantity" validate:"required,min=1"`
		Quality    uint8     `db:"quality" json:"quality" validate:"min=0"`
		ResourceId uint64    `db:"resource_id" json:"resource_id" validate:"required"`
		Resource   *Resource `db:"resource" json:"resource" validate:"-"`
	}

	Resource struct {
		Id           uint64    `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name         string    `db:"name" json:"name" validate:"required"`
		Image        *string   `db:"image" json:"image" validate:"-"`
		CategoryId   uint64    `db:"category_id" json:"category_id" validate:"required"`
		Category     *Category `db:"category" json:"category" validate:"-"`
		Requirements []*Item   `json:"requirements" validate:"dive"`
	}

	service struct {
		repository Repository
	}
)

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetAll(ctx context.Context) ([]*Resource, error) {
	return s.repository.FetchResources(ctx)
}

func (s *service) GetById(ctx context.Context, id uint64) (*Resource, error) {
	return s.repository.GetById(ctx, id)
}

func (s *service) CreateResource(ctx context.Context, resource *Resource) (*Resource, error) {
	return s.repository.SaveResource(ctx, resource)
}

func (s *service) UpdateResource(ctx context.Context, resource *Resource) (*Resource, error) {
	return s.repository.UpdateResource(ctx, resource)
}
