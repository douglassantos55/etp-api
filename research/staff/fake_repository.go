package staff

import (
	"context"
	"math/rand"
	"time"
)

type fakeRepository struct {
	lastStaffId    uint64
	lastTrainingId uint64
	lastSearchId   uint64
	trainings      map[uint64]*Training
	staff          map[uint64]map[uint64]*Staff
	searches       map[uint64]map[uint64]*Search
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
		1: {
			2: {
				Id:       2,
				Name:     "Jane Doe",
				Skill:    52,
				Talent:   50,
				Salary:   200000,
				Status:   HIRED,
				Employer: 1,
			},
		},
	}

	trainings := map[uint64]*Training{
		1: {
			Id:          1,
			Result:      0,
			Investment:  10000000,
			StaffId:     2,
			CompanyId:   1,
			FinishesAt:  time.Now(),
			CompletedAt: time.Now(),
		},
	}

	searches := map[uint64]map[uint64]*Search{
		1: {
			1: {
				Id:         1,
				StartedAt:  time.Now(),
				FinishesAt: time.Now().Add(time.Second),
			},
			42069: {
				Id:         42069,
				StartedAt:  time.Now(),
				FinishesAt: time.Now().Add(time.Second),
			},
		},
	}

	return &fakeRepository{
		lastStaffId:    2,
		lastSearchId:   1,
		lastTrainingId: 1,
		staff:          staff,
		searches:       searches,
		trainings:      trainings,
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
	r.lastStaffId++
	if _, ok := r.staff[companyId]; !ok {
		r.staff[companyId] = make(map[uint64]*Staff)
	}

	staff.Id = r.lastStaffId
	staff.Employer = companyId
	r.staff[companyId][r.lastStaffId] = staff

	return staff, nil
}

func (r *fakeRepository) UpdateStaff(ctx context.Context, staff *Staff) error {
	r.staff[staff.Employer][staff.Id] = staff
	return nil
}

func (r *fakeRepository) StartSearch(ctx context.Context, finishTime time.Time, companyId uint64) (*Search, error) {
	r.lastSearchId++
	search := &Search{Id: r.lastSearchId, FinishesAt: finishTime}
	r.searches[companyId][search.Id] = search
	return search, nil
}

func (r *fakeRepository) DeleteSearch(ctx context.Context, searchId, companyId uint64) error {
	searches, ok := r.searches[companyId]
	if !ok {
		return ErrSearchNotFound
	}
	delete(searches, searchId)
	return nil
}

func (r *fakeRepository) GetTraining(ctx context.Context, trainingId, companyId uint64) (*Training, error) {
	training, ok := r.trainings[trainingId]
	if !ok {
		return nil, ErrTrainingNotFound
	}
	return training, nil
}

func (r *fakeRepository) SaveTraining(ctx context.Context, training *Training) (*Training, error) {
	r.lastTrainingId++
	training.Id = r.lastTrainingId
	return training, nil
}

func (r *fakeRepository) UpdateTraining(ctx context.Context, training *Training) error {
	return nil
}
