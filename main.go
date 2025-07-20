package main

import (
	"encoding/json"
	"fmt"

	"github.com/thmastin/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	err = config.SetUser(cfg, "tate")
	if err != nil {
		fmt.Println(err)
	}
	cfg, err = config.Read()
	if err != nil {
		fmt.Println(err)
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(b))
}
