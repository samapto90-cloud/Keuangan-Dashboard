package main

import "testing"

func TestKegiatanOwnerPortal(t *testing.T) {
	cases := []struct {
		kegiatan string
		want     string
	}{
		{"Administrasi Keuangan Perangkat Daerah", "sekretariat"},
		{"Administrasi Umum Perangkat Daerah", "sekretariat"},
		{"Pengelolaan Pendidikan Anak Usia Dini (PAUD)", "paud"},
		{"Pengelolaan Pendidikan Sekolah Dasar", "sd"},
		{"Pengelolaan Pendidikan Sekolah Menengah Pertama", "smp"},
	}
	for _, c := range cases {
		got := kegiatanOwnerPortal(c.kegiatan)
		if got != c.want {
			t.Fatalf("kegiatanOwnerPortal(%q) = %q, want %q", c.kegiatan, got, c.want)
		}
	}
}

func TestModuleOwnsKegiatan(t *testing.T) {
	if moduleOwnsKegiatan("paud", "Administrasi Keuangan Perangkat Daerah") {
		t.Fatal("PAUD must not own Administrasi Keuangan")
	}
	if !moduleOwnsKegiatan("sekretariat", "Administrasi Keuangan Perangkat Daerah") {
		t.Fatal("Sekretariat must own Administrasi Keuangan")
	}
}
