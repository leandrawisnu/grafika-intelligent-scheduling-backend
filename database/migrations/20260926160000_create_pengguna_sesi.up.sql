CREATE TABLE pengguna (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL CONSTRAINT uni_pengguna_email UNIQUE,
    password_hash   TEXT NOT NULL,
    peran           VARCHAR(30) NOT NULL,
    jurusan_id      UUID REFERENCES jurusan(id),
    aktif           BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pengguna_peran_chk CHECK (peran IN ('admin', 'koor_jurusan')),
    CONSTRAINT pengguna_jurusan_chk CHECK (
        (peran = 'admin' AND jurusan_id IS NULL)
        OR (peran = 'koor_jurusan' AND jurusan_id IS NOT NULL)
    )
);

CREATE TABLE sesi (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pengguna_id     UUID NOT NULL REFERENCES pengguna(id) ON DELETE CASCADE,
    token_hash      CHAR(64) NOT NULL CONSTRAINT uni_sesi_token_hash UNIQUE,
    kedaluwarsa     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sesi_pengguna ON sesi(pengguna_id);
