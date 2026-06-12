package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"movieTicket/backend/internal/database"
	"movieTicket/backend/internal/movie"
)

func main() {
	// 1. Check for command line arguments
	// Usage: go run main.go "Movie Title" 2024
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go \"Movie Title\" <Year>")
		fmt.Println("Example: go run main.go \"The Batman\" 2022")
		return
	}

	title := os.Args[1]
	yearStr := os.Args[2]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		log.Fatalf("Invalid year format: %v", err)
	}

	// 2. Connect to MongoDB
	client := database.ConnectDB()
	repo := movie.NewMovieRepository(client)

	// 3. Prepare the movie data
	newMovie := movie.Movie{
		Title: title,
		Year:  year,
	}

	// 4. Insert into MongoDB
	err = repo.Create(context.Background(), newMovie)
	if err != nil {
		log.Fatalf("Failed to add movie: %v", err)
	}

	fmt.Printf("Successfully added '%s' (%d) to MongoDB! \n", newMovie.Title, newMovie.Year)
}
