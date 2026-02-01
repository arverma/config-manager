package httpapi

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

func storeListNamespaces(ctx context.Context, q querier, limit, offset int) ([]NamespaceWithCount, error) {
	rows, err := q.Query(ctx, `
		SELECT
			n.id, n.name, n.created_at, n.updated_at,
			COUNT(c.id) AS config_count
		FROM namespaces n
		LEFT JOIN configs c
			ON c.namespace = n.name AND c.deleted_at IS NULL
		GROUP BY n.id, n.name, n.created_at, n.updated_at
		ORDER BY n.name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]NamespaceWithCount, 0, limit)
	for rows.Next() {
		var id pgtype.UUID
		var n NamespaceWithCount
		var createdAt, updatedAt pgtype.Timestamptz
		var count int64
		if err := rows.Scan(&id, &n.Name, &createdAt, &updatedAt, &count); err != nil {
			return nil, err
		}
		n.ID = uuidToString(id)
		n.CreatedAt = createdAt.Time
		n.UpdatedAt = updatedAt.Time
		n.ConfigCount = int(count)
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func storeNamespaceExists(ctx context.Context, q querier, name string) (bool, error) {
	var ok bool
	if err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM namespaces WHERE name = $1)`, name).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}
