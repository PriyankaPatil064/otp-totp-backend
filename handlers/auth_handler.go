package handlers

import (
	"mfa-backend/config"
	"mfa-backend/models"
	"mfa-backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Session states
const (
	StateAwaitingOTP   = "AWAITING_OTP"
	StateAwaitingTOTP  = "AWAITING_TOTP"
	StateAuthenticated = "AUTHENTICATED"
)

// OTP Handlers

func SendOTP(c *gin.Context) {
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otp, err := services.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	err = services.StoreOTP(req.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully", "otp": otp})
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, msg, err := services.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

// TOTP Handlers

func GenerateTOTP(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	secret, qrBase64, err := services.GenerateTOTPSecret(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":       secret,
		"qr_code":      "data:image/png;base64," + qrBase64,
		"instructions": "Scan the QR code with Google Authenticator or Authy",
	})
}

func VerifyTOTP(c *gin.Context) {
	var req models.VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, err := services.VerifyTOTP(req.Email, req.Passcode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify TOTP"})
		return
	}

	if !success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP passcode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "TOTP verified successfully"})
}

// MFA Flow Handlers

func LoginStep1(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Password != "password123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	sessionID := uuid.New().String()
	sessionKey := "session:" + sessionID

	err := config.RDB.HSet(config.Ctx, sessionKey, map[string]interface{}{
		"email": req.Email,
		"state": StateAwaitingOTP,
	}).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}
	config.RDB.Expire(config.Ctx, sessionKey, 10*time.Minute)

	otp, err := services.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}
	err = services.StoreOTP(req.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Step 1 successful. Please verify OTP.",
		"session_id": sessionID,
		"next_step":  "POST /login/step2",
		"demo_otp":   otp,
	})
}

func LoginStep2(c *gin.Context) {
	var req models.MFAStep2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionKey := "session:" + req.SessionID
	sessionData, err := config.RDB.HGetAll(config.Ctx, sessionKey).Result()
	if err != nil || len(sessionData) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	if sessionData["state"] != StateAwaitingOTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session state"})
		return
	}

	email := sessionData["email"]
	success, msg, err := services.VerifyOTP(email, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying OTP"})
		return
	}
	if !success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	config.RDB.HSet(config.Ctx, sessionKey, "state", StateAwaitingTOTP)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Step 2 successful. Please verify TOTP.",
		"next_step": "POST /login/step3",
	})
}

func LoginStep3(c *gin.Context) {
	var req models.MFAStep3Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionKey := "session:" + req.SessionID
	sessionData, err := config.RDB.HGetAll(config.Ctx, sessionKey).Result()
	if err != nil || len(sessionData) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	if sessionData["state"] != StateAwaitingTOTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session state"})
		return
	}

	email := sessionData["email"]
	success, err := services.VerifyTOTP(email, req.Passcode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying TOTP"})
		return
	}
	if !success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP passcode"})
		return
	}

	config.RDB.HSet(config.Ctx, sessionKey, "state", StateAuthenticated)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful!", "user": email})
}
