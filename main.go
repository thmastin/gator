package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/thmastin/gator/internal/config"
	"github.com/thmastin/gator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println(err)
	}

	dbQueries := database.New(db)

	var currentState state
	currentState.db = dbQueries
	currentState.cfg = &cfg

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	cliInput := os.Args
	if len(cliInput) < 2 {
		fmt.Println("not enough arguments provided")
		os.Exit(1)
	}
	cliCommand := cliInput[1]
	cliArgs := cliInput[2:]

	userCommand := command{
		name:      cliCommand,
		arguments: cliArgs,
	}

	err = cmds.run(&currentState, userCommand)
	if err != nil {
		output := fmt.Sprintf("error running command: %v", err)
		fmt.Println(output)
		os.Exit(1)
	}

}

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("you must enter a username")
	}
	_, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		fmt.Println("user not found")
		os.Exit(1)
	}

	err = config.SetUser(*s.cfg, cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("you must enter a username")
	}

	name := cmd.arguments[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})

	if err != nil {
		fmt.Println("Error creating user:", err)
		os.Exit(1)
	}

	err = config.SetUser(*s.cfg, cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Println("User created successfully!")
	fmt.Printf("User data: %+v\n", user)

	return nil

}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Println("Error resetting database: ", err)
		os.Exit(1)
	}
	fmt.Println("Database reset")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users", err)
		os.Exit(1)
	}
	if len(users) < 1 {
		fmt.Print("No users found")
	} else {
		for _, user := range users {
			if user.Name == s.cfg.CurrentUserName {
				fmt.Println("*", user.Name, "(current)")
			} else {
				fmt.Println("*", user.Name)
			}
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("no time frame set")
	}
	time_between_requests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.arguments[0])
	ticker := time.NewTicker(time_between_requests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return errors.New("you must enter a name and url")
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
		Url:       cmd.arguments[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to create feed: %v", err)
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unalbe to follow feed: %v", err)
	}

	fmt.Println("Feed successfully created")
	fmt.Printf("Feed data: %+v\n", feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds: %v", err)
	}
	for i := range feeds {
		fmt.Printf("Name: %s URL: %s User: %s\n", feeds[i].Name, feeds[i].Url, feeds[i].UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("you must enter a url")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("unagle to get feed: %v", err)
	}
	feed_follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unalbe to follow feed: %v", err)
	}
	fmt.Println("Feed successfully followed")
	fmt.Printf("Feed Follow data: %+v\n", feed_follow)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("cannot get feeds that %s is following: %v", user.Name, err)
	}
	if len(follows) < 1 {
		fmt.Printf("User %s, is not following any feeds. Use the 'addfeed' command to add one\n", user.Name)
		return nil
	}
	for i := range follows {
		fmt.Println(follows[i].FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("you must enter a url to unfollow")
	}

	params := database.UnfollowFeedParams{Name: user.Name, Url: cmd.arguments[0]}
	err := s.db.UnfollowFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("cannot unfollow feed: %v", err)
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limitArg int32
	if len(cmd.arguments) < 1 {
		limitArg = 2
	} else {
		parsedInt, err := strconv.ParseInt(cmd.arguments[0], 10, 32)
		if err != nil {
			return errors.New("you must provide an whole number for the limit")
		} else {
			limitArg = int32(parsedInt)
		}

	}
	params := database.GetPostsForUserParams{ID: user.ID, Limit: limitArg}
	fmt.Printf("Getting posts for %s\n", user.Name)
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	if len(posts) < 1 {
		fmt.Printf("No posts to display for %s\n", user.Name)
	} else {
		for i := range len(posts) {
			fmt.Printf("Feed Name: %v\nTitle: %v\nPublished At: %v\nDescription: %v\nLink: %v\n\n", posts[i].FeedName, posts[i].Title, posts[i].PublishedAt, posts[i].Description.String, posts[i].Url)
		}
	}
	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	command, exists := c.handlers[cmd.name]
	if exists {
		err := command(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return errors.New("command does not exist")
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failure to create request: %v", err)
	}

	client := &http.Client{}
	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %v", err)
	}

	var newFeed RSSFeed
	err = xml.Unmarshal(body, &newFeed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	newFeed.Channel.Title = html.UnescapeString(newFeed.Channel.Title)
	newFeed.Channel.Description = html.UnescapeString(newFeed.Channel.Description)
	for i := range newFeed.Channel.Item {
		newFeed.Channel.Item[i].Title = html.UnescapeString(newFeed.Channel.Item[i].Title)
		newFeed.Channel.Item[i].Description = html.UnescapeString(newFeed.Channel.Item[i].Description)
	}

	return &newFeed, nil

}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	params := database.MarkFeedFetchedParams{LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: time.Now(), ID: feed.ID}
	err = s.db.MarkFeedFetched(context.Background(), params)
	if err != nil {
		return err
	}
	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("\nFetching items from: %v\n\n", feed.Name)
	for i := range fetchedFeed.Channel.Item {
		item := fetchedFeed.Channel.Item[i]
		var descriptionParam sql.NullString
		if item.Description != "" {
			descriptionParam = sql.NullString{String: item.Description, Valid: true}
		} else {
			descriptionParam = sql.NullString{Valid: false}
		}
		var publshedAtParam time.Time
		if item.PubDate != "" {
			parsedTime, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err != nil {
				publshedAtParam = time.Time{}
			} else {
				publshedAtParam = parsedTime
			}
		}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: descriptionParam,
			PublishedAt: publshedAtParam,
			FeedID:      feed.ID,
		}
		post, err := s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			return err
		}
		fmt.Printf("New Post: \n%+v\n", post)

	}
	return nil

}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}
