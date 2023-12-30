package financing

import (
	"api/company"
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
	ErrCreditorNotFound          = server.NewBusinessRuleError("creditor not found")
)

type (
	Bond struct {
		Id           int64   `db:"id" json:"id" goqu:"skipinsert"`
		Amount       int64   `db:"amount" json:"amount"`
		InterestRate float64 `db:"interest_rate" json:"interest_rate"`
		CompanyId    int64   `db:"company_id" json:"-"`
		Purchased    int64   `db:"purchased" json:"purchased"`

		Company   *company.Company `db:"company" json:"company,omitempty" goqu:"skipinsert,skipupdate"`
		Creditors []*Creditor      `json:"creditors,omitempty" goqu:"skipinsert,skipupdate"`
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
		GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error)
		GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error)
		EmitBond(ctx context.Context, rate float64, amount, companyId int64) (*Bond, error)
		BuyBond(ctx context.Context, amount, bondId, companyId int64) (*Bond, *Creditor, error)
		PayBondInterest(ctx context.Context, creditor *Creditor, bond *Bond) error
		BuyBackBond(ctx context.Context, amount, bondId, creditorId, companyId int64) (*Creditor, error)
	}

	service struct {
		repository Repository
		companySvc company.Service
	}
)

func (b *Bond) GetCreditor(creditorId int64) (*Creditor, error) {
	for _, creditor := range b.Creditors {
		if creditor.Id == uint64(creditorId) {
			return creditor, nil
		}
	}
	return nil, ErrCreditorNotFound
}

func (c *Creditor) GetPrincipal() int64 {
	return c.Principal - c.PrincipalPaid
}

func (c *Creditor) GetInterest() int64 {
	return int64(c.InterestRate * float64(c.GetPrincipal()))
}

func NewService(repository Repository, companySvc company.Service) Service {
	return &service{repository, companySvc}
}

func (s *service) GetBonds(ctx context.Context, page, limit uint) ([]*Bond, error) {
	return s.repository.GetBonds(ctx, page, limit)
}

func (s *service) GetCompanyBonds(ctx context.Context, companyId int64) ([]*Bond, error) {
	return s.repository.GetCompanyBonds(ctx, companyId)
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

func (s *service) PayBondInterest(ctx context.Context, creditor *Creditor, bond *Bond) error {
	emissor, err := s.companySvc.GetById(ctx, uint64(bond.CompanyId))
	if err != nil {
		return err
	}

	if emissor.AvailableCash < int(creditor.GetInterest()) {
		// TODO: notify creditor that there was no payment
	} else {
		err = s.repository.PayBondInterest(ctx, bond, creditor)
		if err != nil {
			return err
		}
		// TODO: notify creditor that payment was executed
	}

	return nil
}

func (s *service) BuyBond(ctx context.Context, amount, bondId, companyId int64) (*Bond, *Creditor, error) {
	bond, err := s.repository.GetBond(ctx, bondId)
	if err != nil {
		return nil, nil, err
	}

	if amount > (bond.Amount - bond.Purchased) {
		return nil, nil, ErrAmountHigherThanAvailable
	}

	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, nil, err
	}

	if company.AvailableCash < int(amount) {
		return nil, nil, ErrNotEnoughCash
	}

	creditor := &Creditor{
		Company:      company,
		Principal:    amount,
		InterestRate: bond.InterestRate,
		PayableFrom:  time.Now().Add(2 * Week),
	}

	creditor, err = s.repository.SaveCreditor(ctx, bond, creditor)
	if err != nil {
		return nil, nil, err
	}

	return bond, creditor, nil
}

func (s *service) BuyBackBond(ctx context.Context, amount, bondId, creditorId, companyId int64) (*Creditor, error) {
	bond, err := s.repository.GetBond(ctx, bondId)
	if err != nil {
		return nil, err
	}

	if bond.CompanyId != companyId {
		return nil, ErrBondNotFound
	}

	creditor, err := bond.GetCreditor(creditorId)
	if err != nil {
		return nil, err
	}

	if amount > creditor.GetPrincipal() {
		return nil, ErrAmountHigherThanPrincipal
	}

	company, err := s.companySvc.GetById(ctx, uint64(companyId))
	if err != nil {
		return nil, err
	}

	if company.AvailableCash < int(amount) {
		return nil, ErrNotEnoughCash
	}

	return s.repository.BuyBackBond(ctx, amount, creditor, bond)
}
