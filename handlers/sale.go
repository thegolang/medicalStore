package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"medicalstore/database"
	"medicalstore/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaleList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := int64(15)
	skip := int64(page-1) * limit

	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	filter := bson.M{}

	if dateFrom != "" || dateTo != "" {
		dateFilter := bson.M{}
		if dateFrom != "" {
			t, _ := time.Parse("2006-01-02", dateFrom)
			dateFilter["$gte"] = t
		}
		if dateTo != "" {
			t, _ := time.Parse("2006-01-02", dateTo)
			dateFilter["$lte"] = t.Add(24 * time.Hour)
		}
		filter["created_at"] = dateFilter
	}

	total, _ := database.Col("sales").CountDocuments(ctx, filter)
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(limit).SetSkip(skip)
	cursor, _ := database.Col("sales").Find(ctx, filter, opts)
	var sales []models.Sale
	cursor.All(ctx, &sales)

	totalPages := (total + limit - 1) / limit
	sd := sessionData(c)

	c.HTML(http.StatusOK, "sales/index.html", gin.H{
		"title":      "Sales",
		"nav":        "sales",
		"sales":      sales,
		"total":      total,
		"page":       page,
		"totalPages": totalPages,
		"dateFrom":   dateFrom,
		"dateTo":     dateTo,
		"username":   sd["username"],
		"role":       sd["role"],
		"fullName":   sd["fullName"],
	})
}

func SaleNew(c *gin.Context) {
	sd := sessionData(c)
	c.HTML(http.StatusOK, "sales/create.html", gin.H{
		"title":    "New Sale",
		"nav":      "sales",
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func SaleCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session := sessions.Default(c)

	// Parse items from JSON
	itemsJSON := c.PostForm("items")
	var items []struct {
		MedicineID string  `json:"medicine_id"`
		Name       string  `json:"name"`
		Qty        int     `json:"qty"`
		Price      float64 `json:"price"`
	}
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil || len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid items"})
		return
	}

	discount, _ := strconv.ParseFloat(c.PostForm("discount"), 64)
	tax, _ := strconv.ParseFloat(c.PostForm("tax"), 64)
	amountPaid, _ := strconv.ParseFloat(c.PostForm("amount_paid"), 64)

	var saleItems []models.SaleItem
	var subTotal float64

	for _, item := range items {
		medID, _ := primitive.ObjectIDFromHex(item.MedicineID)
		total := float64(item.Qty) * item.Price
		subTotal += total
		saleItems = append(saleItems, models.SaleItem{
			MedicineID:   medID,
			MedicineName: item.Name,
			Quantity:     item.Qty,
			UnitPrice:    item.Price,
			Total:        total,
		})

		// Update stock
		database.Col("medicines").UpdateOne(ctx, bson.M{"_id": medID}, bson.M{
			"$inc": bson.M{"quantity": -item.Qty},
			"$set": bson.M{"updated_at": time.Now()},
		})
	}

	grandTotal := subTotal - discount + tax
	change := amountPaid - grandTotal
	if change < 0 {
		change = 0
	}

	invoiceNum := fmt.Sprintf("INV-%s", time.Now().Format("20060102150405"))

	sale := models.Sale{
		ID:            primitive.NewObjectID(),
		InvoiceNumber: invoiceNum,
		CustomerName:  c.PostForm("customer_name"),
		CustomerPhone: c.PostForm("customer_phone"),
		Items:         saleItems,
		SubTotal:      subTotal,
		Discount:      discount,
		Tax:           tax,
		GrandTotal:    grandTotal,
		PaymentMethod: c.PostForm("payment_method"),
		AmountPaid:    amountPaid,
		Change:        change,
		Notes:         c.PostForm("notes"),
		SoldBy:        session.Get("username").(string),
		CreatedAt:     time.Now(),
	}

	result, err := database.Col("sales").InsertOne(ctx, sale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"invoice_id": result.InsertedID.(primitive.ObjectID).Hex(),
		"invoice":    invoiceNum,
	})
}

func SaleView(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var sale models.Sale
	database.Col("sales").FindOne(ctx, bson.M{"_id": id}).Decode(&sale)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "sales/view.html", gin.H{
		"title":    "Invoice - " + sale.InvoiceNumber,
		"nav":      "sales",
		"sale":     sale,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}
