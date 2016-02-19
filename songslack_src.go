// simple golang program to process Songkick "interested"/"going" events and post them to a given Slack channel

package main

import (
	"encoding/json"
	"fmt"
	"github.com/bluele/slack"
	"github.com/jacobsa/aws"
	"github.com/jacobsa/aws/sdb"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type SongSlackUser struct {
	SongkickUsername string
	SlackUsername    string
}

type Artist struct {
	DisplayName string `json:"displayName"`
}

type Performance struct {
	ArtistData Artist `json:"artist"`
}

type Country struct {
	DisplayName string `json:"displayName"`
}

type MetroArea struct {
	DisplayName string  `json:"displayName"`
	CountryData Country `json:"country"`
}

type Venue struct {
	DisplayName   string    `json:"displayName"`
	MetroAreaData MetroArea `json:"metroArea"`
}

type Start struct {
	Date string `json:"date"`
}

type Event struct {
	DisplayName     string        `json:"displayName"`
	URI             string        `json:"uri"`
	ID              int           `json:"id"`
	Type            string        `json:"type"`
	PerformanceData []Performance `json:"performance"`
	VenueData       Venue         `json:"venue"`
	StartData       Start         `json:"start"`
}

type Results struct {
	EventData []Event `json:"event"`
}

type ResultsPage struct {
	ResultData Results `json:"results"`
}

type jsonRoot struct {
	ResultsPage `json:"resultsPage"`
}

const (
	channelName        = "gigs" // the name of the slack channel you want to post your Songkick events to
	attendanceImGoing  = "im_going"
	attendanceIMightGo = "i_might_go"
	songkickPink       = "#f80046"
)

var (
	// Songkick
	songkick_api_key = os.Getenv("SONGKICK_API_KEY") // your songkick API key
	// Slack
	slack_api_key = os.Getenv("SLACK_API_KEY") // your slack API token
	slackClient   = slack.New(slack_api_key)
	slackChannel  *slack.Channel
	// AWS
	mysdb             sdb.SimpleDB
	myDomain          sdb.Domain
	aws_access_key    = os.Getenv("AWS_KEY")    // your amazon web services access key
	aws_access_secret = os.Getenv("AWS_SECRET") //your amazon web services secret

	messages = map[string]string{
		attendanceImGoing:  "I’m going to",
		attendanceIMightGo: "I’m interested in"}
	sdbKeyFormats = map[string]string{
		attendanceImGoing:  "%s-going-%d",
		attendanceIMightGo: "%s-tracking-%d"}
	users = []SongSlackUser{}
)

func init() {
	awsKey := aws.AccessKey{aws_access_key, aws_access_secret}

	// add your Songkick and Slack usernames here we check for new events for
	// each Songkick user and post it to the a Slack channel for the
	// corresponding Slack user (manually specified in this example, you could of
	// course read this from elsewhere)
	users = append(users, SongSlackUser{"songkick_username_1", "slack_username_1"})
	users = append(users, SongSlackUser{"songkick_username_2", "slack_username_2"})

	// connect to SimpleDB, to store previously posted events you could replace
	// this with the datastore of your choice
	mysdb, _ = sdb.NewSimpleDB(sdb.RegionEuIreland, awsKey)
	// the domain that we want to use with simple db to store our data
	myDomain, _ = mysdb.OpenDomain("songslack")

	var err error
	slackChannel, err = slackClient.FindChannelByName(channelName)
	if err != nil {
		panic(err)
	}
}

func main() {
	// for each user
	for _, u := range users {
		// check "I’m going" and "I might go"
		for _, attendanceType := range []string{attendanceImGoing, attendanceIMightGo} {
			request_url := "http://api.songkick.com/api/3.0/users/" + u.SongkickUsername + "/events.json?apikey=" + songkick_api_key + "&attendance=" + attendanceType
			rsp, err := http.Get(request_url)
			if err != nil {
				panic(err)
			}
			defer rsp.Body.Close()
			body_byte, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				panic(err)
			}

			jsonData := jsonRoot{}
			err = json.Unmarshal(body_byte, &jsonData)

			for _, songkickEvent := range jsonData.ResultsPage.ResultData.EventData {
				sdbKey := fmt.Sprintf(sdbKeyFormats[attendanceType], u.SongkickUsername, songkickEvent.ID)

				sItem, _, err := mysdb.Select("select * from songslack where itemName() = '"+sdbKey+"'", true, nil)
				if err != nil {
					panic(err)
				}
				// if no match, hasn't been posted before so lets send it to Slack
				if len(sItem) == 0 {
					err = postToSlack(u.SlackUsername, attendanceType, songkickEvent)
					if err != nil {
						panic(err)
					}

					// add the item to sdb
					newItem := sdb.ItemName(sdbKey)
					upd := make([]sdb.PutUpdate, 1, 1)
					upd[0] = sdb.PutUpdate{"eventid", strconv.Itoa(songkickEvent.ID), true}
					err = myDomain.PutAttributes(newItem, upd, nil)
					if err != nil {
						panic(err)
					}
				}
			} // end of for _, songkickEvent := range jsonData...
		} // end of for _, attendanceType := range [im_going, i_might_go]
	} // end of for _, u := range users
}

func postToSlack(username string, attendanceType string, songkickEvent Event) error {

	messageBase := messages[attendanceType]
	message := messageBase + "..."

	fallbackFormat := message + " %s"
	fallback := fmt.Sprintf(fallbackFormat, songkickEvent.URI)

	title := songkickEvent.DisplayName

	fields := []*slack.AttachmentField{}

	if songkickEvent.Type == "Festival" {
		lineup := slack.AttachmentField{
			Title: "Line-up",
			Value: lineup(songkickEvent.PerformanceData),
			Short: false}
		fields = append(fields, &lineup)
	}

	fields = append(fields, &slack.AttachmentField{
		Title: "Location",
		Value: songkickLocation(songkickEvent),
		Short: true})

	err := slackClient.ChatPostMessage(slackChannel.Id, message, &slack.ChatPostMessageOpt{
		Username: username,
		Attachments: []*slack.Attachment{
			{
				Fallback:  fallback,
				Title:     title,
				TitleLink: songkickEvent.URI,
				Color:     songkickPink,
				Fields:    fields}},
	})

	if err != nil {
		return err
	} else {
		return nil
	}
}

func songkickLocation(songkickEvent Event) string {
	return fmt.Sprintf("%s, %s", songkickEvent.VenueData.MetroAreaData.DisplayName, songkickEvent.VenueData.MetroAreaData.CountryData.DisplayName)
}

const maxLineupDisplay = 6

func lineup(performances []Performance) string {
	var (
		last   string
		lineup []string
	)

	if len(performances) == 0 {
		return "TBA"
	}

	for _, performance := range performances {
		lineup = append(lineup, performance.ArtistData.DisplayName)
		if len(lineup) >= maxLineupDisplay {
			break
		}
	}

	if len(lineup) == 1 {
		return lineup[0]
	}

	if len(lineup) == 2 {
		return strings.Join(lineup, " and ")
	}

	if len(performances) > maxLineupDisplay {
		last = "more..."
	} else {
		// "pop" last act off lineup via https://github.com/golang/go/wiki/SliceTricks
		last, lineup = lineup[len(lineup)-1], lineup[:len(lineup)-1]
	}

	return strings.Join(lineup, ", ") + ", and " + last
}
