-- Mata pelajaran & guru (singkatan PDF SMKN 4 Ganjil 2026/2027)
-- Idempotent. Jalankan sebelum seed slot.

BEGIN;

INSERT INTO mata_pelajaran (kode, nama, jam_wajib_per_minggu, tingkat) VALUES
  ('BIND',     'Bahasa Indonesia', 4, 10),
  ('BING',     'Bahasa Inggris', 4, 10),
  ('MAT',      'Matematika', 4, 10),
  ('SEJ',      'Sejarah', 2, 10),
  ('PP',       'Pendidikan Pancasila', 2, 10),
  ('PABP',     'Pendidikan Agama dan Budi Pekerti', 3, 10),
  ('PJOK',     'Pendidikan Jasmani, Olahraga, dan Kesehatan', 3, 10),
  ('SENI',     'Seni Budaya', 2, 10),
  ('IPAS',     'Ilmu Pengetahuan Alam dan Sosial', 2, 10),
  ('INKA',     'Informatika', 2, 10),
  ('KKA',      'Koding dan Kecerdasan Artifisial', 2, 10),
  ('MULOK',    'Muatan Lokal', 2, 10),
  ('PJBL',     'Projek Penguatan Profil Pelajar Pancasila', 2, 10),
  ('MBG',      'Makan Bergizi Gratis / Pembinaan', 1, 10),
  ('MDIKLAT',  'Masa Diklat', 2, 10),
  ('TEG',      'Teknik Energi Terbarukan', 2, 10),
  ('DDK-DKV',  'Dasar-Dasar Konsentrasi Keahlian DKV', 6, 10),
  ('DKV',      'Desain Komunikasi Visual', 8, 10),
  ('DDK-ANI',  'Dasar-Dasar Konsentrasi Keahlian Animasi', 6, 10),
  ('ANI-PROD', 'Produktif Animasi', 8, 10),
  ('DDK-TG',   'Dasar-Dasar Konsentrasi Keahlian Teknik Grafika', 6, 10),
  ('TG-PROD',  'Produktif Teknik Grafika', 8, 10),
  ('DDK-TKJ',  'Dasar-Dasar Konsentrasi Keahlian TKJ', 6, 10),
  ('TKJ-PROD', 'Produktif TKJ', 8, 10),
  ('DDK-RPL',  'Dasar-Dasar Konsentrasi Keahlian RPL', 6, 10),
  ('RPL-PROD', 'Produktif RPL', 8, 10),
  ('DDK-PH',   'Dasar-Dasar Konsentrasi Keahlian Perhotelan', 6, 10),
  ('PH-PROD',  'Produktif Perhotelan', 8, 10),
  ('DDK-TL',   'Dasar-Dasar Konsentrasi Keahlian Teknik Logistik', 6, 10),
  ('TL-PROD',  'Produktif Teknik Logistik', 8, 10),
  ('DDK-TM',   'Dasar-Dasar Konsentrasi Keahlian Teknik Mekatronika', 6, 10),
  ('TM-PROD',  'Produktif Teknik Mekatronika', 8, 10),
  ('BK',       'Bimbingan Konseling', 1, 10)
ON CONFLICT (kode) DO NOTHING;

-- Guru: kode inisial dari PDF → NIP unik PDF-{KODE}
INSERT INTO guru (nip, nama_lengkap, jam_maksimal_per_minggu, aktif)
SELECT v.nip, v.nama, 40, true
FROM (VALUES
  ('PDF-AJI', 'Ajik'), ('PDF-SAN', 'Santoso'), ('PDF-NOV', 'Novita'), ('PDF-WUR', 'Wuri'),
  ('PDF-SAR', 'Sari'), ('PDF-RAF', 'Rafli'), ('PDF-YGN', 'Yogini'), ('PDF-TNY', 'Tanya'),
  ('PDF-IBAM', 'Ibam'), ('PDF-HGW', 'Hegi W'), ('PDF-IMA', 'Ima'), ('PDF-DRH', 'Diah'),
  ('PDF-PUG', 'Puji'), ('PDF-ASA', 'Asa'), ('PDF-ERI', 'Eri'), ('PDF-DYH', 'Dyah'),
  ('PDF-WIN', 'Winarti'), ('PDF-SET', 'Setyo'), ('PDF-ELI', 'Eli'), ('PDF-DEV', 'Devi'),
  ('PDF-HAN', 'Hani'), ('PDF-ROS', 'Rosita'), ('PDF-GMA', 'Gema'), ('PDF-IWH', 'Iwan'),
  ('PDF-ISM', 'Ismail'), ('PDF-SDI', 'Sadi'), ('PDF-ZEE', 'Zee'), ('PDF-LABANA', 'Labana'),
  ('PDF-MIT', 'Mita'), ('PDF-ACM', 'Acmad'), ('PDF-SHA', 'Shafa'), ('PDF-JGA', 'Jaga')
) AS v(nip, nama)
ON CONFLICT (nip) DO NOTHING;

COMMIT;
