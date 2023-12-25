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
		Id              int64     `db:"id" json:"id" goqu:"skipinsert"`
		InterestRate    float64   `db:"interest_rate" json:"interest_rate"`
		InterestPaid    int64     `db:"interest_paid" json:"interest_paid"`
		PayableFrom     time.Time `db:"payable_from" json:"payable_from"`
		Principal       int64     `db:"principal" json:"principal"`
		PrincipalPaid   int64     `db:"principal_paid" json:"principal_paid"`
		CompanyId       int64     `db:"company_id" json:"-"`
		DelayedPayments int8      `db:"delayed_payments" json:"delayed_payments"`
	}

	Bond struct {
		Id           int64   `db:"id" json:"id" goqu:"skipinsert"`
		Amount       int64   `db:"amount" json:"amount"`
		InterestRate float64 `db:"interest_rate" json:"interest_rate"`
		CompanyId    int64   `db:"company_id" json:"-"`

		Company   *company.Company `db:"company" json:"company" goqu:"skipinsert,skipupdate"`
		Creditors []*Creditor      `json:"creditors" goqu:"skipinsert,skipupdate"`
	}

	Creditor struct {
		*company.Company `db:"company"`
		InterestRate     float64   `db:"interest_rate" json:"interest_rate"`
		InterestPaid     int64     `db:"interest_paid" json:"interest_paid"`
		PayableFrom      time.Time `db:"payable_from" json:"payable_from"`
		Principal        int64     `db:"principal" json:"principal"`
		PrincipalPaid    int64     `db:"principal_paid" json:"principal_paid"`
		DelayedPayments  int8      `db:"delayed_payments" json:"delayed_payments"`
	}

	Service interface {
		TakeLoan(ctx context.Context, amount, companyId int64) (*Loan, error)
		PayLoanInterest(ctx context.Context, loanId, companyId int64) (bool, error)

		EmitBond(ctx context.Context, rate float64, amount, companyId int64) (*Bond, error)
		PayBondInterest(ctx context.Context, bondId, companyId int64) error
	}

	service struct {
		repository Repository
		companySvc company.Service
	}
)

func (c *Creditor) GetPrincipal() int64 {
	return c.Principal - c.PrincipalPaid
}

func (c *Creditor) GetInterest() int64 {
	return int64(c.InterestRate * float64(c.GetPrincipal()))
}

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

	if amount > company.GetCreditScore() {
		return nil, fmt.Errorf("amount must not be higher than %.2f", float64(company.GetCreditScore()/100))
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

func (s *service) EmitBond(ctx context.Context, rate float64, amount, companyId int64) (*Bond, error) {
	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, err
	}

	if amount > company.GetCreditScore() {
		return nil, fmt.Errorf("amount must not be higher than %.2f", float64(company.GetCreditScore())/100)
	}

	bond, err := s.repository.SaveBond(ctx, &Bond{
		Amount:       amount,
		InterestRate: rate,
		CompanyId:    companyId,
	})

	if err != nil {
		return nil, err
	}

	return bond, nil
}

func (s *service) PayBondInterest(ctx context.Context, bondId, companyId int64) error {
	bond, err := s.repository.GetBond(ctx, bondId, companyId)
	if err != nil {
		return err
	}

	emissor, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return err
	}

	for _, creditor := range bond.Creditors {
		interest := creditor.GetInterest()
		if emissor.AvailableCash < int(interest) {
			// TODO: notify creditor
			continue
		}

		err := s.repository.PayBondInterest(ctx, bond, creditor)
		if err != nil {
			return err
		}
	}

	return nil
}
