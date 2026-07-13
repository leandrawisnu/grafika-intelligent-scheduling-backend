package dto

import "encoding/json"

// Validasi konflik
type HasilValidasi struct {
	JumlahKonflik int              `json:"jumlah_konflik"`
	Konflik       []RingkasanKonflik `json:"konflik"`
	Bersih        bool             `json:"bersih"`
}

type RingkasanKonflik struct {
	ID           string `json:"id"`
	Tipe         string `json:"tipe_konflik"`
	Keparahan    string `json:"tingkat_keparahan"`
	Deskripsi    string `json:"deskripsi"`
	Terselesaikan bool  `json:"terselesaikan"`
	Terdeteksi   string `json:"terdeteksi_pada"`
}

// Jadwal Semester
type BuatJadwalSemesterRequest struct {
	SemesterID string `json:"semester_id"`
}

// Jadwal Semester Jurusan
type TambahJurusanRequest struct {
	JurusanIDs []string `json:"jurusan_ids"`
}

// Jadwal Kelas
type BuatJadwalKelasRequest struct {
	KelasID string `json:"kelas_id"`
}

// Penempatan guru
type TugaskanGuruRequest struct {
	GuruID string `json:"guru_id"`
}

type TugaskanBulkRequest struct {
	Tugas []TugasSlotGuru `json:"tugas"`
}

type TugasSlotGuru struct {
	SlotID string `json:"slot_id"`
	GuruID string `json:"guru_id"`
}

// Transisi status
type TransisiStatusRequest struct {
	Status string `json:"status"`
}

// Slot massal
type SlotMassalRequest struct {
	Slots []BuatSlotRequest `json:"slots"`
}

type BuatSlotRequest struct {
	KelasID         string `json:"kelas_id"`
	MataPelajaranID string `json:"mata_pelajaran_id"`
	HariID          string `json:"hari_id"`
	JamPelajaranID  string `json:"jam_pelajaran_id"`
	RuanganID       string `json:"ruangan_id,omitempty"`
	GuruID          string `json:"guru_id,omitempty"`
	MingguKe        int16  `json:"minggu_ke"`
	Terkunci        bool   `json:"terkunci"`
}

// ML service proxy
type PrediksiKonflikRequest struct {
	JadwalSemesterID string        `json:"jadwal_semester_id"`
	Slots            []SlotUntukML `json:"slots"`
	Guru             []GuruUntukML `json:"guru"`
}

type SlotUntukML struct {
	ID       string `json:"id"`
	Kelas    string `json:"kelas"`
	Mapel    string `json:"mata_pelajaran"`
	Hari     string `json:"hari"`
	Jam      string `json:"jam"`
	Ruangan  string `json:"ruangan"`
	Guru     string `json:"guru"`
	MingguKe int    `json:"minggu_ke"`
}

type GuruUntukML struct {
	ID              string   `json:"id"`
	Nama            string   `json:"nama"`
	JamMaksPerMinggu float64  `json:"jam_maksimal_per_minggu"`
	HariLibur       []string `json:"hari_libur"`
	MataPelajaran   []string `json:"mata_pelajaran"`
}

type HasilPrediksiML struct {
	Konflik []KonflikML `json:"konflik"`
}

type KonflikML struct {
	Tipe      string   `json:"tipe"`
	Keparahan string   `json:"keparahan"`
	Deskripsi string   `json:"deskripsi"`
	SlotIDs   []string `json:"slot_ids"`
	Keyakinan float64  `json:"keyakinan"`
}

type MLResolveRequest struct {
	KonflikID string          `json:"konflik_id"`
	Konteks   json.RawMessage `json:"konteks"`
}

type MLResolveResponse struct {
	Alternatif []MLAlternatif `json:"alternatif"`
}

type MLAlternatif struct {
	Peringkat  int             `json:"peringkat"`
	Keyakinan  float64         `json:"keyakinan"`
	Perubahan  json.RawMessage `json:"perubahan"`
	Penjelasan string          `json:"penjelasan"`
}

type MLQueryRequest struct {
	Pertanyaan    string `json:"pertanyaan"`
	JadwalSemesterID string `json:"jadwal_semester_id"`
}

type MLQueryResponse struct {
	Jawaban   string          `json:"jawaban"`
	DataHasil json.RawMessage `json:"data_hasil"`
}

type AIQueryRequest struct {
	Pertanyaan    string `json:"pertanyaan"`
	JadwalSemesterID string `json:"jadwal_semester_id"`
}

type AIQueryResponse struct {
	Jawaban   string          `json:"jawaban"`
	DataHasil json.RawMessage `json:"data_hasil"`
}
