package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

}
