-- Roster demo Demo - Ganjil 2026/2027 — guru + home room tambahan.
-- Idempotent: aman dijalankan ulang.
--
-- Kenapa file ini ada:
--   Template slot (demo_slots.sql) menjadwalkan ~50 kelas paralel di jam
--   yang sama. Dengan 32 guru / 31 ruangan (bawaan PDF), konflik bentrok
--   tak terhindarkan (satu guru dipakai 50 kelas sekaligus, 250 jam/minggu).
--   File ini menambah 28 guru (total 60) dan 1 home room per kelas
--   (50 ruang, kode R-<kode kelas>) sehingga setiap sel (hari, jam) bisa
--   diisi 50 guru dan 50 ruangan berbeda — syarat jadwal bebas konflik.
--   Lab/bengkel bawaan PDF tetap ada sebagai data master (dipakai sepasang
--   slot konflik ruangan yang disengaja untuk demo).

BEGIN;

INSERT INTO guru (nip, nama_lengkap, jam_maksimal_per_minggu, aktif) VALUES
  ('DEMO-001', 'Andini', 40, true),
  ('DEMO-002', 'Bagus', 40, true),
  ('DEMO-003', 'Citra', 40, true),
  ('DEMO-004', 'Dimas', 40, true),
  ('DEMO-005', 'Eka', 40, true),
  ('DEMO-006', 'Fajar', 40, true),
  ('DEMO-007', 'Gita', 40, true),
  ('DEMO-008', 'Hadi', 40, true),
  ('DEMO-009', 'Intan', 40, true),
  ('DEMO-010', 'Joko', 40, true),
  ('DEMO-011', 'Kirana', 40, true),
  ('DEMO-012', 'Lestari', 40, true),
  ('DEMO-013', 'Made', 40, true),
  ('DEMO-014', 'Nabila', 40, true),
  ('DEMO-015', 'Putri', 40, true),
  ('DEMO-016', 'Rizky', 40, true),
  ('DEMO-017', 'Sinta', 40, true),
  ('DEMO-018', 'Taufik', 40, true),
  ('DEMO-019', 'Utami', 40, true),
  ('DEMO-020', 'Wahyu', 40, true),
  ('DEMO-021', 'Yulia', 40, true),
  ('DEMO-022', 'Zaki', 40, true),
  ('DEMO-023', 'Ratna', 40, true),
  ('DEMO-024', 'Budi', 40, true),
  ('DEMO-025', 'Dewi', 40, true),
  ('DEMO-026', 'Agus', 40, true),
  ('DEMO-027', 'Maya', 40, true),
  ('DEMO-028', 'Rudi', 40, true)
ON CONFLICT (nip) DO NOTHING;

-- Home room: 1 ruangan unik per kelas demo (per-semester).
INSERT INTO ruangan (kode, nama, kapasitas, tipe_ruangan, aktif, semester_id)
SELECT 'R-' || k.kode, 'Ruang ' || k.nama, 36, 'kelas', true, k.semester_id
FROM kelas k
JOIN semester sem ON sem.id = k.semester_id
JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
WHERE ta.nama = 'Demo - 2026/2027' AND sem.semester_ke = 1
ON CONFLICT (semester_id, kode) DO NOTHING;

COMMIT;
