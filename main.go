package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	file, err := os.Open("token.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	token := strings.TrimSpace(scanner.Text()) // remove spaces and new lines

	fmt.Println("Token loaded: ", token[:5]+"...") // show first few chars

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating discord session:", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	var guessers = make(map[string]int)

	var guessersWaitingContinue = make(map[string]bool)

	var playingRPS = make(map[string]string)

	var rpsWaitingContinue = make(map[string]bool)

	var weatherDescriptions = map[int]string{
		0:  "Clear sky ☀️",
		1:  "Mainly clear 🌤️",
		2:  "Partly cloudy ⛅",
		3:  "Overcast ☁️",
		45: "Fog 🌫️",
		48: "Depositing rime fog 🌫️❄️",
		51: "Light drizzle 🌦️",
		53: "Moderate drizzle 🌧️",
		55: "Dense drizzle 🌧️💧",
		56: "Freezing light drizzle 🌧️❄️",
		57: "Freezing dense drizzle 🌧️🧊",
		61: "Slight rain 🌦️",
		63: "Moderate rain 🌧️",
		65: "Heavy rain 🌧️💦",
		66: "Freezing light rain 🌧️❄️",
		67: "Freezing heavy rain 🌧️🧊",
		71: "Slight snow fall 🌨️",
		73: "Moderate snow fall 🌨️❄️",
		75: "Heavy snow fall ❄️🌨️",
		77: "Snow grains ❄️🌾",
		80: "Slight rain showers 🌦️",
		81: "Moderate rain showers 🌧️☔",
		82: "Violent rain showers 🌧️🌪️",
		85: "Slight snow showers 🌨️",
		86: "Heavy snow showers 🌨️❄️",
		95: "Thunderstorm ⛈️",
		96: "Thunderstorm with slight hail ⛈️🧊",
		99: "Thunderstorm with heavy hail ⛈️🌩️🧊",
	}

	var cmdList = []string{
		"!ping - Sends back pong!",
		"!dice - Roll a six sided die!",
		"!coin - Flip a coin!",
		"!roulette - Put it all on red or black!",
		"!slot - A slot machine with emojis, can you get all 3?",
		"!guess - Guess a number between 1-100! and !cancel to cancel the guessing game.",
		"!rps - Play Rock, Paper, Scissors with the bot! !cancel to cancel the RPS game.",
		"!weather <cityname> - Get the weather and more info about a specific city. e.g (!weather San Francisco)",
		"!time <cityname> - Get the time and more info about a specific city. e.g (!time Detroit) disclaimer: may not work with certain cities as not all cities are tracked.",
		"!define <word> - Get the definition, and more info about a specific word. e.g (!define gravity)",
	}

	// Add message handler
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		userMention := fmt.Sprintf("<@%s>", m.Author.ID)

		// check if the user is currently guessing a num
		if target, ok := guessers[m.Author.ID]; ok {
			// check if use is waiting for continue
			if guessersWaitingContinue[m.Author.ID] {
				if m.Content == "y" {
					// restart game
					target := rand.Intn(100) + 1
					guessers[m.Author.ID] = target
					rpsWaitingContinue[m.Author.ID] = false
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Guess a Number between 1 and 100!", userMention))
				} else if m.Content == "n" || m.Content == "!cancel" {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Game over! Thanks for playing.", userMention))
					delete(guessers, m.Author.ID)
					delete(guessersWaitingContinue, m.Author.ID)
					return
				} else {
					return
				}
			}

			// cancel if yes
			if m.Content == "!cancel" {
				delete(guessers, m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Your guessing game has been cancelled.", userMention))
				return
			}

			// convert their message to a number
			guess, err := strconv.Atoi(m.Content)
			if err != nil {
				return
			}

			if guess > target {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Too high, try again!", userMention))
			} else if guess < target {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Too low, try again!", userMention))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Good job! You guessed it! The number was %d. Would you like to try again? (y/n)", userMention, target))
				guessersWaitingContinue[m.Author.ID] = true
			}
			return
		}

		// check if the user is playing rps
		if botChoice, ok := playingRPS[m.Author.ID]; ok {
			// waiting for continue?
			if rpsWaitingContinue[m.Author.ID] {
				if m.Content == "y" {
					// restart game
					options := []string{"Rock", "Paper", "Scissors"}
					botChoice = options[rand.Intn(len(options))]
					playingRPS[m.Author.ID] = botChoice
					rpsWaitingContinue[m.Author.ID] = false
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, I picked again! Your turn.", userMention))
					return
				} else if m.Content == "n" || m.Content == "!cancel" {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Game over! Thanks for playing.", userMention))
					delete(playingRPS, m.Author.ID)
					delete(rpsWaitingContinue, m.Author.ID)
					return
				} else {
					return
				}
			}

			if m.Content == "!cancel" {
				delete(playingRPS, m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Your RPS game has been cancelled.", userMention))
				return
			}

			userChoice := m.Content

			if userChoice == "r" || userChoice == "rock" || userChoice == "Rock" {
				userChoice = "Rock"
			} else if userChoice == "p" || userChoice == "paper" || userChoice == "Paper" {
				userChoice = "Paper"
			} else if userChoice == "s" || userChoice == "scissors" || userChoice == "Scissors" {
				userChoice = "Scissors"
			} else {
				return
			}

			if userChoice == botChoice {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, It was a tie! Both of us picked %s. Would you like to continue? (y/n)", userMention, botChoice))
				rpsWaitingContinue[m.Author.ID] = true
			} else if (userChoice == "Rock" && botChoice == "Scissors") || (userChoice == "Paper" && botChoice == "Rock") || (userChoice == "Scissors" && botChoice == "Paper") {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, You win! You chose %s and I chose %s! Would you like to continue? (y/n)", userMention, userChoice, botChoice))
				rpsWaitingContinue[m.Author.ID] = true
			} else if (userChoice == "Rock" && botChoice == "Paper") || (userChoice == "Paper" && botChoice == "Scissors") || (userChoice == "Scissors" && botChoice == "Rock") {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s You lose! You chose %s and I chose %s! Would you like to continue? (y/n)", userMention, userChoice, botChoice))
				rpsWaitingContinue[m.Author.ID] = true
			} else {
				return
			}
			return
		}

		if strings.HasPrefix(m.Content, "!weather") {
			if m.Content == "!cancel" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Your weather request has been cancelled.", userMention))
				return
			}

			city := strings.TrimSpace(strings.TrimPrefix(m.Content, "!weather"))
			if city == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, please provide a city name. Example: !weather San Francisco", userMention))
				return
			}

			// call api
			resp, err := http.Get("https://weathery-service.onrender.com/weather?city=" + url.QueryEscape(city))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to fetch weather! Error: %v", userMention, err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Error: %s", userMention, string(body)))
				return
			}

			// json struct
			var data struct {
				Temperature   float64 `json:"temperature"`
				WeatherCode   int     `json:"weather_code"`
				Humidity      float64 `json:"humidity"`
				Precipitation float64 `json:"precipitation"`
				WindSpeed     float64 `json:"windSpeed"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to parse weather data", userMention))
				return
			}

			description := weatherDescriptions[data.WeatherCode]
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Weather in **%s**:\n🌡️ %.1f °C\n🌧️ %.1f mm rain\n💧 %.0f%% humidity\n💨 %.1f km/h wind\n☁️ %s",
				userMention,
				city,
				data.Temperature,
				data.Precipitation,
				data.Humidity,
				data.WindSpeed,
				description))

		} else if strings.HasPrefix(m.Content, "!time") {
			if m.Content == "!cancel" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Your time request has been cancelled.", userMention))
				return
			}

			city := strings.TrimSpace(strings.TrimPrefix(m.Content, "!time"))
			if city == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a city name. Example: !time San Francisco", userMention))
				return
			}

			// call api
			resp, err := http.Get("https://clickclock-service.onrender.com/time?city=" + url.QueryEscape(city))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to fetch time! Error: %v", userMention, err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Error: %s", userMention, string(body)))
				return
			}

			// json struct
			var data struct {
				Time      string `json:"time"`         // HH:MM
				Timezone  string `json:"timezone"`     // e.g. Asia/Dubai
				UTCOffset string `json:"utc_offset"`   // e.g. +04:00
				ISO       string `json:"iso_datetime"` // full ISO timestamp
				Date      string `json:"Date"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to parse time data", userMention))
				return
			}

			cityData := strings.Split(data.Timezone, "/")[1]
			regionData := strings.Split(data.Timezone, "/")[0]
			timezone := "UTC" + data.UTCOffset

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Time in **%s**: \n Time: %s \n Timezone: %s \n Region: %s \n Date: %s",
				userMention,
				cityData,
				data.Time,
				timezone,
				regionData,
				data.Date))
		} else if strings.HasPrefix(m.Content, "!define") {
			if m.Content == "!cancel" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Your definition request has been cancelled.", userMention))
				return
			}

			word := strings.TrimSpace(strings.TrimPrefix(m.Content, "!define"))
			if word == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a valid word. Example: !define gravity", userMention))
				return
			}

			// call api
			resp, err := http.Get("https://easydefine-service.onrender.com/define?word=" + url.QueryEscape(word))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to fetch definition! Error: %v", userMention, err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Error: %s", userMention, string(body)))
				return
			}

			// json struct
			var data struct {
				Word     string `json:"word"`
				Meanings []struct {
					PartOfSpeech string   `json:"partOfSpeech"`
					Definitions  []string `json:"definitions"`
					Synonyms     []string `json:"synonyms"`
				} `json:"meanings"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to parse definition data", userMention))
				return
			}

			// build the message
			if len(data.Meanings) == 0 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, No definitions found for '%s'.", userMention, data.Word))
				return
			}

			msg := fmt.Sprintf("%s, **%s**\n", userMention, data.Word)
			for _, meaning := range data.Meanings {
				msg += fmt.Sprintf("_%s_\n", meaning.PartOfSpeech)
				for i, def := range meaning.Definitions {
					if i > 2 {
						msg += "...\n"
						break
					}
					msg += fmt.Sprintf("• %s\n", def)
				}
				if len(meaning.Synonyms) > 0 {
					msg += fmt.Sprintf("**Synonyms:** %s\n", strings.Join(meaning.Synonyms, ", "))
				}
				msg += "\n"
			}
			s.ChannelMessageSend(m.ChannelID, msg)
		}

		// handle commands
		switch m.Content {
		case "!ping":
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Pong!", userMention))
		case "!dice":
			num := rand.Intn(6) + 1
			str := strconv.Itoa(num)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Rolled %s!", userMention, str))
		case "!coin":
			options := []string{"Heads", "Tails"}
			index := rand.Intn(len(options))
			choice := options[index]
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s!", userMention, choice))
		case "!roulette":
			options := []string{"Red", "Black"}
			index := rand.Intn(len(options))
			choice := options[index]
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s!", userMention, choice))
		case "!slot":
			options := []string{"💀", "🧌", "🤠", "🤓", "👄", "🧏‍♂️"}
			index1 := rand.Intn(len(options))
			index2 := rand.Intn(len(options))
			index3 := rand.Intn(len(options))
			choice1 := options[index1]
			choice2 := options[index2]
			choice3 := options[index3]

			if choice1 == choice2 && choice2 == choice3 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s %s %s", userMention, choice1, choice2, choice3))
				s.ChannelMessageSend(m.ChannelID, "You Win!!")
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s %s %s", userMention, choice1, choice2, choice3))
			}
		case "!guess":
			target := rand.Intn(100) + 1
			guessers[m.Author.ID] = target
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Guess a Number between 1 and 100!", userMention))
		case "!rps":
			options := []string{"Rock", "Paper", "Scissors"}
			index := rand.Intn(len(options))
			botChoice := options[index]
			playingRPS[m.Author.ID] = botChoice
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, I have picked. Your turn!", userMention))
		case "!help":
			commands := strings.Join(cmdList, "\n")
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, here are all the currently available commands: \n%s", userMention, commands))
		}
	})

	// Open connection
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is running! Press CTRL+C to exit.")

	// Wait for exit
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	dg.Close()
}
