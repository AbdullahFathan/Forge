package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCaptureRecordsOnlyWhenCalled(t *testing.T) {
	c := &Capture{}
	id := uuid.New()
	require.NoError(t, c.Record(context.Background(), Record{
		ActorID: id, EntityType: "Project", EntityID: id, Action: "CREATED", After: map[string]string{"name": "A"},
	}))
	require.Len(t, c.Items, 1)
	require.Equal(t, "CREATED", c.Items[0].Action)
}

func TestMarshalRawMessageAndCSV(t *testing.T) {
	raw := json.RawMessage(`{"a":1}`)
	got := marshal(raw)
	require.JSONEq(t, `{"a":1}`, string(got))
	require.Nil(t, marshal(nil))

	var buf bytes.Buffer
	err := WriteCSV(&buf, []Public{{
		UserName:   "Amina",
		EntityName: "Ship API",
		EntityType: "Task", Action: "UPDATED", Before: json.RawMessage(`{}`), After: json.RawMessage(`{"x":1}`),
	}})
	require.NoError(t, err)
	require.Contains(t, buf.String(), "entity_name")
	require.Contains(t, buf.String(), "Amina")
	require.Contains(t, buf.String(), "Ship API")
	require.Contains(t, buf.String(), "UPDATED")
}

func TestLikeContainsEscapesWildcards(t *testing.T) {
	require.Equal(t, `%100\%\_done%`, likeContains(`100%_done`))
}

func TestPresentWithoutDBKeepsRows(t *testing.T) {
	id := uuid.New()
	got := (&Service{}).Present([]Log{{ID: id, EntityType: "Project", Action: "CREATED"}})
	require.Len(t, got, 1)
	require.Equal(t, id, got[0].ID)
	require.Empty(t, got[0].UserName)
}

func TestNoop(t *testing.T) {
	require.NoError(t, Noop{}.Record(context.Background(), Record{}))
}

func TestPtr(t *testing.T) {
	require.Nil(t, Ptr(uuid.Nil))
	id := uuid.New()
	require.Equal(t, id, *Ptr(id))
}
