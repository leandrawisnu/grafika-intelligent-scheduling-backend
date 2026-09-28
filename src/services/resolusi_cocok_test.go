package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
)

func TestCocokkanGuruBentrokUtamakanPengajarMapelYangSama(t *testing.T) {
	hari := uuid.New()
	jam := uuid.New()
	mapel := uuid.New()
	guruA := uuid.New()
	guruSama := uuid.New()
	guruLain := uuid.New()
	slotA := uuid.New()
	slotB := uuid.New()
	slotLain := uuid.New()

	k := konteksCocok{
		guru: []models.Guru{
			{BaseModel: models.BaseModel{ID: guruLain}, NamaLengkap: "Budi", JamMaksimalPerMinggu: 10, Aktif: true},
			{BaseModel: models.BaseModel{ID: guruSama}, NamaLengkap: "Sari", JamMaksimalPerMinggu: 10, Aktif: true},
		},
		slots: []models.SlotJadwal{
			slotUji(slotA, guruA, mapel, hari, jam, uuid.Nil, "Matematika", "Senin", 3),
			slotUji(slotB, guruA, mapel, hari, jam, uuid.Nil, "Matematika", "Senin", 3),
			slotUji(slotLain, guruSama, mapel, uuid.New(), uuid.New(), uuid.Nil, "Matematika", "Selasa", 1),
		},
		libur: map[string]bool{},
	}
	idB := slotB
	usulan := cocokkan(k, models.Konflik{TipeKonflik: "guru_bentrok", SlotBID: &idB})
	if len(usulan) == 0 {
		t.Fatal("usulan kosong")
	}
	if usulan[0].Label != "Pindah ke Sari" {
		t.Fatalf("usulan pertama %q, ingin guru yang sudah mengajar mapel yang sama", usulan[0].Label)
	}
	if usulan[0].Perubahan[0].NewTeacherID != guruSama.String() {
		t.Fatalf("guru usulan %s", usulan[0].Perubahan[0].NewTeacherID)
	}
}

func TestCocokkanMelewatiGuruSibukLiburAtauPenuh(t *testing.T) {
	hari := uuid.New()
	jam := uuid.New()
	mapel := uuid.New()
	guruA := uuid.New()
	sibuk := uuid.New()
	libur := uuid.New()
	penuh := uuid.New()
	bebas := uuid.New()
	slotB := uuid.New()

	k := konteksCocok{
		guru: []models.Guru{
			{BaseModel: models.BaseModel{ID: sibuk}, NamaLengkap: "Sibuk", JamMaksimalPerMinggu: 10, Aktif: true},
			{BaseModel: models.BaseModel{ID: libur}, NamaLengkap: "Libur", JamMaksimalPerMinggu: 10, Aktif: true},
			{BaseModel: models.BaseModel{ID: penuh}, NamaLengkap: "Penuh", JamMaksimalPerMinggu: 1, Aktif: true},
			{BaseModel: models.BaseModel{ID: bebas}, NamaLengkap: "Bebas", JamMaksimalPerMinggu: 10, Aktif: true},
		},
		slots: []models.SlotJadwal{
			slotUji(slotB, guruA, mapel, hari, jam, uuid.Nil, "Bahasa", "Senin", 2),
			slotUji(uuid.New(), sibuk, mapel, hari, jam, uuid.Nil, "Bahasa", "Senin", 2),
			slotUji(uuid.New(), penuh, uuid.New(), uuid.New(), uuid.New(), uuid.Nil, "Lain", "Rabu", 1),
		},
		libur: map[string]bool{libur.String() + "|" + hari.String(): true},
	}
	idB := slotB
	usulan := cocokkan(k, models.Konflik{TipeKonflik: "guru_hari_libur", SlotAID: &idB})
	if len(usulan) != 1 || usulan[0].Label != "Pindah ke Bebas" {
		t.Fatalf("usulan = %+v", usulan)
	}
}

func TestCocokkanRuanganDanJam(t *testing.T) {
	hari := uuid.New()
	jam1 := uuid.New()
	jam2 := uuid.New()
	ruangA := uuid.New()
	ruangSama := uuid.New()
	ruangBeda := uuid.New()
	slotA := uuid.New()
	slotB := uuid.New()

	ruang := []models.Ruangan{
		{BaseModel: models.BaseModel{ID: ruangSama}, Nama: "Lab 2", TipeRuangan: "lab", Aktif: true},
		{BaseModel: models.BaseModel{ID: ruangBeda}, Nama: "Kelas 1", TipeRuangan: "kelas", Aktif: true},
	}
	jam := []models.JamPelajaran{
		{BaseModel: models.BaseModel{ID: jam1}, JamKe: 1, Istirahat: false},
		{BaseModel: models.BaseModel{ID: jam2}, JamKe: 2, Istirahat: false},
	}

	slotBentrok := slotUji(slotB, uuid.New(), uuid.New(), hari, jam1, ruangA, "IPA", "Senin", 1)
	slotBentrok.Ruangan = &models.Ruangan{TipeRuangan: "lab", Nama: "Lab 1"}
	kRuang := konteksCocok{
		slots: []models.SlotJadwal{slotBentrok},
		ruang: ruang,
		libur: map[string]bool{},
	}
	idB := slotB
	usulRuang := cocokkan(kRuang, models.Konflik{TipeKonflik: "ruangan_bentrok", SlotBID: &idB})
	if len(usulRuang) != 1 || usulRuang[0].Label != "Pindah ke Lab 2" {
		t.Fatalf("ruangan = %+v", usulRuang)
	}

	slotKelas := slotUji(slotA, uuid.New(), uuid.New(), hari, jam1, ruangA, "IPA", "Senin", 1)
	kJam := konteksCocok{
		slots: []models.SlotJadwal{slotKelas},
		jam:   jam,
		libur: map[string]bool{},
	}
	idA := slotA
	usulJam := cocokkan(kJam, models.Konflik{TipeKonflik: "kelas_bentrok", SlotAID: &idA})
	if len(usulJam) != 1 || usulJam[0].Perubahan[0].NewTimeSlotID != jam2.String() {
		t.Fatalf("jam = %+v", usulJam)
	}
	if usulJam[0].Label != "Pindah ke Senin jam ke-2" {
		t.Fatalf("label jam %q", usulJam[0].Label)
	}
}

func slotUji(id, guru, mapel, hari, jam, ruang uuid.UUID, namaMapel, namaHari string, jamKe int16) models.SlotJadwal {
	return models.SlotJadwal{
		BaseModel:       models.BaseModel{ID: id},
		GuruID:          guru,
		MataPelajaranID: mapel,
		HariID:          hari,
		JamPelajaranID:  jam,
		RuanganID:       ruang,
		KelasID:         uuid.New(),
		MataPelajaran:   &models.MataPelajaran{Nama: namaMapel},
		Hari:            &models.Hari{Nama: namaHari},
		JamPelajaran:    &models.JamPelajaran{JamKe: jamKe},
		MingguKe:        1,
	}
}
