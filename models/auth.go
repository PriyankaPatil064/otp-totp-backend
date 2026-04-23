package models

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type OTPRequest struct {
	Email string `json:"email" binding:"required"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

type VerifyTOTPRequest struct {
	Email    string `json:"email" binding:"required"`
	Passcode string `json:"passcode" binding:"required"`
}

type MFAStep2Request struct {
	SessionID string `json:"session_id" binding:"required"`
	OTP       string `json:"otp" binding:"required"`
}

type MFAStep3Request struct {
	SessionID string `json:"session_id" binding:"required"`
	Passcode  string `json:"passcode" binding:"required"`
}
