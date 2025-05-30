// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package pgwirecancel_test

import (
	"context"
	gosql "database/sql"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security/username"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/testutils/pgurlutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// TestCancelQuery uses the pgwire-level query cancellation protocol provided
// by lib/pq to make sure that canceling a query works correctly.
func TestCancelQuery(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	cancelCtx, cancel := context.WithCancel(context.Background())
	args := base.TestServerArgs{
		Knobs: base.TestingKnobs{
			SQLExecutor: &sql.ExecutorTestingKnobs{
				BeforeExecute: func(ctx context.Context, stmt string, descriptors *descs.Collection) {
					if strings.Contains(stmt, "pg_sleep") {
						cancel()
					}
				},
			},
		},
	}
	s, db, _ := serverutils.StartServer(t, args)
	defer s.Stopper().Stop(cancelCtx)
	defer db.Close()

	// Cancellation should stop the query.
	var b bool
	err := db.QueryRowContext(cancelCtx, "select pg_sleep(30)").Scan(&b)
	require.EqualError(t, err, "pq: query execution canceled")

	// Context is already canceled, so error should come before execution.
	var i int
	err = db.QueryRowContext(cancelCtx, "select 1").Scan(&i)
	require.EqualError(t, err, "context canceled")
}

// TestCancelQueryOtherNode uses the pgwire-level query cancellation protocol
// to make sure cancel requests are forwarded to the correct node. It sets up
// a very simple load balancer so that the cancel request is sent to a
// different node than the node with the SQL session.
func TestCancelQueryOtherNode(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx, cancel := context.WithCancel(context.Background())
	args := base.TestServerArgs{
		Knobs: base.TestingKnobs{
			SQLExecutor: &sql.ExecutorTestingKnobs{
				BeforeExecute: func(ctx context.Context, stmt string, descriptors *descs.Collection) {
					if strings.Contains(stmt, "pg_sleep") {
						cancel()
					}
				},
			},
		},
	}
	tc := serverutils.StartCluster(t, 3, base.TestClusterArgs{ServerArgs: args})
	defer tc.Stopper().Stop(ctx)

	proxy, err := net.Listen("tcp", util.TestAddr.String())
	require.NoError(t, err)

	node0, err := net.Dial("tcp", tc.Server(0).AdvSQLAddr())
	require.NoError(t, err)
	defer node0.Close()
	node1, err := net.Dial("tcp", tc.Server(1).AdvSQLAddr())
	require.NoError(t, err)
	defer node1.Close()

	gotSecondConn := false
	group := ctxgroup.WithContext(ctx)
	group.GoCtx(func(ctx context.Context) error {
		// The forwarder only expects to receive two connections: one for the
		// SQL session, and one for the cancel request. After that, the forwarder
		// stops serving connections.
		for i := 0; i < 2; i++ {
			i := i
			clientConn, err := proxy.Accept()
			if err != nil {
				return err
			}
			var crdbConn net.Conn
			if i == 0 {
				// The first connection is routed to node0.
				crdbConn = node0
			} else if i == 1 {
				// The first connection is routed to node1.
				gotSecondConn = true
				crdbConn = node1
			}
			group.GoCtx(func(ctx context.Context) error {
				return ctxgroup.GoAndWait(
					ctx,
					func(ctx context.Context) error {
						_, err := io.Copy(crdbConn, clientConn)
						crdbConn.Close()
						return err
					},
					func(ctx context.Context) error {
						_, err := io.Copy(clientConn, crdbConn)
						clientConn.Close()
						return err
					},
				)
			})
		}
		return nil
	})

	pgURL, cleanup := pgurlutils.PGUrl(
		t,
		proxy.Addr().String(),
		"TestCancelQueryOtherNode",
		url.User(username.RootUser),
	)
	defer cleanup()
	db, err := gosql.Open("postgres", pgURL.String())
	require.NoError(t, err)
	defer db.Close()

	// The cancel will be sent before the query completes.
	var b bool
	err = db.QueryRowContext(ctx, "select pg_sleep(5)").Scan(&b)
	require.EqualError(t, err, "pq: query execution canceled")

	// The simple proxy doesn't close connections cleanly, so we ignore the error
	// it returns.
	_ = group.Wait()

	// Check this after the previous goroutines finish to avoid a data race.
	require.Truef(t, gotSecondConn, "expected cancel request to arrive on a different connection")

}

// TestCancelCopyTo uses the pgwire-level query cancellation protocol provided
// by pgx to make sure that canceling COPY TO works correctly.
func TestCancelCopyTo(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx := context.Background()
	skip.UnderStress(t, "flaky")

	s := serverutils.StartServerOnly(t, base.TestServerArgs{})
	defer s.Stopper().Stop(ctx)

	pgURL, cleanup := pgurlutils.PGUrl(
		t,
		s.AdvSQLAddr(),
		"TestCancelCopyTo",
		url.User(username.RootUser),
	)
	defer cleanup()

	conn, err := pgx.Connect(ctx, pgURL.String())
	require.NoError(t, err)

	g := ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		_, err := conn.Exec(ctx, "COPY (SELECT pg_sleep(1) FROM ROWS FROM (generate_series(1, 60)) AS i) TO STDOUT")
		return err
	})

	time.Sleep(1 * time.Second)
	err = conn.PgConn().CancelRequest(ctx)
	require.NoError(t, err)

	err = g.Wait()
	require.ErrorContains(t, err, "query execution canceled")
}
