package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/thmastin/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	var currentState state
	currentState.pointer = &cfg

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)

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
	pointer *config.Config
}

type command struct {
	name      string
	arguments []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("you must enter a username")
	}
	err := config.SetUser(*s.pointer, cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
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
