package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Jeanpigi/blog/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	dbUser string
	dbPass string
	dbHost string
	dbName string

	// Db es una variable global que mantendrá la conexión a la base de datos
	Db *sql.DB
)

// InitDB inicializa la conexión a la base de datos
func InitDB() {
	// Verificar si existe un archivo .env antes de intentar cargarlo
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ No se pudo cargar el archivo .env, se usarán las variables del sistema.")
		}
	} else {
		log.Println("⚠️ No se encontró el archivo .env, usando variables de entorno del sistema.")
	}

	// Obtener variables de entorno (ya sea desde .env o el sistema)
	dbUser = os.Getenv("MYSQL_USER")
	dbPass = os.Getenv("MYSQL_ROOT_PASSWORD")
	dbHost = os.Getenv("MYSQL_HOST")
	dbName = os.Getenv("MYSQL_DATABASE")

	// Formatear la conexión
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?tls=false", dbUser, dbPass, dbHost, dbName)

	// Intentar conectar a la base de datos
	var err error
	Db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("❌ Error al conectar a la base de datos:", err)
	}

	// Configurar el pool de conexiones
	Db.SetMaxOpenConns(25)
	Db.SetMaxIdleConns(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Conexión a la base de datos establecida.")
}

// CloseDB cierra la conexión a la base de datos
func CloseDB() {
	if Db != nil {
		err := Db.Close()
		if err != nil {
			log.Println("❌ Error al cerrar la conexión a la base de datos:", err)
		}
		log.Println("✅ Conexión a la base de datos cerrada.")
	}
}

func GetUserByUsername(username string) (*models.User, error) {
	query := "SELECT * FROM Users WHERE Username = ?"
	row := Db.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // El usuario no existe
		}
		log.Println("Error al obtener usuario de la base de datos:", err)
		return nil, err
	}

	return user, nil
}

func InsertUser(user *models.User) error {
	query := "INSERT INTO Users (Username, Password) VALUES (?, ?)"
	_, err := Db.Exec(query, user.Username, user.Password)
	if err != nil {
		log.Println("Error al insertar usuario en la base de datos:", err)
		return err
	}

	return nil
}

func GetAllPosts() ([]*models.Post, error) {
	// IMPORTANTE: agrega el ORDER BY CreatedAt DESC
	query := "SELECT * FROM Posts ORDER BY CreatedAt DESC"
	rows, err := Db.Query(query)
	if err != nil {
		log.Println("Error al obtener posts de la base de datos:", err)
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Description,
			&post.Content,
			&post.AuthorID,
			&post.CreatedAt,
			&post.Categoria,
		)
		if err != nil {
			log.Println("Error al escanear post:", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error al leer las filas de posts:", err)
		return nil, err
	}

	return posts, nil
}

func FindPostByID(postID string) (*models.Post, error) {
	query := "SELECT * FROM Posts WHERE ID = ?"
	row := Db.QueryRow(query, postID)

	var post models.Post
	err := row.Scan(&post.ID, &post.Title, &post.Description, &post.Content, &post.AuthorID, &post.CreatedAt, &post.Categoria)
	if err != nil {
		log.Printf("Error al buscar el post con ID %s: %v", postID, err)
		return nil, err
	}

	return &post, nil
}

func FindPostsByCategory(categoria string) ([]*models.Post, error) {
	var posts []*models.Post
	query := "SELECT * FROM Posts WHERE Category = ? ORDER BY CreatedAt DESC"
	rows, err := Db.Query(query, categoria)
	if err != nil {
		log.Printf("Error al buscar posts con la categoría %s: %v", categoria, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Description, &post.Content, &post.AuthorID, &post.CreatedAt, &post.Categoria); err != nil {
			log.Printf("Error al escanear post: %v", err)
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error post rows: %v", err)
		return nil, err
	}

	return posts, nil
}

func InsertPost(post *models.Post) error {
	query := "INSERT INTO Posts (Title, Description, Content, AuthorID, CreatedAt, Category) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := Db.Exec(query, post.Title, post.Description, post.Content, post.AuthorID, post.CreatedAt, post.Categoria)
	if err != nil {
		log.Println("Error al insertar el post en la base de datos:", err)
		return err
	}

	return nil
}

func UpdatePost(postID int, post *models.Post) error {
	query := "UPDATE Posts SET Title = ?, Description = ?, Content = ?, Category = ?, AuthorID = ? WHERE id = ?"
	_, err := Db.Exec(query, post.Title, post.Description, post.Content, post.Categoria, post.AuthorID, postID)
	if err != nil {
		log.Println("Error al actualizar post en la base de datos:", err)
		return err
	}

	return nil
}

func DeletePost(postID int) error {
	query := "DELETE FROM Posts WHERE ID = ?"
	_, err := Db.Exec(query, postID)
	if err != nil {
		log.Println("Error al eliminar post en la base de datos:", err)
		return err
	}

	return nil
}

// InsertVisit guarda una visita en la base de datos
func InsertVisit(visit *models.Visit) error {
	query := `
	INSERT INTO Visits (ip, user_agent, page, country, region, city, latitude, longitude)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, visit.IP, visit.UserAgent, visit.Page, visit.Country, visit.Region, visit.City, visit.Latitude, visit.Longitude)
	if err != nil {
		log.Println("❌ Error al registrar visita en la base de datos:", err)
		return err
	}
	return nil
}

// GetAllVisits obtiene las últimas visitas
func GetAllVisits() ([]models.Visit, error) {
	query := "SELECT id, DATE_FORMAT(timestamp, '%Y-%m-%d %H:%i:%s'), ip, user_agent, page, country, region, city, latitude, longitude FROM Visits ORDER BY timestamp DESC LIMIT 20"
	rows, err := Db.Query(query)
	if err != nil {
		log.Println("❌ Error al obtener visitas:", err)
		return nil, err
	}
	defer rows.Close()

	var visits []models.Visit
	for rows.Next() {
		var visit models.Visit
		var timestamp string // Cambiar de time.Time a string

		// Escanear datos con timestamp como string
		err := rows.Scan(&visit.ID, &timestamp, &visit.IP, &visit.UserAgent, &visit.Page, &visit.Country, &visit.Region, &visit.City, &visit.Latitude, &visit.Longitude)
		if err != nil {
			log.Println("❌ Error al escanear visita:", err)
			continue
		}

		// Convertir timestamp string a time.Time
		visit.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			log.Println("⚠️ Error al convertir timestamp:", err)
			continue
		}

		visits = append(visits, visit)
	}

	return visits, nil
}
