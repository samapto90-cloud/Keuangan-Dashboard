package main

import (
        "embed"
        "encoding/json"
        "fmt"
        "io/fs"
        "log"
        "net/http"
        "os"
        "strconv"
        "strings"
)

//go:embed index.html
var indexHTML []byte

//go:embed kop-disdik.png
var kopDisdikPNG []byte

//go:embed assets/portal-tanjiro.png
var portalTanjiroPNG []byte

//go:embed assets/portal-nezuko.png
var portalNezukoPNG []byte

//go:embed assets/portal-zenitsu.png
var portalZenitsuPNG []byte

//go:embed assets/logo-batam.png
var logoBatamPNG []byte

//go:embed assets/op-runners/*
var opRunnersFS embed.FS

//go:embed assets/naruto-runners/*
var narutoRunnersFS embed.FS

//go:embed assets/doraemon-runners/*
var doraemonRunnersFS embed.FS

//go:embed assets/naruto-smp-runners/*
var narutoSmpRunnersFS embed.FS

//go:embed assets/frozen-runners/*
var frozenRunnersFS embed.FS

//go:embed assets/gundam-icons/*
var gundamIconsFS embed.FS

//go:embed assets/ds-kas-runners/*
var dsKasRunnersFS embed.FS

type PotonganItem struct {
	Jenis     string  `json:"jenis"`
	Tarif     float64 `json:"tarif"`
	Nilai     float64 `json:"nilai"`
	Kategori  string  `json:"kategori"`
	KodeMAP   string  `json:"kode_map,omitempty"`
	MasaPajak string  `json:"masa_pajak,omitempty"`
	IDBilling string  `json:"id_billing,omitempty"`
}

type Transaction struct {
        ID               int     `json:"id"`
        Tanggal          string  `json:"tanggal"`
        Kegiatan         string  `json:"kegiatan"`
        SubKegiatan      string  `json:"sub_kegiatan"`
        KodeRekening     string  `json:"kode_rekening"`
        Penerima         string  `json:"penerima"`
        NoBPK            string  `json:"no_bpk"`
        NoBAST           string  `json:"no_bast"`
        NoKontrak        string  `json:"no_kontrak"`
        Pekerjaan        string  `json:"pekerjaan"`
        Uraian           string  `json:"uraian"`
        JenisPajak       string  `json:"jenis_pajak"`
        JenisPotongan    string  `json:"jenis_potongan"`
        PotonganPajak    []PotonganItem `json:"potongan_pajak,omitempty"`
        Nilai            float64 `json:"nilai"`
        Pajak            float64 `json:"pajak"`
        NilaiPotongan    float64 `json:"nilai_potongan"`
	NTPN             string  `json:"ntpn"`
	KodeBilling      string  `json:"kode_billing"`
	NTB              string  `json:"ntb"`
	PenggunaAnggaran string  `json:"pengguna_anggaran"`
	PPTK             string  `json:"pptk"`
	PPTKnip          string  `json:"pptk_nip"`
	Bendahara        string  `json:"bendahara"`
	NamaRekening     string  `json:"nama_rekening"`
	NoRekening       string  `json:"no_rekening"`
	Bank             string  `json:"bank"`
	NPWP             string  `json:"npwp"`
	NamaWP           string  `json:"nama_wp"`
	BPP              string  `json:"bpp"`
	NoNP2D           string  `json:"no_np2d"`
}

type DashboardStats struct {
        TotalTransaksi   int            `json:"total_transaksi"`
        TotalNilai       float64        `json:"total_nilai"`
        TotalPajak       float64        `json:"total_pajak"`
        NilaiBersih      float64        `json:"nilai_bersih"`
        TotalPagu        float64        `json:"total_pagu"`
        Realisasi        float64        `json:"realisasi"`
        SisaPagu         float64        `json:"sisa_pagu"`
        PersenRealisasi  float64        `json:"persen_realisasi"`
        NilaiPerKegiatan []KegiatanStat `json:"nilai_per_kegiatan"`
        NilaiPerPPTK     []PPTKStat     `json:"nilai_per_pptk"`
        RecentTransaksi  []Transaction  `json:"recent_transaksi"`
        MonthlyStats     []MonthlyStat  `json:"monthly_stats"`
}

type KegiatanStat struct {
        Kegiatan string  `json:"kegiatan"`
        Total    float64 `json:"total"`
        Count    int     `json:"count"`
}

type PPTKStat struct {
        PPTK  string  `json:"pptk"`
        Total float64 `json:"total"`
        Pagu  float64 `json:"pagu"`
        Count int     `json:"count"`
}

type MonthlyStat struct {
        Bulan string  `json:"bulan"`
        Nilai float64 `json:"nilai"`
        Pajak float64 `json:"pajak"`
}

type Pejabat struct {
        Nama string `json:"nama"`
        Nip  string `json:"nip"`
}

type AppSettings struct {
        PA               Pejabat            `json:"pa"`
        Bendahara        Pejabat            `json:"bendahara"`
        AnggaranKegiatan map[string]float64 `json:"anggaran_kegiatan"`
        Rak              []RakRow           `json:"rak"`
        RakMeta          RakMeta            `json:"rak_meta"`
}

func normalizeTransactionTax(t *Transaction) {
        if len(t.PotonganPajak) == 0 {
                if t.JenisPajak != "" || t.Pajak > 0 {
                        t.PotonganPajak = append(t.PotonganPajak, PotonganItem{
                                Jenis: t.JenisPajak, Nilai: t.Pajak, Kategori: "pajak",
                        })
                }
                if t.JenisPotongan != "" || t.NilaiPotongan > 0 {
                        t.PotonganPajak = append(t.PotonganPajak, PotonganItem{
                                Jenis: t.JenisPotongan, Nilai: t.NilaiPotongan, Kategori: "potongan",
                        })
                }
                return
        }
        var totalPajak, totalPotongan float64
        var jenisPajakParts, jenisPotonganParts []string
        for _, item := range t.PotonganPajak {
                if item.Kategori == "potongan" {
                        totalPotongan += item.Nilai
                        if item.Jenis != "" {
                                jenisPotonganParts = append(jenisPotonganParts, item.Jenis)
                        }
                } else {
                        totalPajak += item.Nilai
                        if item.Jenis != "" {
                                jenisPajakParts = append(jenisPajakParts, item.Jenis)
                        }
                }
        }
        t.Pajak = totalPajak
        t.NilaiPotongan = totalPotongan
        t.JenisPajak = strings.Join(jenisPajakParts, "; ")
        t.JenisPotongan = strings.Join(jenisPotonganParts, "; ")
}

var allowedOrigin string

func cors(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                origin := allowedOrigin
                if origin == "" {
                        origin = "*"
                }
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token, X-SIPKEU-App")
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusOK)
                        return
                }
                next(w, r)
        }
}

func withSecurityHeaders(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("X-Content-Type-Options", "nosniff")
                w.Header().Set("X-Frame-Options", "SAMEORIGIN")
                w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
                next.ServeHTTP(w, r)
        })
}

func servePNG(data []byte) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "image/png")
                w.Header().Set("Cache-Control", "public, max-age=86400")
                w.Write(data)
        }
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expires", "0")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(data)
}

func handleTransactions(w http.ResponseWriter, r *http.Request) {
        if getSession(r) == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        mod := moduleFromRequest(r)
        switch r.Method {
        case http.MethodGet:
                mod.mu.Lock()
                result := make([]Transaction, len(mod.txs))
                copy(result, mod.txs)
                mod.mu.Unlock()
                jsonResponse(w, http.StatusOK, result)

        case http.MethodPost:
                var t Transaction
                if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
                        jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                        return
                }
                normalizeTransactionTax(&t)
                mod.mu.Lock()
                t.ID = mod.nextID
                mod.nextID++
                mod.txs = append(mod.txs, t)
                mod.mu.Unlock()
                persistModule(mod)
                jsonResponse(w, http.StatusCreated, t)

        default:
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
        }
}

func handleTransactionByID(w http.ResponseWriter, r *http.Request) {
        if getSession(r) == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        path := strings.TrimPrefix(r.URL.Path, "/data/transactions/")
        id, err := strconv.Atoi(path)
        if err != nil {
                jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
                return
        }

        mod := moduleFromRequest(r)
        switch r.Method {
        case http.MethodPut:
                var updated Transaction
                if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
                        jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                        return
                }
                normalizeTransactionTax(&updated)
                mod.mu.Lock()
                found := false
                for i, t := range mod.txs {
                        if t.ID == id {
                                updated.ID = id
                                mod.txs[i] = updated
                                found = true
                                break
                        }
                }
                mod.mu.Unlock()
                if !found {
                        jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Not found"})
                        return
                }
                persistModule(mod)
                jsonResponse(w, http.StatusOK, updated)

        case http.MethodDelete:
                deleteTransactionByID(w, mod, id)

        default:
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
        }
}

func deleteTransactionByID(w http.ResponseWriter, mod *SipkeuModule, id int) {
        mod.mu.Lock()
        found := false
        for i, t := range mod.txs {
                if t.ID == id {
                        mod.txs = append(mod.txs[:i], mod.txs[i+1:]...)
                        found = true
                        break
                }
        }
        mod.mu.Unlock()
        if !found {
                jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Transaksi tidak ditemukan"})
                return
        }
        persistModule(mod)
        jsonResponse(w, http.StatusOK, map[string]string{"message": "Deleted"})
}

func handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
                return
        }
        if getSession(r) == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        var payload struct {
                ID int `json:"id"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
                jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                return
        }
        mod := moduleFromRequest(r)
        deleteTransactionByID(w, mod, payload.ID)
}

func handleDeleteBulkTransactions(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
                return
        }
        if getSession(r) == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        var payload struct {
                IDs []int `json:"ids"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
                jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                return
        }
        if len(payload.IDs) == 0 {
                jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Tidak ada transaksi yang dipilih"})
                return
        }
        delSet := make(map[int]bool, len(payload.IDs))
        for _, id := range payload.IDs {
                delSet[id] = true
        }
        mod := moduleFromRequest(r)
        mod.mu.Lock()
        kept := mod.txs[:0]
        deleted := 0
        for _, t := range mod.txs {
                if delSet[t.ID] {
                        deleted++
                        continue
                }
                kept = append(kept, t)
        }
        mod.txs = kept
        remaining := len(mod.txs)
        mod.mu.Unlock()
        persistModule(mod)
        jsonResponse(w, http.StatusOK, map[string]interface{}{
                "deleted": deleted,
                "total":   remaining,
                "message": fmt.Sprintf("%d transaksi terpilih dihapus. Sisa %d transaksi.", deleted, remaining),
        })
}

func handleDeleteAllTransactions(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
                return
        }
        mod := moduleFromRequest(r)
        mod.mu.Lock()
        count := len(mod.txs)
        mod.txs = []Transaction{}
        mod.mu.Unlock()
        persistModule(mod)
        jsonResponse(w, http.StatusOK, map[string]interface{}{
                "deleted": count,
                "message": fmt.Sprintf("%d transaksi berhasil dihapus", count),
        })
}

func handleImport(w http.ResponseWriter, r *http.Request) {
        if getSession(r) == nil || getSession(r).Role != "admin" {
                jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Admin"})
                return
        }
        if r.Method != http.MethodPost {
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
                return
        }
        var items []Transaction
        if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
                jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                return
        }
        mod := moduleFromRequest(r)
        mod.mu.Lock()
        for i := range items {
                normalizeTransactionTax(&items[i])
                items[i].ID = mod.nextID
                mod.nextID++
                mod.txs = append(mod.txs, items[i])
        }
        total := len(mod.txs)
        mod.mu.Unlock()
        persistModule(mod)
        jsonResponse(w, http.StatusOK, map[string]interface{}{
                "imported": len(items),
                "total":    total,
                "message":  fmt.Sprintf("%d transaksi ditambahkan. Total kini %d transaksi (data lama tetap tersimpan).", len(items), total),
        })
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
        if getSession(r) == nil || getSession(r).Role != "admin" {
                jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Admin"})
                return
        }
        if r.Method != http.MethodGet {
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
                return
        }

        mod := moduleFromRequest(r)
        mod.mu.Lock()
        data := make([]Transaction, len(mod.txs))
        copy(data, mod.txs)
        mod.mu.Unlock()

        stats := DashboardStats{}
        stats.TotalTransaksi = len(data)

        kegiatanMap := map[string]*KegiatanStat{}
        pptkMap := map[string]*PPTKStat{}
        monthlyMap := map[string]*MonthlyStat{}

        var totalPotongan float64
        for _, t := range data {
                stats.TotalNilai += t.Nilai
                stats.TotalPajak += t.Pajak
                totalPotongan += t.NilaiPotongan

                k, ok := kegiatanMap[t.Kegiatan]
                if !ok {
                        kegiatanMap[t.Kegiatan] = &KegiatanStat{Kegiatan: t.Kegiatan}
                        k = kegiatanMap[t.Kegiatan]
                }
                k.Total += t.Nilai
                k.Count++

                if t.PPTK != "" {
                        p, ok := pptkMap[t.PPTK]
                        if !ok {
                                pptkMap[t.PPTK] = &PPTKStat{PPTK: t.PPTK}
                                p = pptkMap[t.PPTK]
                        }
                        p.Total += t.Nilai
                        p.Count++
                }

                bulan := ""
                if len(t.Tanggal) >= 7 {
                        bulan = t.Tanggal[:7]
                }
                if bulan != "" {
                        m, ok := monthlyMap[bulan]
                        if !ok {
                                monthlyMap[bulan] = &MonthlyStat{Bulan: bulan}
                                m = monthlyMap[bulan]
                        }
                        m.Nilai += t.Nilai
                        m.Pajak += t.Pajak
                }
        }

        stats.NilaiBersih = stats.TotalNilai - stats.TotalPajak - totalPotongan
        stats.Realisasi = stats.TotalNilai

        mod.mu.Lock()
        for _, r := range mod.settings.Rak {
                stats.TotalPagu += r.Anggaran
                if r.PPTK != "" {
                        p, ok := pptkMap[r.PPTK]
                        if !ok {
                                pptkMap[r.PPTK] = &PPTKStat{PPTK: r.PPTK}
                                p = pptkMap[r.PPTK]
                        }
                        p.Pagu += r.Anggaran
                }
        }
        mod.mu.Unlock()
        stats.SisaPagu = stats.TotalPagu - stats.Realisasi
        if stats.TotalPagu > 0 {
                stats.PersenRealisasi = (stats.Realisasi / stats.TotalPagu) * 100
        }

        for _, v := range kegiatanMap {
                stats.NilaiPerKegiatan = append(stats.NilaiPerKegiatan, *v)
        }
        for _, v := range pptkMap {
                stats.NilaiPerPPTK = append(stats.NilaiPerPPTK, *v)
        }
        for _, v := range monthlyMap {
                stats.MonthlyStats = append(stats.MonthlyStats, *v)
        }

        recent := data
        if len(recent) > 5 {
                recent = recent[len(recent)-5:]
        }
        stats.RecentTransaksi = recent

        jsonResponse(w, http.StatusOK, stats)
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
        sess := getSession(r)
        if sess == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        mod := moduleFromRequest(r)
        switch r.Method {
        case http.MethodGet:
                mod.mu.Lock()
                copyAnggaran := cloneAnggaranMap(mod.settings.AnggaranKegiatan)
                rakCopy := cloneRakRows(mod.settings.Rak)
                out := AppSettings{
                        PA:               mod.settings.PA,
                        Bendahara:        mod.settings.Bendahara,
                        AnggaranKegiatan: copyAnggaran,
                        Rak:              rakCopy,
                        RakMeta:          mod.settings.RakMeta,
                }
                mod.mu.Unlock()
                jsonResponse(w, http.StatusOK, out)

        case http.MethodPut:
                if sess.Role != "admin" {
                        jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Akses hanya untuk Admin"})
                        return
                }
                var incoming AppSettings
                if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
                        jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                        return
                }
                mod.mu.Lock()
                if incoming.PA.Nama != "" {
                        mod.settings.PA = incoming.PA
                }
                if incoming.Bendahara.Nama != "" {
                        mod.settings.Bendahara = incoming.Bendahara
                }
                if incoming.AnggaranKegiatan != nil {
                        if mod.settings.AnggaranKegiatan == nil {
                                mod.settings.AnggaranKegiatan = map[string]float64{}
                        }
                        for k, v := range incoming.AnggaranKegiatan {
                                mod.settings.AnggaranKegiatan[k] = v
                        }
                }
                mod.mu.Unlock()
                persistModule(mod)
                jsonResponse(w, http.StatusOK, map[string]string{"message": "Pengaturan berhasil disimpan"})

        default:
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
        }
}

func addSampleData(mod *SipkeuModule) {
        samples := []Transaction{
                {
                        Tanggal: "2026-01-10", Kegiatan: "Administrasi Keuangan Perangkat Daerah",
                        SubKegiatan: "Penyediaan Administrasi Pelaksanaan Tugas ASN",
                        KodeRekening: "5.1.02.02.001.00080",
                        Penerima: "RAMA WARNI, MM",
                        NoBPK: "0001/BPK/UP/1.01.0.00.0.00.01.0000/B01/01/2026",
                        NoBAST: "0001/BAST/DISDIK/I/2026", NoKontrak: "",
                        Pekerjaan: "Belanja Honorarium Penanggungjawaban Pengelola Keuangan",
                        Uraian: "Pembayaran honorarium pengelola keuangan Dinas Pendidikan bulan Januari 2026",
                        JenisPajak: "PPh 21 (5%)", JenisPotongan: "Iuran Wajib Pegawai", Nilai: 5000000, Pajak: 250000, NilaiPotongan: 0,
                        NTPN: "1234567890123456", KodeBilling: "820260100000001", NTB: "BND2026010001",
                        PenggunaAnggaran: "HENDRI ARULAN, S.Pd",
                        PPTK: "RAMA WARNI, MM", PPTKnip: "NIP. 19721203 199802 2 005",
                        Bendahara: "ELDINA SRIDHANTY, SE",
                },
                {
                        Tanggal: "2026-01-20", Kegiatan: "Penyediaan Jasa Penunjang Urusan Pemerintahan Daerah",
                        SubKegiatan: "Penyediaan Jasa Pelayanan Umum Kantor",
                        KodeRekening: "5.1.02.02.001.00067",
                        Penerima: "BANK RIAU KEPRI SYARIAH",
                        NoBPK: "0029/BPK/UP/1.01.0.00.0.00.01.0000/B01/05/2026",
                        NoBAST: "0029/BAST/DISDIK/I/2026", NoKontrak: "",
                        Pekerjaan: "Belanja Pembayaran Pajak, Bea, dan Perizinan",
                        Uraian: "Retribusi Sampah Periode April 2026 DINAS PENDIDIKAN 26987 2605100395",
                        JenisPajak: "", Nilai: 120000, Pajak: 0,
                        NTPN: "", KodeBilling: "", NTB: "",
                        PenggunaAnggaran: "HENDRI ARULAN, S.Pd",
                        PPTK: "RAMA WARNI, MM", PPTKnip: "NIP. 19721203 199802 2 005",
                        Bendahara: "ELDINA SRIDHANTY, SE",
                },
                {
                        Tanggal: "2026-02-05", Kegiatan: "Administrasi Umum Perangkat Daerah",
                        SubKegiatan: "Penyediaan Peralatan dan Perlengkapan Kantor",
                        KodeRekening: "5.1.02.01.001.00024",
                        Penerima: "CV. Maju Jaya",
                        NoBPK: "0005/BPK/UP/1.01.0.00.0.00.01.0000/B01/02/2026",
                        NoBAST: "0005/BAST/DISDIK/II/2026", NoKontrak: "027/SPK/DISDIK/II/2026",
                        Pekerjaan: "Belanja Alat/Bahan untuk Kegiatan Kantor-Alat Tulis Kantor",
                        Uraian: "Pengadaan ATK untuk keperluan operasional kantor Dinas Pendidikan Kota Batam TA 2026",
                        JenisPajak: "PPh 22 (1,5%)", Nilai: 3500000, Pajak: 52500,
                        NTPN: "2345678901234567", KodeBilling: "820260200000002", NTB: "BND2026020001",
                        PenggunaAnggaran: "HENDRI ARULAN, S.Pd",
                        PPTK: "ARIOS ZEUS SANDRY, S.KOM", PPTKnip: "NIP. 19820404 200903 1 002",
                        Bendahara: "ELDINA SRIDHANTY, SE",
                },
                {
                        Tanggal: "2026-03-01", Kegiatan: "Pemeliharaan Barang Milik Daerah Penunjang Urusan Pemerintahan Daerah",
                        SubKegiatan: "Pemeliharaan Peralatan dan Mesin Lainnya",
                        KodeRekening: "5.1.02.03.002.00405",
                        Penerima: "CV. Tekno Mandiri",
                        NoBPK: "0010/BPK/UP/1.01.0.00.0.00.01.0000/B01/03/2026",
                        NoBAST: "0010/BAST/DISDIK/III/2026", NoKontrak: "027/SPK/DISDIK/III/2026",
                        Pekerjaan: "Belanja Pemeliharaan Komputer-Komputer Unit-Personal Computer",
                        Uraian: "Pemeliharaan dan servis 5 unit PC di ruang tata usaha dan kepala bidang",
                        JenisPajak: "PPh 23 (2%)", Nilai: 8500000, Pajak: 170000,
                        NTPN: "3456789012345678", KodeBilling: "820260300000003", NTB: "BND2026030001",
                        PenggunaAnggaran: "HENDRI ARULAN, S.Pd",
                        PPTK: "ARIOS ZEUS SANDRY, S.KOM", PPTKnip: "NIP. 19820404 200903 1 002",
                        Bendahara: "ELDINA SRIDHANTY, SE",
                },
                {
                        Tanggal: "2026-03-20", Kegiatan: "Pengadaan Barang Milik Daerah Penunjang Urusan Pemerintah Daerah",
                        SubKegiatan: "Pengadaan Mebel",
                        KodeRekening: "5.2.02.05.003.00001",
                        Penerima: "UD. Furniture Prima",
                        NoBPK: "0015/BPK/UP/1.01.0.00.0.00.01.0000/B01/03/2026",
                        NoBAST: "0015/BAST/DISDIK/III/2026", NoKontrak: "027/SPK/DISDIK/III/2026-02",
                        Pekerjaan: "Belanja Modal Meja Kerja Pejabat",
                        Uraian: "Pengadaan 3 unit meja kerja pejabat untuk ruang kepala dinas dan kepala bidang",
                        JenisPajak: "PPh Ps.4(2) Final (10%)", Nilai: 25000000, Pajak: 2500000,
                        NTPN: "4567890123456789", KodeBilling: "820260300000004", NTB: "BND2026030002",
                        PenggunaAnggaran: "HENDRI ARULAN, S.Pd",
                        PPTK: "RAMA WARNI, MM", PPTKnip: "NIP. 19721203 199802 2 005",
                        Bendahara: "ELDINA SRIDHANTY, SE",
                },
        }

        for _, s := range samples {
                s.ID = mod.nextID
                mod.nextID++
                mod.txs = append(mod.txs, s)
        }
}

func initSampleAnggaran(mod *SipkeuModule) {
        kegiatanTotals := map[string]float64{}
        for _, t := range mod.txs {
                kegiatanTotals[t.Kegiatan] += t.Nilai
        }
        mod.mu.Lock()
        defer mod.mu.Unlock()
        if mod.settings.AnggaranKegiatan == nil {
                mod.settings.AnggaranKegiatan = map[string]float64{}
        }
        for keg, total := range kegiatanTotals {
                if _, ok := mod.settings.AnggaranKegiatan[keg]; !ok {
                        padded := int(total*1.25/1000000) + 1
                        if padded < 100 {
                                padded = 100
                        }
                        mod.settings.AnggaranKegiatan[keg] = float64(padded) * 1000000
                }
        }
}

func main() {
        port := os.Getenv("PORT")
        if port == "" {
                port = "3000"
        }
        allowedOrigin = strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
        initAuth()

        mux := http.NewServeMux()

        mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{"status":"ok"}`))
        })

        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "text/html; charset=utf-8")
                w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
                w.WriteHeader(http.StatusOK)
                w.Write(indexHTML)
        })

        mux.HandleFunc("/assets/kop-disdik.png", func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "image/png")
                w.Header().Set("Cache-Control", "public, max-age=86400")
                w.Write(kopDisdikPNG)
        })
        mux.HandleFunc("/assets/portal-tanjiro.png", servePNG(portalTanjiroPNG))
        mux.HandleFunc("/assets/portal-nezuko.png", servePNG(portalNezukoPNG))
        mux.HandleFunc("/assets/portal-zenitsu.png", servePNG(portalZenitsuPNG))
        mux.HandleFunc("/assets/logo-batam.png", servePNG(logoBatamPNG))
        if opSub, err := fs.Sub(opRunnersFS, "assets/op-runners"); err == nil {
                mux.Handle("/assets/op-runners/", http.StripPrefix("/assets/op-runners/", http.FileServer(http.FS(opSub))))
        }
        if nrSub, err := fs.Sub(narutoRunnersFS, "assets/naruto-runners"); err == nil {
                mux.Handle("/assets/naruto-runners/", http.StripPrefix("/assets/naruto-runners/", http.FileServer(http.FS(nrSub))))
        }
        if drSub, err := fs.Sub(doraemonRunnersFS, "assets/doraemon-runners"); err == nil {
                mux.Handle("/assets/doraemon-runners/", http.StripPrefix("/assets/doraemon-runners/", http.FileServer(http.FS(drSub))))
        }
        if nsSub, err := fs.Sub(narutoSmpRunnersFS, "assets/naruto-smp-runners"); err == nil {
                mux.Handle("/assets/naruto-smp-runners/", http.StripPrefix("/assets/naruto-smp-runners/", http.FileServer(http.FS(nsSub))))
        }
        if frSub, err := fs.Sub(frozenRunnersFS, "assets/frozen-runners"); err == nil {
                mux.Handle("/assets/frozen-runners/", http.StripPrefix("/assets/frozen-runners/", http.FileServer(http.FS(frSub))))
        }
        if gdSub, err := fs.Sub(gundamIconsFS, "assets/gundam-icons"); err == nil {
                mux.Handle("/assets/gundam-icons/", http.StripPrefix("/assets/gundam-icons/", http.FileServer(http.FS(gdSub))))
        }
        if dsSub, err := fs.Sub(dsKasRunnersFS, "assets/ds-kas-runners"); err == nil {
                mux.Handle("/assets/ds-kas-runners/", http.StripPrefix("/assets/ds-kas-runners/", http.FileServer(http.FS(dsSub))))
        }

        mux.HandleFunc("/data/auth/login", cors(handleLogin))
        mux.HandleFunc("/data/auth/logout", cors(requireAuth(handleLogout)))
        mux.HandleFunc("/data/auth/me", cors(requireAuth(handleMe)))

        mux.HandleFunc("/data/transactions", cors(handleTransactions))
        mux.HandleFunc("/data/transactions/import", cors(handleImport))
        mux.HandleFunc("/data/transactions/delete", cors(requireAuth(handleDeleteTransaction)))
        mux.HandleFunc("/data/transactions/delete-bulk", cors(requireAuth(handleDeleteBulkTransactions)))
        mux.HandleFunc("/data/transactions/delete-all", cors(requireAdmin(handleDeleteAllTransactions)))
        mux.HandleFunc("/data/transactions/", cors(handleTransactionByID))
        mux.HandleFunc("/data/dashboard", cors(handleDashboard))
        mux.HandleFunc("/data/settings", cors(handleSettings))
        mux.HandleFunc("/data/import/anggaran", cors(requireAdmin(handleImportAnggaran)))
        mux.HandleFunc("/data/kas-belanja", cors(requireAuth(handleKasBelanja)))
        mux.HandleFunc("/data/kas-belanja/import-rak", cors(requireAuth(handleKasImportRAK)))
        mux.HandleFunc("/data/kas-belanja/realisasi", cors(requireAuth(handleKasSaveRealisasi)))
        mux.HandleFunc("/data/kas-belanja/realisasi/unlock", cors(requireAuth(handleKasUnlockRealisasi)))

        initSipkeuModules()
        initStorage()
        loadAllModulesFromDisk()
        loadKasFromDisk()

        sek := sipkeuModules["sekretariat"]
        paud := sipkeuModules["paud"]

        if !moduleHasData(sek) {
                addSampleData(sek)
                normalizeModuleIDs(sek)
                persistModule(sek)
        }
        tryLoadDefaultAnggaran()
        sek.mu.Lock()
        hasAnggaranSek := len(sek.settings.AnggaranKegiatan) > 0
        sek.mu.Unlock()
        if !hasAnggaranSek {
                initSampleAnggaran(sek)
                persistModule(sek)
        }
        paud.mu.Lock()
        hasAnggaranPaud := len(paud.settings.AnggaranKegiatan) > 0
        paud.mu.Unlock()
        if !hasAnggaranPaud && hasAnggaranSek {
                sek.mu.Lock()
                paud.settings.Rak = cloneRakRows(sek.settings.Rak)
                paud.settings.AnggaranKegiatan = cloneAnggaranMap(sek.settings.AnggaranKegiatan)
                sek.mu.Unlock()
                persistModule(paud)
        }

        fmt.Printf("%s\n", storageInfo())
        fmt.Printf("Aplikasi Penatausahaan Keuangan berjalan di http://localhost:%s\n", port)
        log.Fatal(http.ListenAndServe(":"+port, withSecurityHeaders(mux)))
}
