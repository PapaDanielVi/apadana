// Command sqldb demonstrates a per-tenant database/sql connection manager.
//
// Each tenant maps to its own DSN, and the manager opens and caches one
// *sql.DB per tenant on first use. This example registers a tiny no-op driver
// so it runs without any external dependency. In a real program you would
// import a real driver, for example:
//
//	import _ "github.com/mattn/go-sqlite3" // or modernc.org/sqlite
//
// and pass its driver name to NewManager.
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log/slog"

	"github.com/PapaDanielVi/apadana/v2/pkg/mt"
	"github.com/PapaDanielVi/apadana/v2/pkg/sqldb"
)

// noopDriver stands in for a real driver so the example is self-contained.
type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("noop driver: connect with a real driver in production")
}

func main() {
	sql.Register("noop", noopDriver{})

	mgr := sqldb.NewManager("noop", map[string]string{
		"acme":   "file:acme.db",
		"globex": "file:globex.db",
	})
	defer func() {
		if err := mgr.Close(); err != nil {
			slog.Error("close", "error", err)
		}
	}()

	for _, tenant := range []string{"acme", "globex", "acme"} {
		ctx := mt.InjectTID(context.Background(), tenant)
		db, err := mgr.Get(ctx)
		if err != nil {
			slog.Error("get db", "tenant", tenant, "error", err)
			continue
		}
		fmt.Printf("tenant %-7s -> *sql.DB %p\n", tenant, db)
	}
}
