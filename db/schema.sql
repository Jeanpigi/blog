-- Esquema para desarrollo local, reconstruido a partir de las queries de db/db.go.
-- Compáralo con producción (SHOW CREATE TABLE ...) antes de usarlo para migrar.
--
--   docker run -d --name blog-mysql -e MYSQL_ROOT_PASSWORD=dev -e MYSQL_DATABASE=blog -p 3306:3306 mysql:8
--   docker exec -i blog-mysql mysql -uroot -pdev blog < db/schema.sql

CREATE TABLE IF NOT EXISTS Users (
  ID       INT AUTO_INCREMENT PRIMARY KEY,
  Username VARCHAR(20)  NOT NULL UNIQUE,
  Password VARCHAR(100) NOT NULL -- hash bcrypt
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS Posts (
  ID          INT AUTO_INCREMENT PRIMARY KEY,
  Title       VARCHAR(255) NOT NULL,
  Description VARCHAR(500) NOT NULL,
  Content     MEDIUMTEXT   NOT NULL, -- HTML de Quill (puede incluir imágenes base64)
  AuthorID    INT          NOT NULL,
  CreatedAt   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  Category    VARCHAR(20)  NOT NULL, -- 'Tech' | 'Historias'
  KEY idx_posts_category_created (Category, CreatedAt),
  KEY idx_posts_created (CreatedAt),
  CONSTRAINT fk_posts_author FOREIGN KEY (AuthorID) REFERENCES Users (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS Visits (
  id         INT AUTO_INCREMENT PRIMARY KEY,
  timestamp  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ip         VARCHAR(45)  NOT NULL,
  user_agent VARCHAR(512) NOT NULL DEFAULT '',
  page       VARCHAR(255) NOT NULL DEFAULT '',
  country    VARCHAR(100) NOT NULL DEFAULT '',
  region     VARCHAR(100) NOT NULL DEFAULT '',
  city       VARCHAR(100) NOT NULL DEFAULT '',
  latitude   DOUBLE       NOT NULL DEFAULT 0,
  longitude  DOUBLE       NOT NULL DEFAULT 0,
  KEY idx_visits_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Portafolio
CREATE TABLE IF NOT EXISTS experiences (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  title       VARCHAR(255) NOT NULL,
  place       VARCHAR(255) NOT NULL,
  description TEXT         NOT NULL,
  start_year  INT          NOT NULL,
  end_year    INT          NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS education (
  id         INT AUTO_INCREMENT PRIMARY KEY,
  title      VARCHAR(255) NOT NULL,
  place      VARCHAR(255) NOT NULL,
  start_year INT          NOT NULL,
  end_year   INT          NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS projects (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  title       VARCHAR(255) NOT NULL,
  description TEXT         NOT NULL,
  image_url   VARCHAR(512) NOT NULL DEFAULT '',
  project_url VARCHAR(512) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
