package financing

import (
	"api/company"
	"context"
)

type fakeRepository struct {
	lastId      int64
	data        map[int64]*Loan
	companyRepo company.Repository
}

func NewFakeRepository(companyRepo company.Repository) Repository {
	return &fakeRepository{
		lastId:      4,
		companyRepo: companyRepo,
		data: map[int64]*Loan{
			1: {
				Id:              1,
				Principal:       1_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 2,
				PrincipalPaid:   300_000_00,
			},
			2: {
				Id:              2,
				Principal:       1_000_000_00,
				CompanyId:       3,
				InterestRate:    0.1,
				DelayedPayments: 3,
				PrincipalPaid:   500_000_00,
			},
			3: {
				Id:              3,
				Principal:       4_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 3,
			},
			4: {
				Id:              4,
				Principal:       4_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 3,
			},
		},
	}
}

func (r *fakeRepository) GetLoan(ctx context.Context, loanId, companyId int64) (*Loan, error) {
	loan, ok := r.data[loanId]
	if !ok {
		return nil, ErrLoanNotFound
	}
	return loan, nil
}

func (r *fakeRepository) SaveLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	r.lastId++
	loan.Id = r.lastId
	return loan, nil
}

func (r *fakeRepository) UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	r.data[loan.Id] = loan
	return loan, nil
}

func (r *fakeRepository) PayInterest(ctx context.Context, loan *Loan) error {
	loan.DelayedPayments = 0
	loan.InterestPaid += loan.GetInterest()
	r.data[loan.Id] = loan
	return nil
}

func (r *fakeRepository) ForcePrincipalPayment(ctx context.Context, terrains []int8, loan *Loan) error {
	loan.PrincipalPaid = loan.GetPrincipal()
	r.data[loan.Id] = loan

	company, _ := r.companyRepo.GetById(ctx, uint64(loan.CompanyId))
	company.AvailableTerrains -= int8(len(terrains))

	return nil
}
