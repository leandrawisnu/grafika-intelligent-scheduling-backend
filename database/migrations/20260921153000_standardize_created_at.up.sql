BEGIN;

-- Rename legacy Indonesian timestamp columns (existing DBs from init before standardization).
DO $$
DECLARE
    t text;
    tables_both text[] := ARRAY[
        'tahun_ajaran', 'semester', 'jurusan', 'guru', 'mata_pelajaran', 'kelas', 'ruangan',
        'jadwal_semester', 'jadwal_kelas', 'slot_jadwal', 'sesi_chat_ai', 'memori_chatbot'
    ];
    tables_created_only text[] := ARRAY[
        'hari', 'jam_pelajaran', 'hari_libur_guru', 'jadwal_semester_jurusan',
        'resolusi_ai', 'log_audit_jadwal', 'pesan_chat_ai', 'feedback_chat_ai'
    ];
BEGIN
    FOREACH t IN ARRAY tables_both LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'dibuat_pada'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN dibuat_pada TO created_at', t);
        END IF;
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'diperbarui_pada'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN diperbarui_pada TO updated_at', t);
        END IF;
    END LOOP;

    FOREACH t IN ARRAY tables_created_only LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'dibuat_pada'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN dibuat_pada TO created_at', t);
        END IF;
    END LOOP;
END $$;

-- Tables that previously had only dibuat_pada: add updated_at when missing.
ALTER TABLE hari ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE jam_pelajaran ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE hari_libur_guru ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE jadwal_semester_jurusan ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE resolusi_ai ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE log_audit_jadwal ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- konflik: align with GORM BaseModel (terdeteksi_pada stays separate).
ALTER TABLE konflik ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE konflik ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

COMMIT;
