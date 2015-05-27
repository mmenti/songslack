// simple golang program to process Songkick "interested"/"going" events and post them to a given Slack channel

package main

import (
	"encoding/json"
	"fmt"
	"github.com/Bowery/slack"
	"github.com/jacobsa/aws"
	"github.com/jacobsa/aws/sdb"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	client            *slack.Client
	songkick_api_key  = "YOUR SONGKICK API KEY"
	slack_api_key     = "YOUR SLACK API KEY"
	aws_access_key    = "YOUR AMAZON WEB SERVICES ACCESS KEY"
	aws_access_secret = "YOUR AMAZON WEB SERVICES SECRET"
	slack_channel     = "#gigs" // the name of the slack channel you want to post your Songkick events to
)

func main() {

	type Event struct {
		DisplayName string `json:"displayName"`
		URI         string `json:"uri"`
		ID          int    `json:"id"`
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

	type SongSlackUser struct {
		SongkickUsername string
		SlackUsername    string
	}

	awsKey := aws.AccessKey{aws_access_key, aws_access_secret}

	// add your Songkick and Slack usernames here
	// we check for new events for each Songkick user
	// and post it to the a Slack channel for the corresponding Slack user
	// (manually specified in this example, you could of course read this from elsewhere)
	users := make([]SongSlackUser, 2, 2)
	users[0] = SongSlackUser{"songkick_username_1", "slack_username_1"}
	users[1] = SongSlackUser{"songkick_username_2", "slack_username_2"}

	// connect to SimpleDB, to store previously posted events
	// you could replace this with the datastore of your choice
	mysdb, _ := sdb.NewSimpleDB(sdb.RegionEuIreland, awsKey)
	// the domain that we want to use with simple db to store our data
	myDomain, _ := mysdb.OpenDomain("songslack")

	client = slack.NewClient(slack_api_key)

	for _, u := range users {

		// check "I'm interested in" Songkick events for each user
		request_url := "http://api.songkick.com/api/3.0/users/" + u.SongkickUsername + "/events.json?apikey=" + songkick_api_key + "&attendance=i_might_go"
		rsp, err := http.Get(request_url)
		if err != nil {
			panic(err)
		}
		defer rsp.Body.Close()
		body_byte, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			panic(err)
		}

		var jsonData jsonRoot
		err = json.Unmarshal(body_byte, &jsonData)

		for _, v := range jsonData.ResultsPage.ResultData.EventData {
			sdb_key := u.SongkickUsername + "-tracking-" + strconv.Itoa(v.ID)
			sItem, _, err := mysdb.Select("select * from songslack where itemName() = '"+sdb_key+"'", true, nil)
			if err != nil {
				panic(err)
			}
			// if no match, hasn't been posted before so lets send it to Slack
			if len(sItem) == 0 {
				err := client.SendMessage(slack_channel, fmt.Sprintf("I'm interested in %s %s", v.DisplayName, v.URI), u.SlackUsername)
				if err != nil {
					panic(err)
				}
				// add the item to sdb
				newItem := sdb.ItemName(sdb_key)
				upd := make([]sdb.PutUpdate, 1, 1)
				upd[0] = sdb.PutUpdate{"eventid", strconv.Itoa(v.ID), true}
				err = myDomain.PutAttributes(newItem, upd, nil)
				if err != nil {
					panic(err)
				}
			}
		}

		// check "I'm going" Songkick events for each user
		request_url = "http://api.songkick.com/api/3.0/users/" + u.SongkickUsername + "/events.json?apikey=" + songkick_api_key
		rsp, err = http.Get(request_url)
		if err != nil {
			panic(err)
		}
		defer rsp.Body.Close()
		body_byte, err = ioutil.ReadAll(rsp.Body)
		if err != nil {
			panic(err)
		}

		jsonData = jsonRoot{}
		err = json.Unmarshal(body_byte, &jsonData)

		for _, v := range jsonData.ResultsPage.ResultData.EventData {
			sdb_key := u.SongkickUsername + "-going-" + strconv.Itoa(v.ID)
			sItem, _, err := mysdb.Select("select * from songslack where itemName() = '"+sdb_key+"'", true, nil)
			if err != nil {
				panic(err)
			}
			// if no match, hasn't been posted before so lets send it to Slack
			if len(sItem) == 0 {
				err := client.SendMessage(slack_channel, fmt.Sprintf("I'm going to %s %s", v.DisplayName, v.URI), u.SlackUsername)
				if err != nil {
					panic(err)
				}
				// add the item to sdb
				newItem := sdb.ItemName(sdb_key)
				upd := make([]sdb.PutUpdate, 1, 1)
				upd[0] = sdb.PutUpdate{"eventid", strconv.Itoa(v.ID), true}
				err = myDomain.PutAttributes(newItem, upd, nil)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
