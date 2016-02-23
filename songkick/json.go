package songkick

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

type JsonRoot struct {
	ResultsPage `json:"resultsPage"`
}
