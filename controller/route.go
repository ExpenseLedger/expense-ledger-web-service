package controller

import (
	"github.com/expenseledger/web-service/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Route data structure to hold a path and the corresponding handler
type Route struct {
	Path    string
	Handler func(context *gin.Context)
}

// InitRoutes ...
func InitRoutes() *gin.Engine {
	configs := config.GetConfigs()
	router := gin.Default()
	router.Use(cors.New(getCorsConfig()))

	router.GET("/", getRoot)
	walletRoute := router.Group("/wallet")
	categoryRoute := router.Group("/category")
	transactionRoute := router.Group("/transaction")

	walletRoute.Use(validateHeader)
	categoryRoute.Use(validateHeader)
	transactionRoute.Use(validateHeader)

	walletRoute.POST("/create", createWallet)
	walletRoute.POST("/get", getWallet)
	walletRoute.POST("/delete", deleteWallet)
	walletRoute.POST("/list", listWallets)
	walletRoute.POST("/listTypes", listWalletTypes)
	walletRoute.POST("/init", initWallets)

	categoryRoute.POST("/create", createCategory)
	categoryRoute.POST("/get", getCategory)
	categoryRoute.POST("/delete", deleteCategory)
	categoryRoute.POST("/list", listCategories)
	categoryRoute.POST("/init", initCategories)

	transactionRoute.POST("/createExpense", createExpense)
	transactionRoute.POST("/createIncome", createIncome)
	transactionRoute.POST("/createTransfer", createTransfer)
	transactionRoute.POST("/get", getTransaction)
	transactionRoute.POST("/delete", deleteTransaction)
	transactionRoute.POST("/list", listTransactions)
	transactionRoute.POST("/listTypes", listTransactionTypes)

	if configs.Mode != "PRODUCTION" {
		walletRoute.POST("/clear", clearWallets)
		categoryRoute.POST("/clear", clearCategories)
		transactionRoute.POST("/clear", clearTransactions)
	}

	return router
}
