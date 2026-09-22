-- Hapus semester ganjil duplikat (semester_ke=2 salah label) + jadwal kosong.
-- Data seed (kelas + slot) ada di semester_ke=1. Idempotent.

BEGIN;

DELETE FROM jadwal_semester js
USING semester sem, tahun_ajaran ta
WHERE js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = '2026/2027'
  AND sem.semester_ke = 2
  AND NOT EXISTS (
    SELECT 1 FROM jadwal_kelas jk WHERE jk.jadwal_semester_id = js.id
  );

DELETE FROM semester sem
USING tahun_ajaran ta
WHERE sem.tahun_ajaran_id = ta.id
  AND ta.nama = '2026/2027'
  AND sem.semester_ke = 2
  AND NOT EXISTS (SELECT 1 FROM kelas k WHERE k.semester_id = sem.id)
  AND NOT EXISTS (SELECT 1 FROM jadwal_semester js WHERE js.semester_id = sem.id);

COMMIT;
