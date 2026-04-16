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

func CategoryList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, _ := database.Col("categories").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	cursor.All(ctx, &categories)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "categories/index.html", gin.H{
		"title":      "Categories",
		"nav":        "categories",
		"categories": categories,
		"username":   sd["username"],
		"role":       sd["role"],
		"fullName":   sd["fullName"],
	})
}

func CategoryCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cat := models.Category{
		ID:          primitive.NewObjectID(),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	database.Col("categories").InsertOne(ctx, cat)

	// Return updated list for HTMX
	cursor, _ := database.Col("categories").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	cursor.All(ctx, &categories)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "categories/table.html", gin.H{
		"categories": categories,
		"role":       sd["role"],
	})
}

func CategoryUpdate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	active := c.PostForm("active") == "true"

	database.Col("categories").UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"name":        c.PostForm("name"),
			"description": c.PostForm("description"),
			"active":      active,
			"updated_at":  time.Now(),
		},
	})

	cursor, _ := database.Col("categories").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	cursor.All(ctx, &categories)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "categories/table.html", gin.H{
		"categories": categories,
		"role":       sd["role"],
	})
}

func CategoryDelete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	database.Col("categories").DeleteOne(ctx, bson.M{"_id": id})

	cursor, _ := database.Col("categories").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	var categories []models.Category
	cursor.All(ctx, &categories)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "categories/table.html", gin.H{
		"categories": categories,
		"role":       sd["role"],
	})
}
