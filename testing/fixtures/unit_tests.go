package fixtures

import (
	"strings"

	"github.com/humper/tor_exit_nodes/models"
)

// TestRows is a slice of models.TorExitNode used for testing.
var TestRows = []*models.TorExitNode{}

func init() {
	ips := strings.Split(MockEndpoints["/tor/clean"], "\n")
	for _, ip := range ips {
		TestRows = append(TestRows, &models.TorExitNode{IP: ip})
	}
}

func FilterRows(allowedIPs []string) []*models.TorExitNode {
	filtered := []*models.TorExitNode{}
	for _, row := range TestRows {
		for _, ip := range allowedIPs {
			if row.IP == ip {
				filtered = append(filtered, row)
			}
		}
	}
	return filtered
}
