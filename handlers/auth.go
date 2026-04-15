package handlers

import (
	"context"
	"net/http"
	"time"

	"medicalstore/database"
	"medicalstore/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Medical Store",
	})
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Username and password are required",
			"title": "Login - Medical Store",
		})
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := database.Col("users").FindOne(ctx, bson.M{"username": username, "active": true}).Decode(&user)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Invalid username or password",
			"title": "Login - Medical Store",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Invalid username or password",
			"title": "Login - Medical Store",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID.Hex())
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("full_name", user.FullName)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func ProfilePage(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(string)

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, _ := primitiveObjID(userID)
	database.Col("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user)

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":    "My Profile",
		"user":     user,
		"nav":      "profile",
		"username": session.Get("username"),
		"role":     session.Get("role"),
		"fullName": session.Get("full_name"),
	})
}

func UpdateProfile(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(string)

	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	newPassword := c.PostForm("new_password")

	objID, _ := primitiveObjID(userID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"full_name":  fullName,
			"email":      email,
			"updated_at": time.Now(),
		},
	}

	if newPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err == nil {
			update["$set"].(bson.M)["password"] = string(hash)
		}
	}

	database.Col("users").UpdateOne(ctx, bson.M{"_id": objID}, update)
	session.Set("full_name", fullName)
	session.Save()

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":    "My Profile",
		"success":  "Profile updated successfully",
		"nav":      "profile",
		"username": session.Get("username"),
		"role":     session.Get("role"),
		"fullName": fullName,
	})
}
