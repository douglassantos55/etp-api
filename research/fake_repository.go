package research

import (
	"api/research/staff"
	"context"
	"time"
)

type (
	fakeRepository struct {
		lastId     uint64
		busyStaff  map[uint64]bool
		qualities  map[uint64]map[uint64]Quality
		researches map[uint64]*Research
	}
)

func NewFakeRepository() Repository {
	busyStaff := map[uint64]bool{1: true, 2: true}
	qualities := map[uint64]map[uint64]Quality{
		3: {1: {Quality: 2, Patents: 0, ResourceId: 1}},
	}

	researches := map[uint64]*Research{
		1: {
			Id:         1,
			CompanyId:  3,
			Investment: 40000000,
			ResourceId: 1,
			AssignedStaff: []*staff.Staff{
				{Skill: 10},
				{Skill: 20},
			},
		},
	}

	return &fakeRepository{
		lastId:     1,
		busyStaff:  busyStaff,
		qualities:  qualities,
		researches: researches,
	}
}

func (r *fakeRepository) GetQuality(ctx context.Context, resourceId, companyId uint64) (Quality, error) {
	quality, ok := r.qualities[companyId][resourceId]
	if !ok {
		return Quality{}, nil
	}
	return quality, nil
}

func (r *fakeRepository) GetResearch(ctx context.Context, researchId uint64) (*Research, error) {
	research, ok := r.researches[researchId]
	if !ok {
		return nil, ErrResearchNotFound
	}
	return research, nil
}

func (r *fakeRepository) IsStaffBusy(ctx context.Context, staffIds []uint64, companyId uint64) (bool, error) {
	for _, staffId := range staffIds {
		busy, ok := r.busyStaff[staffId]
		if ok && busy {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepository) SaveResearch(ctx context.Context, finishesAt time.Time, investment int, staffIds []uint64, resourceId, companyId uint64) (*Research, error) {
	research := &Research{
		Id:         r.lastId,
		CompanyId:  companyId,
		ResourceId: resourceId,
		Investment: investment,
		FinishesAt: finishesAt,
	}

	return research, nil
}

func (r *fakeRepository) CompleteResearch(ctx context.Context, research *Research) (*Research, error) {
	r.researches[research.Id] = research
	return research, nil
}
