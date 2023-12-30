package financing

import (
	"api/financing/bonds"
	"api/financing/loans"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, loans loans.Service, bonds bonds.Service) {
}
