const BASE = "https://sakubijak.com";

const login = await fetch(BASE + "/data/auth/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "admin", password: "admin2026" }),
});
const lj = await login.json();
console.log("login status:", login.status, "token?", !!lj.token);
const token = lj.token;

for (const bln of ["juni", "januari", "februari"]) {
const res = await fetch(BASE + "/data/kas-belanja?bulan=" + bln, {
  headers: { Authorization: "Bearer " + token },
});
console.log("\n== bulan:", bln, "status:", res.status);
const d = await res.json();
console.log("rak_rows:", (d.rak_rows || []).length);
const total = (d.report || []).find(r => r.uraian === "TOTAL BELANJA");
console.log("anggaran_kas:", total?.anggaran_kas, "realisasi:", total?.realisasi);
}
if (false) {
const res = await fetch(BASE + "/data/kas-belanja?bulan=februari", {
  headers: { Authorization: "Bearer " + token },
});
console.log("kas status:", res.status);
const d = await res.json();
console.log("rak_rows:", (d.rak_rows || []).length);
console.log("total_pagu:", d.total_pagu);
console.log("bulan:", d.bulan);
console.log("report rows:", (d.report || []).length);
const total = (d.report || []).find(r => r.uraian === "TOTAL BELANJA");
console.log("TOTAL BELANJA row:", total);
console.log("realisasi months:", Object.keys(d.realisasi || {}).filter(k => d.realisasi[k] && Object.keys(d.realisasi[k]).length));
}
