package main

import (
	"mfa-backend/config"
	"mfa-backend/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Redis
	config.InitRedis()

	r := gin.Default()

	// General health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	// OTP Routes
	r.POST("/send-otp", handlers.SendOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)

	// TOTP Routes
	r.GET("/generate-totp", handlers.GenerateTOTP)
	r.POST("/verify-totp", handlers.VerifyTOTP)

	// Start server
	r.Run(":8080")
}
