package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"column:dibuat_pada" json:"dibuat_pada"`
	UpdatedAt time.Time      `gorm:"column:diperbarui_pada" json:"diperbarui_pada"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ---- Data Master ----

type TahunAjaran struct {
	BaseModel
	Nama           string    `gorm:"uniqueIndex;not null;column:nama" json:"nama"`
	TanggalMulai   time.Time `gorm:"type:date;not null;column:tanggal_mulai" json:"tanggal_mulai"`
	TanggalSelesai time.Time `gorm:"type:date;not null;column:tanggal_selesai" json:"tanggal_selesai"`
	Aktif          bool      `gorm:"default:true;column:aktif" json:"aktif"`
}

func (TahunAjaran) TableName() string { return "tahun_ajaran" }

type Semester struct {
	BaseModel
	TahunAjaranID  uuid.UUID  `gorm:"not null;column:tahun_ajaran_id" json:"tahun_ajaran_id"`
	TahunAjaran    *TahunAjaran
	Nama           string    `gorm:"not null;column:nama" json:"nama"`
	SemesterKe     int16     `gorm:"not null;column:semester_ke" json:"semester_ke"`
	TanggalMulai   time.Time `gorm:"type:date;not null;column:tanggal_mulai" json:"tanggal_mulai"`
	TanggalSelesai time.Time `gorm:"type:date;not null;column:tanggal_selesai" json:"tanggal_selesai"`
	Aktif          bool      `gorm:"default:true;column:aktif" json:"aktif"`
}

func (Semester) TableName() string { return "semester" }

type Jurusan struct {
	BaseModel
	Kode string `gorm:"uniqueIndex;not null;column:kode;size:20" json:"kode"`
	Nama string `gorm:"not null;column:nama;size:100" json:"nama"`
}

func (Jurusan) TableName() string { return "jurusan" }

type Guru struct {
	BaseModel
	NIP                  string    `gorm:"uniqueIndex;not null;column:nip;size:30" json:"nip"`
	NamaLengkap          string    `gorm:"not null;column:nama_lengkap;size:150" json:"nama_lengkap"`
	JamMaksimalPerMinggu float64   `gorm:"not null;default:40.0;column:jam_maksimal_per_minggu" json:"jam_maksimal_per_minggu"`
	Aktif                bool      `gorm:"default:true;column:aktif" json:"aktif"`
}

func (Guru) TableName() string { return "guru" }

type MataPelajaran struct {
	BaseModel
	Kode              string    `gorm:"uniqueIndex;not null;column:kode;size:20" json:"kode"`
	Nama              string    `gorm:"not null;column:nama;size:150" json:"nama"`
	JamWajibPerMinggu float64   `gorm:"not null;column:jam_wajib_per_minggu" json:"jam_wajib_per_minggu"`
	Tingkat           int16     `gorm:"not null;column:tingkat" json:"tingkat"`
}

func (MataPelajaran) TableName() string { return "mata_pelajaran" }

type Kelas struct {
	BaseModel
	Kode       string    `gorm:"uniqueIndex;not null;column:kode;size:30" json:"kode"`
	Nama       string    `gorm:"not null;column:nama;size:100" json:"nama"`
	Tingkat    int16     `gorm:"not null;column:tingkat" json:"tingkat"`
	JurusanID  uuid.UUID `gorm:"type:uuid;column:jurusan_id" json:"jurusan_id"`
	Jurusan    *Jurusan
	SemesterID uuid.UUID `gorm:"not null;type:uuid;column:semester_id" json:"semester_id"`
	Semester   *Semester
}

func (Kelas) TableName() string { return "kelas" }

type Ruangan struct {
	BaseModel
	Kode        string `gorm:"uniqueIndex;not null;column:kode;size:20" json:"kode"`
	Nama        string `gorm:"not null;column:nama;size:100" json:"nama"`
	Kapasitas   int    `gorm:"not null;default:30;column:kapasitas" json:"kapasitas"`
	TipeRuangan string `gorm:"default:kelas;column:tipe_ruangan;size:30" json:"tipe_ruangan"`
	Aktif       bool   `gorm:"default:true;column:aktif" json:"aktif"`
}

func (Ruangan) TableName() string { return "ruangan" }

type Hari struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nama      string    `gorm:"uniqueIndex;not null;column:nama;size:20" json:"nama"`
	UrutanHari int16    `gorm:"uniqueIndex;not null;column:urutan_hari" json:"urutan_hari"`
	AkhirPekan bool     `gorm:"default:false;column:akhir_pekan" json:"akhir_pekan"`
	DibuatPada time.Time `gorm:"column:dibuat_pada" json:"dibuat_pada"`
}

func (Hari) TableName() string { return "hari" }

type JamPelajaran struct {
	BaseModel
	JamKe        int16  `gorm:"uniqueIndex;not null;column:jam_ke" json:"jam_ke"`
	WaktuMulai   string `gorm:"type:time;not null;column:waktu_mulai" json:"waktu_mulai"`
	WaktuSelesai string `gorm:"type:time;not null;column:waktu_selesai" json:"waktu_selesai"`
	Istirahat    bool   `gorm:"default:false;column:istirahat" json:"istirahat"`
}

func (JamPelajaran) TableName() string { return "jam_pelajaran" }

// ---- Kendala Guru ----

type HariLiburGuru struct {
	BaseModel
	GuruID     uuid.UUID `gorm:"not null;type:uuid;column:guru_id" json:"guru_id"`
	Guru       *Guru
	HariID     uuid.UUID `gorm:"not null;type:uuid;column:hari_id" json:"hari_id"`
	Hari       *Hari
	SemesterID uuid.UUID `gorm:"not null;type:uuid;column:semester_id" json:"semester_id"`
	Semester   *Semester
	Alasan     string    `gorm:"column:alasan;size:255" json:"alasan"`
}

func (HariLiburGuru) TableName() string { return "hari_libur_guru" }

type KualifikasiGuru struct {
	BaseModel
	GuruID          uuid.UUID `gorm:"not null;type:uuid;column:guru_id" json:"guru_id"`
	Guru            *Guru
	MataPelajaranID uuid.UUID `gorm:"not null;type:uuid;column:mata_pelajaran_id" json:"mata_pelajaran_id"`
	MataPelajaran   *MataPelajaran
	TingkatKeahlian string    `gorm:"default:berkualifikasi;column:tingkat_keahlian;size:20" json:"tingkat_keahlian"`
}

func (KualifikasiGuru) TableName() string { return "kualifikasi_guru" }

// ---- JADWAL BARU ----

// Master jadwal per semester
type JadwalSemester struct {
	BaseModel
	SemesterID   uuid.UUID              `gorm:"not null;column:semester_id" json:"semester_id"`
	Semester     *Semester
	Status       string                 `gorm:"not null;default:draf;column:status;size:20" json:"status"`
	BebasKonflik bool                   `gorm:"default:false;column:bebas_konflik" json:"bebas_konflik"`
	Jurusan      []JadwalSemesterJurusan `gorm:"foreignKey:JadwalSemesterID" json:"jurusan,omitempty"`
	JadwalKelas  []JadwalKelas           `gorm:"foreignKey:JadwalSemesterID" json:"jadwal_kelas,omitempty"`
}

func (JadwalSemester) TableName() string { return "jadwal_semester" }

// Jurusan yang terdaftar di jadwal semester
type JadwalSemesterJurusan struct {
	BaseModel
	JadwalSemesterID uuid.UUID `gorm:"not null;type:uuid;column:jadwal_semester_id" json:"jadwal_semester_id"`
	JadwalSemester   *JadwalSemester
	JurusanID        uuid.UUID  `gorm:"not null;type:uuid;column:jurusan_id" json:"jurusan_id"`
	Jurusan          *Jurusan
}

func (JadwalSemesterJurusan) TableName() string { return "jadwal_semester_jurusan" }

// Jadwal per kelas (versioned)
type JadwalKelas struct {
	BaseModel
	JadwalSemesterID uuid.UUID    `gorm:"not null;type:uuid;column:jadwal_semester_id" json:"jadwal_semester_id"`
	JadwalSemester   *JadwalSemester
	JurusanID        uuid.UUID    `gorm:"not null;type:uuid;column:jurusan_id" json:"jurusan_id"`
	Jurusan          *Jurusan
	KelasID          uuid.UUID    `gorm:"not null;type:uuid;column:kelas_id" json:"kelas_id"`
	Kelas            *Kelas
	Versi            int          `gorm:"not null;default:1;column:versi" json:"versi"`
	IsActive         bool         `gorm:"default:false;column:is_active" json:"is_active"`
	SlotJadwal       []SlotJadwal  `gorm:"foreignKey:JadwalKelasID" json:"slot_jadwal,omitempty"`
}

func (JadwalKelas) TableName() string { return "jadwal_kelas" }

// Slot jadwal (FK ke jadwal_kelas)
type SlotJadwal struct {
	BaseModel
	JadwalKelasID    uuid.UUID     `gorm:"not null;type:uuid;column:jadwal_kelas_id" json:"jadwal_kelas_id"`
	JadwalKelas      *JadwalKelas
	KelasID          uuid.UUID     `gorm:"not null;type:uuid;column:kelas_id" json:"kelas_id"`
	Kelas            *Kelas
	MataPelajaranID  uuid.UUID     `gorm:"not null;type:uuid;column:mata_pelajaran_id" json:"mata_pelajaran_id"`
	MataPelajaran    *MataPelajaran
	HariID           uuid.UUID     `gorm:"not null;type:uuid;column:hari_id" json:"hari_id"`
	Hari             *Hari
	JamPelajaranID   uuid.UUID     `gorm:"not null;type:uuid;column:jam_pelajaran_id" json:"jam_pelajaran_id"`
	JamPelajaran     *JamPelajaran
	RuanganID        uuid.UUID     `gorm:"type:uuid;column:ruangan_id" json:"ruangan_id"`
	Ruangan          *Ruangan
	GuruID           uuid.UUID     `gorm:"type:uuid;column:guru_id" json:"guru_id"`
	Guru             *Guru
	MingguKe         int16         `gorm:"default:1;column:minggu_ke" json:"minggu_ke"`
	Terkunci         bool          `gorm:"default:false;column:terkunci" json:"terkunci"`
}

func (SlotJadwal) TableName() string { return "slot_jadwal" }

// ---- Konflik ----

type Konflik struct {
	BaseModel
	JadwalSemesterID  uuid.UUID  `gorm:"not null;type:uuid;column:jadwal_semester_id" json:"jadwal_semester_id"`
	JadwalSemester    *JadwalSemester
	TipeKonflik       string     `gorm:"not null;column:tipe_konflik;size:40" json:"tipe_konflik"`
	TingkatKeparahan  string     `gorm:"not null;default:kesalahan;column:tingkat_keparahan;size:15" json:"tingkat_keparahan"`
	SlotAID           *uuid.UUID `gorm:"type:uuid;column:slot_a_id" json:"slot_a_id"`
	SlotA             *SlotJadwal `gorm:"foreignKey:SlotAID"`
	SlotBID           *uuid.UUID `gorm:"type:uuid;column:slot_b_id" json:"slot_b_id"`
	SlotB             *SlotJadwal `gorm:"foreignKey:SlotBID"`
	GuruID            *uuid.UUID `gorm:"type:uuid;column:guru_id" json:"guru_id"`
	Guru              *Guru
	Deskripsi         string     `gorm:"not null;type:text;column:deskripsi" json:"deskripsi"`
	DetailJSON        *string    `gorm:"type:jsonb;column:detail_json" json:"detail_json"`
	Terselesaikan     bool       `gorm:"default:false;column:terselesaikan" json:"terselesaikan"`
	DiselesaikanOleh  *string    `gorm:"column:diselesaikan_oleh;size:20" json:"diselesaikan_oleh"`
	TerdeteksiPada    time.Time  `gorm:"default:now();column:terdeteksi_pada" json:"terdeteksi_pada"`
	TerselesaikanPada *time.Time `gorm:"column:terselesaikan_pada" json:"terselesaikan_pada"`
}

func (Konflik) TableName() string { return "konflik" }

// ---- Resolusi AI ----

type ResolusiAI struct {
	BaseModel
	KonflikID            uuid.UUID `gorm:"not null;type:uuid;column:konflik_id" json:"konflik_id"`
	JadwalSemesterID     uuid.UUID `gorm:"not null;type:uuid;column:jadwal_semester_id" json:"jadwal_semester_id"`
	JadwalSemester       *JadwalSemester
	Peringkat            int16     `gorm:"not null;column:peringkat" json:"peringkat"`
	SkorKeyakinan        float64   `gorm:"not null;column:skor_keyakinan" json:"skor_keyakinan"`
	UsulanPerubahanJSON  string    `gorm:"not null;type:jsonb;column:usulan_perubahan_json" json:"usulan_perubahan_json"`
	Penjelasan           string    `gorm:"not null;type:text;column:penjelasan" json:"penjelasan"`
	Diterima             bool      `gorm:"default:false;column:diterima" json:"diterima"`
}

func (ResolusiAI) TableName() string { return "resolusi_ai" }

// ---- Log Audit ----

type LogAuditJadwal struct {
	BaseModel
	JadwalSemesterID uuid.UUID `gorm:"not null;type:uuid;column:jadwal_semester_id" json:"jadwal_semester_id"`
	JadwalSemester   *JadwalSemester
	Aksi             string    `gorm:"not null;column:aksi;size:50" json:"aksi"`
	PerubahanJSON    string    `gorm:"not null;type:jsonb;column:perubahan_json" json:"perubahan_json"`
	DilakukanOleh    string    `gorm:"default:sistem;column:dilakukan_oleh;size:100" json:"dilakukan_oleh"`
}

func (LogAuditJadwal) TableName() string { return "log_audit_jadwal" }
