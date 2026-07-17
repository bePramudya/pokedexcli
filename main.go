package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; i < 1; i = 0 {
		fmt.Print("Pokedex > ")

		scanner.Scan()

		words := scanner.Text()

		if len(words) == 0 {
			continue
		}

		command, exist := registry()[words]

		if !exist {
			fmt.Print("Unknown command \n")
			continue
		}

		err := command.callback(&config)
		if err != nil {
			fmt.Println(err)
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Invalid input: %s \n", err)
		}
	}
}

func commandExit(c *Config) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config) error {
	fmt.Print("Welcome to the Pokedex!\n")
	fmt.Print("Usage:\n\n")

	for _, reg := range registry() {
		fmt.Print(reg.name, ": ", reg.description, "\n")
	}

	return nil
}

func helperMapsCommands(c *Config, target string) error {
	res, err := http.Get(target)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		fmt.Printf("API response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
		return err
	}

	var areas LocationAreas
	err = json.Unmarshal(body, &areas)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, area := range areas.Results {
		fmt.Print(area.Name, "\n")
	}

	c.Previous = areas.Previous
	c.Next = areas.Next
	return nil
}

func commandMap(c *Config) error {
	if c.Next == "" {
		return fmt.Errorf("map search reached limit")
	}

	err := helperMapsCommands(c, c.Next)
	if err != nil {
		return err
	}

	return nil
}

func commandMapBack(c *Config) error {
	if c.Previous == "" {
		return fmt.Errorf("There's no previous map search")
	}

	err := helperMapsCommands(c, c.Previous)
	if err != nil {
		return err
	}

	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *Config) error
}

func registry() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Display explorable map",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous map search",
			callback:    commandMapBack,
		},
	}
}

type Config struct {
	Next     string
	Previous string
}

var config = Config{
	Next:     "https://pokeapi.co/api/v2/location-area/",
	Previous: "",
}

type LocationAreas struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}
