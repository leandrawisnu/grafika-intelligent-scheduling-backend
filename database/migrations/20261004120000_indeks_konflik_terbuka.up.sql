CREATE INDEX IF NOT EXISTS idx_konflik_terbuka_ringkasan
  ON konflik (jadwal_semester_id, tipe_konflik, tingkat_keparahan)
  WHERE terselesaikan = false;
