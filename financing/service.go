package financing

import (
	"api/company"
	"context"
	"fmt"
	"time"
)

const (
	Week                 = 7 * 24 * time.Hour
	MAX_DELAYED_PAYMENTS = 4
)

type (
	Loan struct {
		Id           int64   `db:"id" json:"id" goqu:"skipinsert"`
		InterestRate float64 `db:"interest_rate" json:"interest_rate"`
		InterestPaid int64   `db:"interest_paid" json:"interest_paid"`
		// From when principal may be paid
		PayableFrom     time.Time `db:"payable_from" json:"payable_from"`
		Principal       int64     `db:"amount" json:"amount"`
		PrincipalPaid   int64     `db:"principal_paid" json:"principal_paid"`
		DelayedPayments int8      `db:"delayed_payments" json:"delayed_payments"`
		CompanyId       int64     `db:"company_id" json:"-"`
	}

	Service interface {
		TakeLoan(ctx context.Context, amount, companyId int64) (*Loan, error)
		PayInterest(ctx context.Context, loan *Loan) error
	}

	service struct {
		repository Repository
		companySvc company.Service
	}
)

func (l *Loan) GetPrincipal() int64 {
	return l.Principal - l.PrincipalPaid
}

func (l *Loan) GetInterest() int64 {
	return int64(float64(l.GetPrincipal()) * l.InterestRate)
}

func NewService(repository Repository, companySvc company.Service) Service {
	return &service{repository, companySvc}
}

func (s *service) TakeLoan(ctx context.Context, amount int64, companyId int64) (*Loan, error) {
	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, err
	}

	total := 1_000_000_00 + (500_000_00 * (int(company.AvailableTerrains-1) / 5)) + (100_000_00 * int(company.AvailableTerrains))
	if amount > int64(total) {
		return nil, fmt.Errorf("amount must not be higher than %.2f", float64(total/100))
	}

	loan, err := s.repository.SaveLoan(ctx, &Loan{
		Principal:   amount,
		CompanyId:   companyId,
		PayableFrom: time.Now().Add(3 * Week),
		// TODO: use interest rate relative to inflation
		InterestRate: 0.15,
	})

	if err != nil {
		return nil, err
	}

	return loan, nil
}

func (s *service) PayInterest(ctx context.Context, loan *Loan) error {
	company, err := s.companySvc.GetById(ctx, uint64(loan.CompanyId))
	if err != nil {
		return err
	}

	principal := loan.GetPrincipal()
	interest := loan.GetInterest()

	// If can't pay 4 consecutive installments lose terrains to cover the debt
	if company.AvailableCash < int(interest) {
		loan.DelayedPayments++

		if loan.DelayedPayments >= MAX_DELAYED_PAYMENTS {
			total := 0
			for i := company.AvailableTerrains; i > 0; i++ {
				total += 1_000_000_00 + 500_000_00*((int(i)-1)/5) + (100_000_00 * int(i))

				if total >= int(principal) {
					if err := s.repository.ForcePrincipalPayment(ctx, i, loan); err != nil {
						return err
					}
					break
				}
			}

			// TODO: notify company about the whole thing
			return nil
		}

		if _, err := s.repository.UpdateLoan(ctx, loan); err != nil {
			return err
		}

		// TODO: notify company
		return nil
	}

	loan.DelayedPayments = 0
	return s.repository.PayInterest(ctx, loan)
}
