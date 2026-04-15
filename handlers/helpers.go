package handlers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func primitiveObjID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

func sessionData(c *gin.Context) gin.H {
	session := sessions.Default(c)
	return gin.H{
		"username": session.Get("username"),
		"role":     session.Get("role"),
		"fullName": session.Get("full_name"),
	}
}
