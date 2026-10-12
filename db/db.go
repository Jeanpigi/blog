package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Jeanpigi/blog/internal/analytics"
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
		Timeout:              10 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		Params: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_unicode_ci",
		},
	}
	dsn := cfg.FormatDSN()
	log.Printf("🔗 Conectando a DB: %s@[host]/%s", dbUser, dbName)

	var err error
	Db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Error al conectar a la base de datos:", err)
	}

	// El MySQL de Hostinger corta las conexiones inactivas antes de lo que el
	// pool esperaba ("connection reset by peer" en los logs): una conexión
	// reutilizada después de eso hace fallar la petición. Se descartan las
	// inactivas a los 30s y se renuevan todas cada 2 min.
	Db.SetMaxOpenConns(25)
	Db.SetMaxIdleConns(5)
	Db.SetConnMaxLifetime(2 * time.Minute)
	Db.SetConnMaxIdleTime(30 * time.Second)

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

// lightSelect lista posts en el idioma pedido (primer parámetro ?): usa la
// traducción si existe y si no el original en español. reading_min se calcula
// sobre el contenido que realmente se va a leer.
const lightSelect = `
	SELECT p.ID,
	       COALESCE(t.title, p.Title),
	       COALESCE(t.description, p.Description),
	       p.AuthorID,
	       DATE_FORMAT(p.CreatedAt, '%Y-%m-%d %H:%i:%s'),
	       p.Category,
	       GREATEST(1, CEIL(CHAR_LENGTH(COALESCE(t.content, p.Content)) / 1200)) AS reading_min,
	       IF(t.post_id IS NULL, 'es', t.lang),
	       EXISTS(SELECT 1 FROM post_translations x WHERE x.post_id = p.ID AND x.lang = 'en')
	FROM Posts p
	LEFT JOIN post_translations t ON t.post_id = p.ID AND t.lang = ?`

func GetPostsListPaged(lang string, limit, offset int) ([]models.PostListItem, error) {
	query := lightSelect + ` ORDER BY p.CreatedAt DESC, p.ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, lang, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListRows(rows)
}

func GetAllPostsList(lang string) ([]models.PostListItem, error) {
	query := lightSelect + ` ORDER BY p.CreatedAt DESC, p.ID DESC`
	rows, err := Db.Query(query, lang)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListRows(rows)
}

func FindPostsByCategoryListPaged(category, lang string, limit, offset int) ([]models.PostListItem, error) {
	query := lightSelect + ` WHERE p.Category = ? ORDER BY p.CreatedAt DESC, p.ID DESC LIMIT ? OFFSET ?`
	rows, err := Db.Query(query, lang, category, limit, offset)
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
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.AuthorID, &it.CreatedAt, &it.Categoria, &it.ReadingMin, &it.Lang, &it.HasEN); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// -------- Traducciones de posts --------

// GetPostTranslations devuelve las traducciones de un post por idioma.
func GetPostTranslations(postID int) (map[string]*models.PostTranslation, error) {
	rows, err := Db.Query(`SELECT lang, title, description, content FROM post_translations WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]*models.PostTranslation{}
	for rows.Next() {
		var lang string
		t := &models.PostTranslation{}
		if err := rows.Scan(&lang, &t.Title, &t.Description, &t.Content); err != nil {
			return nil, err
		}
		out[lang] = t
	}
	return out, rows.Err()
}

// SavePostTranslation crea o reemplaza la traducción de un post.
func SavePostTranslation(postID int, lang string, t *models.PostTranslation) error {
	_, err := Db.Exec(`
		INSERT INTO post_translations (post_id, lang, title, description, content) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE title = VALUES(title), description = VALUES(description), content = VALUES(content)`,
		postID, lang, t.Title, t.Description, t.Content)
	return err
}

// DeletePostTranslation borra la traducción de un post en un idioma.
func DeletePostTranslation(postID int, lang string) error {
	_, err := Db.Exec(`DELETE FROM post_translations WHERE post_id = ? AND lang = ?`, postID, lang)
	return err
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

// InsertPost inserta el post y devuelve su ID autogenerado.
func InsertPost(post *models.Post) (int64, error) {
	const query = `INSERT INTO Posts (Title, Description, Content, AuthorID, CreatedAt, Category) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := Db.Exec(query, post.Title, post.Description, post.Content, post.AuthorID, post.CreatedAt, post.Categoria)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdatePost(postID int, post *models.Post) error {
	const query = `UPDATE Posts SET Title=?, Description=?, Content=?, Category=? WHERE ID=?`
	_, err := Db.Exec(query, post.Title, post.Description, post.Content, post.Categoria, postID)
	return err
}

func DeletePost(postID int) error {
	if _, err := Db.Exec("DELETE FROM post_translations WHERE post_id = ?", postID); err != nil {
		return err
	}
	_, err := Db.Exec("DELETE FROM Posts WHERE ID = ?", postID)
	return err
}

// -------- Visitas --------

func InsertVisit(visit *models.Visit) error {
	const query = `INSERT INTO Visits (ip, user_agent, page, country, region, city, latitude, longitude, view_id, referrer, device)
	               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	var viewID interface{} // NULL si no viene: la columna es UNIQUE
	if visit.ViewID != "" {
		viewID = visit.ViewID
	}
	_, err := Db.Exec(query, visit.IP, visit.UserAgent, visit.Page, visit.Country, visit.Region, visit.City,
		visit.Latitude, visit.Longitude, viewID, visit.Referrer, visit.Device)
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

// VisitStats resume el tráfico para el dashboard.
type VisitStats struct {
	Total     int `json:"total"`
	Last7Days int `json:"last_7_days"`
	Today     int `json:"today"`
	UniqueIPs int `json:"unique_ips"`
}

func GetVisitStats() (VisitStats, error) {
	const query = `
		SELECT COUNT(*),
		       COALESCE(SUM(timestamp >= NOW() - INTERVAL 7 DAY), 0),
		       COALESCE(SUM(timestamp >= CURDATE()), 0),
		       COUNT(DISTINCT ip)
		FROM Visits WHERE NOT REGEXP_LIKE(user_agent, ?, 'i')`
	var st VisitStats
	err := Db.QueryRow(query, analytics.BotSQLPattern).Scan(&st.Total, &st.Last7Days, &st.Today, &st.UniqueIPs)
	return st, err
}

// -------- Portafolio --------

func GetAllExperiences() ([]models.Experience, error) {
	rows, err := Db.Query(`SELECT id, title, place, description, start_year, end_year,
		title_en, place_en, COALESCE(description_en, '')
		FROM experiences ORDER BY (end_year = 0) DESC, end_year DESC, start_year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Experience
	for rows.Next() {
		var e models.Experience
		if err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.Description, &e.StartYear, &e.EndYear, &e.TitleEN, &e.PlaceEN, &e.DescriptionEN); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func GetAllEducation() ([]models.Education, error) {
	rows, err := Db.Query(`SELECT id, title, place, start_year, end_year, title_en, place_en
		FROM education ORDER BY (end_year = 0) DESC, end_year DESC, start_year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Education
	for rows.Next() {
		var e models.Education
		if err := rows.Scan(&e.ID, &e.Title, &e.Place, &e.StartYear, &e.EndYear, &e.TitleEN, &e.PlaceEN); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func GetAllProjects() ([]models.Project, error) {
	rows, err := Db.Query(`SELECT id, title, description, image_url, project_url, title_en, COALESCE(description_en, '')
		FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.ImageURL, &p.ProjectURL, &p.TitleEN, &p.DescriptionEN); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
