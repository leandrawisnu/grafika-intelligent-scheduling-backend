package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/pkg/storage"
	"github.com/grafika-scheduling/backend/src/dto"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

const batasBerkasImporAI = 15 * 1024 * 1024

// LayananImporAI menjalankan job AI import dokumen: parse (LlamaParse via ML
// service) lalu petakan ke master (LLM via ML service), dan menerapkan hasilnya
// secara transaksional.
type LayananImporAI struct {
	db    *gorm.DB
	ml    *mlclient.Client
	objek *storage.Client
}

func NewLayananImporAI(db *gorm.DB, ml *mlclient.Client, objek *storage.Client) *LayananImporAI {
	return &LayananImporAI{db: db, ml: ml, objek: objek}
}

// JalankanAnalisis memproses satu dokumen impor: parse bila teks belum ada,
// lalu petakan ke master dan simpan rencana tervalidasi. Aman dipanggil dari
// goroutine; semua kegagalan disimpan pada baris job, bukan panic.
func (s *LayananImporAI) JalankanAnalisis(jobID uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			s.gagal(jobID, fmt.Sprintf("Analisis berhenti tak terduga: %v", r))
		}
	}()

	var job models.DokumenImpor
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return
	}

	teks := ""
	if job.TeksMarkdown != nil {
		teks = strings.TrimSpace(*job.TeksMarkdown)
	}
	if teks == "" {
		s.simpan(jobID, map[string]any{"status": "memproses", "tahap": "parsing", "pesan": nil})
		isi, err := s.bacaBerkas(&job)
		if err != nil {
			s.gagal(jobID, err.Error())
			return
		}
		hasil, err := s.ml.ParseDokumen(job.NamaBerkas, isi)
		if err != nil {
			s.gagal(jobID, "Parsing dokumen gagal: "+err.Error())
			return
		}
		teks = strings.TrimSpace(hasil.Markdown)
		if teks == "" {
			teks = strings.TrimSpace(hasil.Teks)
		}
		if teks == "" {
			s.gagal(jobID, "Dokumen tidak menghasilkan teks yang bisa diproses.")
			return
		}
		s.simpan(jobID, map[string]any{"teks_markdown": teks})
	}

	s.simpan(jobID, map[string]any{"status": "memproses", "tahap": "memetakan", "pesan": nil})
	konteks, err := s.konteksPemetaan(&job)
	if err != nil {
		s.gagal(jobID, "Gagal menyiapkan konteks pemetaan: "+err.Error())
		return
	}
	raw, err := s.ml.PetakanDokumen(teks, konteks)
	if err != nil {
		s.gagal(jobID, "Analisis AI gagal: "+err.Error())
		return
	}
	var plan dto.ImportPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		s.gagal(jobID, "Jawaban analisis AI tidak dapat dibaca.")
		return
	}
	kat, err := s.katalogValidasi(job.SemesterID, job.JadwalSemesterID)
	if err != nil {
		s.gagal(jobID, "Gagal memuat katalog validasi: "+err.Error())
		return
	}
	plan = ValidasiRencana(plan, kat)
	rencanaJSON, err := json.Marshal(plan)
	if err != nil {
		s.gagal(jobID, "Gagal menyimpan rencana impor.")
		return
	}
	s.simpan(jobID, map[string]any{
		"status": "siap", "tahap": "siap", "pesan": nil, "rencana_json": string(rencanaJSON),
	})
}

func (s *LayananImporAI) simpan(jobID uuid.UUID, data map[string]any) {
	s.db.Model(&models.DokumenImpor{}).Where("id = ?", jobID).Updates(data)
}

func (s *LayananImporAI) gagal(jobID uuid.UUID, pesan string) {
	if len(pesan) > 500 {
		pesan = pesan[:500]
	}
	s.simpan(jobID, map[string]any{"status": "gagal", "pesan": pesan})
}

func (s *LayananImporAI) bacaBerkas(job *models.DokumenImpor) ([]byte, error) {
	if s.objek == nil {
		return nil, fmt.Errorf("penyimpanan berkas (MinIO) belum dikonfigurasi")
	}
	ctx, batal := context.WithTimeout(context.Background(), 60*time.Second)
	defer batal()
	sumber, err := s.objek.Get(ctx, job.KunciBerkas)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca berkas impor: %w", err)
	}
	defer sumber.Close()
	isi, err := io.ReadAll(io.LimitReader(sumber, batasBerkasImporAI+1))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca berkas impor: %w", err)
	}
	return isi, nil
}

// konteksPemetaan menyusun katalog ringkas + daftar slot terpakai untuk prompt LLM.
func (s *LayananImporAI) konteksPemetaan(job *models.DokumenImpor) (map[string]any, error) {
	semester := map[string]any{}
	if job.SemesterID != nil {
		var sem models.Semester
		if err := s.db.First(&sem, "id = ?", *job.SemesterID).Error; err != nil {
			return nil, fmt.Errorf("semester tidak ditemukan")
		}
		semester = map[string]any{"id": sem.ID, "nama": sem.Nama, "semester_ke": sem.SemesterKe}
	}

	var hari []models.Hari
	if err := s.db.Order("urutan_hari").Find(&hari).Error; err != nil {
		return nil, err
	}
	hariList := make([]map[string]any, 0, len(hari))
	for _, h := range hari {
		hariList = append(hariList, map[string]any{"id": h.ID, "nama": h.Nama, "akhir_pekan": h.AkhirPekan})
	}

	var jam []models.JamPelajaran
	if err := s.db.Order("jam_ke").Find(&jam).Error; err != nil {
		return nil, err
	}
	jamList := make([]map[string]any, 0, len(jam))
	for _, j := range jam {
		jamList = append(jamList, map[string]any{
			"id": j.ID, "jam_ke": j.JamKe, "mulai": j.WaktuMulai, "istirahat": j.Istirahat,
		})
	}

	kelasQ := s.db.Model(&models.Kelas{}).Order("kode")
	if job.SemesterID != nil {
		kelasQ = kelasQ.Where("semester_id = ?", *job.SemesterID)
	}
	var kelas []models.Kelas
	if err := kelasQ.Find(&kelas).Error; err != nil {
		return nil, err
	}
	kelasList := make([]map[string]any, 0, len(kelas))
	for _, k := range kelas {
		kelasList = append(kelasList, map[string]any{"id": k.ID, "nama": k.Nama, "kode": k.Kode})
	}

	var guru []models.Guru
	if err := s.db.Where("aktif = ?", true).Order("nama_lengkap").Find(&guru).Error; err != nil {
		return nil, err
	}
	guruList := make([]map[string]any, 0, len(guru))
	for _, g := range guru {
		guruList = append(guruList, map[string]any{"id": g.ID, "nama": g.NamaLengkap, "nip": g.NIP})
	}

	var mapel []models.MataPelajaran
	if err := s.db.Order("kode").Find(&mapel).Error; err != nil {
		return nil, err
	}
	mapelList := make([]map[string]any, 0, len(mapel))
	for _, m := range mapel {
		mapelList = append(mapelList, map[string]any{"id": m.ID, "nama": m.Nama, "kode": m.Kode})
	}

	ruanganQ := s.db.Model(&models.Ruangan{}).Order("kode")
	if job.SemesterID != nil {
		ruanganQ = ruanganQ.Where("semester_id = ?", *job.SemesterID)
	}
	var ruangan []models.Ruangan
	if err := ruanganQ.Find(&ruangan).Error; err != nil {
		return nil, err
	}
	ruanganList := make([]map[string]any, 0, len(ruangan))
	for _, r := range ruangan {
		ruanganList = append(ruanganList, map[string]any{"id": r.ID, "nama": r.Nama, "kode": r.Kode})
	}

	jurusanQ := s.db.Model(&models.Jurusan{}).Order("kode")
	if job.SemesterID != nil {
		jurusanQ = jurusanQ.Where("semester_id = ?", *job.SemesterID)
	}
	var jurusan []models.Jurusan
	if err := jurusanQ.Find(&jurusan).Error; err != nil {
		return nil, err
	}
	jurusanList := make([]map[string]any, 0, len(jurusan))
	for _, j := range jurusan {
		jurusanList = append(jurusanList, map[string]any{"id": j.ID, "nama": j.Nama, "kode": j.Kode})
	}

	slot := []string{}
	if job.JadwalSemesterID != nil {
		set, err := NewLayananJadwal(s.db).KunciSlotAktif(*job.JadwalSemesterID)
		if err != nil {
			return nil, err
		}
		for k := range set {
			slot = append(slot, k)
		}
		sort.Strings(slot)
	}

	return map[string]any{
		"target":   job.Target,
		"semester": semester,
		"katalog": map[string]any{
			"hari":           hariList,
			"jam":            jamList,
			"kelas":          kelasList,
			"guru":           guruList,
			"mata_pelajaran": mapelList,
			"ruangan":        ruanganList,
			"jurusan":        jurusanList,
		},
		"slot_terpakai": slot,
	}, nil
}

// KatalogValidasi adalah himpunan id yang sah plus ref master usulan,
// dipakai ValidasiRencana secara murni (tanpa akses DB) agar mudah diuji.
type KatalogValidasi struct {
	Hari         map[string]bool
	Jam          map[string]bool
	Kelas        map[string]bool
	Mapel        map[string]bool
	Guru         map[string]bool
	Ruangan      map[string]bool
	Jurusan      map[string]bool
	SlotTerpakai map[string]bool
	RefGuru      map[string]bool
	RefMapel     map[string]bool
	RefRuangan   map[string]bool
	RefKelas     map[string]bool
	RefJurusan   map[string]bool
}

func (s *LayananImporAI) katalogValidasi(semesterID, jsID *uuid.UUID) (KatalogValidasi, error) {
	kat := KatalogValidasi{
		Hari: map[string]bool{}, Jam: map[string]bool{}, Kelas: map[string]bool{},
		Mapel: map[string]bool{}, Guru: map[string]bool{}, Ruangan: map[string]bool{},
		Jurusan: map[string]bool{}, SlotTerpakai: map[string]bool{},
		RefGuru: map[string]bool{}, RefMapel: map[string]bool{}, RefRuangan: map[string]bool{},
		RefKelas: map[string]bool{}, RefJurusan: map[string]bool{},
	}

	var hari []models.Hari
	if err := s.db.Find(&hari).Error; err != nil {
		return kat, err
	}
	for _, h := range hari {
		kat.Hari[h.ID.String()] = true
	}

	var jam []models.JamPelajaran
	if err := s.db.Find(&jam).Error; err != nil {
		return kat, err
	}
	for _, j := range jam {
		kat.Jam[j.ID.String()] = true
	}

	var guru []models.Guru
	if err := s.db.Find(&guru).Error; err != nil {
		return kat, err
	}
	for _, g := range guru {
		kat.Guru[g.ID.String()] = true
	}

	var mapel []models.MataPelajaran
	if err := s.db.Find(&mapel).Error; err != nil {
		return kat, err
	}
	for _, m := range mapel {
		kat.Mapel[m.ID.String()] = true
	}

	if semesterID != nil {
		var kelas []models.Kelas
		if err := s.db.Where("semester_id = ?", *semesterID).Find(&kelas).Error; err != nil {
			return kat, err
		}
		for _, k := range kelas {
			kat.Kelas[k.ID.String()] = true
		}
		var ruangan []models.Ruangan
		if err := s.db.Where("semester_id = ?", *semesterID).Find(&ruangan).Error; err != nil {
			return kat, err
		}
		for _, r := range ruangan {
			kat.Ruangan[r.ID.String()] = true
		}
		var jurusan []models.Jurusan
		if err := s.db.Where("semester_id = ?", *semesterID).Find(&jurusan).Error; err != nil {
			return kat, err
		}
		for _, j := range jurusan {
			kat.Jurusan[j.ID.String()] = true
		}
	}

	if jsID != nil {
		set, err := NewLayananJadwal(s.db).KunciSlotAktif(*jsID)
		if err != nil {
			return kat, err
		}
		for k := range set {
			kat.SlotTerpakai[k] = true
		}
	}
	return kat, nil
}

// ValidasiRencana membersihkan rencana dari id/ref yang tidak sah dan
// menghitung ulang status setiap baris secara deterministik.
func ValidasiRencana(plan dto.ImportPlan, kat KatalogValidasi) dto.ImportPlan {
	plan.MasterUsulan, kat = bersihkanMaster(plan.MasterUsulan, kat)

	for i := range plan.BarisJadwal {
		b := &plan.BarisJadwal[i]
		b.HariID = pilihID(b.HariID, kat.Hari)
		b.JamPelajaranID = pilihID(b.JamPelajaranID, kat.Jam)
		b.KelasID = pilihID(b.KelasID, kat.Kelas)
		b.MataPelajaranID = pilihID(b.MataPelajaranID, kat.Mapel)
		b.GuruID = pilihID(b.GuruID, kat.Guru)
		b.RuanganID = pilihID(b.RuanganID, kat.Ruangan)
		b.KelasRef = pilihRef(b.KelasRef, kat.RefKelas)
		b.MapelRef = pilihRef(b.MapelRef, kat.RefMapel)
		b.GuruRef = pilihRef(b.GuruRef, kat.RefGuru)
		b.RuanganRef = pilihRef(b.RuanganRef, kat.RefRuangan)
		if b.KelasID != "" {
			b.KelasRef = ""
		}
		if b.MataPelajaranID != "" {
			b.MapelRef = ""
		}
		if b.GuruID != "" {
			b.GuruRef = ""
		}
		if b.RuanganID != "" {
			b.RuanganRef = ""
		}

		lengkap := b.HariID != "" && b.JamPelajaranID != "" &&
			(b.KelasID != "" || b.KelasRef != "") &&
			(b.MataPelajaranID != "" || b.MapelRef != "")
		pakaiRef := b.KelasRef != "" || b.MapelRef != "" || b.GuruRef != "" || b.RuanganRef != ""
		switch {
		case !lengkap:
			b.Status = "perlu_pilihan"
		case pakaiRef:
			b.Status = "akan_dibuat"
		case kat.SlotTerpakai[b.KelasID+"|"+b.HariID+"|"+b.JamPelajaranID]:
			b.Status = "sudah_ada"
		default:
			b.Status = "siap"
		}
	}
	return plan
}

func bersihkanMaster(m dto.MasterUsulanPlan, kat KatalogValidasi) (dto.MasterUsulanPlan, KatalogValidasi) {
	out := dto.MasterUsulanPlan{
		Guru:          make([]dto.MasterGuruPlan, 0, len(m.Guru)),
		MataPelajaran: make([]dto.MasterMapelPlan, 0, len(m.MataPelajaran)),
		Ruangan:       make([]dto.MasterRuanganPlan, 0, len(m.Ruangan)),
		Kelas:         make([]dto.MasterKelasPlan, 0, len(m.Kelas)),
		Jurusan:       make([]dto.MasterJurusanPlan, 0, len(m.Jurusan)),
	}
	refGuru := map[string]bool{}
	refMapel := map[string]bool{}
	refRuangan := map[string]bool{}
	refKelas := map[string]bool{}
	refJurusan := map[string]bool{}

	for _, item := range m.Guru {
		item.Ref = strings.TrimSpace(item.Ref)
		item.Nama = strings.TrimSpace(item.Nama)
		item.NIP = strings.TrimSpace(item.NIP)
		if item.Ref == "" || item.Nama == "" || refGuru[item.Ref] {
			continue
		}
		refGuru[item.Ref] = true
		out.Guru = append(out.Guru, item)
	}
	for _, item := range m.MataPelajaran {
		item.Ref = strings.TrimSpace(item.Ref)
		item.Nama = strings.TrimSpace(item.Nama)
		item.Kode = strings.TrimSpace(item.Kode)
		if item.Ref == "" || item.Nama == "" || refMapel[item.Ref] {
			continue
		}
		refMapel[item.Ref] = true
		out.MataPelajaran = append(out.MataPelajaran, item)
	}
	for _, item := range m.Ruangan {
		item.Ref = strings.TrimSpace(item.Ref)
		item.Nama = strings.TrimSpace(item.Nama)
		item.Kode = strings.TrimSpace(item.Kode)
		if item.Ref == "" || item.Nama == "" || refRuangan[item.Ref] {
			continue
		}
		refRuangan[item.Ref] = true
		out.Ruangan = append(out.Ruangan, item)
	}
	for _, item := range m.Jurusan {
		item.Ref = strings.TrimSpace(item.Ref)
		item.Nama = strings.TrimSpace(item.Nama)
		item.Kode = strings.TrimSpace(item.Kode)
		if item.Ref == "" || item.Nama == "" || refJurusan[item.Ref] {
			continue
		}
		refJurusan[item.Ref] = true
		out.Jurusan = append(out.Jurusan, item)
	}
	for _, item := range m.Kelas {
		item.Ref = strings.TrimSpace(item.Ref)
		item.Nama = strings.TrimSpace(item.Nama)
		item.Kode = strings.TrimSpace(item.Kode)
		item.JurusanID = pilihID(item.JurusanID, kat.Jurusan)
		if item.Ref == "" || item.Nama == "" || refKelas[item.Ref] {
			continue
		}
		if item.JurusanID == "" {
			item.JurusanRef = pilihRef(item.JurusanRef, refJurusan)
		}
		refKelas[item.Ref] = true
		out.Kelas = append(out.Kelas, item)
	}

	kat.RefGuru = refGuru
	kat.RefMapel = refMapel
	kat.RefRuangan = refRuangan
	kat.RefKelas = refKelas
	kat.RefJurusan = refJurusan
	return out, kat
}

func pilihID(id string, sah map[string]bool) string {
	id = strings.TrimSpace(id)
	if id == "" || !sah[id] {
		return ""
	}
	return id
}

func pilihRef(ref string, sah map[string]bool) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || !sah[ref] {
		return ""
	}
	return ref
}

// TerapkanRencana menerapkan rencana final (hasil edit user) dalam satu transaksi:
// buat master yang disetujui, simpan slot jadwal, lalu deteksi konflik.
func (s *LayananImporAI) TerapkanRencana(job *models.DokumenImpor, req dto.TerapkanImporRequest) (*dto.HasilTerapkanImpor, error) {
	if job.Status == "diterapkan" {
		return nil, fmt.Errorf("rencana ini sudah diterapkan sebelumnya")
	}
	if job.Status != "siap" {
		return nil, fmt.Errorf("rencana belum siap diterapkan")
	}
	if job.JadwalSemesterID == nil && len(req.Rencana.BarisJadwal) > 0 {
		return nil, fmt.Errorf("impor jadwal membutuhkan jadwal semester aktif")
	}

	semesterID := job.SemesterID
	if req.SemesterID != "" {
		if id, err := uuid.Parse(req.SemesterID); err == nil {
			semesterID = &id
		}
	}
	if semesterID != nil {
		var sem models.Semester
		if err := s.db.Select("id").First(&sem, "id = ?", *semesterID).Error; err != nil {
			semesterID = nil
		}
	}

	kat, err := s.katalogValidasi(semesterID, job.JadwalSemesterID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat katalog validasi: %w", err)
	}
	plan := ValidasiRencana(req.Rencana, kat)

	butuhSemester := len(plan.MasterUsulan.Kelas) > 0 || len(plan.MasterUsulan.Ruangan) > 0 || len(plan.MasterUsulan.Jurusan) > 0
	if butuhSemester && semesterID == nil {
		return nil, fmt.Errorf("kelas/ruangan/jurusan baru membutuhkan semester yang jelas")
	}

	hasil := &dto.HasilTerapkanImpor{Peringatan: []string{}}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		refJurusan, jumlah, pesan, err := pastikanJurusan(tx, plan.MasterUsulan.Jurusan, semesterID)
		if err != nil {
			return err
		}
		hasil.JumlahMaster += jumlah
		hasil.Peringatan = append(hasil.Peringatan, pesan...)

		refGuru, jumlah, pesan, err := pastikanGuru(tx, plan.MasterUsulan.Guru)
		if err != nil {
			return err
		}
		hasil.JumlahMaster += jumlah
		hasil.Peringatan = append(hasil.Peringatan, pesan...)

		refMapel, jumlah, pesan, err := pastikanMapel(tx, plan.MasterUsulan.MataPelajaran)
		if err != nil {
			return err
		}
		hasil.JumlahMaster += jumlah
		hasil.Peringatan = append(hasil.Peringatan, pesan...)

		refRuangan, jumlah, pesan, err := pastikanRuangan(tx, plan.MasterUsulan.Ruangan, semesterID)
		if err != nil {
			return err
		}
		hasil.JumlahMaster += jumlah
		hasil.Peringatan = append(hasil.Peringatan, pesan...)

		refKelas, jumlah, pesan, err := pastikanKelas(tx, plan.MasterUsulan.Kelas, semesterID, refJurusan)
		if err != nil {
			return err
		}
		hasil.JumlahMaster += jumlah
		hasil.Peringatan = append(hasil.Peringatan, pesan...)

		baris, dilewati := susunBarisImpor(plan.BarisJadwal, refKelas, refMapel, refGuru, refRuangan)
		hasil.Dilewati += dilewati
		if len(baris) > 0 {
			if semesterID == nil {
				return fmt.Errorf("semester tidak diketahui untuk menyimpan slot jadwal")
			}
			jumlahSlot, err := NewLayananJadwal(tx).SimpanImpor(*job.JadwalSemesterID, *semesterID, baris)
			if err != nil {
				if strings.Contains(err.Error(), "tidak ada baris baru") {
					jumlahSlot = 0
				} else {
					return fmt.Errorf("gagal menyimpan slot jadwal: %w", err)
				}
			}
			hasil.JumlahSlot = jumlahSlot
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if hasil.JumlahSlot > 0 && job.JadwalSemesterID != nil {
		konflik, errKonflik := NewLayananKonflik(s.db).DeteksiKonflik(*job.JadwalSemesterID)
		if errKonflik != nil {
			hasil.Peringatan = append(hasil.Peringatan, "Slot tersimpan, tetapi deteksi konflik gagal: "+errKonflik.Error())
		} else {
			for _, k := range konflik {
				if k.TingkatKeparahan == "kesalahan" {
					hasil.Konflik.Kesalahan++
				} else {
					hasil.Konflik.Peringatan++
				}
			}
			hasil.Konflik.Jumlah = len(konflik)
		}
	}

	hasilJSON, _ := json.Marshal(hasil)
	s.simpan(job.ID, map[string]any{
		"status": "diterapkan", "tahap": "selesai", "pesan": nil, "hasil_terapkan_json": string(hasilJSON),
	})
	return hasil, nil
}

func susunBarisImpor(rows []dto.BarisJadwalPlan, refKelas, refMapel, refGuru, refRuangan map[string]uuid.UUID) ([]BarisImpor, int) {
	out := make([]BarisImpor, 0, len(rows))
	dilewati := 0
	for _, b := range rows {
		if b.Status != "siap" && b.Status != "akan_dibuat" {
			dilewati++
			continue
		}
		kelasID := uuidDari(b.KelasID, refKelas[b.KelasRef])
		mapelID := uuidDari(b.MataPelajaranID, refMapel[b.MapelRef])
		hariID, errHari := uuid.Parse(strings.TrimSpace(b.HariID))
		jamID, errJam := uuid.Parse(strings.TrimSpace(b.JamPelajaranID))
		if kelasID == uuid.Nil || mapelID == uuid.Nil || errHari != nil || errJam != nil {
			dilewati++
			continue
		}
		out = append(out, BarisImpor{
			KelasID:         kelasID,
			MataPelajaranID: mapelID,
			HariID:          hariID,
			JamPelajaranID:  jamID,
			GuruID:          uuidOpsionalDari(b.GuruID, refGuru[b.GuruRef]),
			RuanganID:       uuidOpsionalDari(b.RuanganID, refRuangan[b.RuanganRef]),
		})
	}
	return out, dilewati
}

func uuidDari(teks string, dariRef uuid.UUID) uuid.UUID {
	if id, err := uuid.Parse(strings.TrimSpace(teks)); err == nil {
		return id
	}
	return dariRef
}

func uuidOpsionalDari(teks string, dariRef uuid.UUID) *uuid.UUID {
	id := uuidDari(teks, dariRef)
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func pastikanJurusan(tx *gorm.DB, items []dto.MasterJurusanPlan, semesterID *uuid.UUID) (map[string]uuid.UUID, int, []string, error) {
	ref := map[string]uuid.UUID{}
	if len(items) == 0 {
		return ref, 0, nil, nil
	}
	if semesterID == nil {
		return ref, 0, nil, fmt.Errorf("jurusan baru membutuhkan semester yang jelas")
	}
	dibuat := 0
	peringatan := []string{}
	for _, item := range items {
		if item.Aksi != "buat" {
			continue
		}
		nama := strings.TrimSpace(item.Nama)
		if nama == "" {
			continue
		}
		kode := strings.TrimSpace(item.Kode)
		var ada models.Jurusan
		if kode == "" {
			if err := tx.Where("semester_id = ? AND nama ILIKE ?", *semesterID, nama).First(&ada).Error; err == nil {
				ref[item.Ref] = ada.ID
				continue
			}
			kode = kodeDariNama(nama, "JRS")
		} else if err := tx.Where("semester_id = ? AND (kode = ? OR nama ILIKE ?)", *semesterID, kode, nama).First(&ada).Error; err == nil {
			ref[item.Ref] = ada.ID
			continue
		}
		kode = pastikanKodeUnik(tx, "jurusan", kode, "semester_id = ?", *semesterID)
		baru := models.Jurusan{Kode: kode, Nama: nama, SemesterID: *semesterID}
		if err := tx.Create(&baru).Error; err != nil {
			return ref, dibuat, peringatan, fmt.Errorf("gagal membuat jurusan %s: %w", nama, err)
		}
		ref[item.Ref] = baru.ID
		dibuat++
	}
	return ref, dibuat, peringatan, nil
}

func pastikanGuru(tx *gorm.DB, items []dto.MasterGuruPlan) (map[string]uuid.UUID, int, []string, error) {
	ref := map[string]uuid.UUID{}
	if len(items) == 0 {
		return ref, 0, nil, nil
	}
	dibuat := 0
	peringatan := []string{}
	for _, item := range items {
		if item.Aksi != "buat" {
			continue
		}
		nama := strings.TrimSpace(item.Nama)
		if nama == "" {
			continue
		}
		nip := strings.TrimSpace(item.NIP)
		var ada models.Guru
		if nip != "" {
			if err := tx.Where("nip = ?", nip).First(&ada).Error; err == nil {
				ref[item.Ref] = ada.ID
				continue
			}
		} else if err := tx.Where("nama_lengkap ILIKE ?", nama).First(&ada).Error; err == nil {
			ref[item.Ref] = ada.ID
			continue
		}
		if nip == "" {
			nip = nipOtomatis(tx)
			peringatan = append(peringatan, fmt.Sprintf("Guru %s dibuat dengan NIP sementara %s — lengkapi di Data Guru.", nama, nip))
		}
		baru := models.Guru{NIP: nip, NamaLengkap: nama, JamMaksimalPerMinggu: 40, Aktif: true}
		if err := tx.Create(&baru).Error; err != nil {
			return ref, dibuat, peringatan, fmt.Errorf("gagal membuat guru %s: %w", nama, err)
		}
		ref[item.Ref] = baru.ID
		dibuat++
	}
	return ref, dibuat, peringatan, nil
}

func pastikanMapel(tx *gorm.DB, items []dto.MasterMapelPlan) (map[string]uuid.UUID, int, []string, error) {
	ref := map[string]uuid.UUID{}
	if len(items) == 0 {
		return ref, 0, nil, nil
	}
	dibuat := 0
	peringatan := []string{}
	for _, item := range items {
		if item.Aksi != "buat" {
			continue
		}
		nama := strings.TrimSpace(item.Nama)
		if nama == "" {
			continue
		}
		kode := strings.TrimSpace(item.Kode)
		var ada models.MataPelajaran
		if kode == "" {
			if err := tx.Where("nama ILIKE ?", nama).First(&ada).Error; err == nil {
				ref[item.Ref] = ada.ID
				continue
			}
			kode = kodeDariNama(nama, "MAPEL")
		} else if err := tx.Where("kode = ? OR nama ILIKE ?", kode, nama).First(&ada).Error; err == nil {
			ref[item.Ref] = ada.ID
			continue
		}
		kode = pastikanKodeUnik(tx, "mata_pelajaran", kode, "")
		jam := item.JamWajibPerMinggu
		if jam <= 0 {
			jam = 2
		}
		tingkat := item.Tingkat
		if tingkat <= 0 {
			tingkat = 10
		}
		baru := models.MataPelajaran{Kode: kode, Nama: nama, JamWajibPerMinggu: jam, Tingkat: tingkat}
		if err := tx.Create(&baru).Error; err != nil {
			return ref, dibuat, peringatan, fmt.Errorf("gagal membuat mata pelajaran %s: %w", nama, err)
		}
		ref[item.Ref] = baru.ID
		dibuat++
	}
	return ref, dibuat, peringatan, nil
}

func pastikanRuangan(tx *gorm.DB, items []dto.MasterRuanganPlan, semesterID *uuid.UUID) (map[string]uuid.UUID, int, []string, error) {
	ref := map[string]uuid.UUID{}
	if len(items) == 0 {
		return ref, 0, nil, nil
	}
	if semesterID == nil {
		return ref, 0, nil, fmt.Errorf("ruangan baru membutuhkan semester yang jelas")
	}
	dibuat := 0
	peringatan := []string{}
	for _, item := range items {
		if item.Aksi != "buat" {
			continue
		}
		nama := strings.TrimSpace(item.Nama)
		if nama == "" {
			continue
		}
		kode := strings.TrimSpace(item.Kode)
		var ada models.Ruangan
		if kode == "" {
			if err := tx.Where("semester_id = ? AND nama ILIKE ?", *semesterID, nama).First(&ada).Error; err == nil {
				ref[item.Ref] = ada.ID
				continue
			}
			kode = kodeDariNama(nama, "RNG")
		} else if err := tx.Where("semester_id = ? AND (kode = ? OR nama ILIKE ?)", *semesterID, kode, nama).First(&ada).Error; err == nil {
			ref[item.Ref] = ada.ID
			continue
		}
		kode = pastikanKodeUnik(tx, "ruangan", kode, "semester_id = ?", *semesterID)
		kapasitas := item.Kapasitas
		if kapasitas <= 0 {
			kapasitas = 30
		}
		tipe := strings.TrimSpace(item.TipeRuangan)
		if tipe == "" {
			tipe = "kelas"
		}
		baru := models.Ruangan{
			Kode: kode, Nama: nama, Kapasitas: kapasitas,
			TipeRuangan: tipe, Aktif: true, SemesterID: *semesterID,
		}
		if err := tx.Create(&baru).Error; err != nil {
			return ref, dibuat, peringatan, fmt.Errorf("gagal membuat ruangan %s: %w", nama, err)
		}
		ref[item.Ref] = baru.ID
		dibuat++
	}
	return ref, dibuat, peringatan, nil
}

func pastikanKelas(tx *gorm.DB, items []dto.MasterKelasPlan, semesterID *uuid.UUID, refJurusan map[string]uuid.UUID) (map[string]uuid.UUID, int, []string, error) {
	ref := map[string]uuid.UUID{}
	if len(items) == 0 {
		return ref, 0, nil, nil
	}
	if semesterID == nil {
		return ref, 0, nil, fmt.Errorf("kelas baru membutuhkan semester yang jelas")
	}
	dibuat := 0
	peringatan := []string{}
	for _, item := range items {
		if item.Aksi != "buat" {
			continue
		}
		nama := strings.TrimSpace(item.Nama)
		if nama == "" {
			continue
		}
		kode := strings.TrimSpace(item.Kode)
		var ada models.Kelas
		if kode == "" {
			if err := tx.Where("semester_id = ? AND nama ILIKE ?", *semesterID, nama).First(&ada).Error; err == nil {
				ref[item.Ref] = ada.ID
				continue
			}
			kode = kodeDariNama(nama, "KLS")
		} else if err := tx.Where("semester_id = ? AND (kode = ? OR nama ILIKE ?)", *semesterID, kode, nama).First(&ada).Error; err == nil {
			ref[item.Ref] = ada.ID
			continue
		}
		jurusanID := uuidDari(item.JurusanID, refJurusan[item.JurusanRef])
		if jurusanID == uuid.Nil {
			var daftar []models.Jurusan
			tx.Where("semester_id = ?", *semesterID).Limit(2).Find(&daftar)
			if len(daftar) == 1 {
				jurusanID = daftar[0].ID
			}
		}
		if jurusanID == uuid.Nil {
			peringatan = append(peringatan, fmt.Sprintf("Kelas %s dilewati: jurusan tidak dapat ditentukan.", nama))
			continue
		}
		tingkat := item.Tingkat
		if tingkat <= 0 {
			tingkat = 10
		}
		kode = pastikanKodeUnik(tx, "kelas", kode, "semester_id = ?", *semesterID)
		baru := models.Kelas{
			Kode: kode, Nama: nama, Tingkat: tingkat,
			JurusanID: jurusanID, SemesterID: *semesterID,
		}
		if err := tx.Create(&baru).Error; err != nil {
			return ref, dibuat, peringatan, fmt.Errorf("gagal membuat kelas %s: %w", nama, err)
		}
		ref[item.Ref] = baru.ID
		dibuat++
	}
	return ref, dibuat, peringatan, nil
}

func pastikanKodeUnik(tx *gorm.DB, tabel, kode, syarat string, args ...any) string {
	dasar := kode
	for n := 2; ; n++ {
		var jumlah int64
		q := tx.Table(tabel).Where("kode = ?", kode)
		if syarat != "" {
			q = q.Where(syarat, args...)
		}
		q.Count(&jumlah)
		if jumlah == 0 {
			return kode
		}
		pemotong := dasar
		if len(pemotong) > 6 {
			pemotong = pemotong[:6]
		}
		kode = fmt.Sprintf("%s%d", pemotong, n)
	}
}

func kodeDariNama(nama, cadangan string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(nama) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
		if b.Len() >= 8 {
			break
		}
	}
	if b.Len() == 0 {
		return cadangan
	}
	return b.String()
}

func nipOtomatis(tx *gorm.DB) string {
	for {
		nip := "AI-" + strings.ToUpper(uuid.NewString()[:8])
		var jumlah int64
		tx.Model(&models.Guru{}).Where("nip = ?", nip).Count(&jumlah)
		if jumlah == 0 {
			return nip
		}
	}
}
