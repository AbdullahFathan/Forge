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
	err := WriteCSV(&buf, []Log{{
		UserID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		EntityID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		EntityType: "Task", Action: "UPDATED", Before: json.RawMessage(`{}`), After: json.RawMessage(`{"x":1}`),
	}})
	require.NoError(t, err)
	require.Contains(t, buf.String(), "entity_type")
	require.Contains(t, buf.String(), "UPDATED")
}

func TestNoop(t *testing.T) {
	require.NoError(t, Noop{}.Record(context.Background(), Record{}))
}

func TestPtr(t *testing.T) {
	require.Nil(t, Ptr(uuid.Nil))
	id := uuid.New()
	require.Equal(t, id, *Ptr(id))
}
