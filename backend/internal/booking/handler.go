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
	"github.com/segmentio/kafka-go"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For development, allow all origins
	},
}

type BookingHandler struct {
	repo        *BookingRepository
	hub         *Hub
	kafkaWriter *kafka.Writer
}

func NewBookingHandler(repo *BookingRepository, hub *Hub, kw *kafka.Writer) *BookingHandler {
	return &BookingHandler{repo: repo, hub: hub, kafkaWriter: kw}
}

func (h *BookingHandler) logEvent(ctx context.Context, eventType, message string) {
	fmt.Printf("[%s] %s\n", eventType, message)
	_ = h.repo.InsertLog(ctx, eventType, message)
}

func (h *BookingHandler) broadcastUpdate(movieID int, showtime string) {
	ctx := context.Background()
	booked, err := h.repo.GetOccupiedSeats(ctx, movieID, showtime)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch occupied seats for broadcast: %v", err))
		return
	}

	lockedMap, err := h.repo.GetLockedSeats(ctx, movieID, showtime)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch locked seats for broadcast: %v", err))
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

func (h *BookingHandler) sendBookingEvent(ctx context.Context, b Booking) {
	if h.kafkaWriter == nil {
		return
	}

	event := map[string]interface{}{
		"event_type": "BOOKING_CONFIRMED",
		"booking_id": b.ID,
		"user_id":    b.UserID,
		"movie_id":   b.MovieID,
		"seats":      b.Seats,
		"showtime":   b.Showtime,
		"timestamp":  time.Now().Unix(),
	}

	payload, _ := json.Marshal(event)
	err := h.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(b.ID)),
		Value: payload,
	})

	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to send Kafka event: %v", err))
	} else {
		h.logEvent(ctx, "KAFKA EVENT", fmt.Sprintf("Sent BOOKING_CONFIRMED for booking %d", b.ID))
	}
}

func (h *BookingHandler) HandleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logEvent(context.Background(), "SYSTEM ERROR", fmt.Sprintf("WebSocket upgrade failed: %v", err))
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
	ctx := c.Request().Context()

	var b Booking
	if err := c.Bind(&b); err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to bind booking data: %v", err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	b.UserID = uid

	occupied, err := h.repo.GetOccupiedSeats(ctx, b.MovieID, b.Showtime)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to validate seats for user %s: %v", uid, err))
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

	if err := h.repo.Create(ctx, &b); err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to create booking in DB for user %s: %v", uid, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create booking"})
	}

	for _, seat := range b.Seats {
		_, _ = h.repo.UnlockSeat(ctx, b.MovieID, b.Showtime, seat, uid)
	}

	h.logEvent(ctx, "BOOKING SUCCESS", fmt.Sprintf("User %s booked seats %v for Movie %d", uid, b.Seats, b.MovieID))
	h.broadcastUpdate(b.MovieID, b.Showtime)

	// Send Kafka Event
	h.sendBookingEvent(ctx, b)

	return c.JSON(http.StatusCreated, b)
}

func (h *BookingHandler) GetUserBookings(c echo.Context) error {
	uid := c.Get("uid").(string)
	bookings, err := h.repo.GetByUserID(c.Request().Context(), uid)
	if err != nil {
		h.logEvent(c.Request().Context(), "SYSTEM ERROR", fmt.Sprintf("Failed to fetch bookings for user %s: %v", uid, err))
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
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch booked seats: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch booked seats"})
	}

	lockedMap, err := h.repo.GetLockedSeats(ctx, movieID, showtime)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch locked seats: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch locked seats"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"booked": booked,
		"locked": lockedMap,
	})
}

func (h *BookingHandler) LockSeat(c echo.Context) error {
	uid := c.Get("uid").(string)
	ctx := c.Request().Context()

	var req struct {
		MovieID  int    `json:"movie_id"`
		Showtime string `json:"showtime"`
		Seat     string `json:"seat"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	success, err := h.repo.LockSeat(ctx, req.MovieID, req.Showtime, req.Seat, uid)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Redis lock failure for user %s: %v", uid, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Lock service failure"})
	}

	if !success {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Seat already locked by another user"})
	}

	h.broadcastUpdate(req.MovieID, req.Showtime)

	lockUID := uid
	lockReq := req
	fmt.Printf("\n[TIMER] Started 5m lock for Seat %s (User: %s)\n", lockReq.Seat, lockUID)

	time.AfterFunc(5*time.Minute, func() {
		fmt.Printf("\n[TIMER] Fired for Seat %s\n", lockReq.Seat)
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		unlocked, err := h.repo.UnlockSeat(bgCtx, lockReq.MovieID, lockReq.Showtime, lockReq.Seat, lockUID)
		if err == nil && unlocked {
			fmt.Printf(">>> [TIMEOUT SUCCESS] Seat %s is now Available <<<\n", lockReq.Seat)
			h.logEvent(bgCtx, "BOOKING TIMEOUT", fmt.Sprintf("Seat %s for Movie %d timed out for user %s", lockReq.Seat, lockReq.MovieID, lockUID))
			h.broadcastUpdate(lockReq.MovieID, lockReq.Showtime)
		} else {
			fmt.Printf(">>> [TIMEOUT SKIP] Seat %s already released (unlocked=%v, err=%v) <<<\n", lockReq.Seat, unlocked, err)
		}
	})

	return c.JSON(http.StatusOK, map[string]string{"status": "Seat locked"})
}

func (h *BookingHandler) UnlockSeat(c echo.Context) error {
	uid := c.Get("uid").(string)
	ctx := c.Request().Context()

	var req struct {
		MovieID  int    `json:"movie_id"`
		Showtime string `json:"showtime"`
		Seat     string `json:"seat"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	_, err := h.repo.UnlockSeat(ctx, req.MovieID, req.Showtime, req.Seat, uid)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Redis unlock failure for user %s: %v", uid, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unlock service failure"})
	}

	h.logEvent(ctx, "SEAT RELEASED", fmt.Sprintf("Seat %s for Movie %d manually released/unlocked by user %s", req.Seat, req.MovieID, uid))
	h.broadcastUpdate(req.MovieID, req.Showtime)
	return c.JSON(http.StatusOK, map[string]string{"status": "Seat unlocked"})
}

func (h *BookingHandler) GetAllBookings(c echo.Context) error {
	ctx := c.Request().Context()
	bookings, err := h.repo.GetAll(ctx)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch all bookings: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch history"})
	}
	return c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetSystemLogs(c echo.Context) error {
	ctx := c.Request().Context()
	logs, err := h.repo.GetSystemLogs(ctx)
	if err != nil {
		h.logEvent(ctx, "SYSTEM ERROR", fmt.Sprintf("Failed to fetch system logs: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch system logs"})
	}
	return c.JSON(http.StatusOK, logs)
}
