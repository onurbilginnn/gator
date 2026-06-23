package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/onurbilginnn/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, url string) (*RSSFeed, error) {
	defer ctx.Done()
	request, requestErr := http.NewRequestWithContext(ctx, "GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}
	request.Header.Set("User-Agent", "gator")
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	rawFeed, ioErr := io.ReadAll(response.Body)
	if ioErr != nil {
		return nil, ioErr
	}
	var feed RSSFeed
	unmarshalErr := xml.Unmarshal(rawFeed, &feed)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &feed, nil
}

func scrapeFeeds(state *state) error {
	nextFeed, nextFeedErr := state.db.GetNextFeedToFetch(context.Background())
	if nextFeedErr != nil {
		return nextFeedErr
	}
	markError := state.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID,
	})
	if markError != nil {
		return markError
	}
	newFeed, fetchErr := fetchFeed(context.Background(), nextFeed.Url)
	if fetchErr != nil {
		return fetchErr
	}
	for _, item := range newFeed.Channel.Item {
		publishedTime, parseErr := time.Parse(time.RFC1123Z, item.PubDate)
		if parseErr != nil {
			return parseErr
		}
		_, createPostErr := state.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedTime,
			FeedID:      nextFeed.ID,
		})
		if createPostErr != nil {
			fmt.Printf("failed to create post: %v\n", createPostErr)
		}
		fmt.Printf("Scraped post: %s (URL: %s)\n", item.Title, item.Link)
	}
	return nil
}
