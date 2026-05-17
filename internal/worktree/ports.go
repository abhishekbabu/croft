package worktree

import "fmt"

// AllocatePorts assigns each service the lowest free port in the inclusive
// range [low, high] that is not already in the taken set. taken is the union
// of every other worktree's ports; it is read but never mutated.
func AllocatePorts(low, high int, services []string, taken map[int]bool) (map[string]int, error) {
	used := make(map[int]bool, len(taken))
	for p := range taken {
		used[p] = true
	}
	result := make(map[string]int, len(services))
	port := low
	for _, svc := range services {
		for port <= high && used[port] {
			port++
		}
		if port > high {
			return nil, fmt.Errorf("port range %d-%d exhausted: cannot place service %q", low, high, svc)
		}
		result[svc] = port
		used[port] = true
		port++
	}
	return result, nil
}
