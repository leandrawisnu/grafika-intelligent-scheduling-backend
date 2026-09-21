BEGIN;

CREATE TABLE IF NOT EXISTS kualifikasi_guru (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guru_id           UUID NOT NULL REFERENCES guru(id) ON DELETE CASCADE,
    mata_pelajaran_id UUID NOT NULL REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
    tingkat_keahlian  VARCHAR(20) DEFAULT 'berkualifikasi',
    dibuat_pada       TIMESTAMPTZ DEFAULT now(),
    UNIQUE(guru_id, mata_pelajaran_id)
);

COMMIT;
