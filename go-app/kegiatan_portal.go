package main

import "strings"

// kegiatanOwnerPortal menentukan portal pemilik kegiatan berdasarkan nama kegiatan APBD Dinas Pendidikan.
// Kegiatan yang sama bisa ter-import ke beberapa modul; rekap Command Center harus menempatkannya di portal yang benar.
func kegiatanOwnerPortal(kegiatan string) string {
	k := strings.ToLower(strings.TrimSpace(kegiatan))
	if k == "" {
		return ""
	}

	// Sekretariat — kegiatan lintas jenjang / SDM (judul bisa menyebut PAUD tanpa berarti milik portal PAUD)
	if strings.Contains(k, "pemerataan") &&
		(strings.Contains(k, "pendidik") || strings.Contains(k, "tenaga kependidikan")) {
		return "sekretariat"
	}

	// PAUD — kegiatan operasional PAUD/Nonformal (bukan sekadar kata PAUD di judul panjang)
	if strings.Contains(k, "pengelolaan pendidikan anak usia dini") ||
		strings.Contains(k, "pengelolaan pendidikan nonformal") ||
		strings.Contains(k, "penetapan kurikulum muatan lokal pendidikan anak usia dini") ||
		strings.Contains(k, "penerbitan izin paud") {
		return "paud"
	}

	// SMP
	if strings.Contains(k, "menengah perta") ||
		strings.Contains(k, "smp") ||
		strings.Contains(k, "sltp") {
		return "smp"
	}

	// SD
	if strings.Contains(k, "sekolah dasar") ||
		strings.Contains(k, "pendidikan dasar") {
		return "sd"
	}

	// Sekretariat / urusan umum & keuangan dinas
	if strings.Contains(k, "administrasi keuangan") ||
		strings.Contains(k, "administrasi umum") ||
		strings.Contains(k, "penyediaan jasa penunjang") ||
		strings.Contains(k, "pemeliharaan barang milik daerah") ||
		strings.Contains(k, "pengadaan barang milik daerah") ||
		strings.Contains(k, "perencanaan, penganggaran") ||
		strings.Contains(k, "perencanaan penganggaran") {
		return "sekretariat"
	}

	return ""
}

// moduleOwnsKegiatan true jika pagu/rekap kegiatan ini boleh dihitung untuk modul portalID.
func moduleOwnsKegiatan(portalID, kegiatan string) bool {
	owner := kegiatanOwnerPortal(kegiatan)
	if owner == "" {
		return true
	}
	return owner == portalID
}

// rekapPortalForKegiatan portal tampilan di rekap gabungan (bisa berbeda dari modul sumber transaksi).
func rekapPortalForKegiatan(sourcePortalID, kegiatan string) string {
	if owner := kegiatanOwnerPortal(kegiatan); owner != "" {
		return owner
	}
	return sourcePortalID
}
