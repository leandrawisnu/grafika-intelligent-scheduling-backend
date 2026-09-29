UPDATE hari
SET akhir_pekan = true,
    updated_at = now()
WHERE nama = 'Sabtu';
