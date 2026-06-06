import { Client } from "ssh2";

const host = "145.79.14.155";
const port = 65002;
const username = "u657726332";
const password = process.env.SSH_PASSWORD;
if (!password) {
  console.error("Set SSH_PASSWORD");
  process.exit(1);
}

function exec(conn, cmd) {
  return new Promise((resolve, reject) => {
    conn.exec(cmd, (err, stream) => {
      if (err) return reject(err);
      let out = "";
      stream.on("data", (d) => { out += d; process.stdout.write(d); });
      stream.stderr.on("data", (d) => process.stderr.write(d));
      stream.on("close", (code) => (code === 0 || code === null ? resolve(out) : reject(new Error(`exit ${code}`))));
    });
  });
}

const cmd = `bash -lc '
set -a
source ~/hostinger-web/.env 2>/dev/null || true
set +a
TOKEN=$(curl -sf -X POST http://127.0.0.1:8888/data/auth/login \\
  -H "Content-Type: application/json" \\
  -d "{\\"username\\":\\"$SIPKEU_ADMIN_USER\\",\\"password\\":\\"$SIPKEU_ADMIN_PASSWORD\\"}" | sed -n "s/.*\\"token\\":\\"\\([^\\"]*\\)\\".*/\\1/p")
if [ -z "$TOKEN" ]; then echo "FAIL: login"; exit 1; fi
curl -sf -X POST http://127.0.0.1:8888/data/admin/backfill-np2d \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json"
'`;

const conn = new Client();
conn
  .on("ready", async () => {
    try {
      console.log("==> Backfill NP2D on server");
      await exec(conn, cmd);
      conn.end();
    } catch (e) {
      console.error(e);
      conn.end();
      process.exit(1);
    }
  })
  .on("error", (e) => {
    console.error(e);
    process.exit(1);
  })
  .connect({ host, port, username, password });
