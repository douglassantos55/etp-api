package loans

import (
	"api/company"
	"api/financing"
	"api/server"
	"context"
	"fmt"
	"time"
)

const (
	Week                 = 7 * 24 * time.Hour
	MAX_DELAYED_PAYMENTS = 4
)

var (
	ErrNotEnoughCash             = server.NewBusinessRuleError("not enough cash")
	ErrAmountHigherThanPrincipal = server.NewBusinessRuleError("amount is higher than principal")
	ErrAmountHigherThanAvailable = server.NewBusinessRuleError("amount is higher than available")
)

type (
	Loan struct {
		Id              int64     `db:"id" json:"id" goqu:"skipinsert"`
		InterestRate    float64   `db:"interest_rate" json:"interest_rate"`
		InterestPaid    int64     `db:"interest_paid" json:"interest_paid"`
		PayableFrom     time.Time `db:"payable_from" json:"payable_from"`
		Principal       int64     `db:"principal" json:"principal"`
		PrincipalPaid   int64     `db:"principal_paid" json:"principal_paid"`
		CompanyId       int64     `db:"company_id" json:"-"`
		DelayedPayments int8      `db:"delayed_payments" json:"delayed_payments"`
	}

	Service interface {
		GetLoans(ctx context.Context, companyId int64) ([]*Loan, error)
		TakeLoan(ctx context.Context, amount, companyId int64) (*Loan, error)
		PayLoanInterest(ctx context.Context, loanId, companyId int64) (bool, error)
		BuyBackLoan(ctx context.Context, amount, loanId, companyId int64) (*Loan, error)
	}

	service struct {
		repository   Repository
		companySvc   company.Service
		financingSvc financing.Service
	}
)

func (l *Loan) GetPrincipal() int64 {
	return l.Principal - l.PrincipalPaid
}

func (l *Loan) GetInterest() int64 {
	return int64(float64(l.GetPrincipal()) * l.InterestRate)
}

func NewService(repository Repository, companySvc company.Service, financingSvc financing.Service) Service {
	return &service{repository, companySvc, financingSvc}
}

func (s *service) GetLoans(ctx context.Context, companyId int64) ([]*Loan, error) {
	return s.repository.GetLoans(ctx, companyId)
}

func (s *service) BuyBackLoan(ctx context.Context, amount, loanId, companyId int64) (*Loan, error) {
	loan, err := s.repository.GetLoan(ctx, loanId, companyId)
	if err != nil {
		return nil, err
	}

	if amount > loan.GetPrincipal() {
		return nil, ErrAmountHigherThanPrincipal
	}

	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, err
	}

	if company.AvailableCash < int(amount) {
		return nil, ErrNotEnoughCash
	}

	return s.repository.BuyBackLoan(ctx, amount, loan)
}

func (s *service) TakeLoan(ctx context.Context, amount int64, companyId int64) (*Loan, error) {
	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, err
	}

	if amount > company.GetCreditScore() {
		return nil, fmt.Errorf("amount must not be higher than %.2f", float64(company.GetCreditScore()/100))
	}

	rates, err := s.financingSvc.GetEffectiveRates(ctx)
	if err != nil {
		return nil, err
	}

	loan, err := s.repository.SaveLoan(ctx, &Loan{
		Principal:    amount,
		CompanyId:    companyId,
		PayableFrom:  time.Now().Add(4 * Week),
		InterestRate: rates.Interest,
	})

	if err != nil {
		return nil, err
	}

	return loan, nil
}

func (s *service) PayLoanInterest(ctx context.Context, loanId, companyId int64) (bool, error) {
	loan, err := s.repository.GetLoan(ctx, loanId, companyId)
	if err != nil {
		return true, err
	}

	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return true, err
	}

	interest := loan.GetInterest()

	// If can't pay 4 consecutive installments lose terrains to cover the debt
	if company.AvailableCash < int(interest) {
		loan.DelayedPayments++

		if loan.DelayedPayments >= MAX_DELAYED_PAYMENTS {
			return false, s.forcePayment(ctx, loan, company)
		}

		if _, err := s.repository.UpdateLoan(ctx, loan); err != nil {
			return true, err
		}

		// TODO: notify company
		return true, nil
	}

	return true, s.repository.PayLoanInterest(ctx, loan)
}

func (s *service) forcePayment(ctx context.Context, loan *Loan, company *company.Company) error {
	var total int64
	terrains := []int8{}
	principal := loan.GetPrincipal()

	for i := company.AvailableTerrains; i > 0; i++ {
		terrains = append(terrains, i)
		total += company.TerrainValue(i)

		if total >= principal {
			break
		}
	}

	if err := s.repository.ForcePrincipalPayment(ctx, terrains, loan); err != nil {
		return err
	}

	// TODO: notify company about the whole thing
	return nil
}
