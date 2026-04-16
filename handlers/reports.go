package handlers

import (
	"context"
	"net/http"
	"time"

	"medicalstore/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SalesReport struct {
	Date    string  `bson:"_id"`
	Orders  int     `bson:"orders"`
	Revenue float64 `bson:"revenue"`
}

type TopMedicineReport struct {
	Name     string  `bson:"name"`
	Quantity int     `bson:"total_qty"`
	Revenue  float64 `bson:"total_revenue"`
}

func ReportsPage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	period := c.DefaultQuery("period", "week")
	var fromDate time.Time
	switch period {
	case "today":
		fromDate = time.Now().Truncate(24 * time.Hour)
	case "week":
		fromDate = time.Now().AddDate(0, 0, -7)
	case "month":
		fromDate = time.Now().AddDate(0, -1, 0)
	case "year":
		fromDate = time.Now().AddDate(-1, 0, 0)
	default:
		fromDate = time.Now().AddDate(0, 0, -7)
	}

	// Daily sales trend
	salesTrendPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": fromDate}}}},
		{{Key: "$group", Value: bson.M{
			"_id":     bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
			"orders":  bson.M{"$sum": 1},
			"revenue": bson.M{"$sum": "$grand_total"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}
	trendCursor, _ := database.Col("sales").Aggregate(ctx, salesTrendPipeline)
	var salesTrend []SalesReport
	trendCursor.All(ctx, &salesTrend)

	// Top selling medicines
	topMedPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": fromDate}}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":           "$items.medicine_name",
			"name":          bson.M{"$first": "$items.medicine_name"},
			"total_qty":     bson.M{"$sum": "$items.quantity"},
			"total_revenue": bson.M{"$sum": "$items.total"},
		}}},
		{{Key: "$sort", Value: bson.M{"total_qty": -1}}},
		{{Key: "$limit", Value: 10}},
	}
	topMedCursor, _ := database.Col("sales").Aggregate(ctx, topMedPipeline)
	var topMedicines []TopMedicineReport
	topMedCursor.All(ctx, &topMedicines)

	// Payment method breakdown
	paymentPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": fromDate}}}},
		{{Key: "$group", Value: bson.M{
			"_id":     "$payment_method",
			"count":   bson.M{"$sum": 1},
			"revenue": bson.M{"$sum": "$grand_total"},
		}}},
	}
	paymentCursor, _ := database.Col("sales").Aggregate(ctx, paymentPipeline)
	var paymentBreakdown []bson.M
	paymentCursor.All(ctx, &paymentBreakdown)

	// Summary stats
	summaryPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": fromDate}}}},
		{{Key: "$group", Value: bson.M{
			"_id":      nil,
			"total":    bson.M{"$sum": "$grand_total"},
			"orders":   bson.M{"$sum": 1},
			"discount": bson.M{"$sum": "$discount"},
			"tax":      bson.M{"$sum": "$tax"},
		}}},
	}
	summaryCursor, _ := database.Col("sales").Aggregate(ctx, summaryPipeline)
	var summaryResult []bson.M
	summaryCursor.All(ctx, &summaryResult)

	var totalRevenue, totalDiscount, totalTax float64
	var totalOrders int
	if len(summaryResult) > 0 {
		if v, ok := summaryResult[0]["total"].(float64); ok {
			totalRevenue = v
		}
		switch v := summaryResult[0]["orders"].(type) {
		case int32:
			totalOrders = int(v)
		case int64:
			totalOrders = int(v)
		}
		if v, ok := summaryResult[0]["discount"].(float64); ok {
			totalDiscount = v
		}
		if v, ok := summaryResult[0]["tax"].(float64); ok {
			totalTax = v
		}
	}

	sd := sessionData(c)
	c.HTML(http.StatusOK, "reports/index.html", gin.H{
		"title":            "Reports",
		"nav":              "reports",
		"period":           period,
		"salesTrend":       salesTrend,
		"topMedicines":     topMedicines,
		"paymentBreakdown": paymentBreakdown,
		"totalRevenue":     totalRevenue,
		"totalOrders":      totalOrders,
		"totalDiscount":    totalDiscount,
		"totalTax":         totalTax,
		"username":         sd["username"],
		"role":             sd["role"],
		"fullName":         sd["fullName"],
	})
}
