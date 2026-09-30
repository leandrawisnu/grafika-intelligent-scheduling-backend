UPDATE hari
SET akhir_pekan = false,
    updated_at = now()
WHERE nama = 'Sabtu';
