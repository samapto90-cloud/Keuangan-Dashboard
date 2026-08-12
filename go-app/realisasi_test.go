package main

import "testing"

func newTestRealisasiModule() *SipkeuModule {
	mod := &SipkeuModule{
		ID: "smp",
	}
	mod.settings.Rak = []RakRow{
		{
			Kegiatan:     "Pengelolaan Pendidikan Sekolah Menengah Pertama",
			SubKegiatan:  "Sub A",
			KodeRekening: "5.1.02.01",
			Pekerjaan:    "Belanja Alat Tulis",
			PPTK:         "Budi",
			Anggaran:     10_000_000,
		},
		{
			Kegiatan:     "Pengelolaan Pendidikan Sekolah Menengah Pertama",
			SubKegiatan:  "Sub B",
			KodeRekening: "5.1.02.02",
			Pekerjaan:    "Belanja ATK",
			PPTK:         "Ani",
			Anggaran:     5_000_000,
		},
	}
	return mod
}

func approvedTrx(id int, tanggal, sub, kode string, nilai float64, pptk string) Transaction {
	return Transaction{
		ID:           id,
		Tanggal:      tanggal,
		Kegiatan:     "Pengelolaan Pendidikan Sekolah Menengah Pertama",
		SubKegiatan:  sub,
		KodeRekening: kode,
		Pekerjaan:    "Belanja Alat Tulis",
		PPTK:         pptk,
		Nilai:        nilai,
		Status:       trxStatusApproved,
	}
}

func TestRealisasiNoTransactions(t *testing.T) {
	mod := newTestRealisasiModule()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.TotalRealisasi != 0 {
		t.Fatalf("expected 0 realisasi, got %v", report.Summary.TotalRealisasi)
	}
	if report.Summary.TotalTransaksi != 0 {
		t.Fatalf("expected 0 trx, got %d", report.Summary.TotalTransaksi)
	}
}

func TestRealisasiSingleTransaction(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{approvedTrx(1, "2026-03-15", "Sub A", "5.1.02.01", 1_000_000, "Budi")}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.TotalRealisasi != 1_000_000 {
		t.Fatalf("expected 1M, got %v", report.Summary.TotalRealisasi)
	}
	if report.Summary.TotalTransaksi != 1 {
		t.Fatalf("expected 1 trx")
	}
}

func TestRealisasiSumThreeTransactions(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-02-10", "Sub A", "5.1.02.01", 2_000_000, "Budi"),
		approvedTrx(3, "2026-03-10", "Sub A", "5.1.02.01", 3_000_000, "Budi"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.TotalRealisasi != 6_000_000 {
		t.Fatalf("expected 6M, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiFilterSubKegiatan(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-02-10", "Sub B", "5.1.02.02", 2_000_000, "Ani"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", SubKegiatan: "Sub A"})
	if report.Summary.TotalRealisasi != 1_000_000 {
		t.Fatalf("expected 1M for Sub A, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiFilterKodeRekening(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-02-10", "Sub B", "5.1.02.02", 2_000_000, "Ani"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", KodeRekening: "5.1.02.02"})
	if report.Summary.TotalRealisasi != 2_000_000 {
		t.Fatalf("expected 2M, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiFilterPPTK(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-02-10", "Sub B", "5.1.02.02", 2_000_000, "Ani"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", PPTK: "Ani"})
	if report.Summary.TotalRealisasi != 2_000_000 {
		t.Fatalf("expected 2M for Ani, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiFilterBulan(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-03-10", "Sub A", "5.1.02.01", 2_000_000, "Budi"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", Bulan: "03"})
	if report.Summary.TotalRealisasi != 2_000_000 {
		t.Fatalf("expected 2M for March, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiRejectedNotCounted(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		{ID: 2, Tanggal: "2026-02-10", SubKegiatan: "Sub A", KodeRekening: "5.1.02.01", Nilai: 9_000_000, Status: trxStatusRejected},
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.TotalRealisasi != 1_000_000 {
		t.Fatalf("rejected must not count, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiDashboardMatchesTransactionSum(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_500_000, "Budi"),
		approvedTrx(2, "2026-02-10", "Sub B", "5.1.02.02", 2_500_000, "Ani"),
		{ID: 3, Tanggal: "2026-03-10", SubKegiatan: "Sub A", KodeRekening: "5.1.02.01", Nilai: 500_000, Status: trxStatusPending},
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	var sum float64
	for _, tx := range buildRealisasiTransactions(mod, realisasiFilters{Tahun: "2026"}, "") {
		sum += tx.Nilai
	}
	if sum != report.Summary.TotalRealisasi {
		t.Fatalf("sum txs %v != dashboard %v", sum, report.Summary.TotalRealisasi)
	}
}

func TestRealisasiStatusOverBudget(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 12_000_000, "Budi")}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.OverBudgetCount < 1 {
		t.Fatal("expected over budget row")
	}
}

func TestComputeRealisasiStatus(t *testing.T) {
	if computeRealisasiStatus(0, 0, 100) != "belum" {
		t.Fatal("0% should be belum")
	}
	if computeRealisasiStatus(100, 100, 100) != "selesai" {
		t.Fatal("100% should be selesai")
	}
	if computeRealisasiStatus(50, 50, 100) != "rendah" {
		t.Fatal("50% should be rendah")
	}
}

func TestRealisasiFilterDateRange(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		approvedTrx(2, "2026-02-15", "Sub A", "5.1.02.01", 2_000_000, "Budi"),
		approvedTrx(3, "2026-03-20", "Sub A", "5.1.02.01", 3_000_000, "Budi"),
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", Dari: "2026-01-01", Sampai: "2026-03-31"})
	if report.Summary.TotalRealisasi != 6_000_000 {
		t.Fatalf("Q1 expected 6M, got %v", report.Summary.TotalRealisasi)
	}
	report = buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", Dari: "2026-01-01", Sampai: "2026-01-31"})
	if report.Summary.TotalRealisasi != 1_000_000 {
		t.Fatalf("January range expected 1M, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiPendingNotCounted(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{
		approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 1_000_000, "Budi"),
		{ID: 2, Tanggal: "2026-02-10", Kegiatan: "Pengelolaan Pendidikan Sekolah Menengah Pertama", SubKegiatan: "Sub A", KodeRekening: "5.1.02.01", Pekerjaan: "Belanja Alat Tulis", Nilai: 4_000_000, Status: trxStatusPending},
	}
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026"})
	if report.Summary.TotalRealisasi != 1_000_000 {
		t.Fatalf("pending must not count, got %v", report.Summary.TotalRealisasi)
	}
}

func TestRealisasiSisaFollowsActiveBudget(t *testing.T) {
	mod := newTestRealisasiModule()
	mod.mu.Lock()
	mod.txs = []Transaction{approvedTrx(1, "2026-01-10", "Sub A", "5.1.02.01", 4_000_000, "Budi")}
	mod.settings.Rak[0].Anggaran = 8_000_000
	mod.mu.Unlock()
	report := buildRealisasiForModule(mod, realisasiFilters{Tahun: "2026", SubKegiatan: "Sub A", KodeRekening: "5.1.02.01"})
	if report.Summary.TotalAnggaran != 8_000_000 {
		t.Fatalf("active budget expected 8M, got %v", report.Summary.TotalAnggaran)
	}
	if report.Summary.SisaAnggaran != 4_000_000 {
		t.Fatalf("sisa expected 4M, got %v", report.Summary.SisaAnggaran)
	}
}
