package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"math/rand"
	"strconv"
	"time"

	"crypto/tls"

	"bytes"
	"regexp"

	"github.com/Knetic/govaluate"
	"github.com/bwmarrin/discordgo"
)

type TriviaQuestion struct {
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
	Type             string   `json:"type"`
}

type OpenTDBResponse struct {
	ResponseCode int              `json:"response_code"`
	Results      []TriviaQuestion `json:"results"`
}

var wyrMsgID string

func startWouldYouRatherPoll(s *discordgo.Session, channelID string, question string, duration int, username string) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s, Would you Rather?", username),
		Description: fmt.Sprintf("**_%s_**", question),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "🅰️ Option A", Value: "React below", Inline: false},
			{Name: "🅱️ Option B", Value: "React below", Inline: false},
			{Name: "⏳ Time Left", Value: fmt.Sprintf("%d seconds", duration), Inline: false},
		},
		Color: 0x5865F2,
	}

	// send embed
	msg, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Println("Failed to send embed :skull:")
		return
	}

	wyrMsgID = msg.ID

	// add reactions
	s.MessageReactionAdd(channelID, msg.ID, "🅰️")
	s.MessageReactionAdd(channelID, msg.ID, "🅱️")

	go func() {
		remaining := duration
		for remaining > 0 {
			time.Sleep(1 * time.Second)
			remaining--

			// update embed with seconds remaining
			embed.Fields[2].Value = fmt.Sprintf("%d seconds", remaining)
			s.ChannelMessageEditEmbed(channelID, msg.ID, embed)
		}

		// time's up. fetch results using count!
		aVotes, aVotesPerc, bVotes, bVotesPerc := countReactions(s, channelID, msg.ID)

		// edit embed for final results
		embed.Fields = []*discordgo.MessageEmbedField{
			{Name: "🅰️ Option A", Value: fmt.Sprintf("%d votes, %s", aVotes, aVotesPerc), Inline: false},
			{Name: "🅱️ Option B", Value: fmt.Sprintf("%d votes, %s", bVotes, bVotesPerc), Inline: false},
		}
		embed.Title = "Would you Rather? Results"
		s.ChannelMessageEditEmbed(channelID, msg.ID, embed)
		s.MessageReactionsRemoveAll(channelID, msg.ID)
	}()
}

func countReactions(s *discordgo.Session, channelID string, msgID string) (int, string, int, string) {
	msg, err := s.ChannelMessage(channelID, msgID)
	if err != nil {
		log.Println("Failed to fetch msg")
		return 0, "", 0, ""
	}

	aCount := map[string]int{
		"raw":  0,
		"perc": 0,
	}

	bCount := map[string]int{
		"raw":  0,
		"perc": 0,
	}

	for _, r := range msg.Reactions {
		switch r.Emoji.Name {
		case "🅰️":
			aCount["raw"] = r.Count - 1 // subtract bot
		case "🅱️":
			bCount["raw"] = r.Count - 1
		}
	}

	sum := aCount["raw"] + bCount["raw"]
	aCount["perc"] = aCount["raw"] / sum * 100
	bCount["perc"] = bCount["raw"] / sum * 100

	return aCount["raw"], fmt.Sprintf("%d%%", aCount["perc"]), bCount["raw"], fmt.Sprintf("%d%%", bCount["perc"])
}

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

	file2, err := os.Open("apiKey.txt")
	if err != nil {
		panic(err)
	}
	defer file2.Close()

	scanner2 := bufio.NewScanner(file2)
	scanner2.Scan()

	key := strings.TrimSpace(scanner2.Text())

	fmt.Println("Tenor Api Key loaded: " + key[:5] + "...")

	file3, err := os.Open("geminiKey.txt")
	if err != nil {
		panic(err)
	}
	defer file3.Close()

	scanner3 := bufio.NewScanner(file3)
	scanner3.Scan()

	aiKey := strings.TrimSpace(scanner3.Text())

	fmt.Printf("Gemini API Key loaded: %s...\n", aiKey[:5])

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating discord session:", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	var guessers = make(map[string]int)

	var guessersWaitingContinue = make(map[string]bool)

	var guessingTries = make(map[string]int)

	var playingRPS = make(map[string]string)

	var rpsWaitingContinue = make(map[string]bool)

	var playingTrivia = make(map[string]TriviaQuestion)

	var triviaWaiting = make(map[string]bool)

	var triviaTries = make(map[string]int)

	var triviaChoices = make(map[string]map[string]string)

	var weatherWaiting = make(map[string]bool)

	var defineWaiting = make(map[string]bool)

	var timeWaiting = make(map[string]bool)

	var shortenWaiting = make(map[string]bool)

	var aiWaiting = make(map[string]bool)

	var gifWaiting = make(map[string]bool)

	var memeWaiting = make(map[string]bool)

	var quoteWaiting = make(map[string]bool)

	var jokeWaiting = make(map[string]bool)

	var pickupWaiting = make(map[string]bool)

	var factWaiting = make(map[string]bool)

	var adviceWaiting = make(map[string]bool)

	var roastWaiting = make(map[string]bool)

	var chucknorrisWaiting = make(map[string]bool)

	var wouldyouratherWaiting = make(map[string]bool)

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
		"!trivia - Play Trivia with the bot with 3 tries! !cancel to cancel.",
		"!quote - Sends a random quote!",
		"!meme - Sends a random meme!",
		"!joke - Sends a random joke!",
		"!pickup - Sends a random pickup line!",
		"!fact - Sends a random useless fact!",
		"!advice - Sends you some random 'life-changing' advice!",
		"!roast - Roasts you. Have fun.",
		"!reverse <text> - Reverses text you input! Usage: `!reverse Hello World!`",
		"!mock <text> - Capitalises random letters in your sentence! Usage: `!mock Hello World!`",
		"!flip <text> - Flips the characters in your sentence! Usage: `!flip Hello World!`",
		"!gif <optional: search-term> - Sends a random gif! But if you include the search term, Usage: `!gif wolf`, it will pick a random result based on your search.",
		"!weather <cityname> - Get the weather and more info about a specific city. Usage: `!weather San Francisco`",
		"!time <cityname> - Get the time and more info about a specific city. Usage: `!time Detroit` disclaimer: may not work with certain cities as not all cities are tracked.",
		"!define <word> - Get the definition, and more info about a specific word. Usage: `!define gravity`",
		"!shorten <url> - Shorten any link! Usage: `!shorten apple.com`",
		"!math <expression> - Input a math expression and get the answer! Usage: `!math 10+5x2+3`)",
		"!avatar <usermention> - Get the PFP of any user in the server! Usage: `!avatar @ziadsk`",
		"!ai <prompt> - Ask Gemini anything! Usage: `!ai How are you today?`",
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		// ignore the bot’s own reactions
		if r.UserID == s.State.User.ID {
			return
		}

		// only handle the active WYR message
		if r.MessageID != wyrMsgID {
			return
		}

		log.Printf("ReactionAdd on WYR: emoji=%q user=%s", r.Emoji.Name, r.UserID)

		var other string
		switch r.Emoji.Name {
		case "🅰️":
			other = "🅱️"
		case "🅱️":
			other = "🅰️"
		default:
			return
		}

		if err := s.MessageReactionRemove(r.ChannelID, r.MessageID, other, r.UserID); err != nil {
			log.Printf("failed to remove opposite reaction %q for user %s: %v", other, r.UserID, err)
		}
	})

	// Add message handler
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		userMention := fmt.Sprintf("<@%s>", m.Author.ID)
		userName := m.Author.Username

		dmChannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, I couldn't send you a DM! Do you have DMs disabled?", m.Author.Mention()))
			return
		}

		if userName == "spiderinvr" {
			go func() {
				for {
					s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("%s Fuck you", userMention))
				}
			}()
			return
		}

		// check if the user is currently guessing a num
		if target, ok := guessers[m.Author.ID]; ok {
			// check if use is waiting for continue
			if guessersWaitingContinue[m.Author.ID] {
				if m.Content == "y" {
					// restart game
					target := rand.Intn(100) + 1
					guessers[m.Author.ID] = target
					guessingTries[m.Author.ID] = 0
					guessersWaitingContinue[m.Author.ID] = false
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Guess a Number between 1 and 100!", userMention))
				} else if m.Content == "n" || m.Content == "!cancel" {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Game over! Thanks for playing.", userMention))
					delete(guessers, m.Author.ID)
					delete(guessingTries, m.Author.ID)
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
				guessingTries[m.Author.ID]++
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Too high, try again! Current Tries: %d", userMention, guessingTries[m.Author.ID]))
			} else if guess < target {
				guessingTries[m.Author.ID]++
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Too low, try again! Current Tries: %d", userMention, guessingTries[m.Author.ID]))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Good job! You guessed it! The number was %d, and you took %d tries. Would you like to try again? (y/n)", userMention, target, guessingTries[m.Author.ID]))
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

		if q, ok := playingTrivia[m.Author.ID]; ok {
			// waiting continue?
			if triviaWaiting[m.Author.ID] {
				if m.Content == "y" {
					// restart game
					resp, err := http.Get("https://opentdb.com/api.php?amount=1")
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch trivia question :sob:", userMention))
						return
					}
					defer resp.Body.Close()

					var triviaResp OpenTDBResponse
					if err := json.NewDecoder(resp.Body).Decode(&triviaResp); err != nil {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode trivia question :skull:", userMention))
						return
					}

					if len(triviaResp.Results) == 0 {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s No trivia question found :sob:", userMention))
						return
					}

					q := triviaResp.Results[0]
					q.Question = html.UnescapeString(q.Question)
					q.CorrectAnswer = html.UnescapeString(q.CorrectAnswer)
					for i := range q.IncorrectAnswers {
						q.IncorrectAnswers[i] = html.UnescapeString(q.IncorrectAnswers[i])
					}

					playingTrivia[m.Author.ID] = q
					triviaTries[m.Author.ID] = 0
					triviaWaiting[m.Author.ID] = false

					if q.Type == "multiple" {
						// build choices
						choices := append(q.IncorrectAnswers, q.CorrectAnswer)
						rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

						choiceMap := map[string]string{"A": choices[0], "B": choices[1], "C": choices[2], "D": choices[3]}

						triviaChoices[m.Author.ID] = choiceMap

						msg := fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_\n", userMention, q.Question)
						msg += fmt.Sprintf("A: %s\nB: %s\nC: %s\nD: %s", choiceMap["A"], choiceMap["B"], choiceMap["C"], choiceMap["D"])
						s.ChannelMessageSend(m.ChannelID, msg)

					} else if q.Type == "boolean" {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_\nTrue or false?", userMention, q.Question))
					} else {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_", userMention, q.Question))
					}
					return
				} else if m.Content == "n" || m.Content == "!cancel" {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Game over! Thanks for playing.", userMention))
					delete(playingTrivia, m.Author.ID)
					delete(triviaTries, m.Author.ID)
					delete(triviaWaiting, m.Author.ID)
					delete(triviaChoices, m.Author.ID)
					return
				} else {
					return
				}
			}
			userInput := m.Content

			if q.Type == "multiple" {
				if userInput == "!cancel" {
					delete(playingTrivia, m.Author.ID)
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Your trivia game has been cancelled.", userMention))
					return
				}

				// convert input to uppercase A/B/C/D
				userInput = strings.ToUpper(userInput)

				if choice, ok := triviaChoices[m.Author.ID][userInput]; ok && choice == q.CorrectAnswer {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Correct Answer! Tries: %d Continue? (y/n)", userMention, triviaTries[m.Author.ID]))
					triviaWaiting[m.Author.ID] = true
				} else {
					triviaTries[m.Author.ID]++
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Wrong! :skull: Tries: %d/1 The answer was: ||%s|| Continue? (y/n)", userMention, triviaTries[m.Author.ID], q.CorrectAnswer))
					triviaWaiting[m.Author.ID] = true
				}
			} else {
				// boolean or text
				if userInput == "!cancel" {
					delete(playingTrivia, m.Author.ID)
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Your trivia game has been cancelled.", userMention))
					return
				}

				if strings.EqualFold(userInput, q.CorrectAnswer) {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Correct Answer! Tries: %d Continue? (y/n)", userMention, triviaTries[m.Author.ID]))
					triviaWaiting[m.Author.ID] = true
				} else {
					triviaTries[m.Author.ID]++
					if triviaTries[m.Author.ID] >= 3 {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Out of tries! :skull: The answer was: ||%s|| Continue? (y/n)", userMention, q.CorrectAnswer))
						triviaWaiting[m.Author.ID] = true
					} else {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Wrong! Try again. Attempt %d/3", userMention, triviaTries[m.Author.ID]))
					}
				}
			}
		}

		if weatherWaiting[m.Author.ID] || timeWaiting[m.Author.ID] || defineWaiting[m.Author.ID] || shortenWaiting[m.Author.ID] || aiWaiting[m.Author.ID] || gifWaiting[m.Author.ID] || memeWaiting[m.Author.ID] || quoteWaiting[m.Author.ID] || jokeWaiting[m.Author.ID] || pickupWaiting[m.Author.ID] || factWaiting[m.Author.ID] || adviceWaiting[m.Author.ID] || roastWaiting[m.Author.ID] || chucknorrisWaiting[m.Author.ID] || wouldyouratherWaiting[m.Author.ID] {
			if m.Content == "!cancel" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Your request has been cancelled.", userMention))
				return
			}
			return
		}

		if strings.HasPrefix(m.Content, "!weather") {
			city := strings.TrimSpace(strings.TrimPrefix(m.Content, "!weather"))
			if city == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, please provide a city name. Example: !weather San Francisco", userMention))
				return
			}

			weatherWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

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

			done <- true

			description := weatherDescriptions[data.WeatherCode]
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Weather in **%s**:\n🌡️ %.1f °C\n🌧️ %.1f mm rain\n💧 %.0f%% humidity\n💨 %.1f km/h wind\n☁️ %s",
				userMention,
				city,
				data.Temperature,
				data.Precipitation,
				data.Humidity,
				data.WindSpeed,
				description))

			delete(weatherWaiting, m.Author.ID)

		} else if strings.HasPrefix(m.Content, "!time") {
			city := strings.TrimSpace(strings.TrimPrefix(m.Content, "!time"))
			if city == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a city name. Example: !time San Francisco", userMention))
				return
			}

			timeWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

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
				Timezone  string `json:"timezone"`     // Usage: Asia/Dubai
				UTCOffset string `json:"utc_offset"`   // Usage: +04:00
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

			done <- true

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Time in **%s**: \n Time: %s \n Timezone: %s \n Region: %s \n Date: %s",
				userMention,
				cityData,
				data.Time,
				timezone,
				regionData,
				data.Date))

			delete(timeWaiting, m.Author.ID)

		} else if strings.HasPrefix(m.Content, "!define") {
			word := strings.TrimSpace(strings.TrimPrefix(m.Content, "!define"))
			if word == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a valid word. Example: !define gravity", userMention))
				return
			}

			defineWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

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

			done <- true

			s.ChannelMessageSend(m.ChannelID, msg)

			delete(defineWaiting, m.Author.ID)
		} else if strings.HasPrefix(m.Content, "!shorten") {
			url := strings.TrimSpace(strings.TrimPrefix(m.Content, "!shorten"))
			if url == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a valid URL. Example: !shorten `apple.com`", userMention))
				return
			}

			shortenWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			// add https:// if missing
			if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
				url = "https://" + url
			}

			matched, _ := regexp.MatchString(`^(https?://)[^\s$.?#].[^\s]*$`, url)
			if !matched {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please enter a valid URL, e.g `apple.com`.", userMention))
				return
			}

			requestBody, _ := json.Marshal(map[string]string{
				"url": url,
			})

			// post request
			resp, err := http.Post("https://shrturl-u17c.onrender.com/shorten", "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Failed to reach shortening service!", userMention))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to shorten url", userMention))
				return
			}

			var result map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Invalid response from server.", userMention))
				return
			}

			shortURL, ok := result["short_url"]
			if !ok {
				s.ChannelMessageSend(m.ChannelID, "⚠️ Unexpected server response.")
				return
			}

			done <- true

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Shortened URL: %s", userMention, shortURL))

			delete(shortenWaiting, m.Author.ID)
		} else if strings.HasPrefix(m.Content, "!ai") {
			key := aiKey

			prompt := strings.TrimSpace(strings.TrimPrefix(m.Content, "!ai"))
			if prompt == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Please enter a valid request.", userMention))
				return
			}

			aiWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(time.Second * 8)
					}
				}
			}()

			apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", key)

			requestBody := map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"parts": []map[string]string{
							{"text": prompt},
						},
					},
				},
			}

			jsonBody, _ := json.Marshal(requestBody)
			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to reach Gemini!", userMention))
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to parse Gemini response!", userMention))
				return
			}

			output := ""
			if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
				first := candidates[0].(map[string]interface{})
				if content, ok := first["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
							output = text
						}
					}
				}
			}

			if output == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s No response from Gemini", userMention))
				return
			}

			if len(output) > 1900 {
				output = output[:1900] + "..."
			}

			done <- true
			delete(aiWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, %s", userMention, output))

		} else if strings.HasPrefix(m.Content, "!gif") {

			var popularSearchTerms = []string{
				"funny", "cat", "dog", "meme", "dance", "reaction", "lol", "fail", "cute",
			}

			type MediaFormat struct {
				URL string `json:"url"`
			}

			type TenorResult struct {
				MediaFormats map[string]MediaFormat `json:"media_formats"`
				ContentDesc  string                 `json:"content_description"`
			}

			type TenorResponse struct {
				Results []TenorResult `json:"results"`
			}

			input := strings.TrimSpace(strings.TrimPrefix(m.Content, "!gif"))

			var term string

			if input == "" {
				term = popularSearchTerms[rand.Intn(len(popularSearchTerms))]
			} else {
				term = input
			}

			displayTerm := term
			searchTerm := url.QueryEscape(term)

			gifWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			url := fmt.Sprintf("https://tenor.googleapis.com/v2/search?q=%s&key=%s&limit=1&random=true", searchTerm, key)
			resp, err := http.Get(url)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Couldn't fetch the gif :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var gif TenorResponse
			if err := json.NewDecoder(resp.Body).Decode(&gif); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Error decoding Tenor response :skull:", userMention))
				return
			}

			if len(gif.Results) == 0 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, No gifs found wtf :sob:", userMention))
				return
			}

			// try to gte gif/tinygif
			var gifURL string
			if val, ok := gif.Results[0].MediaFormats["gif"]; ok {
				gifURL = val.URL
			} else if val, ok := gif.Results[0].MediaFormats["tinygif"]; ok {
				gifURL = val.URL
			} else if val, ok := gif.Results[0].MediaFormats["mediumgif"]; ok {
				gifURL = val.URL
			}

			if gifURL == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Couldn't find a usable gif :sob:", userMention))
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("%s, GIF: %s", userName, displayTerm),
				Image: &discordgo.MessageEmbedImage{
					URL: gifURL,
				},
				Color: 0xff69b4,
			}

			done <- true
			delete(gifWaiting, m.Author.ID)

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		} else if strings.HasPrefix(m.Content, "!math") {
			exprInput := strings.TrimSpace(strings.TrimPrefix(m.Content, "!math"))
			if exprInput == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Please provide a math expression. Example: `!math 2+2*5`", userMention))
				return
			}

			var exprStr string

			if strings.Contains(exprInput, "x") {
				exprStr = strings.ReplaceAll(exprInput, "x", "*")
				exprStr = strings.ReplaceAll(exprStr, "^", "**")
			} else {
				exprStr = exprInput
			}

			exprInput = strings.ReplaceAll(exprInput, "**", "^")

			// parse and evaluate omg :O
			expr, err := govaluate.NewEvaluableExpression(exprStr)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Invalid Expression! Example: `9/3`", userMention))
				return
			}

			result, err := expr.Evaluate(nil)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, Error Evaluating Expression!", userMention))
				return
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, **`%s`** = **`%v`**", userMention, exprInput, result))
		} else if strings.HasPrefix(m.Content, "!avatar") {
			// get mentioned user with @
			var user *discordgo.User
			if len(m.Mentions) > 0 {
				user = m.Mentions[0] // take the first mentioned user
			} else {
				user = m.Author
			}

			// Get avatar url
			avatarURL := user.AvatarURL("1024") // 1024 px size

			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("%s, %s's PFP", userName, user.Username),
				Image: &discordgo.MessageEmbedImage{
					URL: avatarURL,
				},
				Color: 0x00ffff,
			}

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
			return
		} else if strings.HasPrefix(m.Content, "!reverse") {
			text := strings.TrimSpace(strings.TrimPrefix(m.Content, "!reverse"))
			if text == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Please include text. Useage: `!reverse Hello`", userMention))
				return
			}

			content := m.Content             // full message string
			parts := strings.Fields(content) // split by spaces
			// "args" will be everything after the command
			args := parts[1:]

			input := strings.Join(args, " ")
			runes := []rune(input)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			reversed := string(runes)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", userMention, reversed))
		} else if strings.HasPrefix(m.Content, "!mock") {
			text := strings.TrimSpace(strings.TrimPrefix(m.Content, "!mock"))
			if text == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Please include text. Useage: `!mock Hello`", userMention))
				return
			}

			content := m.Content             // full message string
			parts := strings.Fields(content) // split by spaces
			// "args" will be everything after the command
			args := parts[1:]

			input := strings.Join(args, " ")
			mocked := ""

			for _, c := range input {
				if rand.Intn(2) == 0 {
					mocked += strings.ToLower(string(c))
				} else {
					mocked += strings.ToUpper(string(c))
				}
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", userMention, mocked))
		} else if strings.HasPrefix(m.Content, "!flip") {
			text := strings.TrimSpace(strings.TrimPrefix(m.Content, "!flip"))
			if text == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Please include text. Useage: `!flip Hello`", userMention))
				return
			}

			content := m.Content             // full message string
			parts := strings.Fields(content) // split by spaces
			// "args" will be everything after calling the cmd
			args := parts[1:]

			input := strings.Join(args, " ")

			flipMap := map[rune]rune{
				'a': 'ɐ', 'b': 'q', 'c': 'ɔ', 'd': 'p',
				'e': 'ǝ', 'f': 'ɟ', 'g': 'ƃ', 'h': 'ɥ',
				'i': 'ᴉ', 'j': 'ɾ', 'k': 'ʞ', 'l': 'ʃ',
				'm': 'ɯ', 'n': 'u', 'o': 'o', 'p': 'd',
				'q': 'b', 'r': 'ɹ', 's': 's', 't': 'ʇ',
				'u': 'n', 'v': 'ʌ', 'w': 'ʍ', 'x': 'x',
				'y': 'ʎ', 'z': 'z',
				'A': '∀', 'B': '𐐒', 'C': 'Ɔ', 'D': 'p',
				'E': 'Ǝ', 'F': 'Ⅎ', 'G': 'פ', 'H': 'H',
				'I': 'I', 'J': 'ſ', 'K': 'ʞ', 'L': '˥',
				'M': 'W', 'N': 'N', 'O': 'O', 'P': 'Ԁ',
				'Q': 'Q', 'R': 'ᴚ', 'S': 'S', 'T': '┴',
				'U': '∩', 'V': 'Λ', 'W': 'M', 'X': 'X',
				'Y': '⅄', 'Z': 'Z',
				'1': 'Ɩ', '2': 'ᄅ', '3': 'Ɛ', '4': 'ㄣ',
				'5': 'ϛ', '6': '9', '7': 'ㄥ', '8': '8',
				'9': '6', '0': '0',
				'.': '˙', ',': '\'', '\'': ',',
				'_': '‾', '&': '⅋', '?': '¿', '!': '¡',
			}

			// flip
			flipped := []rune{}

			for _, r := range input {
				if f, ok := flipMap[r]; ok {
					flipped = append([]rune{f}, flipped...)
				} else {
					flipped = append([]rune{r}, flipped...)
				}
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", userMention, string(flipped)))
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
			guessingTries[m.Author.ID] = 0
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Guess a Number between 1 and 100!", userMention))
		case "!rps":
			options := []string{"Rock", "Paper", "Scissors"}
			index := rand.Intn(len(options))
			botChoice := options[index]
			playingRPS[m.Author.ID] = botChoice
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, I have picked. Your turn!", userMention))
		case "!trivia":
			resp, err := http.Get("https://opentdb.com/api.php?amount=1")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch trivia question :sob:", userMention))
				return
			}
			defer resp.Body.Close()

			var triviaResp OpenTDBResponse
			if err := json.NewDecoder(resp.Body).Decode(&triviaResp); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode trivia question :skull:", userMention))
				return
			}

			if len(triviaResp.Results) == 0 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s No trivia question found :sob:", userMention))
				return
			}

			q := triviaResp.Results[0]
			q.Question = html.UnescapeString(q.Question)
			q.CorrectAnswer = html.UnescapeString(q.CorrectAnswer)
			for i := range q.IncorrectAnswers {
				q.IncorrectAnswers[i] = html.UnescapeString(q.IncorrectAnswers[i])
			}

			playingTrivia[m.Author.ID] = q
			triviaTries[m.Author.ID] = 0

			if q.Type == "multiple" {
				// build choices
				choices := append(q.IncorrectAnswers, q.CorrectAnswer)
				rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

				choiceMap := map[string]string{"A": choices[0], "B": choices[1], "C": choices[2], "D": choices[3]}

				triviaChoices[m.Author.ID] = choiceMap

				msg := fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_\n", userMention, q.Question)
				msg += fmt.Sprintf("A: %s\nB: %s\nC: %s\nD: %s", choiceMap["A"], choiceMap["B"], choiceMap["C"], choiceMap["D"])
				s.ChannelMessageSend(m.ChannelID, msg)

			} else if q.Type == "boolean" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_\nTrue or false?", userMention, q.Question))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Trivia Time\n**Questions:**\n_%s_", userMention, q.Question))
			}
		case "!meme":
			memeWaiting[m.Author.ID] = true
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://meme-api.com/gimme")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a meme :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var meme struct {
				PostLink string `json:"postLink"`
				URL      string `json:"url"`
				Title    string `json:"title"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&meme); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode meme :skull:", userMention))
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("%s, %s", userName, meme.Title),
				URL:   meme.PostLink,
				Image: &discordgo.MessageEmbedImage{
					URL: meme.URL,
				},
				Color: 0x00ff00, // green border
			}

			done <- true
			delete(timeWaiting, m.Author.ID)

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		case "!quote":
			quoteWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			client := http.Client{
				Timeout: time.Second * 10,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			req, _ := http.NewRequest("GET", "https://api.quotable.io/random", nil)
			req.Header.Set("User-Agent", "discord-bot (https://github.com/z-sk1, v1.0)")

			resp, err := client.Do(req)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a quote :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Error fetching quote :skull:: %s", userMention, string(body)))
				return
			}

			var quote struct {
				Content string `json:"content"`
				Author  string `json:"author"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode quote :skull:", userMention))
				return
			}

			done <- true
			delete(quoteWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, \n 💬 _%s_\n— **_%s_**", userMention, quote.Content, quote.Author))
		case "!joke":
			jokeWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://v2.jokeapi.dev/joke/Any")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a joke :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var data struct {
				Type     string `json:"type"`
				Setup    string `json:"setup"`
				Delivery string `json:"delivery"`
				Joke     string `json:"joke"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode joke :skull:", userMention))
				return
			}

			done <- true
			delete(jokeWaiting, m.Author.ID)

			if data.Type == "single" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, %s", userMention, data.Joke))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, %s\n||**_%s_**||", userMention, data.Setup, data.Delivery))
			}
		case "!fact":
			factWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://uselessfacts.jsph.pl/random.json?language=en")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a fact :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var data struct {
				Text string `json:"text"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode fact :sob:", userMention))
				return
			}

			done <- true
			delete(factWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, your useless fact:\n_%s_", userMention, data.Text))
		case "!advice":
			adviceWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://api.adviceslip.com/advice")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch life changing advice :sob:", userMention))
				return
			}
			defer resp.Body.Close()

			var adviceData struct {
				Slip struct {
					ID     int    `json:"id"`
					Advice string `json:"advice"`
				} `json:"slip"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&adviceData); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode life-changing advice :sob:", userMention))
				return
			}

			done <- true
			delete(adviceWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, your life-changing advice:\n_%s_", userMention, adviceData.Slip.Advice))
		case "!roast":
			roastWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://evilinsult.com/generate_insult.php?lang=en&type=json")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch insult :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var insultData struct {
				Insult string `json:"insult"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&insultData); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode insult :skull:", userMention))
				return
			}

			done <- true
			delete(roastWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, your roast:\n_%s_", userMention, insultData.Insult))
		case "!chucknorris":
			chucknorrisWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://api.chucknorris.io/jokes/random")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a Chuck Norris joke :sob:", userMention))
				return
			}
			defer resp.Body.Close()

			var result struct {
				Value string `json:"value"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode joke", userMention))
				return
			}

			done <- true
			delete(chucknorrisWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s\n Chuck Norris once said:\n _%s_", userMention, result.Value))
		case "!pickup":
			pickupWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://rizzapi.vercel.app/random")
			if err != nil {
				s.ChannelMessageSend("%s Couldn't fetch a pickup line :sob;", userMention)
				return
			}
			defer resp.Body.Close()

			var result struct {
				Text string `json:"text"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode pickup line :sob:", userMention))
				return
			}

			done <- true
			delete(pickupWaiting, m.Author.ID)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s, your rizzy ahh pickup line:\n_%s_", userMention, result.Text))
		case "!wouldyourather":
			wouldyouratherWaiting[m.Author.ID] = true
			done := make(chan bool)
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						s.ChannelTyping(m.ChannelID)
						time.Sleep(8 * time.Second)
					}
				}
			}()

			resp, err := http.Get("https://api.truthordarebot.xyz/v1/wyr")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Couldn't fetch a Would You Rather Question :skull:", userMention))
				return
			}
			defer resp.Body.Close()

			var result struct {
				Question string `json:"question"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s Failed to decode Would You Rather question :sob:", userMention))
				return
			}

			done <- true
			delete(wouldyouratherWaiting, m.Author.ID)

			startWouldYouRatherPoll(s, m.ChannelID, result.Question, 60, userName)
		case "!help":
			commands := strings.Join(cmdList, "\n")
			s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("%s, here are all the currently available commands: \n%s", userMention, commands))
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
