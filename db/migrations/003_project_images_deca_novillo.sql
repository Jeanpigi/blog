-- Migración 003: las dos imágenes de proyectos que quedaban en ImgBB.
UPDATE projects SET image_url = '/static/img/projects/deca-960.webp'    WHERE image_url = 'https://i.ibb.co/3cMzVqm/Deca.webp';
UPDATE projects SET image_url = '/static/img/projects/novillo-960.webp' WHERE image_url = 'https://i.ibb.co/xSNWgG6/Novillo.webp';
