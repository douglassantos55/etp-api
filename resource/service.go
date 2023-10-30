package resource

type (
	Service interface {
		GetAll() ([]*Resource, error)
		GetById(id uint64) (*Resource, error)
		CreateResource(resource *Resource) (*Resource, error)
		UpdateResource(resource *Resource) (*Resource, error)
	}

	Category struct {
		Id   uint64 `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name string `db:"name" json:"name" validate:"required"`
	}

	Item struct {
		Qty      uint64    `db:"quantity" json:"quantity"`
		Quality  uint8     `db:"quality" json:"quality"`
		Resource *Resource `db:"resource" json:"resource"`
	}

	Resource struct {
		Id         uint64    `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name       string    `db:"name" json:"name" validate:"required"`
		Image      *string   `db:"image" json:"image"`
		CategoryId uint64    `db:"category_id" json:"category_id" validate:"required"`
		Category   *Category `db:"category" json:"category" validate:"-"`
	}

	service struct {
		repository Repository
	}
)

func NewService(repository Repository) Service {
	return &service{repository}
}

func (s *service) GetAll() ([]*Resource, error) {
	return s.repository.FetchResources()
}

func (s *service) GetById(id uint64) (*Resource, error) {
	return s.repository.GetById(id)
}

func (s *service) CreateResource(resource *Resource) (*Resource, error) {
	return s.repository.SaveResource(resource)
}

func (s *service) UpdateResource(resource *Resource) (*Resource, error) {
	return s.repository.UpdateResource(resource)
}
