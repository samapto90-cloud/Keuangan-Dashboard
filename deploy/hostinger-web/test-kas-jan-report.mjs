const BASE = "https://sakubijak.com";
const login = await fetch(BASE + "/data/auth/login", {
  method: "POST", headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "admin", password: "admin2026" }),
});
const token = (await login.json()).token;
const res = await fetch(BASE + "/data/kas-belanja?bulan=januari", {
  headers: { Authorization: "Bearer " + token },
});
const d = await res.json();
console.log("== JANUARI report rows ==");
for (const r of (d.report || [])) {
  console.log(
    (r.kode || "(total)").padEnd(10),
    String(r.uraian).padEnd(26).slice(0,26),
    "lalu=", r.sisa_bulan_lalu.toLocaleString("id-ID"),
    "| angg=", r.anggaran_kas.toLocaleString("id-ID"),
    "| real=", r.realisasi.toLocaleString("id-ID"),
    "| s/d=", r.sisa_sd.toLocaleString("id-ID"),
    "| %=", r.persen.toFixed(2)
  );
}
