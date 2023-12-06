package research

import (
	"api/scheduler"
	"api/server"
	"context"
	"math/rand"
	"time"
)

type Status int

const (
	PENDING Status = iota
	HIRED
)

var (
	NAMES = []string{
		"James", "Robert", "John", "Michael", "David", "William", "Richard",
		"Joseph", "Thomas", "Christopher", "Mary", "Patricia", "Jennifer",
		"Linda", "Elizabeth", "Barbara", "Susan", "Jessica", "Sarah", "Karen",
	}

	LASTNAMES = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
		"Davis", "Rodriguez", "Martinez",
	}
)

type (
	Staff struct {
		Id       uint64  `db:"id" json:"id"`
		Name     string  `db:"name" json:"name"`
		Skill    uint8   `db:"skill" json:"-"`
		Talent   uint8   `db:"talent" json:"-"`
		Salary   uint64  `db:"salary" json:"salary"`
		Status   Status  `db:"status" json:"status"`
		Offer    uint64  `db:"offer" json:"offer"`
		Poacher  *uint64 `db:"poacher_id" json:"-"`
		Employer uint64  `db:"company_id" json:"-"`
	}

	Service interface {
		FindGraduate(ctx context.Context, companyId uint64) (time.Duration, error)
		FindExperienced(ctx context.Context, companyId uint64) (time.Duration, error)

		GetGraduate(ctx context.Context, companyId uint64) (*Staff, error)
		GetExperienced(ctx context.Context, companyId uint64) (*Staff, error)

		HireStaff(ctx context.Context, staffId uint64) (*Staff, error)
		MakeOffer(ctx context.Context, offer, staffId uint64) (*Staff, error)

		IncreaseSalary(ctx context.Context, salary, staffId uint64) (*Staff, error)
	}

	service struct {
		repository Repository
		timer      *scheduler.Scheduler
	}
)

func NewService(repository Repository, timer *scheduler.Scheduler) Service {
	return &service{repository, timer}
}

func (s *service) FindGraduate(ctx context.Context, companyId uint64) (time.Duration, error) {
	duration := 12 * time.Hour

	s.timer.Add(companyId, duration, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := s.GetGraduate(ctx, companyId)

		// TODO: send a message to the socket

		return err
	})

	return duration, nil
}

func (s *service) GetGraduate(ctx context.Context, companyId uint64) (*Staff, error) {
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))

	skill := randomizer.Intn(100)
	talent := randomizer.Intn(100)
	salary := randomizer.Intn(1000) + 1000

	name := NAMES[randomizer.Intn(len(NAMES))] + " " + LASTNAMES[randomizer.Intn(len(LASTNAMES))]

	staff := &Staff{
		Name:   name,
		Skill:  uint8(skill),
		Talent: uint8(talent),
		Salary: uint64(salary),
		Status: PENDING,
	}

	return s.repository.SaveStaff(ctx, staff, companyId)
}

func (s *service) FindExperienced(ctx context.Context, companyId uint64) (time.Duration, error) {
	duration := 12 * time.Hour

	s.timer.Add(companyId, duration, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := s.GetExperienced(ctx, companyId)

		// TODO: send a message to the socket

		return err
	})

	return duration, nil
}

func (s *service) GetExperienced(ctx context.Context, companyId uint64) (*Staff, error) {
	staff, err := s.repository.RandomStaff(ctx, companyId)
	if err != nil {
		return nil, err
	}

	staff.Poacher = &companyId
	err = s.repository.UpdateStaff(ctx, staff)

	return staff, err
}

func (s *service) HireStaff(ctx context.Context, staffId uint64) (*Staff, error) {
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	// If hiring experienced, update the salary and clean up poaching
	// fields
	if staff.Poacher != nil {
		staff.Salary = staff.Offer
		staff.Employer = *staff.Poacher

		staff.Offer = 0
		staff.Poacher = nil
	}

	staff.Status = HIRED
	if err := s.repository.UpdateStaff(ctx, staff); err != nil {
		return nil, err
	}

	return staff, nil
}

func (s *service) MakeOffer(ctx context.Context, offer, staffId uint64) (*Staff, error) {
	// Must save the offer and notify the current employer of the staff
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	if offer <= staff.Salary {
		return nil, server.NewBusinessRuleError("offer is too low")
	}

	staff.Offer = offer
	if err := s.repository.UpdateStaff(ctx, staff); err != nil {
		return nil, err
	}

	// TODO: send message on socket notifying current employer

	s.timer.Add(staffId, 48*time.Hour, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := s.HireStaff(ctx, staffId)
		return err
	})

	return staff, nil
}

func (s *service) IncreaseSalary(ctx context.Context, salary, staffId uint64) (*Staff, error) {
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	if salary <= staff.Salary {
		return nil, server.NewBusinessRuleError("new salary must be higher than current salary")
	}

	if staff.Poacher != nil {
		if salary <= staff.Offer {
			return nil, server.NewBusinessRuleError("new salary must be higher than current offer")
		}
		staff.Offer = 0
		staff.Poacher = nil
	}

	staff.Salary = salary
	if err := s.repository.UpdateStaff(ctx, staff); err != nil {
		return nil, err
	}

	// TODO: send message on socket notifying the poaching company, if that's
	// the case

	return staff, nil
}
