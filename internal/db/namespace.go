package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Namespace struct {
	Key       string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	DefaultNamespace     = "default"
	DefaultNamespaceName = "Default"
)

var (
	namespaces = make(map[string]*Namespace)
)

func LoadNamespaces() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT key, name, created_at, updated_at FROM public.namespaces;"
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return err
	}
	ns, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*Namespace, error) {
		n := &Namespace{}
		err := row.Scan(&n.Key, &n.Name, &n.CreatedAt, &n.UpdatedAt)
		return n, err
	})
	if err != nil {
		return err
	}
	for _, n := range ns {
		namespaces[n.Key] = n
	}
	return nil
}

func AddNamespace(key, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.namespaces(key, name)
	VALUES ($1, $2)
	ON CONFLICT (key) DO NOTHING;`
	_, err := pool.Exec(ctx, sql, key, name)
	if err != nil {
		return err
	}
	return nil
}
