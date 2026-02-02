package app

import (
	"announcer/config"
	"announcer/internal/adapter"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func RunAnnounceBreakfast(cfg *config.Config) {
	// Get today's date in the format used in the CSV (dd/mm/yyyy)
	today := time.Now().Format("02/01/2006")

	log.Printf("Looking for breakfast menu for date: %s", today)

	// Fetch the CSV data
	resp, err := http.Get(cfg.BreakfastLink)
	if err != nil {
		log.Printf("Error fetching CSV data: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// Parse CSV
	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error parsing CSV data: %v", err)
		return
	}

	// Check if we have enough rows (need at least 4 rows: 0, 1, 2=dates, 3=food)
	if len(records) < 4 {
		log.Printf("CSV doesn't have enough rows (need at least 4, got %d)", len(records))
		return
	}

	// Row 3 (index 2) contains dates, Row 4 (index 3) contains food names
	datesRow := records[2]
	foodRow := records[3]

	// Find the column with today's date
	dateColumn := -1
	for i, date := range datesRow {
		if strings.TrimSpace(date) == today {
			dateColumn = i
			break
		}
	}

	if dateColumn == -1 {
		log.Printf("Today's date (%s) not found in the breakfast menu", today)

		// Show available dates for debugging
		var availableDates []string
		for _, date := range datesRow {
			if strings.Contains(date, "/") && len(strings.TrimSpace(date)) > 0 {
				availableDates = append(availableDates, strings.TrimSpace(date))
			}
		}
		log.Printf("Available dates: %v", availableDates)
		return
	}

	// Get the corresponding food item
	if dateColumn >= len(foodRow) {
		log.Printf("Food data not available for column %d", dateColumn)
		return
	}

	todaysFood := strings.TrimSpace(foodRow[dateColumn])
	if todaysFood == "" {
		log.Printf("No food item specified for today (%s)", today)
		return
	}

	// Send announcement to Discord as embed with color sidebar
	title := "🍽️ BREAKFAST ANNOUNCEMENT 🍽️"
	description := fmt.Sprintf("📅 **Date:** %s\n🍜 **Today's breakfast:** %s\n\nEnjoy your meal! 😋", today, todaysFood)
	color := 16753920 // Orange color for sidebar
	footerText := ""

	if err := adapter.SendDiscordEmbed(cfg.DiscordWebhookURL, title, description, color, footerText); err != nil {
		log.Printf("Error sending Discord embed: %v", err)
	}
}
