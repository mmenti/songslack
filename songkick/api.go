package songkick

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	EventTypeFestival  = "Festival"
	AttendanceImGoing  = "im_going"
	AttendanceIMightGo = "i_might_go"
	Pink               = "#f80046"
)

var apiKey = os.Getenv("SONGKICK_API_KEY")

func RequestUserEvents(username string, attendanceType string) JsonRoot {
	requestUrl := "http://api.songkick.com/api/3.0/users/" + username + "/events.json?apikey=" + apiKey + "&attendance=" + attendanceType
	rsp, err := http.Get(requestUrl)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	bodyByte, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}

	jsonData := JsonRoot{}
	err = json.Unmarshal(bodyByte, &jsonData)
	if err != nil {
		panic(err)
	}

	return jsonData
}
