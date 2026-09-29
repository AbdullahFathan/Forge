import { describe, expect, it } from "vitest"

import { fieldChanges, formatAuditAction, formatAuditValue } from "@/lib/audit-diff"

describe("formatAuditAction", () => {
  it("names the actions people see in the log", () => {
    expect(formatAuditAction("STATUS_CHANGED")).toBe("Status changed")
    expect(formatAuditAction("CREATED")).toBe("Created")
    expect(formatAuditAction("TASK_ASSIGNED")).toBe("Task assigned")
    expect(formatAuditValue("ACTIVE")).toBe("Active")
    expect(formatAuditValue("HIGH")).toBe("High")
  })
})

describe("fieldChanges", () => {
  it("keeps only fields that differ", () => {
    const changes = fieldChanges(
      { ID: "1", Name: "Rapat Kick OFF", Status: "To Do" },
      { ID: "1", Name: "Rapat Kick OFF", Status: "In Progress" },
    )
    expect(changes).toEqual([
      { key: "Status", label: "Status", before: "To Do", after: "In Progress" },
    ])
  })

  it("reads a JSON string snapshot", () => {
    const changes = fieldChanges('{"Status":"To Do"}', { Status: "Done" })
    expect(changes[0]?.after).toBe("Done")
  })

  it("summarizes a related record and omits secrets", () => {
    const changes = fieldChanges(null, {
      Name: "Hidup",
      DeletedAt: null,
      PasswordHash: "secret-hash",
      Owner: { Name: "Super Admin", PasswordHash: "secret-hash", Email: "admin@workspace.local" },
      OwnerID: "92bd59f4-6866-48c5-85a4-4615799fbc9b",
    })
    expect(changes.map((change) => change.key)).toEqual(["Name", "Owner"])
    expect(changes[1]?.after).toBe("Super Admin")
  })

  it("shows a created record as after-only values", () => {
    const changes = fieldChanges(null, { Name: "Kickoff" })
    expect(changes).toEqual([{ key: "Name", label: "Name", before: "—", after: "Kickoff" }])
  })
})
