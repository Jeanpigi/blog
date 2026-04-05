package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Jeanpigi/blog/internal/models"
	"github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func InitDB() {
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_ROOT_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbName := os.Getenv("MYSQL_DATABASE")
	dbPort := os.Getenv("MYSQL_PORTDATABASE")
	if dbPort == "" {
		dbPort = "3306"
	}

	// Usar Config.FormatDSN para escapar correctamente contraseñas con caracteres especiales (?, #, @, etc.)
	cfg := mysql.Config{
		User:                 dbUser,
		Passwd:               dbPass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", dbHost, dbPort),
		DBName:               dbName,
		ParseTime:            true,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_unicode_ci",
		},
	}
	dsn := cfg.FormatDSN()
	log.Printf("🔗 Conectando a DB: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	var err error
	Db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Error al conectar a la base de datos:", err)
	}

	Db.SetMaxOpenConns(25)
	Db.SetMaxIdleConns(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	if err := Db.Ping(); err != nil {
		log.Fatal("❌ No se pudo conectar a la base de datos:", err)
	}
	log.Println("✅ Conexión a la base de datos establecida.")
}

func CloseDB() {
	if Db != nil {
		if err := Db.Close(); err != nil {
			log.Println("❌ Error al cerrar la conexión:", err)
		}
		log.Println("✅ Conexión cerrada.")
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
		return nil, err
	}
	return user, nil
}

func InsertUser(user *models.User) error {
	const query = "INSERT INTO Users (Username, Password) VALUES (?, ?)"
	_, err := Db.Exec(query, user.Username, user.Password)
	return err
}

// -------- Posts (helpers internos) --------

func scanPostRow(rows *sql.Rows) (*models.Post, error) {
	p := new(models.Post)
	if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Content, &p.AuthorID, &p.CreatedAt, &p.Categoria); err != nil {
		return nil, err
	}
	return p, nil
}

// -------- Posts: listados completos (con Content) --------

func GetAllPosts() ([]*models.Post, error) {
	const query = `
		SELECT ID, Title, Description, Content, AuthorID,
		       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'), Category
		FROM Posts ORDER BY CreatedAt DESC, ID DESC`
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func GetPostsPaged(limit, offset int) ([]*models.Post, error) {
	const query = `
		SELECT ID, Title, Description, Content, AuthorID,
		       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'), Category
		FROM Posts ORDER BY CreatedAt DESC, ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func scanPostRows(rows *sql.Rows) ([]*models.Post, error) {
	var posts []*models.Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// -------- Posts: listados livianos (sin Content, reading_min en SQL) --------

const lightSelect = `
	SELECT ID, Title, Description, AuthorID,
	       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'),
	       Category,
	       GREATEST(1, CEIL(CHAR_LENGTH(Content) / 1200)) AS reading_min
	FROM Posts`

func GetPostsListPaged(limit, offset int) ([]models.PostListItem, error) {
	query := lightSelect + ` ORDER BY CreatedAt DESC, ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListRows(rows)
}

func GetAllPostsList() ([]models.PostListItem, error) {
	query := lightSelect + ` ORDER BY CreatedAt DESC, ID DESC`
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListRows(rows)
}

func FindPostsByCategoryListPaged(category string, limit, offset int) ([]models.PostListItem, error) {
	query := lightSelect + ` WHERE Category = ? ORDER BY CreatedAt DESC, ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListRows(rows)
}

func scanListRows(rows *sql.Rows) ([]models.PostListItem, error) {
	var items []models.PostListItem
	for rows.Next() {
		var it models.PostListItem
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.AuthorID, &it.CreatedAt, &it.Categoria, &it.ReadingMin); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// CountPosts devuelve el total de posts (para paginación).
func CountPosts() (int, error) {
	var n int
	return n, Db.QueryRow("SELECT COUNT(*) FROM Posts").Scan(&n)
}

func CountPostsByCategory(category string) (int, error) {
	var n int
	return n, Db.QueryRow("SELECT COUNT(*) FROM Posts WHERE Category = ?", category).Scan(&n)
}

// -------- Post individual --------

func FindPostsByCategory(categoria string) ([]*models.Post, error) {
	const query = `
		SELECT ID, Title, Description, Content, AuthorID,
		       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'), Category
		FROM Posts WHERE Category = ? ORDER BY CreatedAt DESC, ID DESC`
	rows, err := Db.Query(query, categoria)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func FindPostsByCategoryPaged(categoria string, limit, offset int) ([]*models.Post, error) {
	const query = `
		SELECT ID, Title, Description, Content, AuthorID,
		       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'), Category
		FROM Posts WHERE Category = ? ORDER BY CreatedAt DESC, ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, categoria, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func FindPostByID(postID string) (*models.Post, error) {
	const query = `
		SELECT ID, Title, Description, Content, AuthorID,
		       DATE_FORMAT(CreatedAt, '%Y-%m-%d %H:%i:%s'), Category
		FROM Posts WHERE ID = ?`
	row := Db.QueryRow(query, postID)
	var p models.Post
	if err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Content, &p.AuthorID, &p.CreatedAt, &p.Categoria); err != nil {
		return nil, err
	}
	return &p, nil
}

// -------- Mutaciones --------

func InsertPost(post *models.Post) error {
	const query = `INSERT INTO Posts (Title, Description, Content, AuthorID, CreatedAt, Category) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, post.Title, post.Description, post.Content, post.AuthorID, post.CreatedAt, post.Categoria)
	return err
}

func UpdatePost(postID int, post *models.Post) error {
	const query = `UPDATE Posts SET Title=?, Description=?, Content=?, Category=? WHERE ID=?`
	_, err := Db.Exec(query, post.Title, post.Description, post.Content, post.Categoria, postID)
	return err
}

func DeletePost(postID int) error {
	_, err := Db.Exec("DELETE FROM Posts WHERE ID = ?", postID)
	return err
}

// -------- Visitas --------

func InsertVisit(visit *models.Visit) error {
	const query = `INSERT INTO Visits (ip, user_agent, page, country, region, city, latitude, longitude)
	               VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := Db.Exec(query, visit.IP, visit.UserAgent, visit.Page, visit.Country, visit.Region, visit.City, visit.Latitude, visit.Longitude)
	return err
}

func GetAllVisits() ([]models.Visit, error) {
	const query = `SELECT id, DATE_FORMAT(timestamp, '%Y-%m-%d %H:%i:%s'), ip, user_agent, page,
	                      country, region, city, latitude, longitude
	               FROM Visits ORDER BY timestamp DESC LIMIT 100`
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []models.Visit
	for rows.Next() {
		var v models.Visit
		var ts string
		if err := rows.Scan(&v.ID, &ts, &v.IP, &v.UserAgent, &v.Page, &v.Country, &v.Region, &v.City, &v.Latitude, &v.Longitude); err != nil {
			continue
		}
		v.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// -------- Portafolio --------

func GetAllExperiences() ([]models.Experience, error) {
	rows, err := Db.Query("SELECT id, title, place, description, start_year, end_year FROM experiences ORDER BY end_year DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Experience
	for rows.Next() {
		var e models.Experience
		if err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.Description, &e.StartYear, &e.EndYear); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func GetAllEducation() ([]models.Education, error) {
	rows, err := Db.Query("SELECT id, title, place, start_year, end_year FROM education")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Education
	for rows.Next() {
		var e models.Education
		if err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.StartYear, &e.EndYear); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func GetAllProjects() ([]models.Project, error) {
	rows, err := Db.Query("SELECT id, title, description, image_url, project_url FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.ImageURL, &p.ProjectURL); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
