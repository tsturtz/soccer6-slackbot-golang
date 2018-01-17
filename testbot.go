package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	//"reflect"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nlopes/slack"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	defaultChannel := os.Getenv("SLACK_CHANNEL")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	const (
		DEFAULT_MESSAGE_USERNAME         = "CalvaryBot"
		DEFAULT_MESSAGE_THREAD_TIMESTAMP = ""
		DEFAULT_MESSAGE_REPLY_BROADCAST  = false
		DEFAULT_MESSAGE_ASUSER           = false
		DEFAULT_MESSAGE_PARSE            = ""
		DEFAULT_MESSAGE_LINK_NAMES       = 0
		DEFAULT_MESSAGE_UNFURL_LINKS     = false
		DEFAULT_MESSAGE_UNFURL_MEDIA     = true
		DEFAULT_MESSAGE_ICON_URL         = ""
		DEFAULT_MESSAGE_ICON_EMOJI       = ":soccer:"
		DEFAULT_MESSAGE_MARKDOWN         = true
		DEFAULT_MESSAGE_ESCAPE_TEXT      = true
	)

	// set a time for this message to go out
	api.PostMessage(defaultChannel, "Hello World!", slack.PostMessageParameters{
		Username:    DEFAULT_MESSAGE_USERNAME,
		User:        DEFAULT_MESSAGE_USERNAME,
		AsUser:      DEFAULT_MESSAGE_ASUSER,
		Parse:       DEFAULT_MESSAGE_PARSE,
		LinkNames:   DEFAULT_MESSAGE_LINK_NAMES,
		Attachments: nil,
		UnfurlLinks: DEFAULT_MESSAGE_UNFURL_LINKS,
		UnfurlMedia: DEFAULT_MESSAGE_UNFURL_MEDIA,
		IconURL:     DEFAULT_MESSAGE_ICON_URL,
		IconEmoji:   DEFAULT_MESSAGE_ICON_EMOJI,
		Markdown:    DEFAULT_MESSAGE_MARKDOWN,
		EscapeText:  DEFAULT_MESSAGE_ESCAPE_TEXT,
	})

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:

				fmt.Printf("Message: %v\n", ev)
				info := rtm.GetInfo()
				prefix := fmt.Sprintf("<@%s> ", info.User.ID)

				if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
					respond(rtm, ev, prefix)
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				// Take no action
			}
		}
	}
}

func getMonth(d string) int {
	// month // contains func
	var m int
	c := strings.Contains
	if c(d, "January") {
		m = 1
	} else if c(d, "February") {
		m = 2
	} else if c(d, "March") {
		m = 3
	} else if c(d, "April") {
		m = 4
	} else if c(d, "May") {
		m = 5
	} else if c(d, "June") {
		m = 6
	} else if c(d, "July") {
		m = 7
	} else if c(d, "August") {
		m = 8
	} else if c(d, "September") {
		m = 9
	} else if c(d, "October") {
		m = 10
	} else if c(d, "November") {
		m = 11
	} else if c(d, "December") {
		m = 12
	}
	return m
}

func respond(rtm *slack.RTM, msg *slack.MessageEvent, prefix string) {

	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	fmt.Printf("%s\n", text)
	var response string

	fullSchedule := map[string]bool{
		"schedule":         true,
		"show schedule":    true,
		"full schedule":    true,
		"full":             true,
		"games":            true,
		"all games":        true,
		"all matches":      true,
		"all":              true,
		"upcoming":         true,
		"upcoming games":   true,
		"upcoming matches": true,
	}

	nextGame := map[string]bool{
		"next":                   true,
		"next game":              true,
		"game":                   true,
		"next match":             true,
		"when is the next game?": true,
		"whens the next game?":   true,
	}

	if fullSchedule[text] {
		// scrape soccer6 schedule
		doc, err := goquery.NewDocument("https://soccer6.net/schedule/")
		if err != nil {
			log.Fatal(err)
		}
		// find calvary chapel's games and return all of them
		doc.Find(".schedule-date .team-133").Each(func(index int, item *goquery.Selection) {
			date_ := item.Parent().Parent().Parent().Parent().Parent().Find("h5").Text()
			time_ := item.Parent().Parent().Parent().Parent().Find(".match-info .datetime-dropdown").Text()
			field := item.Parent().Parent().Parent().Parent().Find(".match-info .venue-dropdown a").Text()
			fieldSlice := field[5:12]
			score := item.Parent().Parent().Find(".match-vs .visible-print-inline").Text()
			scoreHome := ""
			scoreAway := ""
			away := item.Parent().HasClass("away-team")
			otherTeam := ""
			if away {
				otherTeam = item.Parent().Parent().Find(".home-team .match-team").Text()
			} else {
				otherTeam = item.Parent().Parent().Find(".away-team .match-team").Text()
			}
			if score == " : " {
				score = ""
			} else {
				scoreSplit := strings.Split(score, " : ")
				scoreHome = strings.TrimSpace(scoreSplit[0])
				scoreAway = strings.TrimSpace(scoreSplit[1])
				score = "- *(" + score + ")*"
			}
			winOrLose := ""
			winOrLoseEmoji := ""
			upcoming := false
			scoreHome_, err := strconv.Atoi(scoreHome)
			if err != nil {
				upcoming = true
			}
			scoreAway_, err := strconv.Atoi(scoreAway)
			if err != nil {
				upcoming = true
			}
			if !upcoming {
				if away {
					if scoreAway_ > scoreHome_ {
						winOrLose = "_Win_"
						winOrLoseEmoji = ":grinning:"
					} else if scoreAway_ < scoreHome_ {
						winOrLose = "_Loss_"
						winOrLoseEmoji = ":unamused:"
					} else if scoreAway_ == scoreHome_ {
						winOrLose = "_Draw_"
						winOrLoseEmoji = ":neutral_face:"
					}
				} else {
					if scoreAway_ < scoreHome_ {
						winOrLose = "_Win_"
						winOrLoseEmoji = ":grinning:"
					} else if scoreAway_ > scoreHome_ {
						winOrLose = "_Loss_"
						winOrLoseEmoji = ":unamused:"
					} else if scoreAway_ == scoreHome_ {
						winOrLose = "_Draw_"
						winOrLoseEmoji = ":neutral_face:"
					}
				}
			}
			response := fmt.Sprintf(">>>%s%s%s\nAt %s, on %s, vs. _%s_ %s %s %s\n", "*", date_, "*", strings.TrimSpace(time_), fieldSlice, strings.TrimSpace(otherTeam), score, winOrLoseEmoji, winOrLose)
			rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
		})
	} else if nextGame[text] {
		// scrape soccer6 schedule
		doc, err := goquery.NewDocument("https://soccer6.net/schedule/")
		if err != nil {
			log.Fatal(err)
		}

		// get current date to compare against game weeks
		now := time.Now().UTC()

		// set flag to only grab next game
		nextGameOnly := true

		// find calvary chapel's games and return the first one
		doc.Find(".schedule-date .team-133").Each(func(index int, item *goquery.Selection) {
			date_ := item.Parent().Parent().Parent().Parent().Parent().Find("h5").Text()
			time_ := item.Parent().Parent().Parent().Parent().Find(".match-info .datetime-dropdown").Text()
			time_ = strings.TrimSpace(time_)
			// scrape and parse date into this format from soccer6
			monthNum := getMonth(date_)
			month := time.Month(monthNum)
			daySplit := strings.Split(date_, " ")
			dayString := daySplit[2]
			day, err := strconv.Atoi(dayString)
			fmt.Println(err)
			hourSplit := strings.Split(time_, ":")
			hourString := hourSplit[0]
			hour, err := strconv.Atoi(hourString)
			fmt.Println(err)
			clockEmoji := ":clock11:"
			if hour == 12 {
				clockEmoji = ":clock12:"
			} else if hour == 1 {
				clockEmoji = ":clock1:"
			}
			// update year (TODO: there is a data-date available on the page - grab that and refactor month and day above)
			gameDate := time.Date(2018, month, day, hour, 0, 0, 0, time.UTC)
			// only show next game
			if (gameDate.After(now) || gameDate.Equal(now)) && nextGameOnly == true {
				field := item.Parent().Parent().Parent().Parent().Find(".match-info .venue-dropdown a").Text()
				fieldSlice := field[5:12]
				fieldEmoji := ":stadium:"
				away := item.Parent().HasClass("away-team")
				otherTeam := ""
				otherTeamEmoji := ":busts_in_silhouette:"
				if away {
					otherTeam = item.Parent().Parent().Find(".home-team .match-team").Text()
				} else {
					otherTeam = item.Parent().Parent().Find(".away-team .match-team").Text()
				}
				response = fmt.Sprintf("The next game is on %s%s%s\n>>>%s %s\n%s %s\n%s _%s_\n", "*", date_, "*", clockEmoji, time_, fieldEmoji, fieldSlice, otherTeamEmoji, strings.TrimSpace(otherTeam))
				nextGameOnly = false
			}
		})
		// handle no upcoming matches
		if nextGameOnly == true {
			response = fmt.Sprintf("%sBummer.%s No upcoming regular season matches til next season. :disappointed:", "*", "*")
		}
		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	} else {
		response := fmt.Sprintf("- Use %snext game%s or simply %snext%s for next game\n- Use %sschedule%s or %sall%s for all games", "*", "*", "*", "*", "*", "*", "*", "*")
		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	}
}
