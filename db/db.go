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

	Db *sql.DB
)

func InitDB() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No se pudo cargar .env, usando variables del sistema.")
		}
	} else {
		log.Println("⚠️ No se encontró .env, usando variables del sistema.")
	}

	dbUser = os.Getenv("MYSQL_USER")
	dbPass = os.Getenv("MYSQL_ROOT_PASSWORD")
	dbHost = os.Getenv("MYSQL_HOST")
	dbName = os.Getenv("MYSQL_DATABASE")

	// parseTime para que MySQL maneje bien DATETIME (aunque escaneamos a string usando DATE_FORMAT);
	// utf8mb4 para emojis/acentos adecuados.
	dataSourceName := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&tls=false",
		dbUser, dbPass, dbHost, dbName,
	)

	var err error
	Db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("❌ Error al conectar a la base de datos:", err)
	}

	Db.SetMaxOpenConns(25)
	Db.SetMaxIdleConns(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Conexión a la base de datos establecida.")
}

func CloseDB() {
	if Db != nil {
		if err := Db.Close(); err != nil {
			log.Println("❌ Error al cerrar la conexión a la base de datos:", err)
		}
		log.Println("✅ Conexión a la base de datos cerrada.")
	}
}

// -------- Usuarios --------

func GetUserByUsername(username string) (*models.User, error) {
	const query = "SELECT ID, Username, Password FROM Users WHERE Username = ?"
	row := Db.QueryRow(query, username)

	user := &models.User{}
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Println("Error al obtener usuario:", err)
		return nil, err
	}
	return user, nil
}

func InsertUser(user *models.User) error {
	const query = "INSERT INTO Users (Username, Password) VALUES (?, ?)"
	if _, err := Db.Exec(query, user.Username, user.Password); err != nil {
		log.Println("Error al insertar usuario:", err)
		return err
	}
	return nil
}

// -------- Posts (helpers comunes) --------

func scanPostRow(rows *sql.Rows) (*models.Post, error) {
	p := new(models.Post)
	// Orden explícito y CreatedAt formateado a string
	// NOTA: la última columna es Category (coincide con p.Categoria)
	if err := rows.Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Content,
		&p.AuthorID,
		&p.CreatedAt, // viene de DATE_FORMAT(...)
		&p.Categoria, // Category
	); err != nil {
		return nil, err
	}
	return p, nil
}

// -------- Posts (listados) --------

func GetAllPosts() ([]*models.Post, error) {
	const query = `
		SELECT
			ID,
			Title,
			Description,
			Content,
			AuthorID,
			DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s') AS CreatedAt,
			Category
		FROM Posts
		ORDER BY CreatedAt DESC, ID DESC`
	rows, err := Db.Query(query)
	if err != nil {
		log.Println("Error al obtener posts:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			log.Println("Error al escanear post:", err)
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetPostsPaged(limit, offset int) ([]*models.Post, error) {
	const query = `
		SELECT
			ID,
			Title,
			Description,
			Content,
			AuthorID,
			DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s') AS CreatedAt,
			Category
		FROM Posts
		ORDER BY CreatedAt DESC, ID DESC
		LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, limit, offset)
	if err != nil {
		log.Println("GetPostsPaged error:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			log.Println("GetPostsPaged scan error:", err)
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func FindPostsByCategory(categoria string) ([]*models.Post, error) {
	const query = `
		SELECT
			ID,
			Title,
			Description,
			Content,
			AuthorID,
			DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s') AS CreatedAt,
			Category
		FROM Posts
		WHERE Category = ?
		ORDER BY CreatedAt DESC, ID DESC`
	rows, err := Db.Query(query, categoria)
	if err != nil {
		log.Printf("Error al buscar posts con categoría %s: %v", categoria, err)
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			log.Println("Error al escanear post:", err)
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func FindPostsByCategoryPaged(categoria string, limit, offset int) ([]*models.Post, error) {
	const query = `
		SELECT
			ID,
			Title,
			Description,
			Content,
			AuthorID,
			DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s') AS CreatedAt,
			Category
		FROM Posts
		WHERE Category = ?
		ORDER BY CreatedAt DESC, ID DESC
		LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, categoria, limit, offset)
	if err != nil {
		log.Printf("FindPostsByCategoryPaged error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			log.Println("FindPostsByCategoryPaged scan error:", err)
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// -------- Post (detalles) --------

func FindPostByID(postID string) (*models.Post, error) {
	const query = `
		SELECT
			ID,
			Title,
			Description,
			Content,
			AuthorID,
			DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s') AS CreatedAt,
			Category
		FROM Posts
		WHERE ID = ?`
	row := Db.QueryRow(query, postID)

	var post models.Post
	if err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
		&post.Categoria,
	); err != nil {
		log.Printf("Error al buscar post %s: %v", postID, err)
		return nil, err
	}
	return &post, nil
}

// -------- Mutaciones --------

func InsertPost(post *models.Post) error {
	const query = `
		INSERT INTO Posts (Title, Description, Content, AuthorID, CreatedAt, Category)
		VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := Db.Exec(query, post.Title, post.Description, post.Content, post.AuthorID, post.CreatedAt, post.Categoria); err != nil {
		log.Println("Error al insertar post:", err)
		return err
	}
	return nil
}

func UpdatePost(postID int, post *models.Post) error {
	const query = `
		UPDATE Posts
		SET Title = ?, Description = ?, Content = ?, Category = ?, AuthorID = ?
		WHERE ID = ?`
	if _, err := Db.Exec(query, post.Title, post.Description, post.Content, post.Categoria, post.AuthorID, postID); err != nil {
		log.Println("Error al actualizar post:", err)
		return err
	}
	return nil
}

func DeletePost(postID int) error {
	const query = "DELETE FROM Posts WHERE ID = ?"
	if _, err := Db.Exec(query, postID); err != nil {
		log.Println("Error al eliminar post:", err)
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

func GetAllExperiences() ([]models.Experience, error) {
	rows, err := Db.Query("SELECT id, title, place, description, start_year, end_year FROM experiences ORDER BY end_year DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []models.Experience
	for rows.Next() {
		var e models.Experience
		err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.Description, &e.StartYear, &e.EndYear)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, e)
	}
	return experiences, nil
}

func GetAllEducation() ([]models.Education, error) {
	rows, err := Db.Query("SELECT id, title, place, start_year, end_year FROM education")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var education []models.Education
	for rows.Next() {
		var e models.Education
		err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.StartYear, &e.EndYear)
		if err != nil {
			return nil, err
		}
		education = append(education, e)
	}
	return education, nil
}

func GetAllProjects() ([]models.Project, error) {
	rows, err := Db.Query("SELECT id, title, description, image_url, project_url FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.ImageURL, &p.ProjectURL)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
