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
	"golang.org/x/crypto/bcrypt"
)

func UserList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, _ := database.Col("users").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"username": 1}))
	var users []models.User
	cursor.All(ctx, &users)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "users/index.html", gin.H{
		"title":    "User Management",
		"nav":      "users",
		"users":    users,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func UserNew(c *gin.Context) {
	sd := sessionData(c)
	c.HTML(http.StatusOK, "users/form.html", gin.H{
		"title":    "Add User",
		"nav":      "users",
		"isNew":    true,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func UserCreate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	password := c.PostForm("password")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "users/form.html", gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  c.PostForm("username"),
		Email:     c.PostForm("email"),
		Password:  string(hash),
		FullName:  c.PostForm("full_name"),
		Role:      c.PostForm("role"),
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	database.Col("users").InsertOne(ctx, user)
	c.Redirect(http.StatusFound, "/users")
}

func UserEdit(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var user models.User
	database.Col("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)

	sd := sessionData(c)
	c.HTML(http.StatusOK, "users/form.html", gin.H{
		"title":    "Edit User",
		"nav":      "users",
		"editUser": user,
		"isNew":    false,
		"username": sd["username"],
		"role":     sd["role"],
		"fullName": sd["fullName"],
	})
}

func UserUpdate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	active := c.PostForm("active") == "true"

	update := bson.M{
		"$set": bson.M{
			"email":      c.PostForm("email"),
			"full_name":  c.PostForm("full_name"),
			"role":       c.PostForm("role"),
			"active":     active,
			"updated_at": time.Now(),
		},
	}

	if pw := c.PostForm("password"); pw != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		update["$set"].(bson.M)["password"] = string(hash)
	}

	database.Col("users").UpdateOne(ctx, bson.M{"_id": id}, update)
	c.Redirect(http.StatusFound, "/users")
}

func UserToggle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var user models.User
	database.Col("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	database.Col("users").UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"active": !user.Active, "updated_at": time.Now()},
	})

	c.Header("HX-Redirect", "/users")
	c.Status(http.StatusOK)
}
