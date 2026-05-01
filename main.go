package main

import (
	"context"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	cardreader "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/reader"
	cardrepository "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/repository"
	cardwriter "writeandinviteco/inviteandco/card/cardinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/card/cardpresentation"
	"writeandinviteco/inviteandco/config"
	customerreader "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/reader"
	customerrepository "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/repository"
	customerwriter "writeandinviteco/inviteandco/customer/customerinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/customer/customerpresentation"
	"writeandinviteco/inviteandco/notification"
	"writeandinviteco/inviteandco/order/orderapplication"
	orderreader "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/reader"
	orderrepository "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/repository"
	orderwriter "writeandinviteco/inviteandco/order/orderinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/order/orderpresentation"
	"writeandinviteco/inviteandco/product/productapplication"
	productreader "writeandinviteco/inviteandco/product/productinfrastructure/postgres/reader"
	productrepository "writeandinviteco/inviteandco/product/productinfrastructure/postgres/repository"
	productwriter "writeandinviteco/inviteandco/product/productinfrastructure/postgres/writer"
	"writeandinviteco/inviteandco/product/productpresentation"
	"writeandinviteco/inviteandco/webui"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	appRoot, err := resolveAppRoot()
	if err != nil {
		log.Fatalf("failed to resolve app paths: %v", err)
	}

	loadEnvFile(appRoot)

	cfg := config.Load()
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		log.Fatal("DATABASE_URL or DB_* settings must be configured")
	}
	adminAuthEnabled := gin.Mode() == gin.ReleaseMode
	if adminAuthEnabled && (cfg.AdminUser == "" || cfg.AdminPass == "") {
		log.Fatal("ADMIN_USER and ADMIN_PASS must be configured")
	}

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
	orderReader := orderreader.New(db)
	orderWriter := orderwriter.New(db)
	productReader := productreader.New(db)
	productWriter := productwriter.New(db)

	// repositories
	cardRepo := cardrepository.NewCardRepository(cardReader, cardWriter)
	customerRepo := customerrepository.NewCustomerRepository(customerReader, customerWriter)
	orderRepo := orderrepository.NewOrderRepository(orderReader, orderWriter)
	productRepo := productrepository.NewProductRepository(productReader, productWriter)

	var emailSender orderapplication.EmailSender
	if cfg.ResendAPIKey != "" {
		emailSender = notification.NewResendSender(cfg.ResendAPIKey, cfg.ResendFromEmail)
	}

	// handlers
	productService := productapplication.NewService(productRepo)
	productHandler := productpresentation.NewHandler(productService, cardRepo, filepath.Join(appRoot, "static", "uploads"))
	orderService := orderapplication.NewService(
		db,
		cardRepo,
		customerRepo,
		orderRepo,
		customerWriter,
		orderWriter,
		emailSender,
		cfg.AdminEmail,
	)
	orderHandler := orderpresentation.NewOrderHandler(orderService, filepath.Join(appRoot, "static", "payment-proofs"))
	cardHandler := cardpresentation.NewCardHandler(cardRepo, productService)
	customerHandler := customerpresentation.NewCustomerHandler(customerRepo)

	// router
	router := gin.Default()
	router.MaxMultipartMemory = 20 << 20
	safeStaticPath := makeSafeStaticPathFunc(appRoot)
	router.SetFuncMap(template.FuncMap{
		"add":            func(a, b int) int { return a + b },
		"safeImagePath":  makeSafeImagePathFunc(appRoot),
		"safeStaticPath": safeStaticPath,
		"isPDFPath": func(raw string) bool {
			return strings.HasSuffix(strings.ToLower(strings.TrimSpace(raw)), ".pdf")
		},
	})
	router.LoadHTMLGlob(filepath.Join(appRoot, "templates", "*"))
	router.Static("/static", filepath.Join(appRoot, "static"))
	router.NoRoute(func(c *gin.Context) {
		webui.RenderError(c, 404, "Page Not Found", "The page you requested could not be found.")
	})
	router.NoMethod(func(c *gin.Context) {
		webui.RenderError(c, 404, "Page Not Found", "The page you requested could not be found.")
	})

	// routes
	router.GET("/", cardHandler.ListCards)
	router.GET("/about", customerHandler.AboutPage)
	router.GET("/contact", customerHandler.ContactPage)
	router.GET("/shipping-info", customerHandler.ShippingInfoPage)
	router.GET("/returns-exchanges", customerHandler.ReturnsExchangesPage)
	router.GET("/my-account", customerHandler.MyAccountPage)
	router.GET("/terms-of-use", customerHandler.TermsOfUsePage)
	router.GET("/privacy-policy", customerHandler.PrivacyPolicyPage)
	router.GET("/search", cardHandler.SearchCards)
	router.GET("/card/:id", cardHandler.CardDetail)
	router.GET("/product/:id", productHandler.Detail)
	router.GET("/checkout", cardHandler.Checkout)
	router.POST("/customize", orderHandler.CustomizePage)
	router.POST("/review", orderHandler.ReviewPage)
	router.GET("/order-confirmation/:id", orderHandler.OrderConfirmation)
	router.GET("/order/:id", orderHandler.OrderStatus)
	router.GET("/order/:id/payment", orderHandler.BankTransferPage)
	router.POST("/order/:id/payment-proof", orderHandler.SubmitPaymentProof)
	router.GET("/track-order", orderHandler.TrackOrderPage)
	router.GET("/collections/:category", cardHandler.ListCardsByCategory)
	router.POST("/order", orderHandler.CreateOrder)

	admin := router.Group("/admin")
	if adminAuthEnabled {
		admin.Use(gin.BasicAuth(gin.Accounts{
			cfg.AdminUser: cfg.AdminPass,
		}))
	} else {
		log.Println("Admin basic auth disabled outside release mode for local testing")
	}
	admin.GET("/orders", orderHandler.AdminOrders)
	admin.GET("/orders/:id", orderHandler.AdminOrderDetail)
	admin.POST("/orders/:id/status", orderHandler.AdminUpdateOrderStatus)
	admin.POST("/orders/:id/payment", orderHandler.AdminPaymentAction)
	admin.GET("/products", productHandler.List)
	admin.GET("/products/new", productHandler.NewForm)
	admin.POST("/products", productHandler.Create)
	admin.GET("/products/:id/edit", productHandler.EditForm)
	admin.POST("/products/:id/edit", productHandler.Update)
	admin.POST("/products/:id", productHandler.Update)
	admin.POST("/products/:id/delete", productHandler.Delete)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadEnvFile(appRoot string) {
	candidates := []string{
		".env",
		filepath.Join(appRoot, ".env"),
	}

	for _, candidate := range candidates {
		if err := godotenv.Load(candidate); err == nil {
			return
		}
	}
}

func resolveAppRoot() (string, error) {
	return os.Getwd()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func makeSafeImagePathFunc(appRoot string) func(string) string {
	safeStaticPath := makeSafeStaticPathFunc(appRoot)
	return func(raw string) string {
		fallback := "/static/sample.jpg"
		clean := strings.TrimSpace(raw)
		if clean == "" {
			return fallback
		}
		if strings.HasPrefix(clean, "https://") || strings.HasPrefix(clean, "http://") {
			return clean
		}
		safePath := safeStaticPath(raw)
		if safePath == "" {
			return fallback
		}
		return safePath
	}
}

func makeSafeStaticPathFunc(appRoot string) func(string) string {
	return func(raw string) string {
		clean := strings.TrimSpace(raw)
		if clean == "" {
			return ""
		}
		if strings.HasPrefix(clean, "https://") || strings.HasPrefix(clean, "http://") {
			return ""
		}
		if !strings.HasPrefix(clean, "/static/") {
			return ""
		}

		relative := filepath.Clean(strings.TrimPrefix(clean, "/"))
		staticPrefix := "static" + string(os.PathSeparator)
		if relative != "static" && !strings.HasPrefix(relative, staticPrefix) {
			return ""
		}

		if _, err := os.Stat(filepath.Join(appRoot, relative)); err != nil {
			return ""
		}

		return "/" + filepath.ToSlash(relative)
	}
}
