package main

import (
        "encoding/json"
        _ "embed"
        "fmt"
        "log"
        "net/http"
        "os"
        "strconv"
        "strings"
        "sync"
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
        RecentTransaksi  []Transaction  `json:"recent_transaksi"`
        MonthlyStats     []MonthlyStat  `json:"monthly_stats"`
}

type KegiatanStat struct {
        Kegiatan string  `json:"kegiatan"`
        Total    float64 `json:"total"`
        Count    int     `json:"count"`
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
}

var (
        transactions []Transaction
        nextID       = 1
        mu           sync.Mutex
        appSettings  = AppSettings{
                PA:        Pejabat{Nama: "HENDRI ARULAN, S.Pd", Nip: "NIP. 19670119 199103 1 009"},
                Bendahara: Pejabat{Nama: "ELDINA SRIDHANTY, SE", Nip: "NIP. 19810610 201001 2 002"},
                AnggaranKegiatan: map[string]float64{},
        }
        settingsMu sync.RWMutex
)

func cors(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusOK)
                        return
                }
                next(w, r)
        }
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
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(data)
}

func handleTransactions(w http.ResponseWriter, r *http.Request) {
        if getSession(r) == nil {
                jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
                return
        }
        switch r.Method {
        case http.MethodGet:
                mu.Lock()
                result := make([]Transaction, len(transactions))
                copy(result, transactions)
                mu.Unlock()
                jsonResponse(w, http.StatusOK, result)

        case http.MethodPost:
                var t Transaction
                if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
                        jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                        return
                }
                mu.Lock()
                t.ID = nextID
                nextID++
                transactions = append(transactions, t)
                mu.Unlock()
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

        switch r.Method {
        case http.MethodPut:
                var updated Transaction
                if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
                        jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
                        return
                }
                mu.Lock()
                found := false
                for i, t := range transactions {
                        if t.ID == id {
                                updated.ID = id
                                transactions[i] = updated
                                found = true
                                break
                        }
                }
                mu.Unlock()
                if !found {
                        jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Not found"})
                        return
                }
                jsonResponse(w, http.StatusOK, updated)

        case http.MethodDelete:
                mu.Lock()
                found := false
                for i, t := range transactions {
                        if t.ID == id {
                                transactions = append(transactions[:i], transactions[i+1:]...)
                                found = true
                                break
                        }
                }
                mu.Unlock()
                if !found {
                        jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Not found"})
                        return
                }
                jsonResponse(w, http.StatusOK, map[string]string{"message": "Deleted"})

        default:
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
        }
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
        mu.Lock()
        for i := range items {
                items[i].ID = nextID
                nextID++
                transactions = append(transactions, items[i])
        }
        mu.Unlock()
        jsonResponse(w, http.StatusOK, map[string]interface{}{
                "imported": len(items),
                "message":  fmt.Sprintf("%d transaksi berhasil diimpor", len(items)),
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

        mu.Lock()
        data := make([]Transaction, len(transactions))
        copy(data, transactions)
        mu.Unlock()

        stats := DashboardStats{}
        stats.TotalTransaksi = len(data)

        kegiatanMap := map[string]*KegiatanStat{}
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

        settingsMu.RLock()
        for _, r := range appSettings.Rak {
                stats.TotalPagu += r.Anggaran
        }
        settingsMu.RUnlock()
        stats.SisaPagu = stats.TotalPagu - stats.Realisasi
        if stats.TotalPagu > 0 {
                stats.PersenRealisasi = (stats.Realisasi / stats.TotalPagu) * 100
        }

        for _, v := range kegiatanMap {
                stats.NilaiPerKegiatan = append(stats.NilaiPerKegiatan, *v)
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
        switch r.Method {
        case http.MethodGet:
                settingsMu.RLock()
                copyAnggaran := map[string]float64{}
                for k, v := range appSettings.AnggaranKegiatan {
                        copyAnggaran[k] = v
                }
                rakCopy := make([]RakRow, len(appSettings.Rak))
                copy(rakCopy, appSettings.Rak)
                out := AppSettings{
                        PA:               appSettings.PA,
                        Bendahara:        appSettings.Bendahara,
                        AnggaranKegiatan: copyAnggaran,
                        Rak:              rakCopy,
                }
                settingsMu.RUnlock()
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
                settingsMu.Lock()
                if incoming.PA.Nama != "" {
                        appSettings.PA = incoming.PA
                }
                if incoming.Bendahara.Nama != "" {
                        appSettings.Bendahara = incoming.Bendahara
                }
                if incoming.AnggaranKegiatan != nil {
                        if appSettings.AnggaranKegiatan == nil {
                                appSettings.AnggaranKegiatan = map[string]float64{}
                        }
                        for k, v := range incoming.AnggaranKegiatan {
                                appSettings.AnggaranKegiatan[k] = v
                        }
                }
                settingsMu.Unlock()
                jsonResponse(w, http.StatusOK, map[string]string{"message": "Pengaturan berhasil disimpan"})

        default:
                jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
        }
}

func addSampleData() {
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
                        NoBPK: "0029/BPK/UP/1.01.0.00.0.00.01.0000/B03/05/2026",
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
                        NoBPK: "0005/BPK/UP/1.01.0.00.0.00.01.0000/B02/02/2026",
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
                        NoBPK: "0010/BPK/UP/1.01.0.00.0.00.01.0000/B03/03/2026",
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
                        NoBPK: "0015/BPK/UP/1.01.0.00.0.00.01.0000/B04/03/2026",
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
                s.ID = nextID
                nextID++
                transactions = append(transactions, s)
        }
}

func initSampleAnggaran() {
        kegiatanTotals := map[string]float64{}
        for _, t := range transactions {
                kegiatanTotals[t.Kegiatan] += t.Nilai
        }
        settingsMu.Lock()
        defer settingsMu.Unlock()
        if appSettings.AnggaranKegiatan == nil {
                appSettings.AnggaranKegiatan = map[string]float64{}
        }
        for keg, total := range kegiatanTotals {
                if _, ok := appSettings.AnggaranKegiatan[keg]; !ok {
                        padded := int(total*1.25/1000000) + 1
                        if padded < 100 {
                                padded = 100
                        }
                        appSettings.AnggaranKegiatan[keg] = float64(padded) * 1000000
                }
        }
}

func main() {
        port := os.Getenv("PORT")
        if port == "" {
                port = "3000"
        }

        mux := http.NewServeMux()

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

        mux.HandleFunc("/data/auth/login", cors(handleLogin))
        mux.HandleFunc("/data/auth/logout", cors(requireAuth(handleLogout)))
        mux.HandleFunc("/data/auth/me", cors(requireAuth(handleMe)))

        mux.HandleFunc("/data/transactions", cors(handleTransactions))
        mux.HandleFunc("/data/transactions/import", cors(handleImport))
        mux.HandleFunc("/data/transactions/", cors(handleTransactionByID))
        mux.HandleFunc("/data/dashboard", cors(handleDashboard))
        mux.HandleFunc("/data/settings", cors(handleSettings))
        mux.HandleFunc("/data/import/anggaran", cors(requireAdmin(handleImportAnggaran)))

        addSampleData()
        tryLoadDefaultAnggaran()
        settingsMu.RLock()
        hasAnggaran := len(appSettings.AnggaranKegiatan) > 0
        settingsMu.RUnlock()
        if !hasAnggaran {
                initSampleAnggaran()
        }

        fmt.Printf("Aplikasi Penatausahaan Keuangan berjalan di http://localhost:%s\n", port)
        log.Fatal(http.ListenAndServe(":"+port, mux))
}
