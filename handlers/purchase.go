package handlers

import (
	"context"
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

func PurchaseList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status := c.Query("status")
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := int64(15)
	skip := int64(page-1) * limit

	total, _ := database.Col("purchases").CountDocuments(ctx, filter)
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(limit).SetSkip(skip)
	cursor, _ := database.Col("purchases").Find(ctx, filter, opts)
	var purchases []models.Purchase
	cursor.All(ctx, &purchases)

	totalPages := (total + limit - 1) / limit
	sd := sessionData(c)

	c.HTML(http.StatusOK, "purchases/index.html", gin.H{
		"title":      "Purchase Orders",
		"nav":        "purchases",
		"purchases":  purchases,
		"total":      total,
		"page":       page,
		"totalPages": totalPages,
		"status":     status,
		"username":   sd["username"],
		"role":       sd["role"],
		"fullName":   sd["fullName"],
	})
}

func PurchaseNew(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	supCursor, _ := database.Col("suppliers").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var suppliers []models.Supplier
	supCursor.All(ctx, &suppliers)

	medCursor, _ := database.Col("medicines").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var medicines []models.Medicine
	medCursor.All(ctx, &medicines)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "purchases/create.html", gin.H{
		"title":     "New Purchase Order",
		"nav":       "purchases",
		"suppliers": suppliers,
		"medicines": medicines,
		"username":  sd["username"],
		"role":      sd["role"],
		"fullName":  sd["fullName"],
	})
}

func PurchaseCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session := sessions.Default(c)
	supID, _ := primitive.ObjectIDFromHex(c.PostForm("supplier_id"))

	var sup models.Supplier
	database.Col("suppliers").FindOne(ctx, bson.M{"_id": supID}).Decode(&sup)

	medicineIDs := c.PostFormArray("medicine_id[]")
	quantities := c.PostFormArray("quantity[]")
	prices := c.PostFormArray("unit_price[]")
	batches := c.PostFormArray("batch_number[]")
	expiries := c.PostFormArray("expiry_date[]")

	var purchaseItems []models.PurchaseItem
	var total float64

	for i, medIDStr := range medicineIDs {
		if i >= len(quantities) || i >= len(prices) {
			break
		}
		medID, _ := primitive.ObjectIDFromHex(medIDStr)
		var med models.Medicine
		database.Col("medicines").FindOne(ctx, bson.M{"_id": medID}).Decode(&med)

		qty, _ := strconv.Atoi(quantities[i])
		price, _ := strconv.ParseFloat(prices[i], 64)
		itemTotal := float64(qty) * price
		total += itemTotal

		expiry := time.Now().AddDate(1, 0, 0)
		if i < len(expiries) {
			expiry, _ = time.Parse("2006-01-02", expiries[i])
		}
		batch := ""
		if i < len(batches) {
			batch = batches[i]
		}

		purchaseItems = append(purchaseItems, models.PurchaseItem{
			MedicineID:   medID,
			MedicineName: med.Name,
			Quantity:     qty,
			UnitPrice:    price,
			Total:        itemTotal,
			BatchNumber:  batch,
			ExpiryDate:   expiry,
		})
	}

	poNum := fmt.Sprintf("PO-%s", time.Now().Format("20060102150405"))

	purchase := models.Purchase{
		ID:           primitive.NewObjectID(),
		PONumber:     poNum,
		SupplierID:   supID,
		SupplierName: sup.Name,
		Items:        purchaseItems,
		TotalAmount:  total,
		Status:       "pending",
		Notes:        c.PostForm("notes"),
		OrderedBy:    session.Get("username").(string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	database.Col("purchases").InsertOne(ctx, purchase)
	c.Redirect(http.StatusFound, "/purchases")
}

func PurchaseReceive(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var purchase models.Purchase
	database.Col("purchases").FindOne(ctx, bson.M{"_id": id}).Decode(&purchase)

	if purchase.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase already processed"})
		return
	}

	now := time.Now()
	// Update stock for each item
	for _, item := range purchase.Items {
		database.Col("medicines").UpdateOne(ctx, bson.M{"_id": item.MedicineID}, bson.M{
			"$inc": bson.M{"quantity": item.Quantity},
			"$set": bson.M{
				"batch_number": item.BatchNumber,
				"expiry_date":  item.ExpiryDate,
				"updated_at":   now,
			},
		})
	}

	database.Col("purchases").UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"status":      "received",
			"received_at": now,
			"updated_at":  now,
		},
	})

	c.Header("HX-Redirect", "/purchases")
	c.Status(http.StatusOK)
}

func PurchaseCancel(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	database.Col("purchases").UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"status": "cancelled", "updated_at": time.Now()},
	})

	c.Header("HX-Redirect", "/purchases")
	c.Status(http.StatusOK)
}

func PurchaseView(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var purchase models.Purchase
	database.Col("purchases").FindOne(ctx, bson.M{"_id": id}).Decode(&purchase)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "purchases/view.html", gin.H{
		"title":    "Purchase Order - " + purchase.PONumber,
		"nav":      "purchases",
		"purchase": purchase,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}
