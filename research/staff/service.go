package staff

import (
	"api/notification"
	"api/scheduler"
	"api/server"
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

type Status int

const SEARCH_DURATION = 12 * time.Hour
const TRAINING_DURATION = 8 * time.Hour

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
	Search struct {
		Id         uint64    `db:"id" json:"-"`
		StartedAt  time.Time `db:"started_at" json:"-"`
		FinishesAt time.Time `db:"finishes_at" json:"finishes_at"`
		CompanyId  uint64    `db:"company_id" json:"-"`
	}

	Training struct {
		Id          uint64    `db:"id" goqu:"skipinsert" json:"id"`
		Result      uint8     `db:"result" json:"result,omitempty"`
		Investment  uint64    `db:"investment" json:"-"`
		StaffId     uint64    `db:"staff_id" json:"-"`
		CompanyId   uint64    `db:"company_id" json:"-"`
		FinishesAt  time.Time `db:"finishes_at" json:"finishes_at,omitempty"`
		CompletedAt time.Time `db:"completed_at" json:"completed_at,omitempty"`
	}

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
		FindGraduate(ctx context.Context, companyId uint64) (*Search, error)
		FindExperienced(ctx context.Context, companyId uint64) (*Search, error)
		CancelSearch(ctx context.Context, searchId, companyId uint64) error

		GetGraduate(ctx context.Context, companyId uint64) (*Staff, error)
		GetExperienced(ctx context.Context, companyId uint64) (*Staff, error)

		HireStaff(ctx context.Context, staffId, companyId uint64) (*Staff, error)
		MakeOffer(ctx context.Context, offer, staffId, companyId uint64) (*Staff, error)

		IncreaseSalary(ctx context.Context, salary, staffId, companyId uint64) (*Staff, error)
		Train(ctx context.Context, staffId, companyId uint64) (*Training, error)
		FinishTraining(ctx context.Context, trainingId, companyId uint64) error
	}

	service struct {
		repository Repository
		timer      *scheduler.Scheduler
		notifier   notification.Notifier
		logger     *log.Logger
	}
)

func NewService(
	repository Repository,
	timer *scheduler.Scheduler,
	notifier notification.Notifier,
	logger *log.Logger,
) Service {
	return &service{repository, timer, notifier, logger}
}

func (s *service) FindGraduate(ctx context.Context, companyId uint64) (*Search, error) {
	finishTime := time.Now().Add(SEARCH_DURATION)
	search, err := s.repository.StartSearch(ctx, finishTime, companyId)
	if err != nil {
		return nil, err
	}

	s.timer.Add(search.Id, SEARCH_DURATION, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := s.repository.DeleteSearch(ctx, search.Id, companyId); err != nil {
			return err
		}

		graduate, err := s.GetGraduate(ctx, companyId)

		message := fmt.Sprintf("%s is available for hire", graduate.Name)
		if err := s.notifier.Notify(ctx, message, int64(companyId)); err != nil {
			s.logger.Printf("Error notifying graduate available for hire: %s\n", err)
		}

		return err

	})

	return search, nil
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

func (s *service) FindExperienced(ctx context.Context, companyId uint64) (*Search, error) {
	finishTime := time.Now().Add(SEARCH_DURATION)
	search, err := s.repository.StartSearch(ctx, finishTime, companyId)
	if err != nil {
		return nil, err
	}

	s.timer.Add(search.Id, SEARCH_DURATION, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.repository.DeleteSearch(ctx, search.Id, companyId); err != nil {
			return err
		}

		candidate, err := s.GetExperienced(ctx, companyId)

		message := fmt.Sprintf("%s is available for hire", candidate.Name)
		if err := s.notifier.Notify(ctx, message, int64(companyId)); err != nil {
			s.logger.Printf("Error notifying experienced available for hire: %s\n", err)
		}

		return err
	})

	return search, nil
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

func (s *service) CancelSearch(ctx context.Context, searchId, companyId uint64) error {
	err := s.repository.DeleteSearch(ctx, searchId, companyId)
	if err != nil {
		return err
	}

	s.timer.Remove(searchId)
	return nil
}

func (s *service) HireStaff(ctx context.Context, staffId, companyId uint64) (*Staff, error) {
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	currentEmployer := staff.Employer

	if (staff.Poacher != nil && *staff.Poacher != companyId) || (staff.Poacher == nil && staff.Employer != companyId) {
		return nil, ErrStaffNotFound
	}

	// If hiring experienced, update the salary and clean up poaching fields
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

	message := fmt.Sprintf("%s accepted your offer", staff.Name)
	if err := s.notifier.Notify(ctx, message, int64(companyId)); err != nil {
		s.logger.Printf("Error notifying poacher of offer accepted: %s\n", err)
	}

	message = fmt.Sprintf("%s accepted the offer and no longer works for us", staff.Name)
	if err := s.notifier.Notify(ctx, message, int64(currentEmployer)); err != nil {
		s.logger.Printf("Error notifying employer of offer accepted: %s\n", err)
	}

	return staff, nil
}

func (s *service) MakeOffer(ctx context.Context, offer, staffId, companyId uint64) (*Staff, error) {
	// Must save the offer and notify the current employer of the staff
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	if staff.Poacher == nil || *staff.Poacher != companyId {
		return nil, ErrStaffNotFound
	}

	if offer <= staff.Salary {
		return nil, server.NewBusinessRuleError("offer is too low")
	}

	staff.Offer = offer
	if err := s.repository.UpdateStaff(ctx, staff); err != nil {
		return nil, err
	}

	message := fmt.Sprintf("%s received an offer of $ %.2f", staff.Name, float64(staff.Offer)/100)
	if err := s.notifier.Notify(ctx, message, int64(companyId)); err != nil {
		s.logger.Printf("Error notifying offer: %s\n", err)
	}

	s.timer.Add(staffId, 48*time.Hour, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := s.HireStaff(ctx, staffId, companyId)
		return err
	})

	return staff, nil
}

func (s *service) IncreaseSalary(ctx context.Context, salary, staffId, companyId uint64) (*Staff, error) {
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	if staff.Employer != companyId {
		return nil, ErrStaffNotFound
	}

	if salary <= staff.Salary {
		return nil, server.NewBusinessRuleError("new salary must be higher than current salary")
	}

	var hasPoacher bool
	if staff.Poacher != nil {
		hasPoacher = true
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

	if hasPoacher {
		message := fmt.Sprintf("%s has declined your offer", staff.Name)
		if err := s.notifier.Notify(ctx, message, int64(companyId)); err != nil {
			s.logger.Printf("Error notifying offer declined: %s\n", err)
		}
	}

	return staff, nil
}

func (s *service) Train(ctx context.Context, staffId, companyId uint64) (*Training, error) {
	staff, err := s.repository.GetStaffById(ctx, staffId)
	if err != nil {
		return nil, err
	}

	if staff.Employer != companyId {
		return nil, ErrStaffNotFound
	}

	// Calculate time (relative to skill)
	duration := TRAINING_DURATION + time.Duration(int64(staff.Skill)/10)*time.Hour

	// Save training
	training, err := s.repository.SaveTraining(ctx, &Training{
		StaffId:    staffId,
		CompanyId:  companyId,
		FinishesAt: time.Now().Add(duration),
		Investment: 1000000 + (1000000 * (uint64(staff.Skill) / 10)),
	})

	if err != nil {
		return nil, err
	}

	s.timer.Add(staffId, TRAINING_DURATION, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		return s.FinishTraining(ctx, training.Id, companyId)
	})

	return training, nil
}

func (s *service) FinishTraining(ctx context.Context, trainingId, companyId uint64) error {
	training, err := s.repository.GetTraining(ctx, trainingId, companyId)
	if err != nil {
		return err
	}

	staff, err := s.repository.GetStaffById(ctx, training.StaffId)
	if err != nil {
		return err
	}

	// Complete training
	training.CompletedAt = time.Now()

	// Calculate points (relative to talent, e.g., rand(0, talent / 10))
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	base := int(math.Min(float64(staff.Talent), float64(staff.Talent/10)))
	training.Result = uint8(randomizer.Intn(base) + (base / 2))

	// Save new skill
	return s.repository.UpdateTraining(ctx, training)
}
