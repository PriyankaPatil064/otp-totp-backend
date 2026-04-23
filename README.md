# 🔐 OTP-TOTP-Backend API (Go + Gin + Redis)

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Gonic](https://img.shields.io/badge/Framework-Gin-0081C9?style=flat)](https://gin-gonic.com/)
[![Redis](https://img.shields.io/badge/Database-Redis-DC382D?style=flat&logo=redis)](https://redis.io/)


A high-performance, secure Multi-Factor Authentication (MFA) backend implemented in Go. This project provides a robust solution for implementing 2FA (Two-Factor Authentication) using both SMS-style OTP and Google Authenticator-style TOTP.

---

## 🚀 Features

- **OTP (One-Time Password)**:
  - Secure 6-digit generation using `crypto/rand`.
  - Redis-backed storage with 60s automatic expiry.
  - Bruteforce protection (max 5 failed attempts).
- **TOTP (Time-based One-Time Password)**:
  - Industry-standard algorithm (`pquerna/otp`).
  - QR Code generation for instant enrollment.
  - Compatible with Google Authenticator, Authy, and Microsoft Authenticator.
- **MFA Login Workflow**:
  - Secure state management in Redis.
  - Sequential validation: Password ➔ OTP ➔ TOTP.
- **Clean Architecture**:
  - Organized into `config`, `handlers`, `models`, and `services`.
  - Beginner-friendly and highly extensible.

---

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Web Framework**: [Gin Gonic](https://gin-gonic.com/)
- **Caching/State**: [Redis](https://redis.io/)
- **MFA/Security**: 
  - `pquerna/otp` (TOTP logic)
  - `go-qrcode` (QR generation)
  - `google/uuid` (Session tracking)

---

## 📂 Project Structure

```text
.
├── config/           # Redis & System configurations
├── handlers/         # API Controllers (Request handling)
├── models/           # Data Structures (Request/Response)
├── services/         # Core Business Logic (OTP/TOTP)
├── main.go           # Application Entry Point
└── README.md         # Documentation
```

---

## 🚦 Getting Started

### Prerequisites
- **Go**: [Download & Install](https://golang.org/dl/)
- **Redis**: Running on `localhost:6379`. (Use Docker or native install)

### Running Locally
1. Clone the repository.
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Start the server:
   ```bash
   go run .
   ```
   *The API will be available at `http://localhost:8080`.*

---

## 📡 API Documentation

### 1. General
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/health` | `GET` | Service status check |

### 2. Standalone MFA Components
| Endpoint | Method | Payload / Query | Description |
| :--- | :--- | :--- | :--- |
| `/send-otp` | `POST` | `{"email": "..."}` | Generates and "sends" a 6-digit OTP |
| `/verify-otp` | `POST` | `{"email": "...", "otp": "..."}` | Validates the OTP against Redis |
| `/generate-totp`| `GET` | `?email=...` | Generates secret and QR code (Base64) |
| `/verify-totp` | `POST` | `{"email": "...", "passcode": "..."}` | Validates TOTP token from app |

### 3. Integrated Login Flow (Complete MFA)
| Step | Endpoint | Payload | Action |
| :--- | :--- | :--- | :--- |
| **1** | `/login/step1` | `{"email": "...", "password": "..."}` | Verify password & start session |
| **2** | `/login/step2` | `{"session_id": "...", "otp": "..."}` | Verify OTP for the active session |
| **3** | `/login/step3` | `{"session_id": "...", "passcode": "..."}` | Verify TOTP for final authentication |

---

## 🧪 Testing with Postman

A Postman collection is included in the repository: `mfa_backend_postman_collection.json`.

1. Open Postman.
2. Click **Import**.
3. Select the `mfa_backend_postman_collection.json` file.
4. Use the **MFA Backend API** collection to test all endpoints.

---

## 📝 Usage Example (cURL)

**Step 1: Password Login**
```bash
curl -X POST http://localhost:8080/login/step1 \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

**Step 2: Generate TOTP QR Code (Enrollment)**
```bash
curl http://localhost:8080/generate-totp?email=user@example.com
```

