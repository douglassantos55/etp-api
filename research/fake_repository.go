package research

import (
	"context"
	"math/rand"
	"time"
)

type fakeRepository struct {
	staff map[uint64]map[uint64]*Staff
}

func NewFakeRepository() Repository {
	staff := map[uint64]map[uint64]*Staff{
		2: {
			1: {
				Id:       1,
				Name:     "John Doe",
				Skill:    52,
				Talent:   50,
				Salary:   1523400,
				Status:   HIRED,
				Offer:    2000000,
				Employer: 2,
			},
		},
	}
	return &fakeRepository{
		staff: staff,
	}
}

func (r *fakeRepository) GetStaff(ctx context.Context, companyId uint64) ([]*Staff, error) {
	staff := make([]*Staff, 0)
	return staff, nil
}

func (r *fakeRepository) GetStaffById(ctx context.Context, staffId uint64) (*Staff, error) {
	for _, staff := range r.staff {
		for _, member := range staff {
			if member.Id == staffId {
				return member, nil
			}
		}
	}

	return nil, ErrStaffNotFound
}

func (r *fakeRepository) RandomStaff(ctx context.Context, companyId uint64) (*Staff, error) {
	for id, staff := range r.staff {
		if id != companyId {
			keys := make([]uint64, 0, len(staff))
			for k := range staff {
				keys = append(keys, k)
			}

			selected := uint64(rand.Intn(len(keys)))
			return staff[keys[selected]], nil
		}
	}

	return nil, ErrNoStaffFound
}

func (r *fakeRepository) SaveStaff(ctx context.Context, staff *Staff, companyId uint64) (*Staff, error) {
	id := uint64(len(r.staff[companyId]) + 1)
	if _, ok := r.staff[companyId]; !ok {
		r.staff[companyId] = make(map[uint64]*Staff)
	}

	staff.Id = id
	staff.Employer = companyId
	r.staff[companyId][id] = staff

	return staff, nil
}

func (r *fakeRepository) UpdateStaff(ctx context.Context, staff *Staff) error {
	r.staff[staff.Employer][staff.Id] = staff
	return nil
}

func (r *fakeRepository) StartSearch(ctx context.Context, finishTime time.Time, companyId uint64) (*Search, error) {
	return &Search{FinishesAt: finishTime}, nil
}

func (r *fakeRepository) DeleteSearch(ctx context.Context, searchId uint64) error {
	return nil
}
