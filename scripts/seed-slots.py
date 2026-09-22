#!/usr/bin/env python3
"""
Generate slot_jadwal SQL dari template mingguan per jurusan + tingkat.
Sumber pola: singkatan mapel PDF Ganjil 2026/2027 SMKN 4 (bukan parse sel per sel).

Usage:
  python3 scripts/seed-slots.py > database/seeds/generated_slots.sql
  ./scripts/seed-ganjil-slots.sh
"""

from __future__ import annotations

import sys
from textwrap import dedent

# (hari, jam_ke, mapel_kode, guru_nip, ruangan_kode | None)
Slot = tuple[str, int, str, str, str | None]

UMUM_X: list[Slot] = [
    ("Senin", 1, "BIND", "PDF-SAN", "R-DKV-1"),
    ("Senin", 2, "MAT", "PDF-NOV", "R-DKV-2"),
    ("Senin", 3, "BING", "PDF-WUR", "R-DKV-3"),
    ("Senin", 4, "INKA", "PDF-ERI", "R-DKV-4"),
    ("Senin", 6, "PP", "PDF-SAR", "R-DKV-5"),
    ("Senin", 7, "SEJ", "PDF-RAF", "R-DKV-6"),
    ("Selasa", 1, "BIND", "PDF-SAN", "R-DKV-1"),
    ("Selasa", 2, "PJOK", "PDF-YGN", "R-DKV-6"),
    ("Selasa", 3, "BING", "PDF-WUR", "R-DKV-2"),
    ("Selasa", 4, "IPAS", "PDF-IMA", "R-DKV-3"),
    ("Selasa", 6, "PABP", "PDF-ASA", "R-DKV-4"),
    ("Selasa", 7, "KKA", "PDF-DRH", "R-DKV-5"),
    ("Rabu", 1, "MAT", "PDF-NOV", "R-DKV-1"),
    ("Rabu", 2, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Rabu", 3, "SENI", "PDF-PUG", "R-DKV-3"),
    ("Rabu", 4, "MULOK", "PDF-TNY", "R-DKV-4"),
    ("Rabu", 6, "BING", "PDF-WUR", "R-DKV-5"),
    ("Rabu", 7, "PJBL", "PDF-ERI", "R-DKV-6"),
    ("Kamis", 1, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Kamis", 2, "MAT", "PDF-NOV", "R-DKV-1"),
    ("Kamis", 3, "SEJ", "PDF-RAF", "R-DKV-3"),
    ("Kamis", 4, "PP", "PDF-SAR", "R-DKV-4"),
    ("Kamis", 6, "INKA", "PDF-ERI", "R-DKV-5"),
    ("Kamis", 7, "TEG", "PDF-HGW", "R-DKV-6"),
    ("Jumat", 1, "BING", "PDF-WUR", "R-DKV-1"),
    ("Jumat", 2, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Jumat", 3, "PABP", "PDF-ASA", "R-DKV-3"),
    ("Jumat", 4, "PJOK", "PDF-YGN", "R-DKV-6"),
    ("Jumat", 6, "KKA", "PDF-DRH", "R-DKV-4"),
    ("Jumat", 7, "BK", "PDF-DYH", "R-DKV-5"),
]

UMUM_XI: list[Slot] = [
    ("Senin", 1, "BIND", "PDF-SAN", "R-DKV-1"),
    ("Senin", 2, "MAT", "PDF-NOV", "R-DKV-2"),
    ("Senin", 3, "BING", "PDF-WUR", "R-DKV-3"),
    ("Senin", 4, "SEJ", "PDF-RAF", "R-DKV-4"),
    ("Senin", 6, "PP", "PDF-SAR", "R-DKV-5"),
    ("Senin", 7, "PJBL", "PDF-ERI", "R-DKV-6"),
    ("Selasa", 1, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Selasa", 2, "PJOK", "PDF-YGN", "R-DKV-6"),
    ("Selasa", 3, "BING", "PDF-WUR", "R-DKV-1"),
    ("Selasa", 4, "INKA", "PDF-ERI", "R-DKV-3"),
    ("Selasa", 6, "PABP", "PDF-ASA", "R-DKV-4"),
    ("Selasa", 7, "KKA", "PDF-DRH", "R-DKV-5"),
    ("Rabu", 1, "MAT", "PDF-NOV", "R-DKV-1"),
    ("Rabu", 2, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Rabu", 3, "MULOK", "PDF-TNY", "R-DKV-3"),
    ("Rabu", 4, "SENI", "PDF-PUG", "R-DKV-4"),
    ("Rabu", 6, "BING", "PDF-WUR", "R-DKV-5"),
    ("Rabu", 7, "TEG", "PDF-HGW", "R-DKV-6"),
    ("Kamis", 1, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Kamis", 2, "MAT", "PDF-NOV", "R-DKV-1"),
    ("Kamis", 3, "PP", "PDF-SAR", "R-DKV-3"),
    ("Kamis", 4, "SEJ", "PDF-RAF", "R-DKV-4"),
    ("Kamis", 6, "INKA", "PDF-ERI", "R-DKV-5"),
    ("Kamis", 7, "PJBL", "PDF-ERI", "R-DKV-6"),
    ("Jumat", 1, "BING", "PDF-WUR", "R-DKV-1"),
    ("Jumat", 2, "BIND", "PDF-SAN", "R-DKV-2"),
    ("Jumat", 3, "PABP", "PDF-ASA", "R-DKV-3"),
    ("Jumat", 4, "PJOK", "PDF-YGN", "R-DKV-6"),
    ("Jumat", 6, "KKA", "PDF-DRH", "R-DKV-4"),
    ("Jumat", 7, "BK", "PDF-DYH", "R-DKV-5"),
]

PRODUK: dict[str, tuple[str, str, str]] = {
    # jurusan: (mapel_x, mapel_xi, ruangan_prefix)
    "DKV": ("DDK-DKV", "DKV", "R-DKV"),
    "ANI": ("DDK-ANI", "ANI-PROD", "LAB-ANIMASI"),
    "TG": ("DDK-TG", "TG-PROD", "LAB-INKA"),
    "TKJ": ("DDK-TKJ", "TKJ-PROD", "LAB-TJKT-1"),
    "RPL": ("DDK-RPL", "RPL-PROD", "LAB-RPL-1"),
    "PH": ("DDK-PH", "PH-PROD", "LAB-PH-1"),
    "TL": ("DDK-TL", "TL-PROD", "R-16"),
    "TM": ("DDK-TM", "TM-PROD", "R-33"),
}

GURU_PRODUK: dict[str, str] = {
    "DKV": "PDF-AJI",
    "ANI": "PDF-ZEE",
    "TG": "PDF-IBAM",
    "TKJ": "PDF-MIT",
    "RPL": "PDF-DEV",
    "PH": "PDF-HAN",
    "TL": "PDF-SHA",
    "TM": "PDF-JGA",
}

KELAS: list[tuple[str, str, int]] = [
    ("X-DKV-A", "DKV", 10), ("X-DKV-B", "DKV", 10), ("X-DKV-C", "DKV", 10),
    ("X-ANI-A", "ANI", 10), ("X-ANI-B", "ANI", 10), ("X-ANI-C", "ANI", 10),
    ("X-TG-A", "TG", 10), ("X-TG-B", "TG", 10), ("X-TG-C", "TG", 10), ("X-TG-D", "TG", 10),
    ("X-TG-E", "TG", 10), ("X-TG-F", "TG", 10), ("X-TG-G", "TG", 10), ("X-TG-H", "TG", 10),
    ("X-TKJ-A", "TKJ", 10), ("X-TKJ-B", "TKJ", 10),
    ("X-RPL-A", "RPL", 10), ("X-RPL-B", "RPL", 10), ("X-RPL-C", "RPL", 10),
    ("X-PH-A", "PH", 10), ("X-PH-B", "PH", 10),
    ("X-TL-A", "TL", 10), ("X-TL-B", "TL", 10),
    ("X-TM-A", "TM", 10), ("X-TM-B", "TM", 10),
    ("XI-DKV-A", "DKV", 11), ("XI-DKV-B", "DKV", 11), ("XI-DKV-C", "DKV", 11),
    ("XI-ANI-A", "ANI", 11), ("XI-ANI-B", "ANI", 11), ("XI-ANI-C", "ANI", 11),
    ("XI-TG-A", "TG", 11), ("XI-TG-B", "TG", 11), ("XI-TG-C", "TG", 11), ("XI-TG-D", "TG", 11),
    ("XI-TG-E", "TG", 11), ("XI-TG-F", "TG", 11), ("XI-TG-G", "TG", 11), ("XI-TG-H", "TG", 11),
    ("XI-TKJ-A", "TKJ", 11), ("XI-TKJ-B", "TKJ", 11),
    ("XI-RPL-A", "RPL", 11), ("XI-RPL-B", "RPL", 11), ("XI-RPL-C", "RPL", 11),
    ("XI-PH-A", "PH", 11), ("XI-PH-B", "PH", 11),
    ("XI-TL-A", "TL", 11), ("XI-TL-B", "TL", 11),
    ("XI-TM-A", "TM", 11), ("XI-TM-B", "TM", 11),
]


def produktif_slots(jurusan: str, tingkat: int) -> list[Slot]:
    mx, mxi, ruang = PRODUK[jurusan]
    mapel = mx if tingkat == 10 else mxi
    guru = GURU_PRODUK[jurusan]
    return [
        ("Senin", 8, mapel, guru, ruang),
        ("Selasa", 8, mapel, guru, ruang),
        ("Rabu", 8, mapel, guru, ruang),
        ("Kamis", 8, mapel, guru, ruang),
        ("Jumat", 8, mapel, guru, ruang),
        ("Senin", 9, mapel, guru, ruang),
        ("Selasa", 9, mapel, guru, ruang),
        ("Rabu", 9, mapel, guru, ruang),
    ]


def slots_for_kelas(jurusan: str, tingkat: int) -> list[Slot]:
    base = UMUM_X if tingkat == 10 else UMUM_XI
    return base + produktif_slots(jurusan, tingkat)


def sql_escape(s: str) -> str:
    return s.replace("'", "''")


def slot_insert(kelas_kode: str, hari: str, jam_ke: int, mapel: str, guru: str, ruang: str | None) -> str:
    ruang_sql = "NULL"
    if ruang:
        ruang_sql = f"(SELECT id FROM ruangan WHERE kode = '{sql_escape(ruang)}' LIMIT 1)"
    return dedent(
        f"""
        INSERT INTO slot_jadwal (
          jadwal_kelas_id, kelas_id, mata_pelajaran_id, hari_id, jam_pelajaran_id,
          ruangan_id, guru_id, minggu_ke, terkunci
        )
        SELECT jk.id, k.id, mp.id, h.id, j.id,
          {ruang_sql},
          g.id, 1, false
        FROM kelas k
        JOIN jadwal_kelas jk ON jk.kelas_id = k.id AND jk.versi = 1 AND jk.is_active = true
        JOIN jadwal_semester js ON js.id = jk.jadwal_semester_id
        JOIN semester sem ON sem.id = js.semester_id
        JOIN tahun_ajaran ta ON ta.id = sem.tahun_ajaran_id
        JOIN mata_pelajaran mp ON mp.kode = '{sql_escape(mapel)}'
        JOIN hari h ON h.nama = '{sql_escape(hari)}'
        JOIN jam_pelajaran j ON j.jam_ke = {jam_ke}
        JOIN guru g ON g.nip = '{sql_escape(guru)}'
        WHERE k.kode = '{sql_escape(kelas_kode)}'
          AND ta.nama = '2026/2027' AND sem.semester_ke = 1
        ON CONFLICT (jadwal_kelas_id, kelas_id, hari_id, jam_pelajaran_id, minggu_ke) DO NOTHING;
        """
    ).strip()


def main() -> int:
    lines = [
        "-- Generated slot_jadwal Ganjil 2026/2027",
        "-- python3 scripts/seed-slots.py",
        "BEGIN;",
        dedent(
            """
            DELETE FROM slot_jadwal sj
            USING jadwal_kelas jk, jadwal_semester js, semester sem, tahun_ajaran ta
            WHERE sj.jadwal_kelas_id = jk.id
              AND jk.jadwal_semester_id = js.id
              AND js.semester_id = sem.id
              AND sem.tahun_ajaran_id = ta.id
              AND ta.nama = '2026/2027' AND sem.semester_ke = 1;
            """
        ).strip(),
    ]

    count = 0
    for kelas_kode, jurusan, tingkat in KELAS:
        for hari, jam_ke, mapel, guru, ruang in slots_for_kelas(jurusan, tingkat):
            lines.append(slot_insert(kelas_kode, hari, jam_ke, mapel, guru, ruang))
            count += 1

    lines.append("COMMIT;")
    lines.append(f"-- total insert attempts: {count}")
    sys.stdout.write("\n".join(lines) + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
