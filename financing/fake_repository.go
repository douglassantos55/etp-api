package financing

import (
	"api/company"
)

type fakeRepository struct {
}

func NewFakeRepository(companyRepo company.Repository) Repository {
	return &fakeRepository{}
}
