package lastfm

type Image struct {
	Size string `json:"size"`
	URI  string `json:"#text"`
}

type Artist struct {
	Images []Image `json:"image"`
}

type JsonRoot struct {
	ArtistData Artist `json:"artist"`
}
