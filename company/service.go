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

type Registration struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Confirm  string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type Company struct {
	Id        uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	Pass      string     `db:"password" json:"-"`
	LastLogin *time.Time `db:"last_login" json:"last_login"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/companies")
	repository := NewRepository(conn)

	group.POST("/register", func(c echo.Context) error {
		registration := new(Registration)
		if err := c.Bind(registration); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if err := c.Validate(registration); err != nil {
			return err
		}

		hashedPassword, err := HashPassword(registration.Password)
		if err != nil {
			return err
		}

		registration.Password = hashedPassword
		company, err := repository.Register(registration)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, company)
	})

	group.POST("/login", func(c echo.Context) error {
		credentials := struct {
			Email string `form:"email" json:"email" validate:"required,email"`
			Pass  string `form:"password" json:"password" validate:"required"`
		}{}

		if err := c.Bind(&credentials); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if err := c.Validate(&credentials); err != nil {
			return err
		}

		company, err := repository.GetByEmail(credentials.Email)
		if err != nil || company == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, server.Response{
				Type:   "failure",
				Status: http.StatusUnauthorized,
				Data:   map[string]any{"error": "invalid credentials"},
			})
		}

		if err := ComparePassword(company.Pass, credentials.Pass); err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, server.Response{
				Type:   "failure",
				Status: http.StatusUnauthorized,
				Data:   map[string]any{"error": "invalid credentials"},
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
