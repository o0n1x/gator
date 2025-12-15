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
	"github.com/o0n1x/aggreGator/internal/rss"
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

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("expected arg 'name' and 'url' but was not found")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Name:      sql.NullString{String: name, Valid: true},
		Url:       sql.NullString{String: url, Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	})
	if err != nil {
		fmt.Printf("Error registering feed as User %v . Error: %v", name, err)
		os.Exit(1)
	}
	feedfollow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		fmt.Printf("DB Error for following,\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created Feed:\nName: %v\nURL: %v\nUsername: %v\n", feed.Name.String, feed.Url.String, user.Name)
	fmt.Printf("%v successfully followed %v\n", feedfollow.UserName, feedfollow.FeedName.String)
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

func HandlerAgg(s *State, cmd Command) error {
	rss, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		fmt.Printf("Error retrieving RSS feed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%v\n", rss)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Error retrieving feeds: %v\n", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID.UUID)
		if err != nil {
			fmt.Printf("Error feed fetching, User with id %v does not exist\n", feed.UserID.UUID)
			os.Exit(1)
		}
		fmt.Printf("* Name: %v\n  URL: %v\n  User: %v\n", feed.Name.String, feed.Url.String, user.Name)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("expected arg 'url' but was not found")
	}
	url := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), sql.NullString{String: url, Valid: true})
	if err != nil {
		fmt.Printf("Error following, feed with url: %v does not exist\n", url)
		os.Exit(1)
	}

	feedfollow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		fmt.Printf("DB Error for following,\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%v successfully followed %v\n", feedfollow.UserName, feedfollow.FeedName.String)
	return nil

}

func HandlerFollowing(s *State, cmd Command, user database.User) error {

	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil {
		fmt.Printf("DB Error for list follows,\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Feeds for the user %v:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("* FeedName: %v \n  User: %v\n", feed.FeedName.String, feed.UserName)
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
