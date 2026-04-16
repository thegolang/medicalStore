package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"medicalstore/database"
	"medicalstore/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MedicineList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	search := c.Query("search")
	category := c.Query("category")
	stockFilter := c.Query("stock")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := int64(15)
	skip := int64((page - 1)) * limit

	filter := bson.M{"active": true}
	if search != "" {
		filter["$or"] = bson.A{
			bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"generic_name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"manufacturer": bson.M{"$regex": search, "$options": "i"}},
		}
	}
	if category != "" {
		catID, err := primitive.ObjectIDFromHex(category)
		if err == nil {
			filter["category_id"] = catID
		}
	}
	if stockFilter == "low" {
		filter["$expr"] = bson.M{"$lte": bson.A{"$quantity", "$min_stock_level"}}
	} else if stockFilter == "expired" {
		filter["expiry_date"] = bson.M{"$lt": time.Now()}
	} else if stockFilter == "expiring" {
		filter["expiry_date"] = bson.M{
			"$gte": time.Now(),
			"$lte": time.Now().AddDate(0, 1, 0),
		}
	}

	total, _ := database.Col("medicines").CountDocuments(ctx, filter)
	opts := options.Find().SetSort(bson.M{"name": 1}).SetLimit(limit).SetSkip(skip)
	cursor, _ := database.Col("medicines").Find(ctx, filter, opts)

	var medicines []models.Medicine
	cursor.All(ctx, &medicines)

	// Load categories for filter dropdown
	catCursor, _ := database.Col("categories").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	catCursor.All(ctx, &categories)

	totalPages := (total + limit - 1) / limit
	sd := sessionData(c)

	c.HTML(http.StatusOK, "medicines/index.html", gin.H{
		"title":       "Medicines",
		"nav":         "medicines",
		"medicines":   medicines,
		"categories":  categories,
		"total":       total,
		"page":        page,
		"totalPages":  totalPages,
		"search":      search,
		"category":    category,
		"stockFilter": stockFilter,
		"username":    sd["username"],
		"role":        sd["role"],
		"fullName":    sd["fullName"],
	})
}

func MedicineNew(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	catCursor, _ := database.Col("categories").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	catCursor.All(ctx, &categories)

	supCursor, _ := database.Col("suppliers").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var suppliers []models.Supplier
	supCursor.All(ctx, &suppliers)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "medicines/form.html", gin.H{
		"title":      "Add Medicine",
		"nav":        "medicines",
		"categories": categories,
		"suppliers":  suppliers,
		"medicine":   models.Medicine{},
		"isNew":      true,
		"username":   sd["username"],
		"role":       sd["role"],
		"fullName":   sd["fullName"],
	})
}

func MedicineCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	catID, _ := primitive.ObjectIDFromHex(c.PostForm("category_id"))
	supID, _ := primitive.ObjectIDFromHex(c.PostForm("supplier_id"))
	qty, _ := strconv.Atoi(c.PostForm("quantity"))
	minStock, _ := strconv.Atoi(c.PostForm("min_stock_level"))
	purchasePrice, _ := strconv.ParseFloat(c.PostForm("purchase_price"), 64)
	sellingPrice, _ := strconv.ParseFloat(c.PostForm("selling_price"), 64)
	expiryDate, _ := time.Parse("2006-01-02", c.PostForm("expiry_date"))
	prescription := c.PostForm("prescription") == "true"

	// Fetch names
	var cat models.Category
	var sup models.Supplier
	database.Col("categories").FindOne(ctx, bson.M{"_id": catID}).Decode(&cat)
	database.Col("suppliers").FindOne(ctx, bson.M{"_id": supID}).Decode(&sup)

	medicine := models.Medicine{
		ID:            primitive.NewObjectID(),
		Name:          c.PostForm("name"),
		GenericName:   c.PostForm("generic_name"),
		CategoryID:    catID,
		CategoryName:  cat.Name,
		SupplierID:    supID,
		SupplierName:  sup.Name,
		Manufacturer:  c.PostForm("manufacturer"),
		BatchNumber:   c.PostForm("batch_number"),
		ExpiryDate:    expiryDate,
		Quantity:      qty,
		MinStockLevel: minStock,
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		Unit:          c.PostForm("unit"),
		Description:   c.PostForm("description"),
		Prescription:  prescription,
		Active:        true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := database.Col("medicines").InsertOne(ctx, medicine)
	if err != nil {
		c.HTML(http.StatusOK, "medicines/form.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/medicines")
}

func MedicineEdit(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var medicine models.Medicine
	database.Col("medicines").FindOne(ctx, bson.M{"_id": id}).Decode(&medicine)

	catCursor, _ := database.Col("categories").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	catCursor.All(ctx, &categories)

	supCursor, _ := database.Col("suppliers").Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.M{"name": 1}))
	var suppliers []models.Supplier
	supCursor.All(ctx, &suppliers)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "medicines/form.html", gin.H{
		"title":      "Edit Medicine",
		"nav":        "medicines",
		"medicine":   medicine,
		"categories": categories,
		"suppliers":  suppliers,
		"isNew":      false,
		"username":   sd["username"],
		"role":       sd["role"],
		"fullName":   sd["fullName"],
	})
}

func MedicineUpdate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	catID, _ := primitive.ObjectIDFromHex(c.PostForm("category_id"))
	supID, _ := primitive.ObjectIDFromHex(c.PostForm("supplier_id"))
	qty, _ := strconv.Atoi(c.PostForm("quantity"))
	minStock, _ := strconv.Atoi(c.PostForm("min_stock_level"))
	purchasePrice, _ := strconv.ParseFloat(c.PostForm("purchase_price"), 64)
	sellingPrice, _ := strconv.ParseFloat(c.PostForm("selling_price"), 64)
	expiryDate, _ := time.Parse("2006-01-02", c.PostForm("expiry_date"))
	prescription := c.PostForm("prescription") == "true"

	var cat models.Category
	var sup models.Supplier
	database.Col("categories").FindOne(ctx, bson.M{"_id": catID}).Decode(&cat)
	database.Col("suppliers").FindOne(ctx, bson.M{"_id": supID}).Decode(&sup)

	update := bson.M{
		"$set": bson.M{
			"name":           c.PostForm("name"),
			"generic_name":   c.PostForm("generic_name"),
			"category_id":    catID,
			"category_name":  cat.Name,
			"supplier_id":    supID,
			"supplier_name":  sup.Name,
			"manufacturer":   c.PostForm("manufacturer"),
			"batch_number":   c.PostForm("batch_number"),
			"expiry_date":    expiryDate,
			"quantity":       qty,
			"min_stock_level": minStock,
			"purchase_price": purchasePrice,
			"selling_price":  sellingPrice,
			"unit":           c.PostForm("unit"),
			"description":    c.PostForm("description"),
			"prescription":   prescription,
			"updated_at":     time.Now(),
		},
	}

	database.Col("medicines").UpdateOne(ctx, bson.M{"_id": id}, update)
	c.Redirect(http.StatusFound, "/medicines")
}

func MedicineDelete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	database.Col("medicines").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"active": false}})

	c.Header("HX-Redirect", "/medicines")
	c.Status(http.StatusOK)
}

func MedicineSearch(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := c.Query("q")
	filter := bson.M{
		"active": true,
		"$or": bson.A{
			bson.M{"name": bson.M{"$regex": q, "$options": "i"}},
			bson.M{"generic_name": bson.M{"$regex": q, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(10).SetSort(bson.M{"name": 1})
	cursor, _ := database.Col("medicines").Find(ctx, filter, opts)
	var medicines []models.Medicine
	cursor.All(ctx, &medicines)

	c.JSON(http.StatusOK, medicines)
}
