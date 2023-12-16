package research

import (
	"api/accounting"
	"api/database"
	"api/research/staff"
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type (
	Repository interface {
		GetQuality(ctx context.Context, resourceId, companyId uint64) (Quality, error)
		IsStaffBusy(ctx context.Context, staffIds []uint64, companyId uint64) (bool, error)

		GetResearch(ctx context.Context, researchId uint64) (*Research, error)
		SaveResearch(ctx context.Context, finishesAt time.Time, investment int, staffIds []uint64, resourceId, companyId uint64) (*Research, error)
		CompleteResearch(ctx context.Context, research *Research) (*Research, error)
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

func (r *goquRepository) GetQuality(ctx context.Context, resourceId, companyId uint64) (Quality, error) {
	var quality Quality

	_, err := r.builder.
		Select(
			goqu.I("quality"),
			goqu.I("patents"),
			goqu.I("resource_id"),
		).
		From(goqu.T("resources_qualities")).
		Where(goqu.And(
			goqu.I("company_id").Eq(companyId),
			goqu.I("resource_id").Eq(resourceId),
		)).
		ScanStructContext(ctx, &quality)

	return quality, err
}

func (r *goquRepository) IsStaffBusy(ctx context.Context, staffIds []uint64, companyId uint64) (bool, error) {
	count := 0

	_, err := r.builder.
		Select(goqu.COUNT("*")).
		From(goqu.T("assigned_staff").As("as")).
		InnerJoin(
			goqu.T("researches").As("r"),
			goqu.On(goqu.I("as.research_id").Eq(goqu.I("r.id"))),
		).
		Where(goqu.And(
			goqu.I("r.completed_at").IsNull(),
			goqu.I("as.staff_id").In(staffIds),
		)).
		ScanValContext(ctx, &count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *goquRepository) GetResearch(ctx context.Context, researchId uint64) (*Research, error) {
	research := new(Research)

	found, err := r.builder.
		Select(goqu.Star()).
		From(goqu.T("researches")).
		Where(goqu.I("id").Eq(researchId)).
		ScanStructContext(ctx, research)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrResearchNotFound
	}

	staff := make([]*staff.Staff, 0)

	err = r.builder.
		Select(goqu.I("s.id"), goqu.I("s.name"), goqu.I("s.skill")).
		From(goqu.T("research_staff").As("s")).
		InnerJoin(
			goqu.T("assigned_staff").As("as"),
			goqu.On(goqu.I("as.staff_id").Eq(goqu.I("s.id"))),
		).
		Where(goqu.I("as.research_id").Eq(research.Id)).
		ScanStructsContext(ctx, &staff)

	if err != nil {
		return nil, err
	}

	research.AssignedStaff = staff

	return research, nil
}

func (r *goquRepository) SaveResearch(ctx context.Context, finishesAt time.Time, investment int, staffIds []uint64, resourceId, companyId uint64) (*Research, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if err := r.accountingRepo.RegisterTransaction(
		&database.DB{TxDatabase: tx},
		accounting.Transaction{
			Classification: accounting.RESEARCH,
			Description:    "Payment of research",
			Value:          -investment,
		},
		companyId,
	); err != nil {
		return nil, err
	}

	result, err := tx.
		Insert(goqu.T("researches")).
		Rows(goqu.Record{
			"patents":     0,
			"investment":  investment,
			"finishes_at": finishesAt,
			"company_id":  companyId,
			"resource_id": resourceId,
		}).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	researchId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	staffRows := make([]goqu.Record, 0)
	for _, staffId := range staffIds {
		staffRows = append(staffRows, goqu.Record{
			"staff_id":    staffId,
			"research_id": researchId,
		})
	}

	_, err = tx.Insert(goqu.T("assigned_staff")).
		Rows(staffRows).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetResearch(ctx, uint64(researchId))
}

func (r *goquRepository) CompleteResearch(ctx context.Context, research *Research) (*Research, error) {
	tx, err := r.builder.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Update(goqu.T("researches")).
		Set(research).
		Where(goqu.I("id").Eq(research.Id)).
		Executor().
		Exec()

	if err != nil {
		return nil, err
	}

	quality, err := r.GetQuality(ctx, research.ResourceId, research.CompanyId)
	if err != nil {
		return nil, err
	}

	newQuality := quality.Quality + (quality.Patents+uint8(research.Patents))/((quality.Quality+1)*100)
	newPatents := (quality.Patents + uint8(research.Patents)) % ((quality.Quality + 1) * 100)

	if newPatents > 0 || newQuality > 0 {
		if quality.ResourceId == 0 {
			_, err = tx.
				Insert(goqu.T("resources_qualities")).
				Rows(goqu.Record{
					"quality":     newQuality,
					"patents":     newPatents,
					"company_id":  research.CompanyId,
					"resource_id": research.ResourceId,
				}).
				Executor().
				Exec()

			if err != nil {
				return nil, err
			}
		} else {
			_, err := tx.
				Update(goqu.T("resources_qualities")).
				Set(goqu.Record{
					"quality": newQuality,
					"patents": newPatents,
				}).
				Where(goqu.And(
					goqu.I("company_id").Eq(research.CompanyId),
					goqu.I("resource_id").Eq(research.ResourceId),
				)).
				Executor().
				Exec()

			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return research, nil
}
