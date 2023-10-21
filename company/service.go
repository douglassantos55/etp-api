package company

import (
	"api/database"
	"api/server"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Company struct {
	Id        uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate" validate:"-"`
	Name      string     `db:"name" json:"name" validate:"required"`
	Email     string     `db:"email" json:"email,omitempty" validate:"required,email"`
	Pass      string     `db:"password" json:"password,omitempty" validate:"required"`
	LastLogin *time.Time `db:"last_login" json:"last_login" validate:"-"`
	CreatedAt string     `db:"created_at" json:"created_at" validate:"-"`
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

	group.POST("/login", func(c echo.Context) error {
		credentials := struct {
			Email string `form:"email" validate:"required,email"`
			Pass  string `form:"password" validate:"required"`
		}{}

		if err := c.Bind(&credentials); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if err := c.Validate(&credentials); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		company, err := repository.GetByEmail(credentials.Email)
		if err != nil || company == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, server.Response{
				Type:   "failure",
				Status: http.StatusUnauthorized,
				Data:   map[string]any{"error": "invalid email"},
			})
		}

		if err := ComparePassword(company.Pass, credentials.Pass); err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, server.Response{
				Type:   "failure",
				Status: http.StatusUnauthorized,
				Data:   map[string]any{"error": "invalid password"},
			})
		}

		c.SetCookie(&http.Cookie{
			Secure:   true,
			HttpOnly: false,
			Name:     "company_id",
			SameSite: http.SameSiteNoneMode,
			Value:    fmt.Sprintf("%d", company.Id),
			Expires:  time.Now().Add(10 * time.Minute),
		})

		return c.JSON(http.StatusOK, server.Response{
			Status:   http.StatusOK,
			Type:     "redirect",
			Location: "/",
		})
	})
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
