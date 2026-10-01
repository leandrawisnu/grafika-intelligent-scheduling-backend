CREATE INDEX IF NOT EXISTS idx_slot_guru_hari_jam
  ON slot_jadwal (guru_id, hari_id, jam_pelajaran_id)
  WHERE guru_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_slot_ruang_hari_jam
  ON slot_jadwal (ruangan_id, hari_id, jam_pelajaran_id)
  WHERE ruangan_id IS NOT NULL;
