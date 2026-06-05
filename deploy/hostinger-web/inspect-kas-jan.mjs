import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  const cmd = `
python3 - <<'PY' 2>&1 || echo "python3 unavailable"
import json,os
p=os.path.expanduser("~/sipkeu-data/kas-belanja.json")
d=json.load(open(p))
real=d.get("realisasi") or {}
sm=d.get("sisa_manual") or {}
jan=real.get("januari") or {}
print("== realisasi januari keys (sorted) ==")
for k in sorted(jan):
    print(f"  {k} = {jan[k]:,.2f}")
print("== sisa_manual januari ==")
print(json.dumps(sm.get("januari") or {}, indent=2))
print("== sisa_manual months ==", list(sm.keys()))
# focus kode
kode="5.1.02.03.002.00035"
print("\\n== kode", kode, "==")
print("realisasi januari:", jan.get(kode))
# sum of rak anggaran for this kode prefix in januari (bulan plan)
rr=d.get("rak_rows") or []
s=0; cnt=0
for r in rr:
    kr=(r.get("kode_rekening") or "").strip()
    if kr==kode or kr.startswith(kode+"."):
        cnt+=1
        b=r.get("bulan") or {}
        s+=b.get("januari",0)
print(f"rak rows match {cnt}, sum bulan januari = {s:,.2f}")
for r in rr:
    kr=(r.get("kode_rekening") or "").strip()
    if kr==kode:
        print("  ROW kegiatan:", (r.get("nama_kegiatan") or "")[:40], "| anggaran:", r.get("anggaran"), "| jan:", (r.get("bulan") or {}).get("januari"))
PY
`;
  conn.exec(cmd, (err, stream) => {
    if (err) { console.error(err); conn.end(); return; }
    stream.on("data", (d) => process.stdout.write(d));
    stream.stderr.on("data", (d) => process.stderr.write(d));
    stream.on("close", () => conn.end());
  });
}).connect({
  host: "145.79.14.155",
  port: 65002,
  username: "u657726332",
  password: process.env.SSH_PASSWORD,
});
