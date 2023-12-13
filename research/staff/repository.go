package staff

import (
	"api/accounting"
	"api/database"
	"context"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
)

var (
	ErrNoStaffFound     = errors.New("no staff found")
	ErrStaffNotFound    = errors.New("staff not found")
	ErrTrainingNotFound = errors.New("training not found")
)

type (
	Repository interface {
		GetStaff(ctx context.Context, companyId uint64) ([]*Staff, error)
		GetStaffById(ctx context.Context, staffId uint64) (*Staff, error)
		RandomStaff(ctx context.Context, companyId uint64) (*Staff, error)

		StartSearch(ctx context.Context, finishTime time.Time, companyId uint64) (*Search, error)
		DeleteSearch(ctx context.Context, searchId uint64) error

		SaveStaff(ctx context.Context, staff *Staff, companyId uint64) (*Staff, error)
		UpdateStaff(ctx context.Context, staff *Staff) error

		GetTraining(ctx context.Context, trainingId, companyId uint64) (*Training, error)
		SaveTraining(ctx context.Context, training *Training) (*Training, error)
		UpdateTraining(ctx context.Context, training *Training) error
	}

	goquRepository struct {
		builder        *goqu.Database
		accountingRepo accounting.Repository
	}
)

func NewRepository(conn *database.Connection, accountingRepo accounting.Repository) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder, accountingRepo}
}

func (r *goquRepository) GetStaff(ctx context.Context, companyId uint64) ([]*Staff, error) {
	staff := make([]*Staff, 0)

	err := r.builder.
		Select(goqu.Star()).
		From(goqu.T("research_staff")).
		Where(goqu.Or(
			goqu.I("company_id").Eq(companyId),
			goqu.I("poacher_id").Eq(companyId),
		)).
		ScanStructsContext(ctx, &staff)

	if err != nil {
		return nil, err
	}

	return staff, nil
}

func (r *goquRepository) GetStaffById(ctx context.Context, staffId uint64) (*Staff, error) {
	staff := new(Staff)

	found, err := r.builder.
		Select(goqu.Star()).
		From(goqu.T("research_staff")).
		Where(goqu.I("id").Eq(staffId)).
		ScanStructContext(ctx, staff)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrStaffNotFound
	}

	return staff, nil
}

func (r *goquRepository) StartSearch(ctx context.Context, finishTime time.Time, companyId uint64) (*Search, error) {
	search := &Search{FinishesAt: finishTime}

	result, err := r.builder.
		Insert(goqu.T("staff_searches")).
		Cols(search).
		Executor().
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	search.Id = uint64(id)
	return search, nil
}

func (r *goquRepository) DeleteSearch(ctx context.Context, searchId uint64) error {
	_, err := r.builder.
		Delete(goqu.T("staff_searchers")).
		Where(goqu.I("id").Eq(searchId)).
		Executor().
		ExecContext(ctx)

	return err
}

func (r *goquRepository) RandomStaff(ctx context.Context, companyId uint64) (*Staff, error) {
	staff := new(Staff)

	found, err := r.builder.
		Select(goqu.Star()).
		From(goqu.T("research_staff")).
		Where(goqu.And(
			goqu.I("poacher_id").IsNull(),
			goqu.I("company_id").Neq(companyId),
		)).
		ScanStructContext(ctx, staff)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrNoStaffFound
	}

	return staff, nil
}

func (r *goquRepository) SaveStaff(ctx context.Context, staff *Staff, companyId uint64) (*Staff, error) {
	result, err := r.builder.
		Insert(goqu.T("research_staff")).
		Rows(goqu.Record{
			"name":       staff.Name,
			"status":     staff.Status,
			"salary":     staff.Salary,
			"skill":      staff.Skill,
			"talent":     staff.Talent,
			"company_id": companyId,
		}).
		Executor().
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	staff.Id = uint64(id)
	return staff, nil
}

func (r *goquRepository) UpdateStaff(ctx context.Context, staff *Staff) error {
	_, err := r.builder.
		Update(goqu.T("research_staff")).
		Set(goqu.Record{
			"status":      staff.Status,
			"salary":      staff.Salary,
			"offer":       staff.Offer,
			"poarcher_id": staff.Poacher,
			"company_id":  staff.Employer,
		}).
		Where(goqu.I("id").Eq(staff.Id)).
		Executor().
		ExecContext(ctx)

	if err != nil {
		return err
	}

	return err
}

func (r *goquRepository) GetTraining(ctx context.Context, trainingId, companyId uint64) (*Training, error) {
	training := new(Training)

	found, err := r.builder.
		Select().
		From(goqu.T("trainings")).
		Where(goqu.And(
			goqu.I("id").Eq(trainingId),
			goqu.I("company_id").Eq(companyId),
		)).
		ScanStructContext(ctx, training)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrTrainingNotFound
	}

	return training, nil
}

func (r *goquRepository) SaveTraining(ctx context.Context, training *Training) (*Training, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Value:          -int(training.Investment),
			Description:    "Staff training",
			Classification: accounting.STAFF_TRAINING,
		},
		training.CompanyId,
	); err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("trainings")).
		Rows(training).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	training.Id = uint64(id)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return training, nil
}

func (r *goquRepository) UpdateTraining(ctx context.Context, training *Training) error {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.
		Update(goqu.T("trainings")).
		Set(training).
		Where(goqu.I("id").Eq(training.Id)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	_, err = tx.
		Update(goqu.T("research_staff")).
		Set(goqu.Record{
			"skill": goqu.Case().When(
				goqu.L("(? + ?) > 100", goqu.I("skill"), training.Result),
				100,
			).Else(goqu.L("? + ?", goqu.I("skill"), training.Result)),
		}).
		Where(goqu.I("id").Eq(training.StaffId)).
		Executor().
		Exec()

	if err != nil {
		return err
	}

	return tx.Commit()
}
