package bonds

import (
	"api/company"
	"context"
)

type fakeRepository struct {
	lastId int64
	bonds  map[int64]*Bond

	companyRepo company.Repository
}

func NewFakeRepository(companyRepo company.Repository) Repository {
	return &fakeRepository{
		lastId:      2,
		companyRepo: companyRepo,
		bonds: map[int64]*Bond{
			1: {
				Id:           1,
				Amount:       1_000_000_00,
				InterestRate: 0.1,
				CompanyId:    1,
				Purchased:    500_000_00,
				Creditors: []*Creditor{
					{
						Company:         &company.Company{Id: 2},
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
	}
}

func (r *fakeRepository) GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error) {
	index := 0
	bonds := make([]*Bond, 0, limit)
	for _, bond := range r.bonds {
		if len(bonds) >= int(limit) {
			break
		}

		if int(page*limit) <= index {
			bonds = append(bonds, bond)
		}

		index++
	}
	return bonds, nil
}

func (r *fakeRepository) GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	bonds := make([]*Bond, 0)
	for _, bond := range r.bonds {
		if bond.CompanyId == companyId {
			bonds = append(bonds, bond)
		}
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
	r.lastId++
	bond.Id = r.lastId
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
	creditor.AvailableCash -= int(creditor.Principal)
	return creditor, nil
}

func (r *fakeRepository) BuyBackBond(ctx context.Context, amount int64, creditor *Creditor, bond *Bond) (*Creditor, error) {
	creditor.PrincipalPaid += amount
	return creditor, nil
}
