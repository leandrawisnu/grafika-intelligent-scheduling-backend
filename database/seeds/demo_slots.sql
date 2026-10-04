-- Seed slot demo Demo - Ganjil 2026/2027 — template mingguan minim konflik.
-- Murni SQL, tanpa generator. Pola mapel dari PDF SMKN 4.
--
-- Syarat: jumlah guru >= jumlah kelas (50), lihat demo_roster.sql.
--
-- Cara tune:
--   * template_umum  — mapel umum per tingkat (kolom mapel_kode saja;
--                      guru & ruangan diisi otomatis injektif per sel).
--   * produktif_map  — mapel produktif per jurusan + tingkat.
--   * produktif_jam  — posisi jam produktif dalam seminggu.
--   * KONFLIK_DEMO   — slot konflik yang disengaja untuk demo fitur
--                      deteksi + resolusi AI. Hapus bloknya bila ingin
--                      jadwal 100% bersih.
--
-- Desain anti-konflik: 50 kelas berjalan paralel di jam yang sama.
-- Guru dipilih round-robin injektif per sel (hari, jam) dan ruangan
-- memakai home room (R-<kode kelas>, unik per kelas), sehingga tidak
-- ada bentrok guru/ruangan/kelas. Beban tiap guru ~32 jam (< 40).
-- Slot KONFLIK_DEMO memberi tepat 4 konflik untuk bahan demo:
--   2x guru_bentrok + 2x ruangan_bentrok.
--
-- Cek cepat setelah seed (harus 2 baris + 2 baris + kosong):
--   SELECT g.nip, h.nama, j.jam_ke, count(*)
--   FROM slot_jadwal sj JOIN guru g ON g.id = sj.guru_id
--   JOIN hari h ON h.id = sj.hari_id JOIN jam_pelajaran j ON j.id = sj.jam_pelajaran_id
--   GROUP BY 1,2,3 HAVING count(*) > 1 ORDER BY 1,2,3;
--   SELECT r.kode, h.nama, j.jam_ke, count(*)
--   FROM slot_jadwal sj JOIN ruangan r ON r.id = sj.ruangan_id
--   JOIN hari h ON h.id = sj.hari_id JOIN jam_pelajaran j ON j.id = sj.jam_pelajaran_id
--   GROUP BY 1,2,3 HAVING count(*) > 1;
--   SELECT g.nip, count(*) FROM slot_jadwal sj
--   JOIN guru g ON g.id = sj.guru_id GROUP BY 1 HAVING count(*) > 40;
--
-- Prasyarat: ganjil_2026_2027.sql + mapel_guru_ganjil_2026.sql +
-- demo_roster.sql sudah dijalankan.
-- Idempotent: aman dijalankan ulang (termasuk setelah validasi konflik
-- dijalankan — tabel konflik/resolusi/log demo ikut dibersihkan).

BEGIN;

-- Bersihkan hasil validasi demo lama agar seed ulang tidak menabrak FK.
DELETE FROM resolusi_ai ra
USING konflik kf, jadwal_semester js, semester sem, tahun_ajaran ta
WHERE ra.konflik_id = kf.id
  AND kf.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1;

DELETE FROM konflik kf
USING jadwal_semester js, semester sem, tahun_ajaran ta
WHERE kf.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1;

DELETE FROM log_audit_jadwal la
USING jadwal_semester js, semester sem, tahun_ajaran ta
WHERE la.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1;

-- Bersihkan slot demo lama agar template selalu sesuai file ini.
DELETE FROM slot_jadwal sj
USING jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta
WHERE sj.jadwal_kelas_id = jk.id
  AND jk.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1;

WITH
-- Mapel umum: berlaku untuk semua jurusan pada tingkat tersebut.
-- Urutan dalam VALUES menentukan urutan jam dalam seminggu.
template_umum(urutan, tingkat, mapel_kode) AS (
  VALUES
    -- Kelas X: 30 sel (Senin–Jumat x jam 1,2,3,4,6,7)
    (1, 10, 'BIND'), (2, 10, 'MAT'), (3, 10, 'BING'),
    (4, 10, 'INKA'), (5, 10, 'PP'), (6, 10, 'SEJ'),
    (7, 10, 'BIND'), (8, 10, 'PJOK'), (9, 10, 'BING'),
    (10, 10, 'IPAS'), (11, 10, 'PABP'), (12, 10, 'KKA'),
    (13, 10, 'MAT'), (14, 10, 'BIND'), (15, 10, 'SENI'),
    (16, 10, 'MULOK'), (17, 10, 'BING'), (18, 10, 'PJBL'),
    (19, 10, 'BIND'), (20, 10, 'MAT'), (21, 10, 'SEJ'),
    (22, 10, 'PP'), (23, 10, 'INKA'), (24, 10, 'TEG'),
    (25, 10, 'BING'), (26, 10, 'BIND'), (27, 10, 'PABP'),
    (28, 10, 'PJOK'), (29, 10, 'KKA'), (30, 10, 'BK'),
    -- Kelas XI: 30 sel
    (1, 11, 'BIND'), (2, 11, 'MAT'), (3, 11, 'BING'),
    (4, 11, 'SEJ'), (5, 11, 'PP'), (6, 11, 'PJBL'),
    (7, 11, 'BIND'), (8, 11, 'PJOK'), (9, 11, 'BING'),
    (10, 11, 'INKA'), (11, 11, 'PABP'), (12, 11, 'KKA'),
    (13, 11, 'MAT'), (14, 11, 'BIND'), (15, 11, 'MULOK'),
    (16, 11, 'SENI'), (17, 11, 'BING'), (18, 11, 'TEG'),
    (19, 11, 'BIND'), (20, 11, 'MAT'), (21, 11, 'PP'),
    (22, 11, 'SEJ'), (23, 11, 'INKA'), (24, 11, 'PJBL'),
    (25, 11, 'BING'), (26, 11, 'BIND'), (27, 11, 'PABP'),
    (28, 11, 'PJOK'), (29, 11, 'KKA'), (30, 11, 'BK')
),
-- Posisi 30 sel umum dalam seminggu (urut sesuai kolom urutan).
slot_umum(hari_nama, jam_ke, no_sel) AS (
  VALUES
    ('Senin', 1, 1), ('Senin', 2, 2), ('Senin', 3, 3),
    ('Senin', 4, 4), ('Senin', 6, 5), ('Senin', 7, 6),
    ('Selasa', 1, 7), ('Selasa', 2, 8), ('Selasa', 3, 9),
    ('Selasa', 4, 10), ('Selasa', 6, 11), ('Selasa', 7, 12),
    ('Rabu', 1, 13), ('Rabu', 2, 14), ('Rabu', 3, 15),
    ('Rabu', 4, 16), ('Rabu', 6, 17), ('Rabu', 7, 18),
    ('Kamis', 1, 19), ('Kamis', 2, 20), ('Kamis', 3, 21),
    ('Kamis', 4, 22), ('Kamis', 6, 23), ('Kamis', 7, 24),
    ('Jumat', 1, 25), ('Jumat', 2, 26), ('Jumat', 3, 27),
    ('Jumat', 4, 28), ('Jumat', 6, 29), ('Jumat', 7, 30)
),
-- Mapel produktif per jurusan + tingkat (disebar ke produktif_jam).
produktif_map(jur_kode, tingkat, mapel_kode) AS (
  VALUES
    ('DKV', 10, 'DDK-DKV'), ('DKV', 11, 'DKV'),
    ('ANI', 10, 'DDK-ANI'), ('ANI', 11, 'ANI-PROD'),
    ('TG', 10, 'DDK-TG'), ('TG', 11, 'TG-PROD'),
    ('TKJ', 10, 'DDK-TKJ'), ('TKJ', 11, 'TKJ-PROD'),
    ('RPL', 10, 'DDK-RPL'), ('RPL', 11, 'RPL-PROD'),
    ('PH', 10, 'DDK-PH'), ('PH', 11, 'PH-PROD'),
    ('TL', 10, 'DDK-TL'), ('TL', 11, 'TL-PROD'),
    ('TM', 10, 'DDK-TM'), ('TM', 11, 'TM-PROD')
),
-- Posisi 8 sel produktif dalam seminggu.
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
-- Daftar kelas demo berurutan (indeks stabil per kode).
kelas_urut(kode, tingkat, jur_kode, idx) AS (
  SELECT k.kode, k.tingkat, ju.kode,
    row_number() OVER (ORDER BY k.kode) - 1
  FROM kelas k
  JOIN jurusan ju ON ju.id = k.jurusan_id
  JOIN semester sem ON sem.id = k.semester_id
  JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
  WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
),
-- Daftar guru berurutan (indeks stabil per nip).
guru_urut(nip, idx) AS (
  SELECT g.nip, row_number() OVER (ORDER BY g.nip) - 1
  FROM guru g
),
jumlah_guru(n) AS (
  SELECT count(*) FROM guru_urut
),
-- Semua sel template per kelas: 30 umum + 8 produktif = 38.
semua_slot(kode_kelas, tingkat, jur_kode, idx_kelas, hari_nama, jam_ke, mapel_kode) AS (
  SELECT ku.kode, ku.tingkat, ku.jur_kode, ku.idx, su.hari_nama, su.jam_ke, tu.mapel_kode
  FROM kelas_urut ku
  JOIN template_umum tu ON tu.tingkat = ku.tingkat
  JOIN slot_umum su ON su.no_sel = tu.urutan
  UNION ALL
  SELECT ku.kode, ku.tingkat, ku.jur_kode, ku.idx, pj.hari_nama, pj.jam_ke, pm.mapel_kode
  FROM kelas_urut ku
  JOIN produktif_map pm ON pm.jur_kode = ku.jur_kode AND pm.tingkat = ku.tingkat
  CROSS JOIN produktif_jam pj
),
-- Penomoran global tiap sel (hari, jam): menentukan offset putaran guru.
sel_nomor(hari_nama, jam_ke, no_global) AS (
  SELECT hari_nama, jam_ke, row_number() OVER (ORDER BY
    CASE hari_nama
      WHEN 'Senin' THEN 1 WHEN 'Selasa' THEN 2 WHEN 'Rabu' THEN 3
      WHEN 'Kamis' THEN 4 WHEN 'Jumat' THEN 5 ELSE 6 END, jam_ke) - 1
  FROM (SELECT DISTINCT hari_nama, jam_ke FROM (
    SELECT hari_nama, jam_ke FROM slot_umum
    UNION SELECT hari_nama, jam_ke FROM produktif_jam
  ) t) t
),
-- Slot rapi: guru round-robin injektif per sel, ruangan home room.
-- Translasi (idx_kelas + konstanta) mod 60 adalah permutasi, jadi 50
-- kelas selalu dapat 50 guru berbeda di tiap sel — tanpa bentrok.
slot_rapi AS (
  SELECT ss.*, gu.nip AS guru_nip, ('R-' || ss.kode_kelas) AS ruangan_kode
  FROM semua_slot ss
  JOIN sel_nomor sn ON sn.hari_nama = ss.hari_nama AND sn.jam_ke = ss.jam_ke
  CROSS JOIN jumlah_guru jg
  JOIN guru_urut gu ON gu.idx = mod(ss.idx_kelas + sn.no_global * 50, jg.n)
)
INSERT INTO slot_jadwal (
  jadwal_kelas_id, kelas_id, mata_pelajaran_id, hari_id, jam_pelajaran_id,
  ruangan_id, guru_id, minggu_ke, terkunci
)
SELECT jk.id, k.id, mp.id, h.id, j.id,
  (SELECT r.id FROM ruangan r WHERE r.kode = sr.ruangan_kode AND r.semester_id = sem.id LIMIT 1),
  g.id, 1, false
FROM kelas k
JOIN jurusan ju ON ju.id = k.jurusan_id
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jadwal_semester js ON js.semester_id = sem.id
JOIN jadwal_kelas jk ON jk.jadwal_semester_id = js.id
  AND jk.kelas_id = k.id AND jk.versi = 1 AND jk.is_active = true
JOIN slot_rapi sr ON sr.kode_kelas = k.kode
JOIN mata_pelajaran mp ON mp.kode = sr.mapel_kode
JOIN hari h ON h.nama = sr.hari_nama
JOIN jam_pelajaran j ON j.jam_ke = sr.jam_ke
JOIN guru g ON g.nip = sr.guru_nip
WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke) DO NOTHING;

-- ============================================================
-- KONFLIK_DEMO — 8 slot konflik yang disengaja untuk demo,
-- menghasilkan tepat 4 baris konflik saat validasi dijalankan:
--   1-2. guru_bentrok    : PDF-SAN mengajar X-DKV-A & X-DKV-B
--          di Senin jam 1 dan 2 (2 baris konflik).
--   3-4. ruangan_bentrok : LAB-BINDO dipakai X-ANI-A & X-ANI-B
--          di Selasa jam 1 dan 2 (2 baris konflik).
-- Hapus blok ini bila ingin jadwal 100% bersih.
-- ============================================================

-- Kosongkan sel target guru-bentrok agar slot demo benar-benar tertulis
-- (ON CONFLICT tidak menimpa slot round-robin yang sudah ada).
-- Sel ruangan-bentrok (X-ANI-A/B Selasa jam 1-2) TIDAK dikosongkan:
-- slot round-robin-nya dipertahankan dan hanya ruangannya di-UPDATE
-- menjadi LAB-BINDO di bawah, sehingga pasangan (guru, sel) utuh.
DELETE FROM slot_jadwal sj
USING jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta,
  kelas k, hari h, jam_pelajaran j
WHERE sj.jadwal_kelas_id = jk.id
  AND jk.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND sj.kelas_id = k.id AND sj.hari_id = h.id AND sj.jam_pelajaran_id = j.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
  AND (
    k.kode IN ('X-DKV-A', 'X-DKV-B') AND h.nama = 'Senin' AND j.jam_ke IN (1, 2)
  );

-- 1-2. Guru yang sama di dua kelas, jam yang sama.
INSERT INTO slot_jadwal (
  jadwal_kelas_id, kelas_id, mata_pelajaran_id, hari_id, jam_pelajaran_id,
  ruangan_id, guru_id, minggu_ke, terkunci
)
SELECT jk.id, k.id, mp.id, h.id, j.id,
  (SELECT r.id FROM ruangan r WHERE r.kode = 'R-' || k.kode AND r.semester_id = sem.id LIMIT 1),
  g.id, 1, false
FROM (VALUES ('X-DKV-A', 1), ('X-DKV-B', 1), ('X-DKV-A', 2), ('X-DKV-B', 2)) AS v(kode, jam)
JOIN kelas k ON k.kode = v.kode
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jadwal_semester js ON js.semester_id = sem.id
JOIN jadwal_kelas jk ON jk.jadwal_semester_id = js.id
  AND jk.kelas_id = k.id AND jk.versi = 1 AND jk.is_active = true
JOIN mata_pelajaran mp ON mp.kode = 'BIND'
JOIN hari h ON h.nama = 'Senin'
JOIN jam_pelajaran j ON j.jam_ke = v.jam
JOIN guru g ON g.nip = 'PDF-SAN'
WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke) DO NOTHING;

-- 3-4. Ruangan yang sama di dua kelas, jam yang sama.
-- Slot round-robin X-ANI-A/B di Selasa jam 1-2 hanya di-UPDATE ruangannya
-- menjadi LAB-BINDO — guru round-robin tidak berubah, sehingga tidak ada
-- grup bentrok guru baru. Melengkapi DELETE sel target di atas.
UPDATE slot_jadwal sj SET ruangan_id = lab.id
FROM jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta,
  kelas k, hari h, jam_pelajaran j,
  (SELECT r.id FROM ruangan r
   JOIN semester s2 ON s2.id = r.semester_id
   JOIN tahun_ajaran t2 ON t2.id = s2.tahun_ajaran_id
   WHERE r.kode = 'LAB-BINDO' AND t2.nama = 'Demo - 2026/2027' AND s2.semester_ke = 1
   LIMIT 1) lab
WHERE sj.jadwal_kelas_id = jk.id
  AND jk.jadwal_semester_id = js.id
  AND js.semester_id = sem.id
  AND sem.tahun_ajaran_id = ta.id
  AND sj.kelas_id = k.id AND sj.hari_id = h.id AND sj.jam_pelajaran_id = j.id
  AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
  AND k.kode IN ('X-ANI-A', 'X-ANI-B')
  AND h.nama = 'Selasa' AND j.jam_ke IN (1, 2);

-- Bersihkan sisa: guru round-robin yang jatuh di Senin jam 1-2 kelas
-- selain X-DKV-A/B adalah PDF-SAN (pemilik sel X-DKV-A/B ikut kehapus
-- semua di DELETE atas) — slot sisa itu membentuk grup beranggotakan 3.
-- Sel tersebut diisi ulang dengan guru yang benar-benar bebas di sel itu,
-- sehingga grid tetap penuh dan hanya tersisa 2 grup guru-bentrok.
WITH stray AS (
  DELETE FROM slot_jadwal sj
  USING jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta,
    kelas k, guru g, hari h, jam_pelajaran j
  WHERE sj.jadwal_kelas_id = jk.id
    AND jk.jadwal_semester_id = js.id
    AND js.semester_id = sem.id
    AND sem.tahun_ajaran_id = ta.id
    AND sj.kelas_id = k.id AND sj.guru_id = g.id
    AND sj.hari_id = h.id AND sj.jam_pelajaran_id = j.id
    AND ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
    AND (
      g.nip = 'PDF-SAN' AND h.nama = 'Senin' AND j.jam_ke IN (1, 2)
        AND k.kode NOT IN ('X-DKV-A', 'X-DKV-B')
    )
  RETURNING sj.kelas_id, sj.mata_pelajaran_id, sj.hari_id, sj.jam_pelajaran_id, sj.ruangan_id
)
INSERT INTO slot_jadwal (
  jadwal_kelas_id, kelas_id, mata_pelajaran_id, hari_id, jam_pelajaran_id,
  ruangan_id, guru_id, minggu_ke, terkunci
)
SELECT jk.id, st.kelas_id, st.mata_pelajaran_id, st.hari_id, st.jam_pelajaran_id,
  st.ruangan_id, bebas.guru_id, 1, false
FROM stray st
JOIN kelas k ON k.id = st.kelas_id
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
JOIN jadwal_semester js ON js.semester_id = sem.id
JOIN jadwal_kelas jk ON jk.jadwal_semester_id = js.id
  AND jk.kelas_id = k.id AND jk.versi = 1 AND jk.is_active = true
CROSS JOIN LATERAL (
  SELECT gu.id AS guru_id
  FROM guru gu
  WHERE NOT EXISTS (
    SELECT 1 FROM slot_jadwal o
    WHERE o.guru_id = gu.id
      AND o.hari_id = st.hari_id
      AND o.jam_pelajaran_id = st.jam_pelajaran_id
      AND o.minggu_ke = 1
  )
  ORDER BY gu.nip
  LIMIT 1
) bebas
WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
ON CONFLICT (jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke) DO NOTHING;

-- Backfill plotting 1:1 dari slot (satu baris plotting per satu baris slot,
-- wajar karena seed demo memakai round-robin injektif per sel).
INSERT INTO plotting (semester_id, kelas_id, hari_id, jam_pelajaran_id, mata_pelajaran_id, guru_id, ruangan_id)
SELECT sem.id, sj.kelas_id, sj.hari_id, sj.jam_pelajaran_id,
  sj.mata_pelajaran_id, sj.guru_id, sj.ruangan_id
FROM slot_jadwal sj
JOIN jadwal_kelas jk ON jk.id = sj.jadwal_kelas_id
JOIN jadwal_semester js ON js.id = jk.jadwal_semester_id
JOIN semester sem ON sem.id = js.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
ON CONFLICT (semester_id, kelas_id, hari_id, jam_pelajaran_id) DO NOTHING;

COMMIT;
