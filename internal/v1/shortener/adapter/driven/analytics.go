package driven

import (
	"context"
	"log"
	"sync"

	"url-shortener/internal/v1/shortener/constant"
	"url-shortener/internal/v1/shortener/domain"
)

// Tracker implements domain.ClickTracker using a buffered Go channel and a
// background goroutine. This decouples click recording from the HTTP request
// path — the redirect returns immediately while the click is counted later.
//
// How it works:
//  1. Track() puts the short code into a buffered channel (non-blocking).
//  2. A background goroutine reads from the channel and writes to the DB.
//  3. Stop() closes the channel and waits for all pending clicks to flush.
type Tracker struct {
	repo domain.URLRepository // Used to persist click counts to the database
	ch   chan string          // Buffered channel that queues short codes to be counted
	wg   sync.WaitGroup       // WaitGroup ensures graceful shutdown waits for the goroutine
}

// NewTracker creates a Tracker with a buffered channel of ChannelBufferSize.
func NewTracker(repo domain.URLRepository) *Tracker {
	return &Tracker{
		repo: repo,
		ch:   make(chan string, constant.ChannelBufferSize),
	}
}

// Track enqueues a click event for the given short code.
// Uses a non-blocking send (select/default) so it never slows down the caller.
// If the channel is full, the click is dropped and a warning is logged.
func (t *Tracker) Track(code string) {
	select {
	case t.ch <- code:
	default:
		log.Printf("click tracker channel full, dropping click for %s", code)
	}
}

// Start launches the background goroutine that reads click events from
// the channel and increments the click count in the database.
func (t *Tracker) Start(ctx context.Context) {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		// "range" over a channel reads until the channel is closed.
		for code := range t.ch {
			if err := t.repo.IncrementClicks(ctx, code); err != nil {
				log.Printf("increment clicks error: %v", err)
			}
		}
	}()
}

// Stop signals the tracker to shut down by closing the channel,
// then waits for all remaining click events to be processed.
func (t *Tracker) Stop() {
	close(t.ch)
	t.wg.Wait()
}
