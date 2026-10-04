package dto

// ImportPlan adalah rencana impor hasil analisis AI (LlamaParse + LLM)
// yang sudah divalidasi terhadap data master oleh backend.
type ImportPlan struct {
	JenisDokumen string            `json:"jenis_dokumen"`
	Ringkasan    string            `json:"ringkasan"`
	Keyakinan    float64           `json:"keyakinan"`
	Peringatan   []string          `json:"peringatan"`
	MasterUsulan MasterUsulanPlan  `json:"master_usulan"`
	BarisJadwal  []BarisJadwalPlan `json:"baris_jadwal"`
}

type MasterUsulanPlan struct {
	Guru          []MasterGuruPlan    `json:"guru"`
	MataPelajaran []MasterMapelPlan   `json:"mata_pelajaran"`
	Ruangan       []MasterRuanganPlan `json:"ruangan"`
	Kelas         []MasterKelasPlan   `json:"kelas"`
	Jurusan       []MasterJurusanPlan `json:"jurusan"`
}

type MasterGuruPlan struct {
	Ref  string `json:"ref"`
	Nama string `json:"nama"`
	NIP  string `json:"nip"`
	Aksi string `json:"aksi"`
}

type MasterMapelPlan struct {
	Ref               string  `json:"ref"`
	Nama              string  `json:"nama"`
	Kode              string  `json:"kode"`
	JamWajibPerMinggu float64 `json:"jam_wajib_per_minggu"`
	Tingkat           int16   `json:"tingkat"`
	Aksi              string  `json:"aksi"`
}

type MasterRuanganPlan struct {
	Ref         string `json:"ref"`
	Nama        string `json:"nama"`
	Kode        string `json:"kode"`
	TipeRuangan string `json:"tipe_ruangan"`
	Kapasitas   int    `json:"kapasitas"`
	Aksi        string `json:"aksi"`
}

type MasterKelasPlan struct {
	Ref        string `json:"ref"`
	Nama       string `json:"nama"`
	Kode       string `json:"kode"`
	Tingkat    int16  `json:"tingkat"`
	JurusanID  string `json:"jurusan_id"`
	JurusanRef string `json:"jurusan_ref"`
	Aksi       string `json:"aksi"`
}

type MasterJurusanPlan struct {
	Ref  string `json:"ref"`
	Nama string `json:"nama"`
	Kode string `json:"kode"`
	Aksi string `json:"aksi"`
}

type BarisJadwalPlan struct {
	Hari            string  `json:"hari"`
	Jam             string  `json:"jam"`
	Kelas           string  `json:"kelas"`
	MataPelajaran   string  `json:"mata_pelajaran"`
	Guru            string  `json:"guru"`
	Ruangan         string  `json:"ruangan"`
	HariID          string  `json:"hari_id"`
	JamPelajaranID  string  `json:"jam_pelajaran_id"`
	KelasID         string  `json:"kelas_id"`
	MataPelajaranID string  `json:"mata_pelajaran_id"`
	GuruID          string  `json:"guru_id"`
	RuanganID       string  `json:"ruangan_id"`
	KelasRef        string  `json:"kelas_ref"`
	MapelRef        string  `json:"mapel_ref"`
	GuruRef         string  `json:"guru_ref"`
	RuanganRef      string  `json:"ruangan_ref"`
	Status          string  `json:"status"`
	Keyakinan       float64 `json:"keyakinan"`
	Catatan         string  `json:"catatan"`
}

// TerapkanImporRequest: rencana final hasil edit user di layar tinjauan.
type TerapkanImporRequest struct {
	Rencana    ImportPlan `json:"rencana"`
	SemesterID string     `json:"semester_id,omitempty"`
}

type RingkasanKonflikTerapkan struct {
	Jumlah     int `json:"jumlah"`
	Kesalahan  int `json:"kesalahan"`
	Peringatan int `json:"peringatan"`
}

type HasilTerapkanImpor struct {
	JumlahMaster int                      `json:"jumlah_master"`
	JumlahSlot   int                      `json:"jumlah_slot"`
	Dilewati     int                      `json:"dilewati"`
	Konflik      RingkasanKonflikTerapkan `json:"konflik"`
	Peringatan   []string                 `json:"peringatan"`
}
