package staff

import (
	"api/database"
	"context"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
)

var (
	ErrNoStaffFound  = errors.New("no staff found")
	ErrStaffNotFound = errors.New("staff not found")
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
	}

	goquRepository struct {
		builder *goqu.Database
		tx      *goqu.TxDatabase
	}
)

func NewRepository(conn *database.Connection) Repository {
	builder := goqu.New(conn.Driver, conn.DB)
	return &goquRepository{builder, nil}
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
