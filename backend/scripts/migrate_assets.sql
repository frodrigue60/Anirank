-- Script de utilidad para migrar URLs absolutas a rutas relativas.
-- Maneja prefijos de S3 (local/túnel) y R2 (producción).

BEGIN;

-- Procedimiento: Intentar limpiar prefijos conocidos de S3 y R2
-- Se hace en orden para asegurar que quede la ruta relativa limpia.

-- 1. Animes
UPDATE animes SET cover = REPLACE(cover, 'https://s3.anirank.work/anirank/', '') WHERE cover LIKE 'https://s3.anirank.work/anirank/%';
UPDATE animes SET cover = REPLACE(cover, 'https://r2.anirank.work/', '') WHERE cover LIKE 'https://r2.anirank.work/%';

UPDATE animes SET banner = REPLACE(banner, 'https://s3.anirank.work/anirank/', '') WHERE banner LIKE 'https://s3.anirank.work/anirank/%';
UPDATE animes SET banner = REPLACE(banner, 'https://r2.anirank.work/', '') WHERE banner LIKE 'https://r2.anirank.work/%';

-- 2. Artistas
UPDATE artists SET avatar = REPLACE(avatar, 'https://s3.anirank.work/anirank/', '') WHERE avatar LIKE 'https://s3.anirank.work/anirank/%';
UPDATE artists SET avatar = REPLACE(avatar, 'https://r2.anirank.work/', '') WHERE avatar LIKE 'https://r2.anirank.work/%';

-- 3. Usuarios
UPDATE users SET avatar = REPLACE(avatar, 'https://s3.anirank.work/anirank/', '') WHERE avatar LIKE 'https://s3.anirank.work/anirank/%';
UPDATE users SET avatar = REPLACE(avatar, 'https://r2.anirank.work/', '') WHERE avatar LIKE 'https://r2.anirank.work/%';

UPDATE users SET banner = REPLACE(banner, 'https://s3.anirank.work/anirank/', '') WHERE banner LIKE 'https://s3.anirank.work/anirank/%';
UPDATE users SET banner = REPLACE(banner, 'https://r2.anirank.work/', '') WHERE banner LIKE 'https://r2.anirank.work/%';

-- 4. Anuncios
UPDATE announcements SET image = REPLACE(image, 'https://s3.anirank.work/anirank/', '') WHERE image LIKE 'https://s3.anirank.work/anirank/%';
UPDATE announcements SET image = REPLACE(image, 'https://r2.anirank.work/', '') WHERE image LIKE 'https://r2.anirank.work/%';

-- 5. Medallas
UPDATE badges SET icon = REPLACE(icon, 'https://s3.anirank.work/anirank/', '') WHERE icon LIKE 'https://s3.anirank.work/anirank/%';
UPDATE badges SET icon = REPLACE(icon, 'https://r2.anirank.work/', '') WHERE icon LIKE 'https://r2.anirank.work/%';

COMMIT;
