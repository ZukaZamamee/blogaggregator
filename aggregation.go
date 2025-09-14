package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/curriculum/blogaggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func scrapeFeeds(s *state) error {
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer parentCancel()

	nextFeed, err := s.db.GetNextFeedToFetch(parentCtx)
	if err != nil {
		return fmt.Errorf("couldn't fetch next feed to fetch: %w", err)
	}

	err = s.db.MarkFeedFetched(parentCtx, nextFeed.ID)
	if err != nil {
		return fmt.Errorf("couldn't update feed as fetched: %w", err)
	}

	rssFeed, err := FetchFeed(parentCtx, nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	for _, item := range rssFeed.Channel.Item {
		insertCtx, cancel := context.WithTimeout(parentCtx, 3*time.Second)

		pub := parsePubDate(item.PubDate)
		if !pub.Valid {
			log.Printf("unparsed PubDate: %v\n", item.PubDate)
		}

		now := time.Now()
		newPost := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: pub,
			FeedID:      nextFeed.ID,
		}

		_, err := s.db.CreatePost(insertCtx, newPost)
		cancel()
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" && pqErr.Constraint == "posts_url_key" {
					// duplicate URL – ignore
					continue
				}
			}
			log.Printf("error: %v\n", err)
		}
	}
	return nil
}
