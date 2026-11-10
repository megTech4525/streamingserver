// cmd/dbcheck/main.go
// Throwaway: confirms the database is really reachable and shows what we're connected to.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"main.go/internal/database"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("1. Connect failed: %v", err)
	}
	defer db.Close()
	fmt.Println("1. Connect returned no error")

	// Ping makes a real round trip to the server.
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("2. Ping failed: %v", err)
	}
	fmt.Println("2. Ping succeeded")

	// A real query proves we can talk SQL, not just open a socket.
	var one int
	if err := db.GetContext(ctx, &one, `SELECT 1`); err != nil {
		log.Fatalf("3. SELECT 1 failed: %v", err)
	}
	fmt.Println("3. SELECT 1 returned", one)

	// Shows WHICH database and user you are really connected as.
	var info struct {
		Database string `db:"db"`
		User     string `db:"usr"`
	}
	if err := db.GetContext(ctx, &info,
		`SELECT current_database() AS db, current_user AS usr`); err != nil {
		log.Fatalf("4. identity query failed: %v", err)
	}
	fmt.Printf("4. Connected to database=%q as user=%q\n", info.Database, info.User)

	// Do the tables the app needs actually exist?
	for _, table := range []string{"users", "streams"} {
		var exists bool
		err := db.GetContext(ctx, &exists,
			`SELECT EXISTS (
			   SELECT 1 FROM information_schema.tables
			   WHERE table_schema = 'public' AND table_name = $1)`, table)
		if err != nil {
			log.Fatalf("5. table check failed: %v", err)
		}
		fmt.Printf("5. table %-8s exists: %v\n", table, exists)
	}

	// Pool settings in effect.
	s := db.Stats()
	fmt.Printf("6. pool: open=%d in_use=%d idle=%d max_open=%d\n",
		s.OpenConnections, s.InUse, s.Idle, s.MaxOpenConnections)

	fmt.Println("\nDATABASE CONNECTION OK")
}
