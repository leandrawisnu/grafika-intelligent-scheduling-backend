CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- DATA MASTER
-- ============================================

CREATE TABLE tahun_ajaran (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama              VARCHAR(50) NOT NULL UNIQUE,
    tanggal_mulai     DATE NOT NULL,
    tanggal_selesai   DATE NOT NULL,
    aktif             BOOLEAN DEFAULT true,
    dibuat_pada       TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE semester (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tahun_ajaran_id   UUID NOT NULL REFERENCES tahun_ajaran(id) ON DELETE CASCADE,
    nama              VARCHAR(50) NOT NULL,
    semester_ke       SMALLINT NOT NULL,
    tanggal_mulai     DATE NOT NULL,
    tanggal_selesai   DATE NOT NULL,
    aktif             BOOLEAN DEFAULT true,
    dibuat_pada       TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tahun_ajaran_id, semester_ke)
);

CREATE TABLE jurusan (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode            VARCHAR(20) NOT NULL UNIQUE,
    nama            VARCHAR(100) NOT NULL,
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE guru (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nip                  VARCHAR(30) NOT NULL UNIQUE,
    nama_lengkap         VARCHAR(150) NOT NULL,
    jam_maksimal_per_minggu  DECIMAL(4,1) NOT NULL DEFAULT 40.0,
    aktif                BOOLEAN DEFAULT true,
    dibuat_pada          TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE mata_pelajaran (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode                   VARCHAR(20) NOT NULL UNIQUE,
    nama                   VARCHAR(150) NOT NULL,
    jam_wajib_per_minggu   DECIMAL(4,1) NOT NULL,
    tingkat                SMALLINT NOT NULL,
    dibuat_pada            TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada        TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE kelas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode            VARCHAR(30) NOT NULL UNIQUE,
    nama            VARCHAR(100) NOT NULL,
    tingkat         SMALLINT NOT NULL,
    jurusan_id      UUID REFERENCES jurusan(id),
    semester_id     UUID NOT NULL REFERENCES semester(id),
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE ruangan (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode            VARCHAR(20) NOT NULL UNIQUE,
    nama            VARCHAR(100) NOT NULL,
    kapasitas       INT NOT NULL DEFAULT 30,
    tipe_ruangan    VARCHAR(30) DEFAULT 'kelas',
    aktif           BOOLEAN DEFAULT true,
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE hari (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama         VARCHAR(20) NOT NULL UNIQUE,
    urutan_hari  SMALLINT NOT NULL UNIQUE,
    akhir_pekan  BOOLEAN DEFAULT false,
    dibuat_pada  TIMESTAMPTZ DEFAULT now()
);

INSERT INTO hari (nama, urutan_hari, akhir_pekan) VALUES
    ('Senin',    1, false),
    ('Selasa',   2, false),
    ('Rabu',     3, false),
    ('Kamis',    4, false),
    ('Jumat',    5, false),
    ('Sabtu',    6, false),
    ('Minggu',   7, true);

CREATE TABLE jam_pelajaran (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jam_ke        SMALLINT NOT NULL UNIQUE,
    waktu_mulai   TIME NOT NULL,
    waktu_selesai TIME NOT NULL,
    istirahat     BOOLEAN DEFAULT false,
    dibuat_pada   TIMESTAMPTZ DEFAULT now()
);

-- ============================================
-- KENDALA GURU
-- ============================================

CREATE TABLE hari_libur_guru (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guru_id       UUID NOT NULL REFERENCES guru(id) ON DELETE CASCADE,
    hari_id       UUID NOT NULL REFERENCES hari(id) ON DELETE CASCADE,
    semester_id   UUID NOT NULL REFERENCES semester(id),
    alasan        VARCHAR(255),
    dibuat_pada   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(guru_id, hari_id, semester_id)
);

CREATE TABLE kualifikasi_guru (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guru_id           UUID NOT NULL REFERENCES guru(id) ON DELETE CASCADE,
    mata_pelajaran_id UUID NOT NULL REFERENCES mata_pelajaran(id) ON DELETE CASCADE,
    tingkat_keahlian  VARCHAR(20) DEFAULT 'berkualifikasi',
    dibuat_pada       TIMESTAMPTZ DEFAULT now(),
    UNIQUE(guru_id, mata_pelajaran_id)
);

-- ============================================
-- JADWAL — STRUKTUR BARU
-- ============================================

-- Master jadwal per semester
CREATE TABLE jadwal_semester (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    semester_id     UUID NOT NULL REFERENCES semester(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'draf',
    bebas_konflik   BOOLEAN DEFAULT false,
    dibuat_pada     TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada TIMESTAMPTZ DEFAULT now(),
    UNIQUE(semester_id)
);

-- Jurusan yang terdaftar di jadwal semester ini
CREATE TABLE jadwal_semester_jurusan (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID NOT NULL REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    jurusan_id          UUID NOT NULL REFERENCES jurusan(id) ON DELETE CASCADE,
    dibuat_pada         TIMESTAMPTZ DEFAULT now(),
    UNIQUE(jadwal_semester_id, jurusan_id)
);

-- Jadwal per kelas (versioned)
CREATE TABLE jadwal_kelas (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID NOT NULL REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    jurusan_id          UUID NOT NULL REFERENCES jurusan(id),
    kelas_id            UUID NOT NULL REFERENCES kelas(id),
    versi               INT NOT NULL DEFAULT 1,
    is_active           BOOLEAN DEFAULT false,
    dibuat_pada         TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada     TIMESTAMPTZ DEFAULT now(),
    UNIQUE(jadwal_semester_id, kelas_id, versi)
);

-- Slot jadwal per kelas (FK ke jadwal_kelas)
CREATE TABLE slot_jadwal (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_kelas_id   UUID NOT NULL REFERENCES jadwal_kelas(id) ON DELETE CASCADE,
    kelas_id          UUID NOT NULL REFERENCES kelas(id),
    mata_pelajaran_id UUID NOT NULL REFERENCES mata_pelajaran(id),
    hari_id           UUID NOT NULL REFERENCES hari(id),
    jam_pelajaran_id  UUID NOT NULL REFERENCES jam_pelajaran(id),
    ruangan_id        UUID REFERENCES ruangan(id),
    guru_id           UUID REFERENCES guru(id),
    minggu_ke         SMALLINT DEFAULT 1,
    terkunci          BOOLEAN DEFAULT false,
    dibuat_pada       TIMESTAMPTZ DEFAULT now(),
    diperbarui_pada   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke),
    UNIQUE(jadwal_kelas_id, ruangan_id, hari_id, jam_pelajaran_id, minggu_ke),
    UNIQUE(jadwal_kelas_id, guru_id, hari_id, jam_pelajaran_id, minggu_ke)
);

-- ============================================
-- KONFLIK
-- ============================================

CREATE TYPE tipe_konflik AS ENUM (
    'guru_bentrok',
    'ruangan_bentrok',
    'kelas_bentrok',
    'guru_kelebihan_jam',
    'guru_hari_libur',
    'jam_mapel_kurang',
    'guru_tidak_berkualifikasi',
    'kapasitas_ruangan_melebihi'
);

CREATE TYPE tingkat_keparahan AS ENUM ('kesalahan', 'peringatan');

CREATE TABLE konflik (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID NOT NULL REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    tipe_konflik        tipe_konflik NOT NULL,
    tingkat_keparahan   tingkat_keparahan NOT NULL DEFAULT 'kesalahan',
    slot_a_id           UUID REFERENCES slot_jadwal(id),
    slot_b_id           UUID REFERENCES slot_jadwal(id),
    guru_id             UUID REFERENCES guru(id),
    deskripsi           TEXT NOT NULL,
    detail_json         JSONB,
    terselesaikan       BOOLEAN DEFAULT false,
    diselesaikan_oleh   VARCHAR(20),
    terdeteksi_pada     TIMESTAMPTZ DEFAULT now(),
    terselesaikan_pada  TIMESTAMPTZ
);

CREATE INDEX idx_konflik_jadwal_semester ON konflik(jadwal_semester_id);
CREATE INDEX idx_konflik_tipe            ON konflik(tipe_konflik);
CREATE INDEX idx_konflik_selesai         ON konflik(jadwal_semester_id, terselesaikan);

-- ============================================
-- RESOLUSI AI
-- ============================================

CREATE TABLE resolusi_ai (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    konflik_id              UUID NOT NULL REFERENCES konflik(id) ON DELETE CASCADE,
    jadwal_semester_id      UUID NOT NULL REFERENCES jadwal_semester(id),
    peringkat               SMALLINT NOT NULL,
    skor_keyakinan          DECIMAL(3,2) NOT NULL,
    usulan_perubahan_json   JSONB NOT NULL,
    penjelasan              TEXT NOT NULL,
    diterima                BOOLEAN DEFAULT false,
    dibuat_pada             TIMESTAMPTZ DEFAULT now(),
    CHECK (skor_keyakinan >= 0 AND skor_keyakinan <= 1)
);

-- ============================================
-- LOG AUDIT
-- ============================================

CREATE TABLE log_audit_jadwal (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_semester_id  UUID NOT NULL REFERENCES jadwal_semester(id) ON DELETE CASCADE,
    aksi                VARCHAR(50) NOT NULL,
    perubahan_json      JSONB NOT NULL,
    dilakukan_oleh      VARCHAR(100) DEFAULT 'sistem',
    dibuat_pada         TIMESTAMPTZ DEFAULT now()
);
