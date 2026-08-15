import { Client } from "ssh2";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../..");

function loadDotEnv(filePath) {
  if (!fs.existsSync(filePath)) return;
  for (const line of fs.readFileSync(filePath, "utf8").split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eq = trimmed.indexOf("=");
    if (eq <= 0) continue;
    const key = trimmed.slice(0, eq).trim();
    if (process.env[key]) continue;
    process.env[key] = trimmed.slice(eq + 1).trim().replace(/^['"]|['"]$/g, "");
  }
}

loadDotEnv(path.join(root, "deploy", ".env"));

const password = process.env.SSH_PASSWORD;
if (!password) {
  console.error("SSH_PASSWORD kosong di deploy/.env");
  process.exit(1);
}

const conn = new Client();
conn.on("keyboard-interactive", (_n, _i, _l, prompts, finish) => {
  finish(prompts.map(() => password));
});

conn
  .on("ready", () => {
    const cmd = `
echo "=== process before ==="
ps aux | grep -E '[k]euangan' || echo "no keuangan process"
echo "=== start ==="
bash ~/hostinger-web/start-remote.sh
echo "=== local health ==="
curl -s -m 8 http://127.0.0.1:8888/health || echo "local health FAIL"
echo
echo "=== last log ==="
tail -30 ~/sipkeu.log 2>/dev/null || true
`;
    conn.exec(cmd, (err, stream) => {
      if (err) {
        console.error(err);
        conn.end();
        process.exit(1);
      }
      stream.on("data", (d) => process.stdout.write(d));
      stream.stderr.on("data", (d) => process.stderr.write(d));
      stream.on("close", (code) => {
        conn.end();
        process.exit(code === 0 || code == null ? 0 : code);
      });
    });
  })
  .on("error", (e) => {
    console.error("SSH error:", e.message);
    process.exit(1);
  })
  .connect({
    host: "145.79.14.155",
    port: 65002,
    username: "u657726332",
    password,
    tryKeyboard: true,
    readyTimeout: 30000,
  });
