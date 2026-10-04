package services

import (
	"reflect"
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
	if usulan[0].Label != "Kelas Uji pindah ke Sari" {
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
	if len(usulan) != 1 || usulan[0].Label != "Kelas Uji pindah ke Bebas" {
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
	if len(usulRuang) != 1 || usulRuang[0].Label != "Kelas Uji pindah ke Lab 2" {
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
	if usulJam[0].Label != "Kelas Uji pindah ke Senin jam ke-2" {
		t.Fatalf("label jam %q", usulJam[0].Label)
	}
}

func TestSlotUntukUsulanHanyaYangBerbagiKunci(t *testing.T) {
	hari := uuid.New()
	jam := uuid.New()
	mapel := uuid.New()
	guru := uuid.New()
	terkait := uuid.New()
	jamSama := uuid.New()
	jauh := uuid.New()
	semua := []models.SlotJadwal{
		slotUji(terkait, guru, mapel, hari, jam, uuid.Nil, "Matematika", "Senin", 3),
		slotUji(jamSama, uuid.New(), mapel, hari, jam, uuid.Nil, "Matematika", "Senin", 3),
		slotUji(jauh, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.Nil, "Lain", "Jumat", 8),
	}
	id := terkait
	hasil := slotUntukUsulan(semua, models.Konflik{TipeKonflik: "guru_bentrok", SlotAID: &id, GuruID: &guru})
	if len(hasil) != 2 {
		t.Fatalf("len = %d, ingin 2 (slot konflik dan slot jam yang sama)", len(hasil))
	}
}

func TestSlotUntukUsulanKelebihanJamMemakaiGuruKonflik(t *testing.T) {
	guru := uuid.New()
	milikGuru := uuid.New()
	jauh := uuid.New()
	semua := []models.SlotJadwal{
		slotUji(milikGuru, guru, uuid.New(), uuid.New(), uuid.New(), uuid.Nil, "IPA", "Senin", 1),
		slotUji(jauh, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.Nil, "Lain", "Jumat", 8),
	}
	hasil := slotUntukUsulan(semua, models.Konflik{TipeKonflik: "guru_kelebihan_jam", GuruID: &guru})
	if len(hasil) != 1 || hasil[0].ID != milikGuru {
		t.Fatalf("hasil = %+v, ingin hanya slot milik guru konflik", hasil)
	}
}

func TestUsulanTidakMenyarankanGuruPenuhYangSlotnyaTersaring(t *testing.T) {
	hari := uuid.New()
	jam := uuid.New()
	guruA := uuid.New()
	penuh := uuid.New()
	bebas := uuid.New()
	slotB := uuid.New()
	slotPenuh := uuid.New()

	semua := []models.SlotJadwal{
		slotUji(slotB, guruA, uuid.New(), hari, jam, uuid.Nil, "Bahasa", "Senin", 2),
		slotUji(slotPenuh, penuh, uuid.New(), uuid.New(), uuid.New(), uuid.Nil, "Lain", "Rabu", 5),
	}
	guru := []models.Guru{
		{BaseModel: models.BaseModel{ID: penuh}, NamaLengkap: "Penuh", JamMaksimalPerMinggu: 1, Aktif: true},
		{BaseModel: models.BaseModel{ID: bebas}, NamaLengkap: "Bebas", JamMaksimalPerMinggu: 10, Aktif: true},
	}
	idB := slotB
	konflik := models.Konflik{TipeKonflik: "guru_hari_libur", SlotAID: &idB}

	k := susunKonteksUsulan(semua, konflik, guru, nil, nil, map[string]bool{})
	if k.slotDariID(&slotPenuh) != nil {
		t.Fatal("slot guru Penuh seharusnya tersaring agar tes ini bermakna")
	}
	usulan := cocokkan(k, konflik)
	if len(usulan) != 1 || usulan[0].Label != "Kelas Uji pindah ke Bebas" {
		t.Fatalf("usulan = %+v, guru yang sudah mencapai jam maksimal tidak boleh disarankan", usulan)
	}
}

func TestUsulanTersaringSamaDenganTanpaSaringan(t *testing.T) {
	senin, selasa, rabu := uuid.New(), uuid.New(), uuid.New()
	j1, j2, j3 := uuid.New(), uuid.New(), uuid.New()
	mat, ipa, bing := uuid.New(), uuid.New(), uuid.New()
	k1, k2, k3, k4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	labA, labB, r1, r2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	gA, gB, gC, gD, gE := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	gPenuh, gBebas := uuid.New(), uuid.New()
	s := make([]uuid.UUID, 13)
	for i := range s {
		s[i] = uuid.New()
	}

	semua := []models.SlotJadwal{
		slotKunci(s[1], gA, mat, k1, senin, j1, labA),
		slotKunci(s[2], gA, ipa, k2, senin, j1, labB),
		slotKunci(s[3], gB, bing, k3, senin, j2, r1),
		slotKunci(s[4], gC, mat, k4, senin, j2, r1),
		slotKunci(s[5], gD, ipa, k1, senin, j3, labA),
		slotKunci(s[6], gE, bing, k2, selasa, j1, r2),
		slotKunci(s[7], gB, mat, k2, selasa, j1, r1),
		slotKunci(s[8], gPenuh, bing, k3, rabu, j2, r2),
		slotKunci(s[9], gPenuh, bing, k3, rabu, j3, r2),
		slotKunci(s[10], gC, ipa, k4, rabu, j1, labB),
		slotKunci(s[11], gE, ipa, k3, selasa, j3, labA),
		slotKunci(s[12], gA, bing, k4, rabu, j3, labB),
	}
	guru := []models.Guru{
		{BaseModel: models.BaseModel{ID: gPenuh}, NamaLengkap: "Penuh", JamMaksimalPerMinggu: 2, Aktif: true},
		{BaseModel: models.BaseModel{ID: gA}, NamaLengkap: "A", JamMaksimalPerMinggu: 10, Aktif: true},
		{BaseModel: models.BaseModel{ID: gB}, NamaLengkap: "B", JamMaksimalPerMinggu: 10, Aktif: true},
		{BaseModel: models.BaseModel{ID: gC}, NamaLengkap: "C", JamMaksimalPerMinggu: 10, Aktif: true},
		{BaseModel: models.BaseModel{ID: gD}, NamaLengkap: "D", JamMaksimalPerMinggu: 10, Aktif: true},
		{BaseModel: models.BaseModel{ID: gE}, NamaLengkap: "E", JamMaksimalPerMinggu: 1, Aktif: true},
		{BaseModel: models.BaseModel{ID: gBebas}, NamaLengkap: "Bebas", JamMaksimalPerMinggu: 10, Aktif: true},
	}
	ruang := []models.Ruangan{
		{BaseModel: models.BaseModel{ID: labA}, Nama: "Lab A", TipeRuangan: "lab", Aktif: true},
		{BaseModel: models.BaseModel{ID: labB}, Nama: "Lab B", TipeRuangan: "lab", Aktif: true},
		{BaseModel: models.BaseModel{ID: r1}, Nama: "R1", TipeRuangan: "kelas", Aktif: true},
		{BaseModel: models.BaseModel{ID: r2}, Nama: "R2", TipeRuangan: "kelas", Aktif: true},
	}
	jam := []models.JamPelajaran{
		{BaseModel: models.BaseModel{ID: j1}, JamKe: 1},
		{BaseModel: models.BaseModel{ID: j2}, JamKe: 2},
		{BaseModel: models.BaseModel{ID: j3}, JamKe: 3},
	}
	libur := map[string]bool{gD.String() + "|" + senin.String(): true}

	kasus := []struct {
		nama    string
		konflik models.Konflik
		// slotHilang wajib tersaring agar kesamaan hasil tidak kebetulan.
		slotHilang uuid.UUID
	}{
		{"guru_bentrok", models.Konflik{TipeKonflik: "guru_bentrok", SlotAID: &s[1], SlotBID: &s[2], GuruID: &gA}, s[9]},
		{"guru_hari_libur", models.Konflik{TipeKonflik: "guru_hari_libur", SlotAID: &s[5], GuruID: &gD}, s[8]},
		{"guru_kelebihan_jam", models.Konflik{TipeKonflik: "guru_kelebihan_jam", GuruID: &gE}, s[4]},
		{"ruangan_bentrok", models.Konflik{TipeKonflik: "ruangan_bentrok", SlotAID: &s[3], SlotBID: &s[4]}, s[5]},
		{"kelas_bentrok", models.Konflik{TipeKonflik: "kelas_bentrok", SlotAID: &s[6], SlotBID: &s[7]}, s[10]},
	}
	for _, tc := range kasus {
		t.Run(tc.nama, func(t *testing.T) {
			tersaring := susunKonteksUsulan(semua, tc.konflik, guru, ruang, jam, libur)
			if len(tersaring.slots) >= len(semua) {
				t.Fatalf("saringan tidak membuang slot (%d dari %d)", len(tersaring.slots), len(semua))
			}
			if tersaring.slotDariID(&tc.slotHilang) != nil {
				t.Fatalf("slot %s seharusnya tersaring", tc.slotHilang)
			}
			penuh := konteksCocok{slots: semua, guru: guru, ruang: ruang, jam: jam, libur: libur}

			dapat := cocokkan(tersaring, tc.konflik)
			ingin := cocokkan(penuh, tc.konflik)
			if len(ingin) == 0 {
				t.Fatal("usulan tanpa saringan kosong; data uji tidak bermakna")
			}
			if !reflect.DeepEqual(dapat, ingin) {
				t.Fatalf("tersaring = %+v\ntanpa saringan = %+v", dapat, ingin)
			}
			for _, u := range dapat {
				for _, p := range u.Perubahan {
					if p.NewTeacherID == gPenuh.String() {
						t.Fatalf("guru yang sudah di batas jam mingguan disarankan: %+v", u)
					}
				}
			}
		})
	}
}

func slotKunci(id, guru, mapel, kelas, hari, jam, ruang uuid.UUID) models.SlotJadwal {
	return models.SlotJadwal{
		BaseModel:       models.BaseModel{ID: id},
		GuruID:          guru,
		MataPelajaranID: mapel,
		KelasID:         kelas,
		HariID:          hari,
		JamPelajaranID:  jam,
		RuanganID:       ruang,
		MingguKe:        1,
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
		Kelas:           &models.Kelas{Nama: "Kelas Uji"},
		MataPelajaran:   &models.MataPelajaran{Nama: namaMapel},
		Hari:            &models.Hari{Nama: namaHari},
		JamPelajaran:    &models.JamPelajaran{JamKe: jamKe},
		MingguKe:        1,
	}
}
