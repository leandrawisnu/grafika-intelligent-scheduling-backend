BEGIN;

ALTER TABLE konflik DROP COLUMN IF EXISTS updated_at;
ALTER TABLE konflik DROP COLUMN IF EXISTS created_at;

ALTER TABLE log_audit_jadwal DROP COLUMN IF EXISTS updated_at;
ALTER TABLE resolusi_ai DROP COLUMN IF EXISTS updated_at;
ALTER TABLE jadwal_semester_jurusan DROP COLUMN IF EXISTS updated_at;
ALTER TABLE hari_libur_guru DROP COLUMN IF EXISTS updated_at;
ALTER TABLE jam_pelajaran DROP COLUMN IF EXISTS updated_at;
ALTER TABLE hari DROP COLUMN IF EXISTS updated_at;

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
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'updated_at'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN updated_at TO diperbarui_pada', t);
        END IF;
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'created_at'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN created_at TO dibuat_pada', t);
        END IF;
    END LOOP;

    FOREACH t IN ARRAY tables_created_only LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = t AND column_name = 'created_at'
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN created_at TO dibuat_pada', t);
        END IF;
    END LOOP;
END $$;

COMMIT;
