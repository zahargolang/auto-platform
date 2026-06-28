package auth_postgres_repository

import (
	"context"
	"fmt"
	"reflect"
	"time"

	core_postgres_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool"
)

// fakeRow — ручная fake-реализация core_postgres_pool.Row. Scan копирует
// заранее заданные значения в переданные указатели через reflect — так же,
// как это в реальности делает pgx.Row, но без поднятия настоящей БД.
type fakeRow struct {
	values []any
	err    error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("fakeRow.Scan: got %d destinations, want %d", len(dest), len(r.values))
	}
	for i, d := range dest {
		dv := reflect.ValueOf(d)
		if dv.Kind() != reflect.Pointer {
			return fmt.Errorf("fakeRow.Scan: destination %d is not a pointer: %T", i, d)
		}
		dv.Elem().Set(reflect.ValueOf(r.values[i]))
	}
	return nil
}

// fakeCommandTag — fake CommandTag для Exec().
type fakeCommandTag struct {
	rows int64
}

func (c *fakeCommandTag) RowsAffected() int64 {
	return c.rows
}

// fakePool — fake core_postgres_pool.Pool. Репозитории auth-service
// используют только QueryRow и Exec — Query() здесь не нужен ни одному
// тесту, поэтому он паникует, если вдруг понадобится (сигнал, что фейк
// нужно расширить).
type fakePool struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error)
}

func (p *fakePool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	panic("fakePool.Query: not used by any auth-service repository method")
}

func (p *fakePool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return p.queryRowFunc(ctx, sql, args...)
}

func (p *fakePool) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	return p.execFunc(ctx, sql, args...)
}

func (p *fakePool) Close() {}

func (p *fakePool) OpTimeout() time.Duration {
	return time.Second
}
