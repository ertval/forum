package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"forum/internal/platform/database"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./data/forum.db"
	conn, err := database.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer conn.Close()

	// Load DB schema
	dbTables, err := loadDBSchema(conn.DB())
	if err != nil {
		log.Fatalf("failed to load DB schema: %v", err)
	}

	// Parse migrations
	migrationsDir := "./migrations"
	migSchema, err := parseMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("failed to parse migrations: %v", err)
	}

	// Compare
	report := compareSchemas(dbTables, migSchema)

	// Print report
	ok := true
	if len(report.tablesOnlyInDB) > 0 || len(report.tablesOnlyInMigrations) > 0 || len(report.extraColumns) > 0 || len(report.missingColumns) > 0 {
		ok = false
	}

	fmt.Println("Schema comparison report:")
	fmt.Println()

	if len(report.tablesOnlyInDB) > 0 {
		fmt.Println("Tables present in DB but not declared in migrations:")
		for _, t := range report.tablesOnlyInDB {
			fmt.Printf(" - %s\n", t)
		}
		fmt.Println()
	}

	if len(report.tablesOnlyInMigrations) > 0 {
		fmt.Println("Tables declared in migrations but missing in DB:")
		for _, t := range report.tablesOnlyInMigrations {
			fmt.Printf(" - %s\n", t)
		}
		fmt.Println()
	}

	if len(report.extraColumns) > 0 {
		fmt.Println("Columns present in DB but not declared in migrations:")
		for table, cols := range report.extraColumns {
			fmt.Printf(" - %s: %s\n", table, strings.Join(cols, ", "))
		}
		fmt.Println()
	}

	if len(report.missingColumns) > 0 {
		fmt.Println("Columns declared in migrations but missing in DB:")
		for table, cols := range report.missingColumns {
			fmt.Printf(" - %s: %s\n", table, strings.Join(cols, ", "))
		}
		fmt.Println()
	}

	if ok {
		fmt.Println("OK: Database schema matches migrations (at least for tables/columns).")
		os.Exit(0)
	} else {
		fmt.Println("MISMATCH: See items above. Decide whether to apply missing migrations, mark migrations as applied, or create cleanup migrations.")
		os.Exit(2)
	}
}

// loadDBSchema returns map[table]map[column]true
func loadDBSchema(db *sql.DB) (map[string]map[string]bool, error) {
	tables := map[string]map[string]bool{}
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols := map[string]bool{}
		pragma := fmt.Sprintf("PRAGMA table_info('%s');", name)
		cr, err := db.Query(pragma)
		if err != nil {
			return nil, err
		}
		for cr.Next() {
			var cid int
			var col string
			var ctype string
			var notnull int
			var dflt interface{}
			var pk int
			if err := cr.Scan(&cid, &col, &ctype, &notnull, &dflt, &pk); err != nil {
				cr.Close()
				return nil, err
			}
			cols[col] = true
		}
		cr.Close()
		tables[name] = cols
	}
	return tables, nil
}

// parseMigrations extracts CREATE TABLE and ALTER TABLE ... ADD COLUMN from Up sections
func parseMigrations(dir string) (map[string]map[string]bool, error) {
	m := map[string]map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if !strings.HasSuffix(strings.ToLower(base), ".sql") {
			return nil
		}
		// skip templates and guide files
		if base == "000_template_migration.sql" || strings.HasPrefix(base, "MIGRATIONS_GUIDE") || strings.HasPrefix(base, "README") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		up := extractUpSQL(string(b))
		if up == "" {
			return nil
		}

		// find CREATE TABLE blocks (robust: allow multiline column lists)
		// pattern captures table name and inner column definitions; allow optional IF NOT EXISTS
		createRe := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:["'` + "`" + `]?([A-Za-z0-9_]+)["'` + "`" + `]?)\s*\(([\s\S]*?)\)`)
		matches := createRe.FindAllStringSubmatch(up, -1)
		for _, mm := range matches {
			name := mm[1]
			colsDef := mm[2]
			cols := extractColumnsFromCreate(colsDef)
			if _, ok := m[name]; !ok {
				m[name] = map[string]bool{}
			}
			for _, c := range cols {
				m[name][c] = true
			}
		}

		// find ALTER TABLE ... ADD COLUMN
		alterRe := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+([A-Za-z0-9_]+)\s+ADD\s+COLUMN\s+([A-Za-z0-9_]+)`)
		amatches := alterRe.FindAllStringSubmatch(up, -1)
		for _, am := range amatches {
			name := am[1]
			col := am[2]
			if _, ok := m[name]; !ok {
				m[name] = map[string]bool{}
			}
			m[name][col] = true
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func extractUpSQL(content string) string {
	upMarker := "-- +migrate Up"
	downMarker := "-- +migrate Down"
	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return ""
	}
	upIdx += len(upMarker)
	downIdx := strings.Index(content, downMarker)
	if downIdx == -1 {
		downIdx = len(content)
	}
	return strings.TrimSpace(content[upIdx:downIdx])
}

func extractColumnsFromCreate(block string) []string {
	// naive split on commas, then extract first token as column name when appropriate
	lines := strings.Split(block, ",")
	cols := []string{}
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		// skip table constraints like PRIMARY KEY(...), UNIQUE(...), FOREIGN KEY
		up := strings.ToUpper(l)
		if strings.HasPrefix(up, "PRIMARY KEY") || strings.HasPrefix(up, "UNIQUE") || strings.HasPrefix(up, "FOREIGN KEY") || strings.HasPrefix(up, "CHECK") {
			continue
		}
		// first token before a space is likely the column name; strip backticks/quotes
		parts := strings.Fields(l)
		if len(parts) == 0 {
			continue
		}
		name := parts[0]
		name = strings.Trim(name, "`\"' ")
		// remove potential trailing '(' from constraints
		if idx := strings.Index(name, "("); idx != -1 {
			name = name[:idx]
		}
		cols = append(cols, name)
	}
	return cols
}

type compareReport struct {
	tablesOnlyInDB         []string
	tablesOnlyInMigrations []string
	extraColumns           map[string][]string
	missingColumns         map[string][]string
}

func compareSchemas(db map[string]map[string]bool, mig map[string]map[string]bool) compareReport {
	rep := compareReport{
		extraColumns:   map[string][]string{},
		missingColumns: map[string][]string{},
	}

	// tables only in DB
	for t := range db {
		if _, ok := mig[t]; !ok {
			rep.tablesOnlyInDB = append(rep.tablesOnlyInDB, t)
		}
	}
	// tables only in migrations
	for t := range mig {
		if _, ok := db[t]; !ok {
			rep.tablesOnlyInMigrations = append(rep.tablesOnlyInMigrations, t)
		}
	}
	// columns differences
	for t, dbCols := range db {
		if migCols, ok := mig[t]; ok {
			// extra in db
			for c := range dbCols {

				if !migCols[c] {
					rep.extraColumns[t] = append(rep.extraColumns[t], c)
				}
			}
			// missing in db
			for c := range migCols {
				if !dbCols[c] {
					rep.missingColumns[t] = append(rep.missingColumns[t], c)
				}
			}
		}
	}

	return rep
}
