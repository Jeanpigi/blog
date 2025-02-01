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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error al cargar el archivo .env:", err)
	}

	dbUser = os.Getenv("MYSQL_USER")
	dbPass = os.Getenv("MYSQL_ROOT_PASSWORD")
	dbHost = os.Getenv("MYSQL_HOST")
	dbName = os.Getenv("MYSQL_DATABASE")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?tls=false", dbUser, dbPass, dbHost, dbName)
	Db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}

	// Configura el pool de conexiones
	Db.SetMaxOpenConns(25)
	Db.SetMaxIdleConns(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Conexión a la base de datos establecida.")
}

// CloseDB cierra la conexión a la base de datos
func CloseDB() {
	if Db != nil {
		err := Db.Close()
		if err != nil {
			log.Println("Error al cerrar la conexión a la base de datos:", err)
		}
		log.Println("Conexión a la base de datos cerrada.")
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
	query := "SELECT * FROM Posts"
	rows, err := Db.Query(query)
	if err != nil {
		log.Println("Error al obtener posts de la base de datos:", err)
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Description, &post.Content, &post.AuthorID, &post.CreatedAt, &post.Categoria)
		if err != nil {
			log.Println("Error al escanear post:", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error al obtener posts de la base de datos:", err)
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
	query := "SELECT * FROM Posts WHERE categoria = ?"
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
