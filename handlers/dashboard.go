package handlers

import (
	"context"
	"net/http"
	"time"

	"medicalstore/database"
	"medicalstore/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DashboardStats struct {
	TotalMedicines   int64
	LowStockCount    int64
	ExpiredCount     int64
	ExpiringSoon     int64
	TodaySales       float64
	TodayOrders      int64
	MonthSales       float64
	TotalSuppliers   int64
	TotalCategories  int64
	RecentSales      []models.Sale
	LowStockItems    []models.Medicine
	TopMedicines     []TopMedicine
}

type TopMedicine struct {
	Name     string
	Quantity int
	Revenue  float64
}

func Dashboard(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := DashboardStats{}

	// Total medicines
	stats.TotalMedicines, _ = database.Col("medicines").CountDocuments(ctx, bson.M{"active": true})

	// Low stock
	stats.LowStockCount, _ = database.Col("medicines").CountDocuments(ctx, bson.M{
		"active": true,
		"$expr":  bson.M{"$lte": bson.A{"$quantity", "$min_stock_level"}},
	})

	// Expired
	stats.ExpiredCount, _ = database.Col("medicines").CountDocuments(ctx, bson.M{
		"active":      true,
		"expiry_date": bson.M{"$lt": time.Now()},
	})

	// Expiring soon (within 30 days)
	stats.ExpiringSoon, _ = database.Col("medicines").CountDocuments(ctx, bson.M{
		"active": true,
		"expiry_date": bson.M{
			"$gte": time.Now(),
			"$lte": time.Now().AddDate(0, 1, 0),
		},
	})

	// Today's sales
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)
	stats.TodayOrders, _ = database.Col("sales").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": todayStart, "$lt": todayEnd},
	})

	// Today's revenue
	todayRevPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": todayStart, "$lt": todayEnd}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$grand_total"}}}},
	}
	todayRevCursor, _ := database.Col("sales").Aggregate(ctx, todayRevPipeline)
	var todayRevResult []bson.M
	todayRevCursor.All(ctx, &todayRevResult)
	if len(todayRevResult) > 0 {
		if v, ok := todayRevResult[0]["total"].(float64); ok {
			stats.TodaySales = v
		}
	}

	// Month revenue
	monthStart := time.Now().AddDate(0, 0, -30)
	monthRevPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": monthStart}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$grand_total"}}}},
	}
	monthRevCursor, _ := database.Col("sales").Aggregate(ctx, monthRevPipeline)
	var monthRevResult []bson.M
	monthRevCursor.All(ctx, &monthRevResult)
	if len(monthRevResult) > 0 {
		if v, ok := monthRevResult[0]["total"].(float64); ok {
			stats.MonthSales = v
		}
	}

	// Supplier and category counts
	stats.TotalSuppliers, _ = database.Col("suppliers").CountDocuments(ctx, bson.M{"active": true})
	stats.TotalCategories, _ = database.Col("categories").CountDocuments(ctx, bson.M{"active": true})

	// Recent sales (last 5)
	recentOpts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(5)
	recentCursor, _ := database.Col("sales").Find(ctx, bson.M{}, recentOpts)
	recentCursor.All(ctx, &stats.RecentSales)

	// Low stock items (top 5)
	lowStockOpts := options.Find().SetSort(bson.M{"quantity": 1}).SetLimit(5)
	lowStockCursor, _ := database.Col("medicines").Find(ctx, bson.M{
		"active": true,
		"$expr":  bson.M{"$lte": bson.A{"$quantity", "$min_stock_level"}},
	}, lowStockOpts)
	lowStockCursor.All(ctx, &stats.LowStockItems)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "Dashboard",
		"nav":      "dashboard",
		"stats":    stats,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}
