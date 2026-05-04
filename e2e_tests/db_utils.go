package e2e_tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	seededRates []struct{ Currency, Date string }
	mu          sync.Mutex
	envOnce     sync.Once
)

func getDB() *sql.DB {
	envOnce.Do(func() {
		// Load .env from current directory or e2e_tests subdir
		godotenv.Load(".env")
		godotenv.Load("e2e_tests/.env")
	})

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	return db
}

func SeedRate(currency string, date string, rate float64) {
	db := getDB()
	defer db.Close()

	query := `
		INSERT INTO currency_conversion_rates (target_currency, rate_date, exchange_rate) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (target_currency, rate_date) 
		DO UPDATE SET exchange_rate = EXCLUDED.exchange_rate, updated_at = NOW()`

	_, err := db.Exec(query, currency, date, rate)
	if err != nil {
		log.Fatalf("Failed to seed rate: %v", err)
	}

	mu.Lock()
	seededRates = append(seededRates, struct{ Currency, Date string }{currency, date})
	mu.Unlock()
}

func CleanupRates() {
	mu.Lock()
	defer mu.Unlock()

	if len(seededRates) == 0 {
		return
	}

	db := getDB()
	defer db.Close()

	for _, r := range seededRates {
		_, err := db.Exec("DELETE FROM currency_conversion_rates WHERE target_currency = $1 AND rate_date = $2", r.Currency, r.Date)
		if err != nil {
			log.Printf("Failed to cleanup rate for %s on %s: %v", r.Currency, r.Date, err)
		}
	}
	seededRates = nil
}
