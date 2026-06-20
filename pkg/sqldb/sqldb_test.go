package sqldb_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
	"github.com/PapaDanielVi/apadana/v2/pkg/sqldb"
)

// fakeDriver is a no-op driver so sql.Open succeeds without a real database.
// sql.Open is lazy, so Open is never called in these tests.
type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("not implemented")
}

const driverName = "sqldb-fake"

// registerDriver registers the fake driver exactly once across all tests.
var registerDriver = sync.OnceFunc(func() {
	sql.Register(driverName, fakeDriver{})
})

func TestManager_Get_PerTenant(t *testing.T) {
	t.Parallel()
	registerDriver()

	mgr := sqldb.NewManager(driverName, map[string]string{
		"acme":   "dsn-acme",
		"globex": "dsn-globex",
	})
	t.Cleanup(func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	acme, err := mgr.Get(mt.InjectTID(context.Background(), "acme"))
	if err != nil {
		t.Fatalf("Get(acme) error = %v", err)
	}
	globex, err := mgr.Get(mt.InjectTID(context.Background(), "globex"))
	if err != nil {
		t.Fatalf("Get(globex) error = %v", err)
	}
	if acme == globex {
		t.Error("each tenant should get a distinct *sql.DB")
	}

	// Same tenant returns the same handle.
	acme2, err := mgr.Get(mt.InjectTID(context.Background(), "acme"))
	if err != nil {
		t.Fatalf("Get(acme) second call error = %v", err)
	}
	if acme != acme2 {
		t.Error("same tenant should reuse the same *sql.DB")
	}
}
