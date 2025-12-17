package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/fatih/color"
	"github.com/o0n1x/gator/internal/config"
	"github.com/o0n1x/gator/internal/database"
	"github.com/o0n1x/gator/internal/rss"
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
	if len(cmd.Args) < 1 {
		return errors.New("expected arg 'time_between_reqs' but was not found")
	}
	duration_text := cmd.Args[0]

	duration, err := time.ParseDuration(duration_text)
	if err != nil {
		fmt.Printf("Error parsing time %v: %v\n", duration_text, err)
		os.Exit(1)
	}

	//TODO not taking advantage of decoupled time checking for concurrent scraping
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s, duration)
	}

}

func scrapeFeeds(s *State, duration time.Duration) {
	nextfeed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Error fetching next feed. Maybe there is no feed to scrape.")
		os.Exit(1)
	}
	if nextfeed.LastFetchedAt.Time.Add(duration).After(time.Now()) {
		fmt.Printf("No Feed to fetch yet.")
		os.Exit(1)
	}
	err = s.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID:        nextfeed.ID,
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		fmt.Printf("Error marking feed %v , Error: %v.\n", nextfeed.Url, err)
		os.Exit(1)
	}

	rss, err := rss.FetchFeed(context.Background(), nextfeed.Url.String)
	fmt.Printf("Fetched from %v\n", nextfeed.Name)
	if err != nil {
		fmt.Printf("Error retrieving RSS feed: %v\n", err)
		os.Exit(1)
	}

	for _, rssitem := range rss.Channel.Item {
		pubdate, err := parsePublishedAt(rssitem.PubDate)

		if err != nil {
			fmt.Printf("Error parsing time %v: %v\n", rssitem.PubDate, err)
			os.Exit(1)
		}
		_, err = s.DB.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			Title:       sql.NullString{String: rssitem.Title, Valid: true},
			Url:         sql.NullString{String: rssitem.Link, Valid: true},
			Description: sql.NullString{String: rssitem.Description, Valid: true},
			PublishedAt: sql.NullTime{Time: pubdate, Valid: true},
			FeedID:      uuid.NullUUID{UUID: nextfeed.ID, Valid: true},
		})
		if err != nil {
			fmt.Printf("Silenced Error couldnt insert post into dbms: %v\n", err)
		}

	}

	//printing rss
	fmt.Printf("Channel Title: %v\n", rss.Channel.Title)
	//fmt.Printf("Channel Description:\n%v\n",rss.Channel.Description)
	fmt.Printf("number of feeds fetched: %v\n", len(rss.Channel.Item))

}

var layouts = []string{
	time.RFC1123Z,
	time.RFC3339,
	"02-01-2006", // dd-mm-yyyy
	"01-02-2006", // mm-dd-yyyy
}

func parsePublishedAt(s string) (time.Time, error) {
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, err
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

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("expected arg 'url' but was not found")
	}
	url := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), sql.NullString{String: url, Valid: true})
	if err != nil {
		fmt.Printf("Error unfollowing, feed with url: %v does not exist\n", url)
		os.Exit(1)
	}

	err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		fmt.Printf("DB Error for unfollowing,\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%v successfully unfollowed %v\n", user.Name, feed.Name)
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

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) > 0 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			fmt.Printf("invalid limit %v: %v\nDefaulted to 2\n", cmd.Args[0], err)
			n = 2
		}
		if n > 20 {
			fmt.Printf("invalid limit %v too large\nDefaulted to 2\n", cmd.Args[0])
			n = 2
		}
		limit = n
	}
	posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		fmt.Printf("DB Error getting list of posts for user,\nError: %v\n", err)
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Printf("   -------------------- \n")
		fmt.Printf(color.YellowString("Title: %v"), "")
		fmt.Printf(color.GreenString("%v\n"), post.Title.String)
		fmt.Printf("	Published Date: %v\n", post.PublishedAt.Time)
		fmt.Printf("	URL: %v\n", post.Url.String)
		fmt.Printf("	Description: %v\n", post.Description.String)

	}
	fmt.Printf("   --- End of Posts --- \n")
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
