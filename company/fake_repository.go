package company

import "context"

type fakeRepository struct {
	data map[uint64]*Company
}

func NewFakeRepository() Repository {
	data := map[uint64]*Company{
		1: {Id: 1, Name: "Test", Email: "admin@test.com", Pass: "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy", AvailableCash: 720, AvailableTerrains: 3},
		2: {Id: 2, Name: "Test 2", Email: "admin@test2.com", Pass: "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy", AvailableCash: 255720, AvailableTerrains: 3},
		3: {Id: 3, Name: "Test 3", Email: "admin@test3.com", Pass: "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy", AvailableCash: 125572000, AvailableTerrains: 3},
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

func (r *fakeRepository) PurchaseTerrain(ctx context.Context, total int, companyId uint64) error {
	r.data[companyId].AvailableTerrains++
	return nil
}
