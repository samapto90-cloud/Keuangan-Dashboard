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
        Nilai            float64 `json:"nilai"`
        Pajak            float64 `json:"pajak"`
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

var (
        transactions []Transaction
        nextID       = 1
        mu           sync.Mutex
)

func cors(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusOK)
                        return
                }
                next(w, r)
        }
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(data)
}

func handleTransactions(w http.ResponseWriter, r *http.Request) {
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

func handleDashboard(w http.ResponseWriter, r *http.Request) {
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

        for _, t := range data {
                stats.TotalNilai += t.Nilai
                stats.TotalPajak += t.Pajak

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

        stats.NilaiBersih = stats.TotalNilai - stats.TotalPajak

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

func addSampleData() {
        samples := []Transaction{
                {
                        Tanggal: "2026-01-10", Kegiatan: "Administrasi Keuangan Perangkat Daerah",
                        SubKegiatan: "Penyediaan Administrasi Pelaksanaan Tugas ASN",
                        KodeRekening: "5.1.02.02.001.00080",
                        Penerima: "RAMA WARNI, MM",
                        NoBPK: "0001/BPK/UP/1.01.0.00.0.00.01.0000/B01/01/2026",
                        NoBAST: "0001/BAST/DISDIK/I/2026",
                        NoKontrak: "",
                        Pekerjaan: "Belanja Honorarium Penanggungjawaban Pengelola Keuangan",
                        Uraian: "Pembayaran honorarium pengelola keuangan Dinas Pendidikan bulan Januari 2026",
                        Nilai: 5000000, Pajak: 500000,
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
                        NoBAST: "0029/BAST/DISDIK/I/2026",
                        NoKontrak: "",
                        Pekerjaan: "Belanja Pembayaran Pajak, Bea, dan Perizinan",
                        Uraian: "Retribusi Sampah Periode April 2026 DINAS PENDIDIKAN 26987 2605100395",
                        Nilai: 120000, Pajak: 12000,
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
                        NoBAST: "0005/BAST/DISDIK/II/2026",
                        NoKontrak: "027/SPK/DISDIK/II/2026",
                        Pekerjaan: "Belanja Alat/Bahan untuk Kegiatan Kantor-Alat Tulis Kantor",
                        Uraian: "Pengadaan ATK untuk keperluan operasional kantor Dinas Pendidikan Kota Batam TA 2026",
                        Nilai: 3500000, Pajak: 350000,
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
                        NoBAST: "0010/BAST/DISDIK/III/2026",
                        NoKontrak: "027/SPK/DISDIK/III/2026",
                        Pekerjaan: "Belanja Pemeliharaan Komputer-Komputer Unit-Personal Computer",
                        Uraian: "Pemeliharaan dan servis 5 unit PC di ruang tata usaha dan kepala bidang",
                        Nilai: 8500000, Pajak: 850000,
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
                        NoBAST: "0015/BAST/DISDIK/III/2026",
                        NoKontrak: "027/SPK/DISDIK/III/2026-02",
                        Pekerjaan: "Belanja Modal Meja Kerja Pejabat",
                        Uraian: "Pengadaan 3 unit meja kerja pejabat untuk ruang kepala dinas dan kepala bidang",
                        Nilai: 25000000, Pajak: 2500000,
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

func main() {
        port := os.Getenv("PORT")
        if port == "" {
                port = "3000"
        }

        mux := http.NewServeMux()

        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "text/html; charset=utf-8")
                w.WriteHeader(http.StatusOK)
                w.Write(indexHTML)
        })

        mux.HandleFunc("/data/transactions", cors(handleTransactions))
        mux.HandleFunc("/data/transactions/", cors(handleTransactionByID))
        mux.HandleFunc("/data/dashboard", cors(handleDashboard))

        addSampleData()

        fmt.Printf("Aplikasi Penatausahaan Keuangan berjalan di http://localhost:%s\n", port)
        log.Fatal(http.ListenAndServe(":"+port, mux))
}
