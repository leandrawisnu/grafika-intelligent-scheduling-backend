CREATE INDEX IF NOT EXISTS idx_jadwal_kelas_semester_aktif
  ON jadwal_kelas (jadwal_semester_id, is_active)
  WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_slot_jadwal_jadwal_kelas
  ON slot_jadwal (jadwal_kelas_id);
