package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 70),
			wantErr:  false,
		},
		{
			name:     "too long password",
			password: strings.Repeat("a", 100),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if !tt.wantErr && got == tt.password {
				t.Error("HashPassword() returned the same password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "invalid_hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()

	token, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Verify token can be validated
	validatedID, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}

	if validatedID != userID {
		t.Errorf("ValidateToken() returned userID = %v, want %v", validatedID, userID)
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	validToken, _ := GenerateToken(userID)

	tests := []struct {
		name    string
		token   string
		wantErr bool
		wantUID uuid.UUID
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
			wantUID: userID,
		},
		{
			name:    "invalid token",
			token:   "invalid.token.here",
			wantErr: true,
			wantUID: uuid.Nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			wantUID: uuid.Nil,
		},
		{
			name:    "malformed token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
			wantErr: true,
			wantUID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantUID {
				t.Errorf("ValidateToken() = %v, want %v", got, tt.wantUID)
			}
		})
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	// Создаем токен с неправильным signing method (RSA вместо HMAC)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: uuid.New(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString([]byte("fake-key"))

	_, err := ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for token with wrong signing method, got nil")
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "length 0",
			length: 0,
		},
		{
			name:   "length 16",
			length: 16,
		},
		{
			name:   "length 32",
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Errorf("GenerateRandomString() error = %v", err)
				return
			}

			expectedLen := tt.length * 2 // hex encoding doubles the length
			if len(got) != expectedLen {
				t.Errorf("GenerateRandomString() length = %v, want %v", len(got), expectedLen)
			}
		})
	}
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	str1, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString() error = %v", err)
	}

	str2, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString() error = %v", err)
	}

	if str1 == str2 {
		t.Error("GenerateRandomString() generated identical strings")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Создаем токен с экспирацией в прошлом
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // expired
		},
		UserID: uuid.New(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	_, err := ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateToken_InvalidTokenStructure(t *testing.T) {
	// Test with token that has invalid claims structure
	invalidToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":    time.Now().Add(time.Hour).Unix(),
		"userID": "not-a-uuid", // Invalid UUID format
	})
	tokenString, _ := invalidToken.SignedString(jwtSecret)

	_, err := ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for token with invalid UUID, got nil")
	}
}
