package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"mfa-backend/config"
	"strconv"
	"time"
)

const OTP_EXPIRY = 60 * time.Second
const MAX_ATTEMPTS = 5

func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func StoreOTP(email, otp string) error {
	key := "otp:" + email
	err := config.RDB.Set(config.Ctx, key, otp, OTP_EXPIRY).Err()
	if err != nil {
		return err
	}

	// Reset attempts on new OTP request
	attemptsKey := "otp_attempts:" + email
	return config.RDB.Set(config.Ctx, attemptsKey, "0", OTP_EXPIRY).Err()
}

func VerifyOTP(email, otp string) (bool, string, error) {
	attemptsKey := "otp_attempts:" + email
	attemptsStr, err := config.RDB.Get(config.Ctx, attemptsKey).Result()
	if err != nil {
		return false, "No OTP found or expired", nil
	}

	attempts, _ := strconv.Atoi(attemptsStr)
	if attempts >= MAX_ATTEMPTS {
		return false, "Too many attempts. Please request a new OTP", nil
	}

	key := "otp:" + email
	storedOTP, err := config.RDB.Get(config.Ctx, key).Result()
	if err != nil {
		return false, "OTP expired", nil
	}

	if storedOTP != otp {
		config.RDB.Incr(config.Ctx, attemptsKey)
		return false, "Invalid OTP", nil
	}

	// Success: Delete OTP and attempts
	config.RDB.Del(config.Ctx, key, attemptsKey)
	return true, "Success", nil
}
