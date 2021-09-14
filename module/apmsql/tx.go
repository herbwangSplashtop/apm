// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apmsql // import "go.elastic.co/apm/module/apmsql"

import (
	"context"
	"database/sql/driver"
)

func newTx(in driver.Tx, conn *conn) driver.Tx {
	tx := &tx{
		Tx:   in,
		conn: conn,
	}
	tx.PrepareContext = conn.PrepareContext
	return tx
}

type tx struct {
	driver.Tx
	conn *conn

	connPrepareContext driver.ConnPrepareContext
}

func (t *tx) PrepareContext(ctx context.Context, query string) (_ driver.Stmt, resultError error) {
	var stmt driver.Stmt
	var err error
	if t.connPrepareContext != nil {
		stmt, err = t.connPrepareContext.PrepareContext(ctx, query)
	} else {
		stmt, err = t.conn.Prepare(query)
		if err == nil {
			select {
			default:
			case <-ctx.Done():
				stmt.Close()
				return nil, ctx.Err()
			}
		}
	}
	if stmt != nil {
		stmt = newStmt(stmt, t.conn, query)
	}
	return stmt, err
}

func (t *tx) Commit() error {
	return t.Tx.Commit()
}

func (t *tx) Rollback() error {
	return t.Tx.Rollback()
}
