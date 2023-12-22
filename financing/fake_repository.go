package financing

import "context"

type fakeRepository struct {
	lastId int64
	data   map[int64]*Loan
}

func NewFakeRepository() Repository {
	return &fakeRepository{
		lastId: 2,
		data: map[int64]*Loan{
			1: {
				Id:              1,
				Principal:       1_000_000_00,
				CompanyId:       1,
				InterestRate:    0.1,
				DelayedPayments: 1,
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
		},
	}
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
	loan.InterestPaid += loan.GetInterest()
	r.data[loan.Id] = loan
	return nil
}
