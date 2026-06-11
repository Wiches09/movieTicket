package user

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	repo *UserRepository
}

func NewUserHandler(r *UserRepository) *UserHandler {
	return &UserHandler{repo: r}
}

func (h *UserHandler) SaveProfile(c echo.Context) error {
	// Pull the UID safely out of the Echo Context (populated by your Firebase Middleware)
	uid := c.Get("uid").(string)
	email := c.Get("email").(string)

	// Bind additional profile data passed from the Nuxt UI frontend form
	var input struct {
		DisplayName string `json:"display_name"`
	}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	profile := UserProfile{
		FirebaseUID: uid,
		DisplayName: input.DisplayName,
		Email:       email,
	}

	// Save to MongoDB
	err := h.repo.UpsertProfile(context.Background(), profile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save profile"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "Profile saved successfully!"})
}
