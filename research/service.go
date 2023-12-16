package research

import (
	"api/company"
	"api/research/staff"
	"api/scheduler"
	"api/server"
	"context"
	"math"
	"math/rand"
	"time"
)

var (
	ErrBusyStaff        = server.NewBusinessRuleError("busy staff")
	ErrNotEnoughCash    = server.NewBusinessRuleError("not enough cash")
	ErrResearchNotFound = server.NewBusinessRuleError("research not found")
)

type (
	Research struct {
		Id            uint64         `db:"id" goqu:"skipinsert" json:"id"`
		Patents       int            `db:"patents" json:"patents,omitempty"`
		Investment    int            `db:"investment" json:"-"`
		FinishesAt    time.Time      `db:"finishes_at" json:"finishes_at"`
		CompletedAt   *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
		AssignedStaff []*staff.Staff `db:"-" goqu:"skipinsert,skipupdate" json:"staff"`
		CompanyId     uint64         `db:"company_id" json:"-"`
		ResourceId    uint64         `db:"resource_id" json:"-"`
	}

	Quality struct {
		Quality    uint8  `db:"quality" json:"quality"`
		Patents    uint8  `db:"patents" json:"patents"`
		ResourceId uint64 `db:"resource_id" json:"-"`
	}

	Service interface {
		GetQuality(ctx context.Context, resourceId, companyId uint64) (Quality, error)
		StartResearch(ctx context.Context, staffIds []uint64, resourceId, companyId uint64) (*Research, error)
		CompleteResearch(ctx context.Context, researchId uint64) (*Research, error)
	}

	service struct {
		repository Repository
		companySvc company.Service
		timer      *scheduler.Scheduler
	}
)

func NewService(repository Repository, companySvc company.Service) Service {
	return &service{repository, companySvc, scheduler.NewScheduler()}
}

func (s *service) GetQuality(ctx context.Context, resourceId, companyId uint64) (Quality, error) {
	return s.repository.GetQuality(ctx, resourceId, companyId)
}

func (s *service) StartResearch(ctx context.Context, staffIds []uint64, resourceId, companyId uint64) (*Research, error) {
	// Chosen staff should not be already researching
	busy, err := s.repository.IsStaffBusy(ctx, staffIds, companyId)
	if err != nil {
		return nil, err
	}

	if busy {
		return nil, ErrBusyStaff
	}

	quality, err := s.repository.GetQuality(ctx, resourceId, companyId)
	if err != nil {
		return nil, err
	}

	// Time to complete should be relative to current resource quality -  Max(48, L*6)
	duration := time.Duration(math.Min(48, float64(quality.Quality)*6)) * time.Hour
	finishesAt := time.Now().Add(duration)

	// Investment should be relative to level - 100k * ((L*2) + (1 / L))
	investment := 10000000 * ((int(quality.Quality) * 2) + (1 / int(math.Max(1, float64(quality.Quality)))))

	investingCompany, err := s.companySvc.GetById(ctx, companyId)
	if err != nil {
		return nil, err
	}

	if investingCompany.AvailableCash < investment {
		return nil, ErrNotEnoughCash
	}

	research, err := s.repository.SaveResearch(ctx, finishesAt, investment, staffIds, resourceId, companyId)
	if err != nil {
		return nil, err
	}

	s.timer.Add(research.Id, duration, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := s.CompleteResearch(ctx, research.Id)
		return err
	})

	return research, nil
}

func (s *service) CompleteResearch(ctx context.Context, researchId uint64) (*Research, error) {
	research, err := s.repository.GetResearch(ctx, researchId)
	if err != nil {
		return nil, err
	}

	// The more skill the staff have, the more likely to get a patent
	totalSkill := 0
	for _, staff := range research.AssignedStaff {
		totalSkill += int(staff.Skill)
	}

	now := time.Now()
	research.CompletedAt = &now

	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	research.Patents = randomizer.Intn(totalSkill/10) + (totalSkill / 10 / len(research.AssignedStaff) % len(research.AssignedStaff))

	if _, err := s.repository.CompleteResearch(ctx, research); err != nil {
		return nil, err
	}

	return research, nil
}
