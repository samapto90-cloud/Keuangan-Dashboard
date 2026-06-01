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
        Penerima         string  `json:"penerima"`
        NoBAST           string  `json:"no_bast"`
        KeteranganUraian string  `json:"keterangan_uraian"`
        Nilai            float64 `json:"nilai"`
        Pajak            float64 `json:"pajak"`
        PenggunaAnggaran string  `json:"pengguna_anggaran"`
        PPTK             string  `json:"pptk"`
        Bendahara        string  `json:"bendahara"`
}

type DashboardStats struct {
        TotalTransaksi  int                `json:"total_transaksi"`
        TotalNilai      float64            `json:"total_nilai"`
        TotalPajak      float64            `json:"total_pajak"`
        NilaiBersih     float64            `json:"nilai_bersih"`
        NilaiPerKegiatan []KegiatanStat    `json:"nilai_per_kegiatan"`
        RecentTransaksi  []Transaction     `json:"recent_transaksi"`
        MonthlyStats     []MonthlyStat     `json:"monthly_stats"`
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
                        Tanggal: "2026-01-10", Kegiatan: "Pengadaan ATK", SubKegiatan: "Pembelian Alat Tulis Kantor",
                        Penerima: "CV. Maju Jaya", NoBAST: "BAST/001/I/2026", KeteranganUraian: "Pembelian ATK untuk kebutuhan kantor bulan Januari",
                        Nilai: 5000000, Pajak: 500000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-01-20", Kegiatan: "Perjalanan Dinas", SubKegiatan: "Perjalanan Dinas Dalam Kota",
                        Penerima: "H. Ridwan Hasan", NoBAST: "BAST/002/I/2026", KeteranganUraian: "Perjalanan dinas koordinasi ke Provinsi",
                        Nilai: 3500000, Pajak: 350000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-02-05", Kegiatan: "Honorarium", SubKegiatan: "Honor Narasumber",
                        Penerima: "Prof. Dr. Wijaya, M.Pd", NoBAST: "BAST/003/II/2026", KeteranganUraian: "Honorarium narasumber pelatihan SDM",
                        Nilai: 8000000, Pajak: 800000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-02-15", Kegiatan: "Pengadaan ATK", SubKegiatan: "Pembelian Toner dan Kertas",
                        Penerima: "UD. Berkah Mandiri", NoBAST: "BAST/004/II/2026", KeteranganUraian: "Pengadaan toner printer dan kertas A4 dan F4",
                        Nilai: 12000000, Pajak: 1200000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-03-01", Kegiatan: "Pemeliharaan", SubKegiatan: "Service Kendaraan Dinas",
                        Penerima: "Bengkel Karya Mandiri", NoBAST: "BAST/005/III/2026", KeteranganUraian: "Service rutin kendaraan dinas roda empat",
                        Nilai: 6500000, Pajak: 650000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-03-20", Kegiatan: "Perjalanan Dinas", SubKegiatan: "Perjalanan Dinas Luar Daerah",
                        Penerima: "Ir. Suharto, M.T", NoBAST: "BAST/006/III/2026", KeteranganUraian: "Perjalanan dinas menghadiri rapat koordinasi nasional",
                        Nilai: 15000000, Pajak: 1500000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-04-10", Kegiatan: "Honorarium", SubKegiatan: "Honor Panitia Kegiatan",
                        Penerima: "Tim Panitia", NoBAST: "BAST/007/IV/2026", KeteranganUraian: "Honorarium panitia pelaksana kegiatan workshop",
                        Nilai: 10000000, Pajak: 1000000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
                },
                {
                        Tanggal: "2026-05-05", Kegiatan: "Pemeliharaan", SubKegiatan: "Pemeliharaan Gedung",
                        Penerima: "CV. Konstruksi Prima", NoBAST: "BAST/008/V/2026", KeteranganUraian: "Pemeliharaan atap gedung kantor yang bocor",
                        Nilai: 25000000, Pajak: 2500000, PenggunaAnggaran: "Drs. Ahmad Fauzi, M.Si", PPTK: "Budi Santoso, S.E", Bendahara: "Siti Rahayu, A.Md",
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
