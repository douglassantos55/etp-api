package company

import (
	"api/auth"
	"api/database"
	"api/server"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
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

		hashedPassword, err := auth.HashPassword(registration.Password)
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
			return echo.NewHTTPError(http.StatusBadRequest, server.ValidationErrors{
				Errors: map[string]string{"email": "invalid credentials"},
			})
		}

		if err := auth.ComparePassword(company.Pass, credentials.Pass); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, server.ValidationErrors{
				Errors: map[string]string{"email": "invalid credentials"},
			})
		}

		token, err := auth.GenerateToken(company.Id, server.GetJwtSecret())
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})
}
