/*
	simple golang program to process Songkick "interested"/"going" events and
	post them to a given Slack channel

	Run with the following environment variables set:

	SONGKICK_API_KEY=   your songkick api key
	SLACK_API_KEY=			your slack api token
	SLACK_CHANNEL_NAME= the name of the slack channel you want to post your Songkick events to
	AWS_KEY=						your amazon web services access key
	AWS_SECRET=         your amazon web services secret

	to specify a pair of usernames to do a dry-run of posts with, set
	DRY_RUN_SLACK_USER=
	DRY_RUN_SONGKICK_USER=
*/

package main

import (
	"github.com/bluele/slack"
	"github.com/jacobsa/aws"
	"github.com/jacobsa/aws/sdb"

	"songslack/songkick"

	"fmt"
	"os"
	"strconv"
)

type SongSlackUser struct {
	SongkickUsername string
	SlackUsername    string
}

var (
	err error

	// Slack
	slackApiKey  = os.Getenv("SLACK_API_KEY")
	channelName  = os.Getenv("SLACK_CHANNEL_NAME")
	slackClient  = slack.New(slackApiKey)
	slackChannel *slack.Channel

	// AWS
	mysdb             sdb.SimpleDB
	myDomain          sdb.Domain
	awsAccessKey    = os.Getenv("AWS_KEY")
	awsAccessSecret = os.Getenv("AWS_SECRET")
	sdbKeyFormats = map[string]string{
		songkick.AttendanceImGoing:  "%s-going-%d",
		songkick.AttendanceIMightGo: "%s-tracking-%d"}

	messages = map[string]string{
		songkick.AttendanceImGoing:  "I’m going to",
		songkick.AttendanceIMightGo: "I’m interested in"}
	users = []SongSlackUser{}

	dryRunUser = SongSlackUser{os.Getenv("DRY_RUN_SONGKICK_USER"), os.Getenv("DRY_RUN_SLACK_USER")}
)

func isDryRun() bool {
	if dryRunUser.SongkickUsername != "" && dryRunUser.SlackUsername != "" {
		return true
	} else {
		return false
	}
}

func init() {
	if isDryRun() {
		users = append(users, dryRunUser)
	}
	// add your Songkick and Slack usernames here we check for new events for
	// each Songkick user and post it to the a Slack channel for the
	// corresponding Slack user (manually specified in this example, you could of
	// course read this from elsewhere)
	if !isDryRun() {
		users = append(users, SongSlackUser{"songkick_username_1", "slack_username_1"})
		users = append(users, SongSlackUser{"songkick_username_2", "slack_username_2"})

		awsKey := aws.AccessKey{awsAccessKey, awsAccessSecret}

		// connect to SimpleDB, to store previously posted events you could replace
		// this with the datastore of your choice
		mysdb, _ = sdb.NewSimpleDB(sdb.RegionEuIreland, awsKey)
		// the domain that we want to use with simple db to store our data
		myDomain, _ = mysdb.OpenDomain("songslack")
	}

	slackChannel, err = slackClient.FindChannelByName(channelName)
	if err != nil {
		panic(err)
	}
}

func main() {
	// for each user
	for _, u := range users {
		// check "I’m going" and "I might go"
		for _, attendanceType := range []string{songkick.AttendanceImGoing, songkick.AttendanceIMightGo} {

			jsonData := songkick.RequestUserEvents(u.SongkickUsername, attendanceType)

			for _, songkickEvent := range jsonData.ResultsPage.ResultData.EventData {
				sdbKey := fmt.Sprintf(sdbKeyFormats[attendanceType], u.SongkickUsername, songkickEvent.ID)

				var sItem []sdb.SelectedItem

				if !isDryRun() {
					sItem, _, err = mysdb.Select("select * from songslack where itemName() = '"+sdbKey+"'", true, nil)
					if err != nil {
						panic(err)
					}
				}

				// if there is no match (hasn't been posted before), let's send it to Slack
				if len(sItem) == 0 {
					err = postToSlack(u.SlackUsername, attendanceType, songkickEvent)
					if err != nil {
						panic(err)
					}

					if !isDryRun() {
						// add the item to sdb
						newItem := sdb.ItemName(sdbKey)
						upd := make([]sdb.PutUpdate, 1, 1)
						upd[0] = sdb.PutUpdate{"eventid", strconv.Itoa(songkickEvent.ID), true}
						err = myDomain.PutAttributes(newItem, upd, nil)
						if err != nil {
							panic(err)
						}
					}
				}
			} // end of for _, songkickEvent := range jsonData...
		} // end of for _, attendanceType := range [im_going, i_might_go]
	} // end of for _, u := range users
}

func postToSlack(username string, attendanceType string, songkickEvent songkick.Event) error {
	messageBase := messages[attendanceType]
	message := messageBase + "..."

	fallbackFormat := message + " %s"
	fallback := fmt.Sprintf(fallbackFormat, songkickEvent.URI)

	title := songkickEvent.DisplayName

	fields := []*slack.AttachmentField{}

	if songkickEvent.Type == "Festival" {
		lineup := slack.AttachmentField{
			Title: "Line-up",
			Value: songkick.FormatLineup(songkickEvent.PerformanceData),
			Short: false}
		fields = append(fields, &lineup)
	}

	fields = append(fields, &slack.AttachmentField{
		Title: "Location",
		Value: songkick.FormatLocation(songkickEvent),
		Short: true})

	err := slackClient.ChatPostMessage(slackChannel.Id, message, &slack.ChatPostMessageOpt{
		Username: username,
		Attachments: []*slack.Attachment{
			{
				Fallback:  fallback,
				Title:     title,
				TitleLink: songkickEvent.URI,
				Color:     songkick.Pink,
				Fields:    fields}},
	})

	if err != nil {
		return err
	} else {
		return nil
	}
}
