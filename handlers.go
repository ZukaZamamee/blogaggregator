package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bootdotdev/curriculum/blogaggregator/internal/database"
	"github.com/google/uuid"
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
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	feedURL := "https://www.wagslane.dev/index.xml"

	rssFeed, err := FetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}

	fmt.Printf("Feed: %+v\n", rssFeed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s name url", cmd.Name)
	}

	name := cmd.Args[0]
	feedURL := cmd.Args[1]
	currentUser := s.cfg.CurrentUserName
	user, err := s.db.GetUser(context.Background(), currentUser)
	if err != nil {
		return fmt.Errorf("error getting current user info %w", err)
	}

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
	fmt.Printf("New Feed Info: %+v", feed)
	return nil
}
