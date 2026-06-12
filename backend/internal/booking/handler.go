package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For development, allow all origins
	},
}

type BookingHandler struct {
	repo *BookingRepository
	hub  *Hub
}

func NewBookingHandler(repo *BookingRepository, hub *Hub) *BookingHandler {
	return &BookingHandler{repo: repo, hub: hub}
}

// broadcastUpdate sends the current seat state for a movie/showtime to all clients
func (h *BookingHandler) broadcastUpdate(movieID int, showtime string) {
	ctx := context.Background()
	booked, err := h.repo.GetOccupiedSeats(ctx, movieID, showtime)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch occupied seats for broadcast: %v\n", err)
		return
	}

	lockedMap, err := h.repo.GetLockedSeats(ctx, movieID, showtime)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch locked seats for broadcast: %v\n", err)
		return
	}

	msg, _ := json.Marshal(map[string]interface{}{
		"type":     "SEAT_UPDATE",
		"movie_id": movieID,
		"showtime": showtime,
		"booked":   booked,
		"locked":   lockedMap,
	})
	h.hub.broadcast <- msg
}

func (h *BookingHandler) HandleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] WebSocket upgrade failed: %v\n", err)
		return err
	}
	h.hub.register <- ws

	defer func() {
		h.hub.unregister <- ws
	}()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
	return nil
}

func (h *BookingHandler) CreateBooking(c echo.Context) error {
	uid := c.Get("uid").(string)

	var b Booking
	if err := c.Bind(&b); err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to bind booking data: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	b.UserID = uid

	occupied, err := h.repo.GetOccupiedSeats(c.Request().Context(), b.MovieID, b.Showtime)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to validate seats for user %s: %v\n", uid, err)
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

	if err := h.repo.Create(c.Request().Context(), &b); err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to create booking in DB for user %s: %v\n", uid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create booking"})
	}

	for _, seat := range b.Seats {
		_ = h.repo.UnlockSeat(c.Request().Context(), b.MovieID, b.Showtime, seat, uid)
	}

	fmt.Printf("[BOOKING SUCCESS] User %s booked seats %v for Movie %d\n", uid, b.Seats, b.MovieID)
	h.broadcastUpdate(b.MovieID, b.Showtime)
	return c.JSON(http.StatusCreated, b)
}

func (h *BookingHandler) GetUserBookings(c echo.Context) error {
	uid := c.Get("uid").(string)
	bookings, err := h.repo.GetByUserID(c.Request().Context(), uid)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch bookings for user %s: %v\n", uid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bookings"})
	}
	return c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetOccupiedSeats(c echo.Context) error {
	movieID, _ := strconv.Atoi(c.QueryParam("movie_id"))
	showtime := c.QueryParam("showtime")

	if movieID == 0 || showtime == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "movie_id and showtime are required"})
	}

	ctx := c.Request().Context()
	booked, err := h.repo.GetOccupiedSeats(ctx, movieID, showtime)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch booked seats: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch booked seats"})
	}

	lockedMap, err := h.repo.GetLockedSeats(ctx, movieID, showtime)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch locked seats: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch locked seats"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"booked": booked,
		"locked": lockedMap,
	})
}

func (h *BookingHandler) LockSeat(c echo.Context) error {
	uid := c.Get("uid").(string)
	var req struct {
		MovieID  int    `json:"movie_id"`
		Showtime string `json:"showtime"`
		Seat     string `json:"seat"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	success, err := h.repo.LockSeat(c.Request().Context(), req.MovieID, req.Showtime, req.Seat, uid)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Redis lock failure for user %s: %v\n", uid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Lock service failure"})
	}

	if !success {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Seat already locked by another user"})
	}

	h.broadcastUpdate(req.MovieID, req.Showtime)

	// Set a timer to manually unlock and broadcast after 5 minutes.
	time.AfterFunc(5*time.Minute, func() {
		// Check if it's still locked by this user before announcing timeout
		// (In case they already booked it or manually unlocked it)
		err := h.repo.UnlockSeat(context.Background(), req.MovieID, req.Showtime, req.Seat, uid)
		if err == nil {
			fmt.Printf("[BOOKING TIMEOUT] Seat %s for Movie %d timed out for user %s\n", req.Seat, req.MovieID, uid)
			h.broadcastUpdate(req.MovieID, req.Showtime)
		}
	})

	return c.JSON(http.StatusOK, map[string]string{"status": "Seat locked"})
}

func (h *BookingHandler) UnlockSeat(c echo.Context) error {
	uid := c.Get("uid").(string)
	var req struct {
		MovieID  int    `json:"movie_id"`
		Showtime string `json:"showtime"`
		Seat     string `json:"seat"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	err := h.repo.UnlockSeat(c.Request().Context(), req.MovieID, req.Showtime, req.Seat, uid)
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Redis unlock failure for user %s: %v\n", uid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unlock service failure"})
	}

	fmt.Printf("[SEAT RELEASED] Seat %s for Movie %d manually released/unlocked by user %s\n", req.Seat, req.MovieID, uid)
	h.broadcastUpdate(req.MovieID, req.Showtime)
	return c.JSON(http.StatusOK, map[string]string{"status": "Seat unlocked"})
}

func (h *BookingHandler) GetAllBookings(c echo.Context) error {
	bookings, err := h.repo.GetAll(c.Request().Context())
	if err != nil {
		fmt.Printf("[SYSTEM ERROR] Failed to fetch all bookings: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch history"})
	}
	return c.JSON(http.StatusOK, bookings)
}
