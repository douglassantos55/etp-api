package company

import (
	"api/database"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Company struct {
	Id    uint64 `db:"id" json:"id" goqu:"skipinsert,skipupdate" validate:"-"`
	Name  string `db:"name" json:"name" validate:"required"`
	Email string `db:"email" json:"email,omitempty" validate:"required,email"`
	Pass  string `db:"password" json:"password,omitempty" validate:"required"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/companies")
	repository := NewRepository(conn)

	group.POST("/register", func(c echo.Context) error {
		company := new(Company)
		if err := c.Bind(company); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if err := c.Validate(company); err != nil {
			return err
		}

		hashedPassword, err := HashPassword(company.Pass)
		if err != nil {
			return err
		}

		company.Pass = hashedPassword
		if err := repository.SaveCompany(company); err != nil {
			return err
		}

		company.Pass = ""
		return c.JSON(http.StatusCreated, company)
	})

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
}
