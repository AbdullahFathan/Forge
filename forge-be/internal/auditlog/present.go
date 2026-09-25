package auditlog

import "github.com/google/uuid"

type namedRow struct {
	ID   uuid.UUID
	Name string
}

func (s *Service) Present(rows []Log) []Public {
	out := make([]Public, len(rows))
	for i, row := range rows {
		out[i] = ToPublic(row)
	}
	if s == nil || s.db == nil || len(rows) == 0 {
		return out
	}
	userNames := s.lookupNames("users", collect(rows, func(row Log) uuid.UUID { return row.UserID }))
	entityNames := s.lookupEntityNames(rows)
	for i, row := range rows {
		out[i].UserName = userNames[row.UserID]
		out[i].EntityName = entityNames[row.EntityType+":"+row.EntityID.String()]
	}
	return out
}

func (s *Service) lookupEntityNames(rows []Log) map[string]string {
	names := map[string]string{}
	put := func(entityType string, found map[uuid.UUID]string) {
		for id, name := range found {
			names[entityType+":"+id.String()] = name
		}
	}
	put("User", s.lookupNames("users", idsFor(rows, "User")))
	put("Project", s.lookupNames("projects", idsFor(rows, "Project")))
	put("Task", s.lookupNames("tasks", idsFor(rows, "Task")))
	put("ProjectMember", s.lookupJoined(`
		SELECT pm.id AS id, COALESCE(u.name, '') AS name
		FROM project_members pm
		LEFT JOIN users u ON u.id = pm.user_id
		WHERE pm.id IN ?`, idsFor(rows, "ProjectMember")))
	put("ResourceAllocation", s.lookupJoined(`
		SELECT ra.id AS id, COALESCE(NULLIF(p.name, ''), NULLIF(u.name, ''), '') AS name
		FROM resource_allocations ra
		LEFT JOIN projects p ON p.id = ra.project_id
		LEFT JOIN users u ON u.id = ra.user_id
		WHERE ra.id IN ?`, idsFor(rows, "ResourceAllocation")))
	return names
}

func (s *Service) lookupNames(table string, ids []uuid.UUID) map[uuid.UUID]string {
	found := map[uuid.UUID]string{}
	if len(ids) == 0 {
		return found
	}
	var rows []namedRow
	if err := s.db.Table(table).Select("id, name").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return found
	}
	for _, row := range rows {
		found[row.ID] = row.Name
	}
	return found
}

func (s *Service) lookupJoined(query string, ids []uuid.UUID) map[uuid.UUID]string {
	found := map[uuid.UUID]string{}
	if len(ids) == 0 {
		return found
	}
	var rows []namedRow
	if err := s.db.Raw(query, ids).Scan(&rows).Error; err != nil {
		return found
	}
	for _, row := range rows {
		found[row.ID] = row.Name
	}
	return found
}

func idsFor(rows []Log, entityType string) []uuid.UUID {
	return collect(rows, func(row Log) uuid.UUID {
		if row.EntityType != entityType {
			return uuid.Nil
		}
		return row.EntityID
	})
}

func collect(rows []Log, pick func(Log) uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0)
	for _, row := range rows {
		id := pick(row)
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
