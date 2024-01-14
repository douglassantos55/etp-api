package main

import (
	"api/accounting"
	"api/building"
	"api/company"
	companyBuilding "api/company/building"
	"api/company/building/production"
	"api/database"
	"api/financing"
	"api/financing/bonds"
	"api/financing/loans"
	"api/market"
	"api/notification"
	"api/research"
	"api/research/staff"
	"api/resource"
	"api/scheduler"
	"api/server"
	"api/warehouse"
	"log"
	"os"
)

func main() {
	conn, err := database.GetConnection(database.SQLITE, "development.db?_loc=UTC")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	svr := server.NewServer()
	timer := scheduler.NewScheduler()

	logFile, err := os.OpenFile("dev.log", os.O_CREATE|os.O_APPEND, 664)
	logger := log.New(logFile, "[DEV]", log.Flags())

	notificationRepo := notification.NewRepository(conn)
	notifier := notification.NewNotifier(notificationRepo)

	resourceRepo := resource.NewRepository(conn)
	resourceSvc := resource.NewService(resourceRepo)
	resource.CreateEndpoints(svr, resourceSvc)

	warehouseRepo := warehouse.NewRepository(conn)
	warehouseSvc := warehouse.NewService(warehouseRepo)
	warehouse.CreateEndpoints(svr, warehouseSvc)

	buildingSvc := building.NewService(building.NewRepository(conn, resourceRepo))
	building.CreateEndpoints(svr, buildingSvc)

	accountingRepo := accounting.NewRepository(conn)
	companyRepo := company.NewRepository(conn, accountingRepo)
	companySvc := company.NewService(companyRepo)

	companyBuildingRepo := companyBuilding.NewBuildingRepository(conn, resourceRepo, warehouseRepo)
	companyBuildingSvc := companyBuilding.NewBuildingService(companyBuildingRepo, warehouseSvc, buildingSvc)
	scheduledBuildingSvc := companyBuilding.NewScheduledBuildingService(companyBuildingSvc, timer)

	researchSvc := research.NewService(research.NewRepository(conn, accountingRepo), companySvc)
	productionRepo := production.NewProductionRepository(conn, accountingRepo, companyBuildingRepo, warehouseRepo)
	productionSvc := production.NewProductionService(productionRepo, companySvc, companyBuildingSvc, warehouseSvc, researchSvc)
	scheduledProductionSvc := production.NewScheduledProductionService(productionSvc, timer)

	company.CreateEndpoints(svr, companySvc)
	production.CreateEndpoints(svr, scheduledProductionSvc, scheduledBuildingSvc, companySvc)

	staffRepo := staff.NewRepository(conn, accountingRepo)
	staffSvc := staff.NewService(staffRepo, timer, notifier, logger)
	staff.CreateEndpoints(svr, staffSvc)

	marketRepo := market.NewRepository(conn, companyRepo, warehouseRepo, accountingRepo)
	marketSvc := market.NewService(marketRepo, companySvc, warehouseSvc, notifier, logger)
	market.CreateEndpoints(svr, marketSvc)

	financingSvc := financing.NewService(financing.NewRepository(conn), notifier, logger)
	financingGroup := financing.CreateEndpoints(svr, financingSvc, companySvc)

	loansRepo := loans.NewRepository(conn, accountingRepo)
	loansSvc := loans.NewService(loansRepo, companySvc, financingSvc, notifier, logger)
	scheduledLoansSvc := loans.NewScheduledService(loansSvc, timer)
	loans.CreateEndpoints(financingGroup, scheduledLoansSvc)

	bondsRepo := bonds.NewRepository(conn, accountingRepo)
	bondsSvc := bonds.NewService(bondsRepo, companySvc, notifier, logger)
	scheduledBondsSvc := bonds.NewScheduledService(bondsSvc, timer)
	bonds.CreateEndpoints(financingGroup, scheduledBondsSvc)

	notificationSvc := notification.NewService(notificationRepo)
	notification.CreateEndpoints(svr, notificationSvc, notifier)

	svr.Start(":1323")
}
