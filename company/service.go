package company

import (
	"api/auth"
	"api/database"
	"api/server"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type (
	Registration struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
		Confirm  string `json:"confirm_password" validate:"required,eqfield=Password"`
	}

	Company struct {
		Id        uint64     `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
		Name      string     `db:"name" json:"name"`
		Email     string     `db:"email" json:"email"`
		Pass      string     `db:"password" json:"-"`
		LastLogin *time.Time `db:"last_login" json:"last_login"`
		CreatedAt time.Time  `db:"created_at" json:"created_at"`
	}

	CompanyBuilding struct {
		Id              uint64  `db:"id" json:"id"`
		Name            string  `db:"name" json:"name"`
		WagesHour       uint64  `db:"wages_per_hour" json:"wages_per_hour"`
		AdminHour       uint64  `db:"admin_per_hour" json:"admin_per_hour"`
		MaintenanceHour uint64  `db:"maintenance_per_hour" json:"maintenance_per_hour"`
		Level           uint16  `db:"level" json:"level"`
		Position        *uint16 `db:"position" json:"position"`
	}
)

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/companies")
	repository := NewRepository(conn)

	group.GET("/:id", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		company, err := repository.GetById(companyId)
		if err != nil {
			return err
		}

		if company == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, company)
	})

	group.GET("/:id/buildings", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildings, err := repository.GetBuildings(companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, buildings)
	})

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
