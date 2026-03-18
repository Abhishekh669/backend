package lib

import (
	"errors"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId              string `json:"user_id"`
	Email               string `json:"email"`
	LastPasswordResetAt int64  `json:"last_password_reset_at"`
	jwt.RegisteredClaims
}

type OrderApprovalClaims struct {
	Id          string `json:"id"`
	TableNumber int    `json:"table_number"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   int64  `json:"created_at"`
	jwt.RegisteredClaims
}

type OrderApprovalDataType struct {
	Id          string `json:"id"`
	TableNumber int    `json:"table_number"`
	PhoneNumber string `json:"phone_number"`
}

type JwtDataType struct {
	UserId              string
	Email               string
	LastPasswordResetAt int64
}

func GenerateJWTToken(jwtData *JwtDataType) (string, error) {
	if jwtData.Email == "" || jwtData.UserId == "" {
		return "", fmt.Errorf("invalid user data")
	}
	secret := config.AppConfig.JWTToken
	if secret == "" {
		return "", fmt.Errorf("invalid jwt secret")
	}

	now := time.Now()

	claims := &Claims{
		UserId:              jwtData.UserId,
		Email:               jwtData.Email,
		LastPasswordResetAt: jwtData.LastPasswordResetAt,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)), // 7 days access token
			Issuer:    "golang-backend",
			Subject:   jwtData.UserId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}

func ParseJwtToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	secret := config.AppConfig.JWTToken
	if secret == "" {
		return nil, errors.New("jwt secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	},
		jwt.WithIssuer("golang-backend"),
		jwt.WithLeeway(2*time.Minute),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errors.New("token expired")
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errors.New("malformed token")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errors.New("token not valid yet")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errors.New("invalid token signature")
		default:
			return nil, errors.New("invalid token")
		}
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserId == "" || claims.Email == "" {
		return nil, errors.New("token missing required claims")
	}

	return claims, nil

}

// GenerateOrderApprovalToken creates a JWT for order approval session
func GenerateOrderApprovalToken(session *OrderApprovalDataType) (string, error) {
	if session.Id == "" || session.TableNumber == 0 || session.PhoneNumber == "" {
		return "", fmt.Errorf("invalid session data")
	}

	secret := config.AppConfig.JWTToken
	if secret == "" {
		return "", fmt.Errorf("jwt secret not configured")
	}

	now := time.Now()
	claims := &OrderApprovalClaims{
		Id:          session.Id,
		TableNumber: session.TableNumber,
		PhoneNumber: session.PhoneNumber,
		CreatedAt:   now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // valid for 24 hours
			Issuer:    "golang-backend",
			Subject:   session.Id,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateOrderApprovalToken parses the token and extracts data without querying the DB
func ValidateOrderApprovalToken(tokenString string) (*OrderApprovalDataType, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	secret := config.AppConfig.JWTToken
	if secret == "" {
		return nil, errors.New("jwt secret not configured")
	}

	// Parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &OrderApprovalClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*OrderApprovalClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check expiration manually if needed
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	// Optionally check the CreatedAt timestamp in claims (e.g., older than 23h)
	createdAt := time.Unix(claims.CreatedAt, 0)
	if time.Since(createdAt) > 23*time.Hour {
		return nil, errors.New("token session expired")
	}

	// Return the extracted data
	return &OrderApprovalDataType{
		Id:          claims.Id,
		TableNumber: claims.TableNumber,
		PhoneNumber: claims.PhoneNumber,
	}, nil
}
