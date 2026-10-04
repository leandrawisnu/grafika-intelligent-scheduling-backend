-- Seed slot demo Ganjil 2026/2027 — template mingguan per kelas.
-- Pengganti scripts/seed-slots.py: murni SQL, tanpa generator.
-- Pola dari singkatan mapel PDF Ganjil 2026/2027 SMKN 4.
--
-- Cara tune: ubah tabel template_umum (mapel umum per tingkat),
-- produktif_map (mapel produktif per jurusan + tingkat), atau
-- produktif_jam (posisi jam produktif). Struktur kelas tidak perlu diubah.
--
-- Prasyarat: database/seeds/ganjil_2026_2027.sql +
-- database/seeds/mapel_guru_ganjil_2026.sql sudah dijalankan.
-- Idempotent: aman dijalankan ulang.

BEGIN;

-- Bersihkan slot demo lama agar template selalu sesuai file ini.
DELETE FROM slot_jadwal sj
USING jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta
WHERE sj.jadwal_kelas_id = jk.id
  AND jk.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = '2026/2027' AND sem.semester_ke = 1;

WITH
-- Mapel umum: berlaku untuk semua jurusan pada tingkat tersebut.
template_umum(hari_nama, jam_ke, tingkat, mapel_kode, guru_nip, ruangan_kode) AS (
  VALUES
    -- Kelas X
    ('Senin', 1, 10, 'BIND', 'PDF-SAN', 'R-DKV-1'),
    ('Senin', 2, 10, 'MAT', 'PDF-NOV', 'R-DKV-2'),
    ('Senin', 3, 10, 'BING', 'PDF-WUR', 'R-DKV-3'),
    ('Senin', 4, 10, 'INKA', 'PDF-ERI', 'R-DKV-4'),
    ('Senin', 6, 10, 'PP', 'PDF-SAR', 'R-DKV-5'),
    ('Senin', 7, 10, 'SEJ', 'PDF-RAF', 'R-DKV-6'),
    ('Selasa', 1, 10, 'BIND', 'PDF-SAN', 'R-DKV-1'),
    ('Selasa', 2, 10, 'PJOK', 'PDF-YGN', 'R-DKV-6'),
    ('Selasa', 3, 10, 'BING', 'PDF-WUR', 'R-DKV-2'),
    ('Selasa', 4, 10, 'IPAS', 'PDF-IMA', 'R-DKV-3'),
    ('Selasa', 6, 10, 'PABP', 'PDF-ASA', 'R-DKV-4'),
    ('Selasa', 7, 10, 'KKA', 'PDF-DRH', 'R-DKV-5'),
    ('Rabu', 1, 10, 'MAT', 'PDF-NOV', 'R-DKV-1'),
    ('Rabu', 2, 10, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Rabu', 3, 10, 'SENI', 'PDF-PUG', 'R-DKV-3'),
    ('Rabu', 4, 10, 'MULOK', 'PDF-TNY', 'R-DKV-4'),
    ('Rabu', 6, 10, 'BING', 'PDF-WUR', 'R-DKV-5'),
    ('Rabu', 7, 10, 'PJBL', 'PDF-ERI', 'R-DKV-6'),
    ('Kamis', 1, 10, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Kamis', 2, 10, 'MAT', 'PDF-NOV', 'R-DKV-1'),
    ('Kamis', 3, 10, 'SEJ', 'PDF-RAF', 'R-DKV-3'),
    ('Kamis', 4, 10, 'PP', 'PDF-SAR', 'R-DKV-4'),
    ('Kamis', 6, 10, 'INKA', 'PDF-ERI', 'R-DKV-5'),
    ('Kamis', 7, 10, 'TEG', 'PDF-HGW', 'R-DKV-6'),
    ('Jumat', 1, 10, 'BING', 'PDF-WUR', 'R-DKV-1'),
    ('Jumat', 2, 10, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Jumat', 3, 10, 'PABP', 'PDF-ASA', 'R-DKV-3'),
    ('Jumat', 4, 10, 'PJOK', 'PDF-YGN', 'R-DKV-6'),
    ('Jumat', 6, 10, 'KKA', 'PDF-DRH', 'R-DKV-4'),
    ('Jumat', 7, 10, 'BK', 'PDF-DYH', 'R-DKV-5'),
    -- Kelas XI
    ('Senin', 1, 11, 'BIND', 'PDF-SAN', 'R-DKV-1'),
    ('Senin', 2, 11, 'MAT', 'PDF-NOV', 'R-DKV-2'),
    ('Senin', 3, 11, 'BING', 'PDF-WUR', 'R-DKV-3'),
    ('Senin', 4, 11, 'SEJ', 'PDF-RAF', 'R-DKV-4'),
    ('Senin', 6, 11, 'PP', 'PDF-SAR', 'R-DKV-5'),
    ('Senin', 7, 11, 'PJBL', 'PDF-ERI', 'R-DKV-6'),
    ('Selasa', 1, 11, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Selasa', 2, 11, 'PJOK', 'PDF-YGN', 'R-DKV-6'),
    ('Selasa', 3, 11, 'BING', 'PDF-WUR', 'R-DKV-1'),
    ('Selasa', 4, 11, 'INKA', 'PDF-ERI', 'R-DKV-3'),
    ('Selasa', 6, 11, 'PABP', 'PDF-ASA', 'R-DKV-4'),
    ('Selasa', 7, 11, 'KKA', 'PDF-DRH', 'R-DKV-5'),
    ('Rabu', 1, 11, 'MAT', 'PDF-NOV', 'R-DKV-1'),
    ('Rabu', 2, 11, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Rabu', 3, 11, 'MULOK', 'PDF-TNY', 'R-DKV-3'),
    ('Rabu', 4, 11, 'SENI', 'PDF-PUG', 'R-DKV-4'),
    ('Rabu', 6, 11, 'BING', 'PDF-WUR', 'R-DKV-5'),
    ('Rabu', 7, 11, 'TEG', 'PDF-HGW', 'R-DKV-6'),
    ('Kamis', 1, 11, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Kamis', 2, 11, 'MAT', 'PDF-NOV', 'R-DKV-1'),
    ('Kamis', 3, 11, 'PP', 'PDF-SAR', 'R-DKV-3'),
    ('Kamis', 4, 11, 'SEJ', 'PDF-RAF', 'R-DKV-4'),
    ('Kamis', 6, 11, 'INKA', 'PDF-ERI', 'R-DKV-5'),
    ('Kamis', 7, 11, 'PJBL', 'PDF-ERI', 'R-DKV-6'),
    ('Jumat', 1, 11, 'BING', 'PDF-WUR', 'R-DKV-1'),
    ('Jumat', 2, 11, 'BIND', 'PDF-SAN', 'R-DKV-2'),
    ('Jumat', 3, 11, 'PABP', 'PDF-ASA', 'R-DKV-3'),
    ('Jumat', 4, 11, 'PJOK', 'PDF-YGN', 'R-DKV-6'),
    ('Jumat', 6, 11, 'KKA', 'PDF-DRH', 'R-DKV-4'),
    ('Jumat', 7, 11, 'BK', 'PDF-DYH', 'R-DKV-5')
),
-- Mapel produktif per jurusan + tingkat (disebar ke produktif_jam).
produktif_map(jur_kode, tingkat, mapel_kode, guru_nip, ruangan_kode) AS (
  VALUES
    ('DKV', 10, 'DDK-DKV', 'PDF-AJI', 'R-DKV'),
    ('DKV', 11, 'DKV', 'PDF-AJI', 'R-DKV'),
    ('ANI', 10, 'DDK-ANI', 'PDF-ZEE', 'LAB-ANIMASI'),
    ('ANI', 11, 'ANI-PROD', 'PDF-ZEE', 'LAB-ANIMASI'),
    ('TG', 10, 'DDK-TG', 'PDF-IBAM', 'LAB-INKA'),
    ('TG', 11, 'TG-PROD', 'PDF-IBAM', 'LAB-INKA'),
    ('TKJ', 10, 'DDK-TKJ', 'PDF-MIT', 'LAB-TJKT-1'),
    ('TKJ', 11, 'TKJ-PROD', 'PDF-MIT', 'LAB-TJKT-1'),
    ('RPL', 10, 'DDK-RPL', 'PDF-DEV', 'LAB-RPL-1'),
    ('RPL', 11, 'RPL-PROD', 'PDF-DEV', 'LAB-RPL-1'),
    ('PH', 10, 'DDK-PH', 'PDF-HAN', 'LAB-PH-1'),
    ('PH', 11, 'PH-PROD', 'PDF-HAN', 'LAB-PH-1'),
    ('TL', 10, 'DDK-TL', 'PDF-SHA', 'R-16'),
    ('TL', 11, 'TL-PROD', 'PDF-SHA', 'R-16'),
    ('TM', 10, 'DDK-TM', 'PDF-JGA', 'R-33'),
    ('TM', 11, 'TM-PROD', 'PDF-JGA', 'R-33')
),
-- Posisi jam produktif dalam seminggu.
produktif_jam(hari_nama, jam_ke) AS (
  VALUES
    ('Senin', 8),
    ('Selasa', 8),
    ('Rabu', 8),
    ('Kamis', 8),
    ('Jumat', 8),
    ('Senin', 9),
    ('Selasa', 9),
    ('Rabu', 9)
),
-- Gabungan template umum + produktif per jurusan.
semua_slot(hari_nama, jam_ke, tingkat, jur_kode, mapel_kode, guru_nip, ruangan_kode) AS (
  SELECT hari_nama, jam_ke, tingkat, 'SEMUA', mapel_kode, guru_nip, ruangan_kode
  FROM template_umum
  UNION ALL
  SELECT pj.hari_nama, pj.jam_ke, pm.tingkat, pm.jur_kode, pm.mapel_kode, pm.guru_nip, pm.ruangan_kode
  FROM produktif_map pm
  CROSS JOIN produktif_jam pj
)
INSERT INTO slot_jadwal (
  jadwal_kelas_id, kelas_id, mata_pelajaran_id, hari_id, jam_pelajaran_id,
  ruangan_id, guru_id, minggu_ke, terkunci
)
SELECT jk.id, k.id, mp.id, h.id, j.id,
  (SELECT id FROM ruangan WHERE kode = s.ruangan_kode LIMIT 1),
  g.id, 1, false
FROM kelas k
JOIN jurusan ju ON ju.id = k.jurusan_id
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jadwal_semester js ON js.semester_id = sem.id
JOIN jadwal_kelas jk ON jk.jadwal_semester_id = js.id
  AND jk.kelas_id = k.id AND jk.versi = 1 AND jk.is_active = true
JOIN semua_slot s ON s.tingkat = k.tingkat
  AND (s.jur_kode = 'SEMUA' OR s.jur_kode = ju.kode)
JOIN mata_pelajaran mp ON mp.kode = s.mapel_kode
JOIN hari h ON h.nama = s.hari_nama
JOIN jam_pelajaran j ON j.jam_ke = s.jam_ke
JOIN guru g ON g.nip = s.guru_nip
WHERE ta.nama = '2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke) DO NOTHING;

COMMIT;
