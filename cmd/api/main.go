package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"main.go/internal/auth"
	"main.go/internal/database"
	"main.go/internal/stream"
	"main.go/internal/user"
)

func check(step string, err error) {
	if err != nil {

		log.Fatalf("FAILED at %s: %v", step, err)
	}
	fmt.Printf("ok  %s\n", step)
}

// randomID makes a throwaway pubic ID like "abc123" for a stream. It's not secure, but it's good enough for a demo.
func randomID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func main() {
	// A timeout so a bad connection can't hang the program forever.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect()
	check("connect to database", err)
	check("ping database", db.PingContext(ctx))

	userRepo := user.NewRepository(db)
	streamRepo := stream.NewRepository(db)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// ---- 1. Create a user ----
	u := &user.User{
		Username:     "test_" + suffix,
		Email:        "test_" + suffix + "@example.com",
		PasswordHash: "not-a-real-hash", // the repository doesn't care; the service hashes
	}

	check("create user", userRepo.CreateUser(ctx, u))
	fmt.Printf("    user id = %d\n", u.ID)

	// ---- 2. Read the user back by email ----
	got, err := userRepo.GetByEmail(ctx, u.Email)
	check("get user by email", err)
	if got.ID != u.ID {
		log.Fatalf("FAILED: got user id %d, want %d", got.ID, u.ID)
	}

	// ---- 3. A missing user must return ErrNotFound ----
	_, err = userRepo.GetByEmail(ctx, "nobody@example.com")
	if !errors.Is(err, user.ErrNotFound) {
		log.Fatalf("FAILED: expected user.ErrNotFound, got %v", err)
	}
	fmt.Println("ok  missing user returns ErrNotFound")
	// ---- 4. Create a stream owned by that user ----
	key, err := auth.GenerateStreamKey()
	check("generate stream key", err)

	s := &stream.Stream{
		UserID:     u.ID,
		Title:      "Scratch test stream",
		StreamKey:  key,
		PlaybackID: randomID("pb_"),
	}
	check("create stream", streamRepo.Create(ctx, s))
	fmt.Printf("    stream id = %d\n", s.ID)

	// ---- 5. Read it back by key ----
	found, err := streamRepo.GetByKey(ctx, key)
	check("get stream by key", err)
	fmt.Printf("title=%q status=%s started_at=%v\n", found.Title, found.Status, found.StartedAt)

	// ---- 6. A wrong key must return ErrNotFound ----
	_, err = streamRepo.GetByKey(ctx, "live_wrong")
	if !errors.Is(err, stream.ErrNotFound) {
		log.Fatalf("FAILED: expected stream.ErrNotFound, got %v", err)
	}
	fmt.Println("ok  wrong key returns ErrNotFound")

	// ---- 7. The same key twice must be rejected (UNIQUE constraint) ----
	dup := &stream.Stream{UserID: u.ID, Title: "dup", StreamKey: key, PlaybackID: randomID("pb_")}
	if err := streamRepo.Create(ctx, dup); err == nil {
		log.Fatal("FAILED: duplicate stream key was accepted, UNIQUE constraint is missing")
	}
	fmt.Println("ok  duplicate stream key rejected")

	// ---- 8. Go live, then check it appears in the live list ----
	check("set status live", streamRepo.SetStatus(ctx, s.ID, stream.StatusLive))

	live, err := streamRepo.ListLive(ctx)
	check("list live streams", err)

	inList := false
	for _, ls := range live {
		if ls.ID == s.ID {
			inList = true
			fmt.Printf("    started_at is now set: %v\n", ls.StartedAt)
		}
	}
	if !inList {
		log.Fatal("FAILED: live stream missing from ListLive")
	}
	fmt.Println("ok  live stream appears in ListLive")

	// ---- 9. Go offline, check it disappears ----
	check("set status offline", streamRepo.SetStatus(ctx, s.ID, stream.StatusOffline))
	live, err = streamRepo.ListLive(ctx)
	check("list live streams again", err)
	for _, ls := range live {
		if ls.ID == s.ID {
			log.Fatal("FAILED: offline stream still in ListLive")
		}
	}
	fmt.Println("ok  offline stream removed from ListLive")

	// ---- 10. Clean up the test rows (streams first: they reference the user) ----
	_, err = db.ExecContext(ctx, `DELETE FROM streams WHERE user_id = $1`, u.ID)
	check("delete test streams", err)
	_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, u.ID)
	check("delete test user", err)

	fmt.Println("\nALL CHECKS PASSED")

}
