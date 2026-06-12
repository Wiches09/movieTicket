package booking

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type BookingHandler struct {
	repo *BookingRepository
}

func NewBookingHandler(repo *BookingRepository) *BookingHandler {
	return &BookingHandler{repo: repo}
}

func (h *BookingHandler) CreateBooking(c echo.Context) error {
	uid := c.Get("uid").(string) // Extracted from Firebase Auth middleware

	var b Booking
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	b.UserID = uid

	// 1. Validation: Check if seats are already occupied
	occupied, err := h.repo.GetOccupiedSeats(c.Request().Context(), b.MovieID, b.Showtime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to validate seats"})
	}

	for _, requestedSeat := range b.Seats {
		for _, takenSeat := range occupied {
			if requestedSeat == takenSeat {
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "Seat " + requestedSeat + " is already occupied",
				})
			}
		}
	}

	// 2. Save booking
	if err := h.repo.Create(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create booking"})
	}

	return c.JSON(http.StatusCreated, b)
}

func (h *BookingHandler) GetUserBookings(c echo.Context) error {
	uid := c.Get("uid").(string)

	bookings, err := h.repo.GetByUserID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bookings"})
	}

	return c.JSON(http.StatusOK, bookings)
}

// GetOccupiedSeats returns a list of seats already booked for a movie/showtime
func (h *BookingHandler) GetOccupiedSeats(c echo.Context) error {
	movieID, _ := strconv.Atoi(c.QueryParam("movie_id"))
	showtime := c.QueryParam("showtime")

	if movieID == 0 || showtime == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "movie_id and showtime are required"})
	}

	occupied, err := h.repo.GetOccupiedSeats(c.Request().Context(), movieID, showtime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch occupied seats"})
	}

	return c.JSON(http.StatusOK, occupied)
}

// GetAllBookings is an Admin only view for history
func (h *BookingHandler) GetAllBookings(c echo.Context) error {
	// Note: In a real app, check for "admin" role here.
	// For now, we rely on the RestrictedHandler in main.go
	ctx := c.Request().Context()

	bookings, err := h.repo.GetAll(ctx)
	if err != nil {
		c.Logger().Errorf("Database error fetching bookings: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch booking history"})
	}

	c.Logger().Infof("Retrieved %d bookings from database", len(bookings))
	return c.JSON(http.StatusOK, bookings)
}
