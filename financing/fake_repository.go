package financing

import "context"

type fakeRepository struct {
}

func NewFakeRepository() Repository {
	return &fakeRepository{}
}

func (r *fakeRepository) SaveLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	return loan, nil
}
