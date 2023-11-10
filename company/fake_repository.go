package company

import (
	"api/database"
	"context"
)

type fakeRepository struct {
	data map[uint64]*Company
}

func NewFakeRepository() Repository {
	data := map[uint64]*Company{
		1: {Id: 1, Name: "Test", Email: "admin@test.com", Pass: "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy", AvailableCash: 720},
	}
	return &fakeRepository{data}
}

func (r *fakeRepository) Register(ctx context.Context, registration *Registration) (*Company, error) {
	id := uint64(len(r.data) + 1)
	company := &Company{
		Id:    id,
		Name:  registration.Name,
		Email: registration.Email,
		Pass:  registration.Password,
	}
	r.data[id] = company
	return r.GetById(ctx, id)
}

func (r *fakeRepository) GetById(ctx context.Context, id uint64) (*Company, error) {
	return r.data[id], nil
}

func (r *fakeRepository) GetByEmail(ctx context.Context, email string) (*Company, error) {
	for _, company := range r.data {
		if company.Email == email {
			return company, nil
		}
	}
	return nil, nil
}

func (r *fakeRepository) RegisterTransaction(tx *database.DB, companyId, classificationId uint64, amount int, description string) error {
	r.data[companyId].AvailableCash += amount
	return nil
}
