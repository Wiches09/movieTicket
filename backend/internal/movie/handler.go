package movie

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type MovieHandler struct {
	repo *MovieRepository
}

func NewMovieHandler(repo *MovieRepository) *MovieHandler {
	return &MovieHandler{repo: repo}
}

func (h *MovieHandler) GetMovies(c echo.Context) error {
	movies, err := h.repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch movies"})
	}
	return c.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) GetMovie(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid movie ID format"})
	}

	movie, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Movie not found"})
	}
	return c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) CreateMovie(c echo.Context) error {
	var m Movie
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if m.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
	}

	err := h.repo.Create(c.Request().Context(), m)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create movie"})
	}

	return c.JSON(http.StatusCreated, m)
}
