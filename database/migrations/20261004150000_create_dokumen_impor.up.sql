BEGIN;

CREATE TABLE dokumen_impor (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    semester_id         UUID REFERENCES semester(id),
    target              VARCHAR(20) NOT NULL DEFAULT 'otomatis',
    nama_berkas         TEXT NOT NULL,
    kunci_berkas        TEXT NOT NULL,
    mime_berkas         VARCHAR(120),
    ukuran_berkas       BIGINT NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'menunggu',
    tahap               VARCHAR(30) NOT NULL DEFAULT 'menunggu',
    pesan               TEXT,
    teks_markdown       TEXT,
    rencana_json        JSONB,
    hasil_terapkan_json JSONB,
    dilakukan_oleh      VARCHAR(150),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (status IN ('menunggu', 'memproses', 'siap', 'diterapkan', 'gagal')),
    CHECK (target IN ('otomatis', 'jadwal', 'master'))
);

CREATE INDEX idx_dokumen_impor_jadwal ON dokumen_impor(jadwal_semester_id);
CREATE INDEX idx_dokumen_impor_status ON dokumen_impor(status);

COMMIT;
