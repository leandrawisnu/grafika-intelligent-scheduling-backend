BEGIN;

CREATE TABLE sesi_chat_ai (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID REFERENCES jadwal_semester(id) ON DELETE SET NULL,
    dilakukan_oleh      VARCHAR(100) NOT NULL DEFAULT 'anonim',
    judul               VARCHAR(200),
    dibuat_pada         TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada     TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_sesi_chat_jadwal ON sesi_chat_ai(jadwal_semester_id);
CREATE INDEX idx_sesi_chat_pengguna ON sesi_chat_ai(dilakukan_oleh, dibuat_pada DESC);

CREATE TABLE pesan_chat_ai (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sesi_chat_id    UUID NOT NULL REFERENCES sesi_chat_ai(id) ON DELETE CASCADE,
    peran           VARCHAR(20) NOT NULL,
    isi             TEXT NOT NULL,
    metadata_json   JSONB,
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    CHECK (peran IN ('pengguna', 'asisten', 'sistem'))
);

CREATE INDEX idx_pesan_chat_sesi ON pesan_chat_ai(sesi_chat_id, dibuat_pada);

CREATE TABLE feedback_chat_ai (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sesi_chat_id    UUID NOT NULL REFERENCES sesi_chat_ai(id) ON DELETE CASCADE,
    pesan_chat_id   UUID REFERENCES pesan_chat_ai(id) ON DELETE SET NULL,
    nilai           SMALLINT,
    jenis           VARCHAR(30) NOT NULL,
    komentar        TEXT,
    dilakukan_oleh  VARCHAR(100) NOT NULL DEFAULT 'anonim',
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    CHECK (nilai IS NULL OR (nilai >= 1 AND nilai <= 5)),
    CHECK (jenis IN ('positif', 'negatif', 'saran', 'laporkan'))
);

CREATE INDEX idx_feedback_chat_sesi ON feedback_chat_ai(sesi_chat_id);
CREATE INDEX idx_feedback_chat_pesan ON feedback_chat_ai(pesan_chat_id);

CREATE TABLE memori_chatbot (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dilakukan_oleh      VARCHAR(100) NOT NULL,
    jadwal_semester_id  UUID REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    kunci               VARCHAR(100) NOT NULL DEFAULT 'umum',
    ringkasan           TEXT NOT NULL,
    fakta_json          JSONB,
    sumber_sesi_chat_id UUID REFERENCES sesi_chat_ai(id) ON DELETE SET NULL,
    aktif               BOOLEAN DEFAULT true,
    dibuat_pada         TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada     TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_memori_chatbot_pengguna ON memori_chatbot(dilakukan_oleh, aktif);
CREATE INDEX idx_memori_chatbot_jadwal ON memori_chatbot(jadwal_semester_id) WHERE jadwal_semester_id IS NOT NULL;

COMMIT;
