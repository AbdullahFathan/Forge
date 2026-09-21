package task

import "github.com/google/uuid"

func WouldCycle(edges map[uuid.UUID][]uuid.UUID, from, to uuid.UUID) bool {
	seen := map[uuid.UUID]bool{}
	var dfs func(uuid.UUID) bool
	dfs = func(n uuid.UUID) bool {
		if n == from {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		for _, nxt := range edges[n] {
			if dfs(nxt) {
				return true
			}
		}
		return false
	}
	return dfs(to)
}
