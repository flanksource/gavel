package database

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSharedReturnsOneProcessHandle(t *testing.T) {
	require.NoError(t, closeShared())
	t.Cleanup(func() { require.NoError(t, closeShared()) })
	clearDatabaseEnvironment(t)

	const callers = 16
	got := make([]*DB, callers)
	errCh := make(chan error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var err error
			got[i], err = Shared(t.Context())
			errCh <- err
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
	for i := 1; i < callers; i++ {
		require.Same(t, got[0], got[i])
	}
	require.True(t, got[0].Disabled())
	require.NoError(t, got[0].Close(), "consumers cannot close the shared handle")
	require.Same(t, got[0], mustShared(t))
}

func TestRequireRejectsDisabledDatabase(t *testing.T) {
	require.NoError(t, closeShared())
	t.Cleanup(func() { require.NoError(t, closeShared()) })
	clearDatabaseEnvironment(t)

	_, err := Require(t.Context(), "native TODO persistence")
	require.ErrorIs(t, err, ErrUnavailable)
	require.Contains(t, err.Error(), EnvDSN)
	require.Contains(t, err.Error(), "gavel system install --embedded")
	require.True(t, errors.Is(err, ErrUnavailable))
}

func TestResetDisabledSharedForTestRejectsEnabledHandle(t *testing.T) {
	require.NoError(t, closeShared())
	processDB.Lock()
	processDB.db = &DB{gorm: &gorm.DB{}, shared: true}
	processDB.Unlock()
	t.Cleanup(func() {
		processDB.Lock()
		processDB.db = nil
		processDB.Unlock()
	})

	err := ResetDisabledSharedForTest()
	require.EqualError(t, err, "cannot reset an enabled shared database")
	processDB.Lock()
	require.NotNil(t, processDB.db, "the enabled process handle must remain registered")
	processDB.Unlock()
}

func mustShared(t *testing.T) *DB {
	t.Helper()
	db, err := Shared(t.Context())
	require.NoError(t, err)
	return db
}
