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
	"api/research"
	"api/resource"
	"api/scheduler"
	"api/server"
	"api/warehouse"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	conn, err := database.GetConnection(database.SQLITE, "development.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	events := make(chan server.Message)
	connections := make(chan server.Connection)
	disconnections := make(chan string)

	svr := server.NewServer()
	timer := scheduler.NewScheduler()

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

	financingSvc := financing.NewService(financing.NewRepository(conn))
	financingGroup := financing.CreateEndpoints(svr, financingSvc, companySvc)

	loansRepo := loans.NewRepository(conn, accountingRepo)
	loansSvc := loans.NewService(loansRepo, companySvc, financingSvc)
	scheduledLoansSvc := loans.NewScheduledService(loansSvc, timer)
	loans.CreateEndpoints(financingGroup, scheduledLoansSvc)

	bondsRepo := bonds.NewRepository(conn, accountingRepo)
	bondsSvc := bonds.NewService(bondsRepo, companySvc)
	scheduledBondsSvc := bonds.NewScheduledService(bondsSvc, timer)
	bonds.CreateEndpoints(financingGroup, scheduledBondsSvc)

	svr.GET("/", server.Greeting(events))
	svr.GET("/private", server.Private(events))
	svr.GET("/ws", server.WS(connections, disconnections))

	go processEvents(events, connections, disconnections)

	svr.Start(":1323")
}

func processEvents(events <-chan server.Message, connections <-chan server.Connection, disconnections <-chan string) {
	sockets := make(map[string]*websocket.Conn)

	for {
		select {
		case message := <-events:
			log.Printf("received event: %v\n", message)

			if message.Token != "" {
				socket, ok := sockets[message.Token]
				if ok {
					if err := socket.WriteMessage(websocket.TextMessage, []byte(message.Message)); err != nil {
						log.Println("write:", err)
					}
				} else {
					log.Printf("socket not found: %s\n", message.Token)
				}
			} else {
				for _, socket := range sockets {
					if err := socket.WriteMessage(websocket.TextMessage, []byte(message.Message)); err != nil {
						log.Println("write:", err)
					}
				}
			}
		case connection := <-connections:
			log.Printf("registering socket for user: %s\n", connection.Token)
			sockets[connection.Token] = connection.Socket
		case token := <-disconnections:
			log.Printf("socket disconnected for user: %s\n", token)
			delete(sockets, token)
		}
	}
}
