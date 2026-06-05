import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  const cmd = `
echo "=== sipkeu-data listing ==="
ls -la ~/sipkeu-data/ 2>&1
echo "=== kas-belanja.json size & head ==="
wc -c ~/sipkeu-data/kas-belanja.json 2>&1
echo "--- rak_rows count (grep) ---"
python3 - <<'PY' 2>&1 || echo "python3 unavailable"
import json,os
p=os.path.expanduser("~/sipkeu-data/kas-belanja.json")
try:
    d=json.load(open(p))
    rr=d.get("rak_rows") or []
    print("rak_rows:", len(rr))
    real=d.get("realisasi") or {}
    print("realisasi months:", [k for k,v in real.items() if v])
    print("version:", d.get("version"), d.get("version_label"))
    print("imported_at:", d.get("imported_at"))
    if rr[:1]:
        print("sample row keys:", list(rr[0].keys()))
        print("sample bulan:", rr[0].get("bulan"))
except Exception as e:
    print("ERR", e)
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
