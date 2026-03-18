package lib

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractTokenFromHeader(c *gin.Context) (string, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		fmt.Println("token is not send")
		return "", fmt.Errorf("invalid session")
	}

	parts := strings.Split(tokenString, " ")

	if len(parts) != 2 {
		fmt.Println("token is not send")
		return "", fmt.Errorf("invalid session")
	}

	if strings.ToLower(parts[0]) != "bearer" {
		fmt.Println("token is not valid")
		return "", fmt.Errorf("invalid session")
	}

	token := strings.TrimSpace(parts[1])

	if token == "" {
		fmt.Println("token is not in correct format")
		return "", fmt.Errorf("invalid session")
	}

	return token, nil
}

func ExtractTokenFromCookie(c *gin.Context) (string, error) {
	// "session_token" is the cookie name — replace with your actual name
	token, err := c.Cookie("session_token")
	if err != nil {
		fmt.Println("token cookie not found")
		return "", fmt.Errorf("invalid session")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		fmt.Println("token cookie is empty")
		return "", fmt.Errorf("invalid session")
	}

	return token, nil
}
