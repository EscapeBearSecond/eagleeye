package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePortsList(data string) ([]int, error) {
	return parsePortsSlice(strings.Split(data, ","))
}

func parsePortsSlice(ranges []string) ([]int, error) {
	var ports []int
	for _, r := range ranges {
		r = strings.TrimSpace(r)

		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid port selection segment: '%s'", r)
			}

			p1, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid port number: '%s'", parts[0])
			}

			p2, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port number: '%s'", parts[1])
			}

			if p1 > p2 || p2 > 65535 {
				return nil, fmt.Errorf("invalid port range: %d-%d", p1, p2)
			}

			for i := p1; i <= p2; i++ {
				ports = append(ports, i)
			}
		} else {
			portNumber, err := strconv.Atoi(r)
			if err != nil || portNumber > 65535 {
				return nil, fmt.Errorf("invalid port number: '%s'", r)
			}
			ports = append(ports, portNumber)
		}
	}

	seen := make(map[int]struct{})
	var dedupedPorts []int
	for _, port := range ports {
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		dedupedPorts = append(dedupedPorts, port)
	}

	return dedupedPorts, nil
}
