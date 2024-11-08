package postgres

import (
	"fmt"
	"path/filepath"
)

var (
	migrationsPath = fmt.Sprintf("file://%s", filepath.Join("..", "..", "..", "migrate", "migrations"))
	seedPath       = filepath.Join("..", "..", "..", "migrate", "tests", "unit_seed.sql")
)
