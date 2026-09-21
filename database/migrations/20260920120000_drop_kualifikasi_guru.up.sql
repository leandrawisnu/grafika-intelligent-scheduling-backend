BEGIN;

DROP TABLE IF EXISTS kualifikasi_guru;

-- Nilai enum `guru_tidak_berkualifikasi` pada DB lama tidak di-drop (PostgreSQL tidak mendukung DROP VALUE).

COMMIT;
