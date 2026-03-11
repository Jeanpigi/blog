// Herramienta de diagnóstico de conexión a la base de datos.
// Uso: go run ./cmd/dbcheck/
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️  No se pudo cargar .env:", err)
		} else {
			log.Println("✅ .env cargado")
		}
	}

	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_ROOT_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbName := os.Getenv("MYSQL_DATABASE")
	dbPort := os.Getenv("MYSQL_PORTDATABASE")
	if dbPort == "" {
		dbPort = "3306"
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════")
	fmt.Println("   Diagnóstico de conexión a DB")
	fmt.Println("══════════════════════════════════════")
	fmt.Printf("  Host:     %s\n", dbHost)
	fmt.Printf("  Puerto:   %s\n", dbPort)
	fmt.Printf("  Usuario:  %s\n", dbUser)
	fmt.Printf("  Base:     %s\n", dbName)
	fmt.Printf("  Password: %s*** (%d chars)\n", maskPass(dbPass), len(dbPass))
	fmt.Println("══════════════════════════════════════")

	if dbUser == "" || dbHost == "" || dbName == "" {
		log.Fatal("❌ Variables faltantes (MYSQL_USER / MYSQL_HOST / MYSQL_DATABASE)")
	}

	addr := fmt.Sprintf("%s:%s", dbHost, dbPort)

	// Intento 1: configuración estándar
	fmt.Println("\n[1/3] Probando configuración estándar...")
	cfg1 := mysql.Config{
		User:                 dbUser,
		Passwd:               dbPass,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbName,
		ParseTime:            true,
		AllowNativePasswords: true,
	}
	if ok, err := tryConnect(cfg1.FormatDSN()); ok {
		fmt.Println("     ✅ CONECTADO con configuración estándar")
		runChecks(cfg1.FormatDSN(), dbName)
		return
	} else {
		fmt.Printf("     ❌ Fallo: %v\n", err)
	}

	// Intento 2: con cleartext passwords (algunos auth plugins lo requieren)
	fmt.Println("\n[2/3] Probando con cleartext passwords...")
	cfg2 := mysql.Config{
		User:                    dbUser,
		Passwd:                  dbPass,
		Net:                     "tcp",
		Addr:                    addr,
		DBName:                  dbName,
		ParseTime:               true,
		AllowNativePasswords:    true,
		AllowCleartextPasswords: true,
	}
	if ok, err := tryConnect(cfg2.FormatDSN()); ok {
		fmt.Println("     ✅ CONECTADO con cleartext passwords")
		runChecks(cfg2.FormatDSN(), dbName)
		return
	} else {
		fmt.Printf("     ❌ Fallo: %v\n", err)
	}

	// Intento 3: sin DBName (a veces el usuario no tiene permisos sobre esa base específica)
	fmt.Println("\n[3/3] Probando sin seleccionar base de datos...")
	cfg3 := mysql.Config{
		User:                    dbUser,
		Passwd:                  dbPass,
		Net:                     "tcp",
		Addr:                    addr,
		ParseTime:               true,
		AllowNativePasswords:    true,
		AllowCleartextPasswords: true,
	}
	if ok, err := tryConnect(cfg3.FormatDSN()); ok {
		fmt.Println("     ✅ Servidor alcanzable, pero sin acceso a la base de datos.")
		fmt.Println("     → Verifica que el usuario tiene permisos en:", dbName)
	} else {
		fmt.Printf("     ❌ Fallo: %v\n", err)
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════")
	fmt.Println("  DIAGNÓSTICO:")
	fmt.Println("  Corre esto en tu cliente MariaDB:")
	fmt.Printf("  SELECT user, plugin, host FROM mysql.user WHERE user = '%s';\n", dbUser)
	fmt.Println()
	fmt.Println("  Si el plugin es 'ed25519', cámbialo con:")
	fmt.Printf("  ALTER USER '%s'@'%%'\n", dbUser)
	fmt.Println("    IDENTIFIED VIA mysql_native_password")
	fmt.Println("    USING PASSWORD('tu-contraseña');")
	fmt.Println("  FLUSH PRIVILEGES;")
	fmt.Println("══════════════════════════════════════")
	os.Exit(1)
}

func tryConnect(dsn string) (bool, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Second)
	if err := db.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

func runChecks(dsn, dbName string) {
	db, _ := sql.Open("mysql", dsn)
	defer db.Close()

	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	fmt.Printf("     Versión: %s\n", version)

	var plugin string
	db.QueryRow("SELECT plugin FROM mysql.user WHERE user = CURRENT_USER()").Scan(&plugin)
	if plugin != "" {
		fmt.Printf("     Auth plugin: %s\n", plugin)
		if plugin == "ed25519" {
			fmt.Println("     ⚠️  Plugin ed25519 detectado — puede causar problemas")
			fmt.Println("     → Cambia a mysql_native_password (ver instrucciones al final)")
		}
	}

	tables := []string{"Users", "Posts", "Visits"}
	fmt.Println("     Tablas:")
	for _, t := range tables {
		var n int
		err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name=?", dbName, t).Scan(&n)
		if err != nil || n == 0 {
			fmt.Printf("       ⚠️  '%s' — no encontrada\n", t)
		} else {
			var rows int
			db.QueryRow("SELECT COUNT(*) FROM `" + t + "`").Scan(&rows)
			fmt.Printf("       ✅ '%s' — %d filas\n", t, rows)
		}
	}
	fmt.Println()
	fmt.Println("  🎉 Conexión exitosa")
	fmt.Println("══════════════════════════════════════")
}

func maskPass(p string) string {
	if len(p) <= 3 {
		return "***"
	}
	return p[:3]
}
