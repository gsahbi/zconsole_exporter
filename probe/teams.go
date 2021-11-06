package probe

import (
	"log"
	"zconsole_exporter/client"
)

type Team struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func probeTeams(c client.ZConsoleAPI) ([]Team, error) {

	var ts []Team

	if err := c.Get("api/teams/v1/teams/all", "", &ts, nil); err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	return ts, nil
}
