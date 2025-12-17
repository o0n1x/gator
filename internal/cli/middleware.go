package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/o0n1x/gator/internal/database"
)

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {

	return func(s *State, cmd Command) error {
		user, err := s.DB.GetUser(context.Background(), s.State.CurrentUserName)
		if err != nil {
			fmt.Printf("Error registering feed as User %v is not found. Error: %v", s.State.CurrentUserName, err)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}
