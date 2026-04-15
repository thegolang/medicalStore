package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"medicalstore/config"
	"medicalstore/database"
	"medicalstore/handlers"
	"medicalstore/middleware"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b float64) float64 { return a * b },
	"div": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"slice": func(s string, start, end int) string {
		if start >= len(s) {
			return ""
		}
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
	"not": func(b bool) bool { return !b },
}

// createRenderer builds a multitemplate renderer so each page gets its own
// template set (layout + page), allowing the "content" block to be unique
// per page without conflicts.
func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	layout := "templates/layout.html"

	// Standalone pages (no layout)
	r.AddFromFilesFuncs("login.html", funcMap, "templates/login.html")

	// Pages that use the layout
	withLayout := map[string][]string{
		"dashboard.html":      {"templates/dashboard.html"},
		"profile.html":        {"templates/profile.html"},

		"medicines/index.html": {"templates/medicines/index.html"},
		"medicines/form.html":  {"templates/medicines/form.html"},

		// categories/index.html embeds the table partial inline, so include it
		"categories/index.html": {"templates/categories/index.html", "templates/categories/table.html"},

		"suppliers/index.html": {"templates/suppliers/index.html"},
		"suppliers/form.html":  {"templates/suppliers/form.html"},

		"sales/index.html":  {"templates/sales/index.html"},
		"sales/create.html": {"templates/sales/create.html"},
		"sales/view.html":   {"templates/sales/view.html"},

		"purchases/index.html":  {"templates/purchases/index.html"},
		"purchases/create.html": {"templates/purchases/create.html"},
		"purchases/view.html":   {"templates/purchases/view.html"},

		"reports/index.html": {"templates/reports/index.html"},

		"users/index.html": {"templates/users/index.html"},
		"users/form.html":  {"templates/users/form.html"},
	}

	for name, files := range withLayout {
		all := append([]string{layout}, files...)
		r.AddFromFilesFuncs(name, funcMap, all...)
	}

	// Partials rendered directly (HTMX swaps, no layout)
	r.AddFromFilesFuncs("categories/table.html", funcMap, "templates/categories/table.html")

	return r
}

func main() {
	cfg := config.Load()
	database.Connect(cfg.MongoURI, cfg.MongoDB)
	defer database.Disconnect()

	r := gin.Default()
	r.HTMLRender = createRenderer()

	// Session store
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("medstore_session", store))

	// Static files
	r.Static("/static", "./static")

	// ── Public routes ──────────────────────────────────
	r.GET("/login", handlers.LoginPage)
	r.POST("/login", handlers.Login)

	// ── Protected routes ───────────────────────────────
	auth := r.Group("/")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("/", handlers.Dashboard)
		auth.GET("/logout", handlers.Logout)

		// Profile
		auth.GET("/profile", handlers.ProfilePage)
		auth.POST("/profile", handlers.UpdateProfile)

		// Medicines
		auth.GET("/medicines", handlers.MedicineList)
		auth.GET("/medicines/new", handlers.MedicineNew)
		auth.POST("/medicines", handlers.MedicineCreate)
		auth.GET("/medicines/search", handlers.MedicineSearch)
		auth.GET("/medicines/:id/edit", handlers.MedicineEdit)
		auth.POST("/medicines/:id", handlers.MedicineUpdate)
		auth.DELETE("/medicines/:id", handlers.MedicineDelete)

		// Categories
		auth.GET("/categories", handlers.CategoryList)
		auth.POST("/categories", handlers.CategoryCreate)
		auth.POST("/categories/:id", handlers.CategoryUpdate)
		auth.DELETE("/categories/:id", handlers.CategoryDelete)

		// Suppliers
		auth.GET("/suppliers", handlers.SupplierList)
		auth.GET("/suppliers/new", handlers.SupplierNew)
		auth.POST("/suppliers", handlers.SupplierCreate)
		auth.GET("/suppliers/:id/edit", handlers.SupplierEdit)
		auth.POST("/suppliers/:id", handlers.SupplierUpdate)
		auth.DELETE("/suppliers/:id", handlers.SupplierDelete)

		// Sales
		auth.GET("/sales", handlers.SaleList)
		auth.GET("/sales/new", handlers.SaleNew)
		auth.POST("/sales", handlers.SaleCreate)
		auth.GET("/sales/:id", handlers.SaleView)

		// Purchases
		auth.GET("/purchases", handlers.PurchaseList)
		auth.GET("/purchases/new", handlers.PurchaseNew)
		auth.POST("/purchases", handlers.PurchaseCreate)
		auth.GET("/purchases/:id", handlers.PurchaseView)
		auth.POST("/purchases/:id/receive", handlers.PurchaseReceive)
		auth.POST("/purchases/:id/cancel", handlers.PurchaseCancel)

		// Reports
		auth.GET("/reports", handlers.ReportsPage)

		// Users (admin only)
		admin := auth.Group("/users")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("", handlers.UserList)
			admin.GET("/new", handlers.UserNew)
			admin.POST("", handlers.UserCreate)
			admin.GET("/:id/edit", handlers.UserEdit)
			admin.POST("/:id", handlers.UserUpdate)
			admin.POST("/:id/toggle", handlers.UserToggle)
		}
	}

	// 404
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/")
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("MedStore running at http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
}
