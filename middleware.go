package main

import (
	"context"
	"fmt"

	"github.com/bootdotdev/curriculum/blogaggregator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUser := s.cfg.CurrentUserName
		user, err := s.db.GetUser(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("error getting current user info %w", err)
		}
		return handler(s, cmd, user)
	}
}
