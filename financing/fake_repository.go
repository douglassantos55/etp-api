package financing

import (
	"api/company"
	"context"
)

type fakeRepository struct {
	lastLoanId  int64
	lastBondId  int64
	loans       map[int64]*Loan
	bonds       map[int64]*Bond
	companyRepo company.Repository
}

func NewFakeRepository(companyRepo company.Repository) Repository {
	return &fakeRepository{
		lastLoanId:  4,
		lastBondId:  2,
		companyRepo: companyRepo,
		bonds: map[int64]*Bond{
			1: {
				Id:           1,
				Amount:       1_000_000_00,
				InterestRate: 0.1,
				CompanyId:    1,
				Creditors: []*Creditor{
					{
						Principal:       500_000_00,
						InterestRate:    0.1,
						InterestPaid:    100_000_00,
						DelayedPayments: 0,
						PrincipalPaid:   100_000_00,
					},
				},
			},
			2: {
				Id:           2,
				Amount:       2_000_000_00,
				InterestRate: 0.1,
				CompanyId:    3,
				Creditors: []*Creditor{
					{
						Principal:       500_000_00,
						InterestRate:    0.1,
						InterestPaid:    100_000_00,
						DelayedPayments: 0,
						PrincipalPaid:   100_000_00,
					},
					{
						Principal:       1_500_000_00,
						InterestRate:    0.1,
						InterestPaid:    0,
						DelayedPayments: 0,
						PrincipalPaid:   0,
					},
				},
			},
		},
		loans: map[int64]*Loan{
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
	loan, ok := r.loans[loanId]
	if !ok {
		return nil, ErrLoanNotFound
	}
	return loan, nil
}

func (r *fakeRepository) SaveLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	r.lastLoanId++
	loan.Id = r.lastLoanId
	return loan, nil
}

func (r *fakeRepository) UpdateLoan(ctx context.Context, loan *Loan) (*Loan, error) {
	r.loans[loan.Id] = loan
	return loan, nil
}

func (r *fakeRepository) PayLoanInterest(ctx context.Context, loan *Loan) error {
	loan.DelayedPayments = 0
	loan.InterestPaid += loan.GetInterest()
	r.loans[loan.Id] = loan
	return nil
}

func (r *fakeRepository) ForcePrincipalPayment(ctx context.Context, terrains []int8, loan *Loan) error {
	loan.PrincipalPaid = loan.GetPrincipal()
	r.loans[loan.Id] = loan

	company, _ := r.companyRepo.GetById(ctx, uint64(loan.CompanyId))
	company.AvailableTerrains -= int8(len(terrains))

	return nil
}

func (r *fakeRepository) GetBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	bonds := make([]*Bond, 0)
	for _, bond := range r.bonds {
		bonds = append(bonds, bond)
	}
	return bonds, nil
}

func (r *fakeRepository) GetBond(ctx context.Context, bondId int64) (*Bond, error) {
	bond, ok := r.bonds[bondId]
	if !ok {
		return nil, ErrBondNotFound
	}
	return bond, nil
}

func (r *fakeRepository) SaveBond(ctx context.Context, bond *Bond) (*Bond, error) {
	r.lastBondId++
	bond.Id = r.lastBondId
	r.bonds[bond.Id] = bond
	return bond, nil
}

func (r *fakeRepository) PayBondInterest(ctx context.Context, bond *Bond, creditor *Creditor) error {
	creditor.DelayedPayments = 0
	creditor.InterestPaid += creditor.GetInterest()

	company, _ := r.companyRepo.GetById(ctx, uint64(bond.CompanyId))
	company.AvailableCash -= int(creditor.GetInterest())

	return nil
}

func (r *fakeRepository) SaveCreditor(ctx context.Context, bond *Bond, creditor *Creditor) (*Creditor, error) {
	return creditor, nil
}
