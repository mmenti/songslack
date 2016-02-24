package lastfm

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

var apiKey = os.Getenv("LASTFM_API_KEY")

func GetImageByMbid(mbid string) string {
	requestUrl := "http://ws.audioscrobbler.com/2.0/?method=artist.getinfo&format=json&mbid=" + mbid + "&api_key=" + apiKey
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

	var result string
	// if there are images, return the 3rd image ("large" size)
	if len(jsonData.ArtistData.Images) > 0 {
		result = jsonData.ArtistData.Images[2].URI
	}
	return result
}
