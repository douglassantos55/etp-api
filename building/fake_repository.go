package building

import (
	"api/resource"
	"errors"
)

type fakeRepository struct {
	data map[uint64]*Building
}

func NewFakeRepository() Repository {
	data := map[uint64]*Building{
		1: {
			Id:   1,
			Name: "Plantation",
			Requirements: []*resource.Item{
				{Qty: 50, Quality: 0, Resource: &resource.Resource{Id: 1}},
			},
		},
		2: {
			Id:   2,
			Name: "Factory",
			Requirements: []*resource.Item{
				{Qty: 150, Quality: 0, Resource: &resource.Resource{Id: 1}},
			},
		},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) GetAll() ([]*Building, error) {
	items := make([]*Building, 0)
	for _, item := range r.data {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeRepository) GetById(id uint64) (*Building, error) {
	building, ok := r.data[id]
	if !ok {
		return nil, errors.New("building not found")
	}
	return building, nil
}
