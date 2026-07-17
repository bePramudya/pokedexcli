package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; i < 1; i = 0 {
		fmt.Print("Pokedex > ")

		scanner.Scan()

		input := scanner.Text()
		if len(input) == 0 {
			continue
		}

		inputs := strings.Fields(input)

		firstInput := inputs[0]
		secondInput := ""
		if len(inputs) > 1 {
			secondInput = inputs[1:][0]
		}

		command, exist := commands()[firstInput]
		if !exist {
			fmt.Print("Unknown command \n")
			continue
		}

		err := command.callback(&config, secondInput)
		if err != nil {
			fmt.Println(err)
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Invalid input: %s \n", err)
		}
	}
}

func commandExit(c *Config, val string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config, val string) error {
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

func commandMap(c *Config, val string) error {
	if c.Next == "" {
		return fmt.Errorf("map search reached limit")
	}

	err := helperMapsCommands(c, c.Next)
	if err != nil {
		return err
	}

	return nil
}

func commandMapBack(c *Config, val string) error {
	if c.Previous == "" {
		return fmt.Errorf("There's no previous map search")
	}

	err := helperMapsCommands(c, c.Previous)
	if err != nil {
		return err
	}

	return nil
}

func commandExplore(c *Config, val string) error {
	if val == "" {
		return fmt.Errorf("No location specified, command: explore <location>")
	}

	url := "https://pokeapi.co/api/v2/location-area/" + val

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode == 404 {
		return fmt.Errorf("location not found")
	} else if res.StatusCode > 299 {
		return fmt.Errorf("API response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}

	var locationDetail = RespLocationDetail{}
	err = json.Unmarshal(body, &locationDetail)
	if err != nil {
		return err
	}

	fmt.Print("Exploring ", locationDetail.Name, "... \n")
	fmt.Print("Found Pokemon: \n")

	for _, encounter := range locationDetail.PokemonEncounters {
		fmt.Print(encounter.Pokemon.Name, "\n")
	}

	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *Config, val string) error
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
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
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
