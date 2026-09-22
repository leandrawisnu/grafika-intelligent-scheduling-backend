-- Seed skeleton: Semester Ganjil 2026/2027 (SMKN 4 Malang)
-- Master + jadwal_semester + semua kelas terdaftar, tanpa slot_jadwal.
-- Idempotent: aman dijalankan ulang.

BEGIN;

-- Tahun ajaran & semester
INSERT INTO tahun_ajaran (nama, tanggal_mulai, tanggal_selesai, aktif)
VALUES ('2026/2027', '2026-07-01', '2027-06-30', true)
ON CONFLICT (nama) DO NOTHING;

INSERT INTO semester (tahun_ajaran_id, nama, semester_ke, tanggal_mulai, tanggal_selesai, aktif)
SELECT ta.id, 'Ganjil 2026/2027', 1, '2026-07-01', '2026-12-31', true
FROM tahun_ajaran ta
WHERE ta.nama = '2026/2027'
ON CONFLICT (tahun_ajaran_id, semester_ke) DO NOTHING;

-- 8 jurusan
INSERT INTO jurusan (kode, nama) VALUES
  ('DKV', 'Desain Komunikasi Visual'),
  ('ANI', 'Animasi'),
  ('TG',  'Teknik Grafika'),
  ('TKJ', 'Teknik Komputer dan Jaringan'),
  ('RPL', 'Rekayasa Perangkat Lunak'),
  ('PH',  'Perhotelan'),
  ('TL',  'Teknik Logistik'),
  ('TM',  'Teknik Mekatronika')
ON CONFLICT (kode) DO NOTHING;

-- Jam pelajaran (grid skeleton)
INSERT INTO jam_pelajaran (jam_ke, waktu_mulai, waktu_selesai, istirahat) VALUES
  (0,  '06:30', '07:00', true),
  (1,  '07:00', '07:45', false),
  (2,  '07:45', '08:30', false),
  (3,  '08:30', '09:15', false),
  (4,  '09:15', '10:00', false),
  (5,  '10:00', '10:15', true),
  (6,  '10:15', '11:00', false),
  (7,  '11:00', '11:45', false),
  (8,  '11:45', '12:30', false),
  (9,  '13:00', '13:45', false),
  (10, '13:45', '14:30', false)
ON CONFLICT (jam_ke) DO NOTHING;

-- Ruangan dari PDF
INSERT INTO ruangan (kode, nama, kapasitas, tipe_ruangan, aktif) VALUES
  ('R-DKV-1',  'R. DKV 1',  36, 'kelas', true),
  ('R-DKV-2',  'R. DKV 2',  36, 'kelas', true),
  ('R-DKV-3',  'R. DKV 3',  36, 'kelas', true),
  ('R-DKV-4',  'R. DKV 4',  36, 'kelas', true),
  ('R-DKV-5',  'R. DKV 5',  36, 'kelas', true),
  ('R-DKV-6',  'R. DKV 6',  36, 'kelas', true),
  ('LAB-ANIMASI', 'LAB ANIMASI', 36, 'lab', true),
  ('LAB-2D',      'LAB 2D',      36, 'lab', true),
  ('LAB-INKA',    'LAB INKA',    36, 'lab', true),
  ('LAB-TJKT-1',  'LAB TJKT 1',  36, 'lab', true),
  ('LAB-TJKT-2',  'LAB TJKT 2',  36, 'lab', true),
  ('LAB-TJKT-4',  'LAB TJKT 4',  36, 'lab', true),
  ('LAB-BINDO',   'LAB BINDO',   36, 'lab', true),
  ('LAB-RPL-1',   'Lab RPL 1',   36, 'lab', true),
  ('LAB-RPL-2',   'Lab RPL 2',   36, 'lab', true),
  ('LAB-RPL-3',   'Lab RPL 3',   36, 'lab', true),
  ('LAB-RPL-4',   'Lab RPL 4',   36, 'lab', true),
  ('LAB-BI',      'Lab BI',      36, 'lab', true),
  ('LAB-BING',    'Lab Bing',    36, 'lab', true),
  ('LAB-PH-1',    'LAB PH 1',    36, 'lab', true),
  ('LAB-PH-2',    'LAB PH 2',    36, 'lab', true),
  ('R-1',   'R. 1',   36, 'kelas', true),
  ('R-2',   'R. 2',   36, 'kelas', true),
  ('R-16',  'R. 16',  36, 'kelas', true),
  ('R-17',  'R. 17',  36, 'kelas', true),
  ('R-24',  'R. 24',  36, 'kelas', true),
  ('R-25',  'R. 25',  36, 'kelas', true),
  ('R-26',  'R. 26',  36, 'kelas', true),
  ('R-33',  'R. 33',  36, 'kelas', true),
  ('R-34',  'R. 34',  36, 'kelas', true),
  ('R-35',  'R. 35',  36, 'kelas', true)
ON CONFLICT (kode) DO NOTHING;

-- 50 kelas (X & XI, semua jurusan — dari PDF)
INSERT INTO kelas (kode, nama, tingkat, jurusan_id, semester_id)
SELECT v.kode, v.nama, v.tingkat, j.id, sem.id
FROM (VALUES
  ('X-DKV-A',  'X DKV A',  10, 'DKV'), ('X-DKV-B',  'X DKV B',  10, 'DKV'), ('X-DKV-C',  'X DKV C',  10, 'DKV'),
  ('X-ANI-A',  'X ANI A',  10, 'ANI'), ('X-ANI-B',  'X ANI B',  10, 'ANI'), ('X-ANI-C',  'X ANI C',  10, 'ANI'),
  ('X-TG-A',   'X TG A',   10, 'TG'),  ('X-TG-B',   'X TG B',   10, 'TG'),
  ('X-TG-C',   'X TG C',   10, 'TG'),  ('X-TG-D',   'X TG D',   10, 'TG'),
  ('X-TG-E',   'X TG E',   10, 'TG'),  ('X-TG-F',   'X TG F',   10, 'TG'),
  ('X-TG-G',   'X TG G',   10, 'TG'),  ('X-TG-H',   'X TG H',   10, 'TG'),
  ('X-TKJ-A',  'X TKJ A',  10, 'TKJ'), ('X-TKJ-B',  'X TKJ B',  10, 'TKJ'),
  ('X-RPL-A',  'X RPL A',  10, 'RPL'), ('X-RPL-B',  'X RPL B',  10, 'RPL'), ('X-RPL-C',  'X RPL C',  10, 'RPL'),
  ('X-PH-A',   'X PH A',   10, 'PH'),  ('X-PH-B',   'X PH B',   10, 'PH'),
  ('X-TL-A',   'X TL A',   10, 'TL'),  ('X-TL-B',   'X TL B',   10, 'TL'),
  ('X-TM-A',   'X TM A',   10, 'TM'),  ('X-TM-B',   'X TM B',   10, 'TM'),
  ('XI-DKV-A', 'XI DKV A', 11, 'DKV'), ('XI-DKV-B', 'XI DKV B', 11, 'DKV'), ('XI-DKV-C', 'XI DKV C', 11, 'DKV'),
  ('XI-ANI-A', 'XI ANI A', 11, 'ANI'), ('XI-ANI-B', 'XI ANI B', 11, 'ANI'), ('XI-ANI-C', 'XI ANI C', 11, 'ANI'),
  ('XI-TG-A',  'XI TG A',  11, 'TG'),  ('XI-TG-B',  'XI TG B',  11, 'TG'),
  ('XI-TG-C',  'XI TG C',  11, 'TG'),  ('XI-TG-D',  'XI TG D',  11, 'TG'),
  ('XI-TG-E',  'XI TG E',  11, 'TG'),  ('XI-TG-F',  'XI TG F',  11, 'TG'),
  ('XI-TG-G',  'XI TG G',  11, 'TG'),  ('XI-TG-H',  'XI TG H',  11, 'TG'),
  ('XI-TKJ-A', 'XI TKJ A', 11, 'TKJ'), ('XI-TKJ-B', 'XI TKJ B', 11, 'TKJ'),
  ('XI-RPL-A', 'XI RPL A', 11, 'RPL'), ('XI-RPL-B', 'XI RPL B', 11, 'RPL'), ('XI-RPL-C', 'XI RPL C', 11, 'RPL'),
  ('XI-PH-A',  'XI PH A',  11, 'PH'),  ('XI-PH-B',  'XI PH B',  11, 'PH'),
  ('XI-TL-A',  'XI TL A',  11, 'TL'),  ('XI-TL-B',  'XI TL B',  11, 'TL'),
  ('XI-TM-A',  'XI TM A',  11, 'TM'),  ('XI-TM-B',  'XI TM B',  11, 'TM')
) AS v(kode, nama, tingkat, jur_kode)
JOIN jurusan j ON j.kode = v.jur_kode
JOIN semester sem ON sem.semester_ke = 1
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id AND ta.nama = '2026/2027'
ON CONFLICT (kode) DO NOTHING;

-- Jadwal semester (satu workbook sekolah)
INSERT INTO jadwal_semester (semester_id, status, bebas_konflik)
SELECT sem.id, 'draf', false
FROM semester sem
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
WHERE ta.nama = '2026/2027' AND sem.semester_ke = 1
ON CONFLICT (semester_id) DO NOTHING;

-- Semua jurusan terdaftar di jadwal
INSERT INTO jadwal_semester_jurusan (jadwal_semester_id, jurusan_id)
SELECT js.id, j.id
FROM jadwal_semester js
JOIN semester sem ON sem.id = js.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jurusan j ON j.kode IN ('DKV', 'ANI', 'TG', 'TKJ', 'RPL', 'PH', 'TL', 'TM')
WHERE ta.nama = '2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_semester_id, jurusan_id) DO NOTHING;

-- Jadwal kelas aktif per kelas (versi 1, slot kosong)
INSERT INTO jadwal_kelas (jadwal_semester_id, jurusan_id, kelas_id, versi, is_active)
SELECT js.id, k.jurusan_id, k.id, 1, true
FROM kelas k
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jadwal_semester js ON js.semester_id = sem.id
WHERE ta.nama = '2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_semester_id, kelas_id, versi) DO NOTHING;

COMMIT;
