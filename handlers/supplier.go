package handlers

import (
	"context"
	"net/http"
	"time"

	"medicalstore/database"
	"medicalstore/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SupplierList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	search := c.Query("search")
	filter := bson.M{}
	if search != "" {
		filter["$or"] = bson.A{
			bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"contact_person": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"email": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	cursor, _ := database.Col("suppliers").Find(ctx, filter, options.Find().SetSort(bson.M{"name": 1}))
	var suppliers []models.Supplier
	cursor.All(ctx, &suppliers)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "suppliers/index.html", gin.H{
		"title":     "Suppliers",
		"nav":       "suppliers",
		"suppliers": suppliers,
		"search":    search,
		"username":  sd["username"],
		"role":      sd["role"],
		"fullName":  sd["fullName"],
	})
}

func SupplierNew(c *gin.Context) {
	sd := sessionData(c)
	c.HTML(http.StatusOK, "suppliers/form.html", gin.H{
		"title":    "Add Supplier",
		"nav":      "suppliers",
		"supplier": models.Supplier{},
		"isNew":    true,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func SupplierCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	supplier := models.Supplier{
		ID:            primitive.NewObjectID(),
		Name:          c.PostForm("name"),
		ContactPerson: c.PostForm("contact_person"),
		Email:         c.PostForm("email"),
		Phone:         c.PostForm("phone"),
		Address:       c.PostForm("address"),
		City:          c.PostForm("city"),
		Active:        true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	database.Col("suppliers").InsertOne(ctx, supplier)
	c.Redirect(http.StatusFound, "/suppliers")
}

func SupplierEdit(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var supplier models.Supplier
	database.Col("suppliers").FindOne(ctx, bson.M{"_id": id}).Decode(&supplier)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "suppliers/form.html", gin.H{
		"title":    "Edit Supplier",
		"nav":      "suppliers",
		"supplier": supplier,
		"isNew":    false,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func SupplierUpdate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	database.Col("suppliers").UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"name":           c.PostForm("name"),
			"contact_person": c.PostForm("contact_person"),
			"email":          c.PostForm("email"),
			"phone":          c.PostForm("phone"),
			"address":        c.PostForm("address"),
			"city":           c.PostForm("city"),
			"updated_at":     time.Now(),
		},
	})
	c.Redirect(http.StatusFound, "/suppliers")
}

func SupplierDelete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	database.Col("suppliers").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"active": false}})

	c.Header("HX-Redirect", "/suppliers")
	c.Status(http.StatusOK)
}
