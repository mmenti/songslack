package songkick

import (
	"fmt"
	"strings"
)

const maxLineupDisplay = 6

func FormatLineup(performances []Performance) string {
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

func FormatLocation(songkickEvent Event) string {
	return fmt.Sprintf("%s, %s", songkickEvent.VenueData.MetroAreaData.DisplayName, songkickEvent.VenueData.MetroAreaData.CountryData.DisplayName)
}
