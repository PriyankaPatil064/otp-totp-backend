package services

import (
	"encoding/base64"
	"mfa-backend/config"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

func GenerateTOTPSecret(email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MFABackend",
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}

	secret := key.Secret()
	url := key.URL()

	// Generate QR code as PNG bytes then encode to base64
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	qrBase64 := base64.StdEncoding.EncodeToString(png)

	// In a real app, you'd store this secret in a database associated with the user.
	// For this demo, we can temporarily store it in Redis or just return it.
	// We'll store it in Redis for verification purposes in this simple backend.
	err = config.RDB.Set(config.Ctx, "totp_secret:"+email, secret, 24*time.Hour).Err()
	if err != nil {
		return "", "", err
	}

	return secret, qrBase64, nil
}

func VerifyTOTP(email, passcode string) (bool, error) {
	secret, err := config.RDB.Get(config.Ctx, "totp_secret:"+email).Result()
	if err != nil {
		return false, err
	}

	valid := totp.Validate(passcode, secret)
	return valid, nil
}
