package main

import (
	"context"
	"html/template"
	"log"

	cardreader "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/reader"
	cardrepository "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/repository"
	cardwriter "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/card/cardpresentation"
	"writeandinviteco/inviteandco/config"
	customerreader "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/reader"
	customerrepository "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/repository"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/customer/customerpresentation"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/order/orderpresentation"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}

	// sqlc queries
	cardReader := cardreader.New(db)
	cardWriter := cardwriter.New(db)
	customerReader := customerreader.New(db)
	customerWriter := customerwriter.New(db)
	orderWriter := orderwriter.New(db)

	// repositories
	cardRepo := cardrepository.NewCardRepository(cardReader, cardWriter)
	customerRepo := customerrepository.NewCustomerRepository(customerReader, customerWriter)

	// handlers
	cardHandler := cardpresentation.NewCardHandler(cardRepo)
	customerHandler := customerpresentation.NewCustomerHandler(customerRepo)
	orderHandler := orderpresentation.NewOrderHandler(
		db,
		cardRepo,
		customerWriter,
		orderWriter,
	)

	// router
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// routes
	router.GET("/", cardHandler.ListCards)
	router.GET("/about", customerHandler.AboutPage)
	router.GET("/contact", customerHandler.ContactPage)
	router.GET("/search", cardHandler.SearchCards)
	router.GET("/card/:id", cardHandler.CardDetail)
	router.GET("/checkout", cardHandler.Checkout)
	router.GET("/collections/:category", cardHandler.ListCardsByCategory)
	router.POST("/order", orderHandler.CreateOrder)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
