-- Migración 004: contenido en inglés.
-- Solo agrega una tabla y columnas; no modifica el contenido en español.
-- Si un texto en inglés queda vacío, el sitio muestra el español.

-- Traducciones de posts (el post original sigue en la tabla Posts, en español).
CREATE TABLE IF NOT EXISTS post_translations (
  post_id     INT          NOT NULL,
  lang        VARCHAR(5)   NOT NULL,
  title       VARCHAR(255) NOT NULL,
  description VARCHAR(500) NOT NULL,
  content     MEDIUMTEXT   NOT NULL,
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (post_id, lang)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Portafolio en inglés
ALTER TABLE experiences
  ADD COLUMN title_en       VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN place_en       VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN description_en TEXT         NULL;

ALTER TABLE education
  ADD COLUMN title_en VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN place_en VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE projects
  ADD COLUMN title_en       VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN description_en TEXT         NULL;

-- Traducciones del contenido actual (se busca por el título en español).
UPDATE experiences SET title_en = 'Administrative Assistant', description_en = 'Developer and inventory administrator.' WHERE title = 'Auxiliar Administrativo';
UPDATE experiences SET title_en = 'ICT Manager', description_en = 'ICT administrator and systems developer.' WHERE title = 'Gestor TIC';
UPDATE experiences SET title_en = 'Educational Projects Analyst', description_en = 'Coordinated the implementation of educational projects as digital resources, supported the teaching staff''s pedagogical work, and designed and ran educational projects for distance and online learning.' WHERE title = 'Analista de Proyectos Educativos';
UPDATE experiences SET title_en = 'Project Systematizer: Agreement Int/Vo 399/17 - CORPOAMAZ - CASA', description_en = 'Consolidated the agreement''s monthly management report and supported logistics preparation for the project team.' WHERE title LIKE 'Sistematizador del Proyecto%';
UPDATE experiences SET title_en = 'Systems Assistant', description_en = 'Selected and organized patient records and entered patient information into the organization''s system.' WHERE title = 'Auxiliar de sistemas';

UPDATE education SET title_en = 'Specialization in ICT for Educational Innovation' WHERE title = 'Especialización TIC para la innovación educativa';
UPDATE education SET title_en = 'Systems Engineer' WHERE title = 'Ingeniero de sistemas';

UPDATE projects SET description_en = 'Online store built with Next.js and Firebase.' WHERE title = 'Valery Tienda';
UPDATE projects SET description_en = 'Interactive web platform to browse and buy tech products.' WHERE title = 'Galeria Movil';
UPDATE projects SET description_en = 'Personal blog created to teach and share the fundamentals of computing.' WHERE title = 'JbearP Blog';
UPDATE projects SET description_en = 'Online store built with React and Firebase.' WHERE title = 'Store Online';
UPDATE projects SET description_en = 'Instagram clone featuring characters from Rick and Morty.' WHERE title = 'RickStangram';
UPDATE projects SET title_en = 'Virtual Reality Project', description_en = 'A solar system project that strengthens seventh-grade students'' research skills about celestial bodies in social studies.' WHERE title = 'Proyecto de Realidad Virtual';
UPDATE projects SET title_en = 'Derivative', description_en = 'Virtual Learning Object (OVA) that teaches the basic concepts of the derivative.' WHERE title = 'Derivada';
UPDATE projects SET title_en = 'Weight on the Moon', description_en = 'Website that calculates your weight on the Moon.' WHERE title = 'Peso en la Luna';
UPDATE projects SET description_en = 'Website for the Doctorate in Education and Environmental Culture in Florencia, Caquetá, built with HTML5, CSS3 and JavaScript.' WHERE title = 'Deca';
UPDATE projects SET description_en = 'Website showcasing the meat distributor Districarnes El Gran Novillo in Florencia, Caquetá.' WHERE title = 'El Gran Novillo';
