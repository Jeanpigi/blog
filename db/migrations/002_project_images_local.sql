-- Migración 002: imágenes de proyectos servidas desde el propio blog.
-- Antes estaban en ImgBB (i.ibb.co), que tarda 20+ s en responder o no responde.
-- Las nuevas están en static/img/projects/ en WebP (480 y 960 px); la plantilla
-- genera el srcset a partir de la versión -960.webp.
-- Solo cambia filas cuya URL coincide exactamente con la antigua; es seguro re-ejecutarla.
--
--   mysql -h HOST -u USUARIO -p BASE < db/migrations/002_project_images_local.sql

UPDATE projects SET image_url = '/static/img/projects/valery-960.webp'           WHERE image_url = 'https://i.ibb.co/R0jN0mr/Valery.png';
UPDATE projects SET image_url = '/static/img/projects/galeria-movil-960.webp'    WHERE image_url = 'https://i.ibb.co/bRqndVhj/Galeria-Movil.png';
UPDATE projects SET image_url = '/static/img/projects/jbearp-blog-960.webp'      WHERE image_url = 'https://i.ibb.co/7VWXBBd/Jbear-P-001.png';
UPDATE projects SET image_url = '/static/img/projects/store-online-960.webp'     WHERE image_url = 'https://i.ibb.co/6vwqZrj/Store-Online-001.png';
UPDATE projects SET image_url = '/static/img/projects/rick-and-morty-960.webp'   WHERE image_url = 'https://i.ibb.co/vwRRcpQ/rickandmorty.png';
UPDATE projects SET image_url = '/static/img/projects/realidad-virtual-960.webp' WHERE image_url = 'https://i.ibb.co/zb8RkYc/Realidad-Virtual.png';
UPDATE projects SET image_url = '/static/img/projects/derivada-960.webp'         WHERE image_url = 'https://i.ibb.co/WsprSgd/Derivada.webp';
UPDATE projects SET image_url = '/static/img/projects/moon-960.webp'             WHERE image_url = 'https://i.ibb.co/DQh9K6C/Moon.webp';

-- Para deshacer, cambia cada par de URLs de lado (la antigua de ImgBB sigue existiendo).
