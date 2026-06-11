package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"movieTicket/backend/internal/auth"
	"movieTicket/backend/internal/database"
	"movieTicket/backend/internal/user"
)

func main() {
	e := echo.New()

	// Global Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS Configuration: Enforce safe rules allowing header options from Nuxt 3 (port 3000)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	// 1. Initialize MongoDB Client Pool Connection
	mongoClient := database.ConnectDB()

	// 2. Initialize User Layer Domain Dependencies
	userRepo := user.NewUserRepository(mongoClient)
	userHandler := user.NewUserHandler(userRepo)

	// 3. Initialize Firebase Auth Middleware Guard Engine
	authGuard, err := auth.NewAuthMiddleware()
	if err != nil {
		e.Logger.Fatalf("Failed to initialize Firebase Admin: %v", err)
	}

	// 4. Global Routing Groups
	api := e.Group("/api")
	{
		// Test Route: Standard protected verification placeholder
		api.GET("/secure-data", func(c echo.Context) error {
			uid := c.Get("uid").(string)
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Access Granted! This data is securely processed via Go Echo.",
				"uid":     uid,
			})
		}, authGuard.RestrictedHandler)

		// MongoDB Profile Route: Upserts frontend data straight into MongoDB documents
		// Wrapped in authGuard.RestrictedHandler to protect database writes
		api.POST("/profile/save", userHandler.SaveProfile, authGuard.RestrictedHandler)
	}

	// Start Server Engine
	e.Logger.Fatal(e.Start(":8080"))
}
