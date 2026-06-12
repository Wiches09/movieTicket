package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

type AuthMiddleware struct {
	FirebaseApp *firebase.App
}

func NewAuthMiddleware() (*AuthMiddleware, error) {
	// Initialize Firebase Admin SDK using your service account/ credentials file
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	return &AuthMiddleware{FirebaseApp: app}, nil
}

// RestrictedHandler intercepts and validates the Firebase JWT Token
func (m *AuthMiddleware) RestrictedHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing authorization header"})
		}

		// Split "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token format"})
		}
		idToken := tokenParts[1]

		// Get the Auth client instance
		client, err := m.FirebaseApp.Auth(context.Background())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Auth service unavailable"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Verify the token validity
		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
		}

		// Save user details into Echo context so endpoints down the road can use them
		c.Set("uid", token.UID)
		c.Set("email", token.Claims["email"])

		return next(c)
	}
}
