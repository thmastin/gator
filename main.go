package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
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
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
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
