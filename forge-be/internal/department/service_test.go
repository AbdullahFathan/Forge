package department

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/pkg/apperr"
)

type fakeStore struct {
	byName map[string]uuid.UUID
	byID   map[uuid.UUID]*Department
}

func newFake() *fakeStore {
	return &fakeStore{byName: map[string]uuid.UUID{}, byID: map[uuid.UUID]*Department{}}
}

func (f *fakeStore) Create(d *Department) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	f.byID[d.ID] = d
	f.byName[d.Name] = d.ID
	return nil
}

func (f *fakeStore) Save(d *Department) error {
	f.byID[d.ID] = d
	f.byName[d.Name] = d.ID
	return nil
}

func (f *fakeStore) GetByID(id uuid.UUID) (*Department, error) {
	d, ok := f.byID[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return d, nil
}

func (f *fakeStore) NameTaken(name string, excludeID *uuid.UUID) (bool, error) {
	id, ok := f.byName[name]
	if !ok {
		return false, nil
	}
	if excludeID != nil && id == *excludeID {
		return false, nil
	}
	return true, nil
}

func (f *fakeStore) List(ListFilter) ([]Department, int64, error) {
	return nil, 0, nil
}

func (f *fakeStore) SoftDelete(id uuid.UUID) error {
	delete(f.byID, id)
	return nil
}

func TestDepartmentUniqueName(t *testing.T) {
	svc := NewService(newFake())
	d, err := svc.Create("Engineering")
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, d.ID)

	_, err = svc.Create("Engineering")
	ae, ok := apperr.As(err)
	require.True(t, ok)
	require.Equal(t, apperr.ErrConflict.Code, ae.Code)
}
