package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/bootdotdev/curriculum/blogaggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	switch err {
	case sql.ErrNoRows:
		return fmt.Errorf("user does not exist, please register user first")
	case nil:
		if err := s.cfg.SetUser(name); err != nil {
			return fmt.Errorf("couldn't set current user: %w", err)
		}
		fmt.Println("User switched successfully!")
		return nil
	default:
		return fmt.Errorf("get user: %w", err)
	}
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	switch err {
	case sql.ErrNoRows:
		newUser := database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      name,
		}
		user, err := s.db.CreateUser(context.Background(), newUser)
		if err != nil {
			return err
		}
		err = s.cfg.SetUser(name)
		if err != nil {
			return fmt.Errorf("couldn't set current user: %w", err)
		}
		fmt.Printf("User %s was successfully registered!\n", user.Name)
		fmt.Println("User Data Debugging:")
		fmt.Printf("%+v\n", user)

	case nil:
		return fmt.Errorf("user already exists")

	default:
		return fmt.Errorf("get user: %w", err)
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users %w", err)
	}
	fmt.Println("successfully deleted all users")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users %w", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time between reqs> ", cmd.Name)
	}

	time_between_reqs := cmd.Args[0]

	timeBetweenRequests, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("error parsing time between requests: %w", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests.String())

	err = scrapeFeeds(s)
	if err != nil {
		log.Printf("error scraping feed %v\n", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			log.Printf("error scraping feed %v\n", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s name url", cmd.Name)
	}

	name := cmd.Args[0]
	feedURL := cmd.Args[1]

	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       feedURL,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), newFeed)
	if err != nil {
		return fmt.Errorf("error creating feed %w", err)
	}
	fmt.Printf("New Feed Info: %+v\n", feed)

	followCmd := command{
		Name: "follow",
		Args: []string{feedURL},
	}

	err = handlerFollowFeed(s, followCmd, user)
	if err != nil {
		return fmt.Errorf("error following new feed %w", err)
	}

	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("* Feed: \"%s\" URL: \"%s\" Creator: %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s \"url\"", cmd.Name)
	}

	feedURL := cmd.Args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error getting feed by URL: %w", err)
	}

	newFeedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	newFeedFollowRow, err := s.db.CreateFeedFollow(context.Background(), newFeedFollow)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && string(pqErr.Code) == "23505" {
			fmt.Println("You're already following that feed.")
			return nil
		}
		return err
	}

	fmt.Printf("Feed Name: %s Current User: %s\n", newFeedFollowRow.FeedName, newFeedFollowRow.UserName)

	return nil
}

func handlerGetFollowingFeeds(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting followed feeds %w", err)
	}

	fmt.Printf("%s's Followed Feeds:\n", user.Name)
	for _, feed := range followedFeeds {
		fmt.Printf("* %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s \"url\"", cmd.Name)
	}

	feedURL := cmd.Args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error getting feed by URL: %w", err)
	}

	feedToUnfollow := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), feedToUnfollow)
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %w", err)
	}

	fmt.Printf("Feed Name: %s unfollowed by User: %s\n", feed.Name, user.Name)

	return nil
}

func handlerBrowsePosts(s *state, cmd command, user database.User) error {
	limit := int32(2) // default

	switch len(cmd.Args) {
	case 0:
		// keep default
	case 1:
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil || n < 1 || n > int(math.MaxInt32) {
			return fmt.Errorf("limit must be a positive integer")
		}
		limit = int32(n)
	default:
		return fmt.Errorf("usage: %s <limit>", cmd.Name)
	}

	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	}

	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting posts for user: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts to display")
		return nil
	}
	for _, post := range posts {
		desc := "(no description)"
		if post.Description.Valid {
			desc = post.Description.String
		}
		pub := "unknown"
		if post.PublishedAt.Valid {
			pub = post.PublishedAt.Time.Format(time.RFC822)
		}
		fmt.Printf("Feed: %s\n%s\n%s\nPublished: %s\n%s\n\n", post.FeedName, post.Title, post.Url, pub, desc)
	}
	return nil
}
