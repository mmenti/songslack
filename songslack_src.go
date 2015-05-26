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
	client *slack.Client
)

var songkick_api_key = "YOUR SONGKICK API KEY"
var slack_api_key = "YOUR SLACK API KEY"
var aws_access_key = "YOUR AMAZON WEB SERVICES ACCESS KEY"
var aws_access_secret = "YOUR AMAZON WEB SERVICES SECRET"

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
	// and post it to the #gigs Slack channel for the corresponding Slack user
	users := make([]SongSlackUser, 2, 2)
	users[0] = SongSlackUser{"songkick_username_1", "slack_username_1"}
	users[1] = SongSlackUser{"songkick_username_2", "slack_username_2"}

	mysdb, _ := sdb.NewSimpleDB(sdb.RegionEuIreland, awsKey)
	// the domain that we want to use with simple db to store our data
	myDomain, _ := mysdb.OpenDomain("songslack")

	client = slack.NewClient(slack_api_key)

	for _, u := range users {

		// interested in
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
			if len(sItem) == 0 {
				err := client.SendMessage("#gigs", fmt.Sprintf("I'm interested in %s %s", v.DisplayName, v.URI), u.SlackUsername)
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

		// going
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

		err = json.Unmarshal(body_byte, &jsonData)

		for _, v := range jsonData.ResultsPage.ResultData.EventData {
			sdb_key := u.SongkickUsername + "-going-" + strconv.Itoa(v.ID)
			sItem, _, err := mysdb.Select("select * from songslack where itemName() = '"+sdb_key+"'", true, nil)
			if err != nil {
				panic(err)
			}
			if len(sItem) == 0 {
				err := client.SendMessage("#gigs", fmt.Sprintf("I'm going to %s %s", v.DisplayName, v.URI), u.SlackUsername)
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
