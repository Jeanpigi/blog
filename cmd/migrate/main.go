// Herramienta de base de datos: copias de seguridad y migraciones.
//
// Lee el .env con godotenv, igual que la app. Así funciona aunque la contraseña
// tenga caracteres especiales (?, #, $…) que bash interpretaría mal al hacer
// `source .env`. La contraseña nunca aparece en la línea de comandos.
//
// Uso (desde la carpeta del proyecto):
//
//	go run ./cmd/migrate status            # qué migraciones faltan
//	go run ./cmd/migrate backup            # copia de Visits, projects, Posts y Users
//	go run ./cmd/migrate backup Visits     # copia de tablas concretas
//	go run ./cmd/migrate up                # aplica las migraciones pendientes
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const migrationsDir = "db/migrations"

type config struct {
	user, pass, host, port, name string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
	}
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("❌ No se pudo leer .env: %v", err)
		}
	}
	cfg := config{
		user: os.Getenv("MYSQL_USER"),
		pass: os.Getenv("MYSQL_ROOT_PASSWORD"),
		host: os.Getenv("MYSQL_HOST"),
		port: os.Getenv("MYSQL_PORTDATABASE"),
		name: os.Getenv("MYSQL_DATABASE"),
	}
	if cfg.port == "" {
		cfg.port = "3306"
	}
	if cfg.user == "" || cfg.host == "" || cfg.name == "" {
		log.Fatal("❌ Faltan MYSQL_USER, MYSQL_HOST o MYSQL_DATABASE en el .env")
	}

	switch os.Args[1] {
	case "status":
		db := open(cfg)
		defer db.Close()
		pending := pendingMigrations(db)
		for _, f := range allMigrations() {
			mark := "✅ aplicada "
			if contains(pending, f) {
				mark = "⏳ pendiente"
			}
			fmt.Printf("  %s  %s\n", mark, f)
		}
	case "up":
		db := open(cfg)
		defer db.Close()
		up(db)
	case "backup":
		tables := os.Args[2:]
		if len(tables) == 0 {
			tables = []string{"Visits", "projects", "Posts", "Users"}
		}
		backup(cfg, tables)
	default:
		usage()
	}
}

func usage() {
	log.Fatal("Uso: go run ./cmd/migrate [status | backup [tablas...] | up]")
}

func open(c config) *sql.DB {
	mc := mysql.Config{
		User: c.user, Passwd: c.pass, Net: "tcp", Addr: c.host + ":" + c.port, DBName: c.name,
		AllowNativePasswords: true,
		MultiStatements:      true, // un archivo .sql puede tener varias sentencias
		Timeout:              10 * time.Second,
		Params:               map[string]string{"charset": "utf8mb4"},
	}
	db, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ No se pudo conectar a la base de datos: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(255) PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		log.Fatalf("❌ No se pudo crear schema_migrations: %v", err)
	}
	return db
}

func allMigrations() []string {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	for i := range files {
		files[i] = filepath.Base(files[i])
	}
	sort.Strings(files)
	return files
}

func pendingMigrations(db *sql.DB) []string {
	applied := map[string]bool{}
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			log.Fatalf("❌ %v", err)
		}
		applied[v] = true
	}
	var pending []string
	for _, f := range allMigrations() {
		if !applied[f] {
			pending = append(pending, f)
		}
	}
	return pending
}

func up(db *sql.DB) {
	pending := pendingMigrations(db)
	if len(pending) == 0 {
		log.Println("✅ No hay migraciones pendientes.")
		return
	}
	for _, f := range pending {
		body, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			log.Fatalf("❌ %v", err)
		}
		log.Printf("→ Aplicando %s…", f)
		if _, err := db.Exec(string(body)); err != nil {
			log.Fatalf("❌ %s falló: %v\n   No se aplicaron las siguientes. Revisa el error antes de reintentar.", f, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, f); err != nil {
			log.Fatalf("❌ %s se aplicó pero no se pudo registrar: %v", f, err)
		}
		log.Printf("  ✅ %s", f)
	}
}

func backup(c config, tables []string) {
	dump, err := exec.LookPath("mariadb-dump")
	if err != nil {
		if dump, err = exec.LookPath("mysqldump"); err != nil {
			log.Fatal("❌ No se encontró mariadb-dump ni mysqldump")
		}
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Fatalf("❌ %v", err)
	}
	out := filepath.Join(dir, fmt.Sprintf("blog-%s.sql", time.Now().Format("20060102-150405")))
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer f.Close()

	args := append([]string{
		"-h", c.host, "-P", c.port, "-u", c.user,
		"--single-transaction", // copia consistente sin bloquear tablas
		"--no-tablespaces",     // los hostings compartidos no dan el permiso PROCESS
		c.name,
	}, tables...)
	cmd := exec.Command(dump, args...) //nolint:gosec // binario y argumentos controlados por la herramienta
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+c.pass)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(out)
		log.Fatalf("❌ La copia falló: %v", err)
	}
	st, _ := f.Stat()
	log.Printf("✅ Copia guardada en %s (%s, %d KB)", out, strings.Join(tables, ", "), st.Size()/1024)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
