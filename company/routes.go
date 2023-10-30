package company

import (
	"api/auth"

	"api/server"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/companies")

	group.GET("/:id", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		company, err := service.GetById(companyId)
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

		buildings, err := service.GetBuildings(companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, buildings)
	})

	group.POST("/:id/buildings", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		data := new(Building)
		if err := c.Bind(data); err != nil {
			return err
		}

		if err := c.Validate(data); err != nil {
			return err
		}

		authenticated, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if companyId != authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		companyBuilding, err := service.AddBuilding(companyId, data.BuildingId, data.Position)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, companyBuilding)
	})

	group.POST("/register", func(c echo.Context) error {
		registration := new(Registration)
		if err := c.Bind(registration); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if err := c.Validate(registration); err != nil {
			return err
		}

		company, err := service.Register(registration)
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

		company, err := service.GetByEmail(credentials.Email)
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
