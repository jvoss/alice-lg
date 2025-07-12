package frrproxy

import (
	"encoding/json"
	"io"
	"time"

	"github.com/alice-lg/alice-lg/pkg/api"
)

func parseRoutesResponseStream(
	body io.Reader,
	config Config,
) (*api.Meta, api.Routes, error) {

	dec := json.NewDecoder(body)
	meta := &api.Meta{}
	routes := api.Routes{}

	throttle := time.Duration(config.StreamParserThrottle) * time.Nanosecond

	// Read opening JSON token (should be `{`)
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, nil, io.ErrUnexpectedEOF
	}

	// Stream over top-level JSON keys
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}

		key, ok := tok.(string)
		if !ok {
			continue
		}

		switch key {
		case "routes":
			// Expecting the routes object
			if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
				return nil, nil, err
			}

			// Iterate over route prefixes
			for dec.More() {
				// Read prefix key
				prefixTok, err := dec.Token()
				if err != nil {
					return nil, nil, err
				}
				prefix, ok := prefixTok.(string)
				if !ok {
					continue
				}

				// Decode the list of path entries
				var entries []map[string]interface{}
				if err := dec.Decode(&entries); err != nil {
					return nil, nil, err
				}

				for _, entry := range entries {
					// Inject the prefix so parseRouteData can pick it up
					entry["network"] = prefix

					// Throttle parsing to reduce CPU
					time.Sleep(throttle)

					route := parseRouteData(entry, prefix, config, false)
					if route != nil {
						routes = append(routes, route)
					}
				}
			}

			// Expect closing '}' for routes
			if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
				return nil, nil, err
			}

		case "tableVersion", "vrfId", "vrfName", "routerId", "defaultLocPrf", "localAS":
			// Skip values we don't need for API meta (yet)
			if err := skipNext(dec); err != nil {
				return nil, nil, err
			}

		default:
			// Ignore unknown keys
			if err := skipNext(dec); err != nil {
				return nil, nil, err
			}
		}
	}

	return meta, routes, nil
}

// skipNext consumes the next value from the decoder without storing it.
func skipNext(dec *json.Decoder) error {
	var skip interface{}
	return dec.Decode(&skip)
}
