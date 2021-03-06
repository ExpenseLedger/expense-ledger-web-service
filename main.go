package main

import (
	"log"

	"github.com/expenseledger/web-service/config"
	"github.com/expenseledger/web-service/controller"
	"github.com/expenseledger/web-service/db"
	"github.com/shopspring/decimal"
)

func main() {
	var err error
	decimal.MarshalJSONWithoutQuotes = true

	err = db.CreateTables()
	if err != nil {
		log.Fatal("Error creating tables")
	} else {
		log.Println("Successfully created tables")
	}

	router := controller.InitRoutes()
	configs := config.GetConfigs()

	err = router.Run(":" + configs.Port)
	if err != nil {
		log.Fatal("Error running the server", err)
	}
}
