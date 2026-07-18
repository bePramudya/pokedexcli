package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	commands := commands()

	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "" {
			continue
		}

		words := strings.Fields(input)
		cmdName := words[0]
		arg := ""
		if len(words) > 1 {
			arg = words[1]
		}

		cmd, ok := commands[cmdName]
		if !ok {
			fmt.Print("Unknown command \n")
			continue
		}

		if err := cmd.callback(&config, arg); err != nil {
			fmt.Println(err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Invalid input: %s \n", err)
	}
}

func commandExit(c *Config, arg string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config, arg string) error {
	fmt.Print("Welcome to the Pokedex!\n")
	fmt.Print("Usage:\n\n")

	for _, reg := range commands() {
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

	var areas RespSearchLocation
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

func commandMap(c *Config, arg string) error {
	if c.Next == "" {
		return fmt.Errorf("map search reached limit")
	}

	err := helperMapsCommands(c, c.Next)
	if err != nil {
		return err
	}

	return nil
}

func commandMapBack(c *Config, arg string) error {
	if c.Previous == "" {
		return fmt.Errorf("There's no previous map search")
	}

	err := helperMapsCommands(c, c.Previous)
	if err != nil {
		return err
	}

	return nil
}

func commandExplore(c *Config, arg string) error {
	if arg == "" {
		return fmt.Errorf("No location specified, command: explore <location>\n")
	}

	url := "https://pokeapi.co/api/v2/location-area/" + arg

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode == 404 {
		return fmt.Errorf("location not found\n")
	} else if res.StatusCode > 299 {
		return fmt.Errorf("API response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}

	// var locationDetail = RespLocationDetail{}
	err = json.Unmarshal(body, &locationDetail)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", locationDetail.Name)
	fmt.Print("Found Pokemon:\n")

	for _, encounter := range locationDetail.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(c *Config, arg string) error {
	if arg == "" {
		return fmt.Errorf("what's you want to catch?")
	}

	// pokemonIsPresent := false
	// for _, encounter := range locationDetail.PokemonEncounters {
	// 	fmt.Print(encounter.Pokemon.Name)
	// 	if encounter.Pokemon.Name == arg {
	// 		pokemonIsPresent = true
	// 	}
	// }

	// if !pokemonIsPresent {
	// 	return fmt.Errorf("Pokemon specified is not in the area")
	// }

	url := "https://pokeapi.co/api/v2/pokemon/" + arg

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Problem fetching pokemon data")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode == 404 {
		return fmt.Errorf("pokemon not found")
	} else if res.StatusCode > 299 {
		return fmt.Errorf("API response failed with status code: %d and \nbody: %s", res.StatusCode, body)
	}

	var pokemon = Pokemon{}
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)

	if rand.IntN(pokemon.BaseExperience) == 0 {
		fmt.Print(pokemon.Name, " was caught!\n")
	} else {
		fmt.Print(pokemon.Name, " escaped!\n")
		return nil
	}

	pokedex[pokemon.Name] = pokemon
	fmt.Print("You may now inspect it with the inspect command.\n")

	return nil
}

func commandInspect(c *Config, arg string) error {
	if arg == "" {
		return fmt.Errorf("you have not caught that pokemon")
	}

	fmt.Printf("Name: %s\n", pokedex[arg].Name)
	fmt.Printf("Height: %d\n", pokedex[arg].Height)
	fmt.Printf("Weight: %d\n", pokedex[arg].Weight)

	fmt.Print("Stats:\n")
	for _, stat := range pokedex[arg].Stats {
		fmt.Printf(" -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Print("Types:\n")
	for _, t := range pokedex[arg].Types {
		fmt.Printf(" -%s \n", t.Type.Name)
	}

	return nil
}

func commandPokedex(c *Config, arg string) error {
	fmt.Print("Your Pokedex: \n")

	for _, pokemon := range pokedex {
		fmt.Printf(" - %s \n", pokemon.Name)
	}

	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *Config, arg string) error
}

func commands() map[string]cliCommand {
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
		"explore": {
			name:        "explore",
			description: "explore a region",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "try to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "check pokemon details",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "check pokedex",
			callback:    commandPokedex,
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

type RespSearchLocation struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type RespLocationDetail struct {
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int           `json:"chance"`
				ConditionValues []interface{} `json:"condition_values"`
				MaxLevel        int           `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

var locationDetail = RespLocationDetail{}

type Pokemon struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Stats  []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	BaseExperience int `json:"base_experience"`
}

var pokedex = make(map[string]Pokemon)

func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words
}
