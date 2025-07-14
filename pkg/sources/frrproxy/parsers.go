package frrproxy

import (
	"strconv"
	"strings"

	"github.com/alice-lg/alice-lg/pkg/api"
)

// Parse a list of FRR returned community strings
func parseBgpCommunityList(list []string) api.Communities {
	var result api.Communities

	for _, item := range list {
		parts := strings.Split(item, ":")
		if len(parts) > 2 {
			var community api.Community

			first, err1 := strconv.Atoi(parts[0])
			second, err2 := strconv.Atoi(parts[1])

			if err1 != nil || err2 != nil {
				continue
			}

			community = append(community, first)
			community = append(community, second)

			if len(parts) == 3 {
				third, err3 := strconv.Atoi(parts[2])

				if err3 != nil {
					continue
				}

				community = append(community, third)
			}

			result = append(result, community)
		} else {
			continue
		}
	}

	return result
}

// Parse the extendedCommunity string
func parseExtBgpCommunities(str string) api.ExtCommunities {
	var result api.ExtCommunities

	extComm := strings.Split(str, " ")
	for _, item := range extComm {
		if item != "" {
			result = append(result, api.ExtCommunity{item})
		}
	}

	return result
}
