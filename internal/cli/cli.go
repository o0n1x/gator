package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/o0n1x/aggreGator/internal/config"
	"github.com/o0n1x/aggreGator/internal/database"
)

type State struct {
	State *config.Config
	DB    *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Commands map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("expected arg 'username' but was not found")
	}
	name := cmd.Args[0]
	_, err := s.DB.GetUser(context.Background(), name)
	if err != nil {
		fmt.Printf("Error login, User %v does not exist\n", name)
		os.Exit(1)
	}
	err = s.State.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("set username as: %s\n", name)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("expected arg 'username' but was not found")
	}
	name := cmd.Args[0]
	_, err := s.DB.GetUser(context.Background(), name)
	if err == nil {
		fmt.Printf("Error registering user: %v \n Error: User already Exists\n", name)
		os.Exit(1)
	}
	user, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Name:      name,
	})
	if err != nil {
		fmt.Printf("Error registering user: %v \n Error: %v", name, err)
		os.Exit(1)
	}
	s.State.SetUser(name)
	fmt.Println("User Registered: ", user)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		fmt.Printf("Error deleting users: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All users deleted Successfully")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Error retrieving users: %v\n", err)
		os.Exit(1)
	}
	for _, user := range users {
		fmt.Printf("* %v", user.Name)
		if user.Name == s.State.CurrentUserName {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	err := c.Commands[cmd.Name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Commands[name] = f
}
