-- Migración 001: medición de visitas desde el navegador (estadísticas del dashboard).
-- Ejecutar UNA vez en producción ANTES de desplegar el código que la usa.
-- Solo AGREGA columnas e índices: no borra ni modifica datos existentes.
--
--   mysql -h HOST -u USUARIO -p BASE < db/migrations/001_visits_analytics.sql

ALTER TABLE Visits
  ADD COLUMN view_id      CHAR(32)     NULL                  COMMENT 'id aleatorio de la vista (track.js)',
  ADD COLUMN referrer     VARCHAR(100) NOT NULL DEFAULT ''   COMMENT 'dominio externo de origen; vacío = directo',
  ADD COLUMN device       VARCHAR(10)  NOT NULL DEFAULT ''   COMMENT 'mobile | tablet | desktop',
  ADD COLUMN duration_sec INT          NOT NULL DEFAULT 0    COMMENT 'segundos con la página visible',
  ADD UNIQUE KEY uq_visits_view_id (view_id),
  ADD KEY idx_visits_page (page);

-- La columna ip pasa a guardar un hash HMAC de la IP (32 hex), no la IP real.
-- Las IPs antiguas las anonimiza la app automáticamente al arrancar.
