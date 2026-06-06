import { Client } from "ssh2";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../..");
const goApp = path.join(root, "go-app");

const host = "145.79.14.155";
const port = 65002;
const username = "u657726332";
const password = process.env.SSH_PASSWORD;
if (!password) {
  console.error("Set SSH_PASSWORD environment variable");
  process.exit(1);
}

const uploads = [
  { local: path.join(goApp, "keuangan-linux-amd64"), remote: "/home/u657726332/sipkeu/keuangan.new", mode: 0o755 },
  { local: path.join(goApp, "Anggaran.xlsx"), remote: "/home/u657726332/sipkeu/Anggaran.xlsx" },
  { local: path.join(__dirname, "start-remote.sh"), remote: "/home/u657726332/hostinger-web/start-remote.sh", mode: 0o755 },
  { local: path.join(__dirname, "keepalive.sh"), remote: "/home/u657726332/hostinger-web/keepalive.sh", mode: 0o755 },
  { local: path.join(__dirname, ".env.production"), remote: "/home/u657726332/hostinger-web/.env" },
  { local: path.join(__dirname, "public_html-proxy.php"), remote: "/home/u657726332/domains/sakubijak.com/public_html/index.php", mode: 0o644 },
  { local: path.join(__dirname, "public_html-htaccess"), remote: "/home/u657726332/domains/sakubijak.com/public_html/.htaccess", mode: 0o644 },
];

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

function upload(sftp, local, remote, mode) {
  return new Promise((resolve, reject) => {
    const read = fs.createReadStream(local);
    const write = sftp.createWriteStream(remote, { mode: mode || 0o644 });
    write.on("close", resolve);
    write.on("error", (e) => reject(new Error(`upload ${remote}: ${e.message}`)));
    read.on("error", (e) => reject(new Error(`read ${local}: ${e.message}`)));
    read.pipe(write);
  });
}

const conn = new Client();
conn.on("keyboard-interactive", (_name, _instructions, _lang, prompts, finish) => {
  finish(prompts.map(() => password));
});
conn
  .on("ready", async () => {
    try {
      console.log("==> Connected");
      await exec(conn, "mkdir -p ~/sipkeu ~/sipkeu-data ~/hostinger-web ~/domains/sakubijak.com/public_html");
      await exec(conn, "pkill -x keuangan || pkill -f sipkeu/keuangan || true");
      await new Promise((r) => setTimeout(r, 2000));
      await new Promise((resolve, reject) => {
        conn.sftp((err, sftp) => {
          if (err) return reject(err);
          (async () => {
            for (const u of uploads) {
              if (!fs.existsSync(u.local)) {
                console.log(`==> Skip (missing): ${path.basename(u.local)}`);
                continue;
              }
              console.log(`==> Upload ${path.basename(u.local)}`);
              await upload(sftp, u.local, u.remote, u.mode);
            }
            sftp.end();
            resolve();
          })().catch(reject);
        });
      });
      await exec(conn, "mv -f ~/sipkeu/keuangan.new ~/sipkeu/keuangan && chmod +x ~/sipkeu/keuangan");
      await exec(conn, "for f in ~/hostinger-web/start-remote.sh ~/hostinger-web/keepalive.sh ~/hostinger-web/.env; do [ -f \"$f\" ] && perl -pi -e 's/\\r\\n/\\n/g' \"$f\" 2>/dev/null || sed -i 's/\\r$//' \"$f\" 2>/dev/null || true; done");
      console.log("==> Install keepalive cron (tiap 2 menit, jika crontab tersedia)");
      await exec(conn, "command -v crontab >/dev/null 2>&1 && ((crontab -l 2>/dev/null | grep -v 'sipkeu keepalive'; echo '*/2 * * * * bash $HOME/hostinger-web/keepalive.sh # sipkeu keepalive') | crontab - && echo cron-ok) || echo 'cron-skip: pasang via hPanel/MCP -> */2 * * * * bash $HOME/hostinger-web/keepalive.sh'");
      console.log("==> Verify public_html proxy");
      const pub = "~/domains/sakubijak.com/public_html";
      await exec(conn, `[ -f ${pub}/index.php ] || (echo "ERROR: ${pub}/index.php hilang — jangan taruh git clone di public_html" && exit 1)`);
      await exec(conn, `[ -f ${pub}/.htaccess ] || (echo "ERROR: ${pub}/.htaccess hilang" && exit 1)`);
      console.log("==> Start SIPKEU");
      await exec(conn, "bash ~/hostinger-web/start-remote.sh");
      console.log("\nDeploy OK: https://sakubijak.com:8888");
      conn.end();
    } catch (e) {
      console.error("Deploy failed:", e.message);
      conn.end();
      process.exit(1);
    }
  })
  .on("error", (e) => {
    console.error("SSH error:", e.message);
    process.exit(1);
  })
  .connect({ host, port, username, password, tryKeyboardInteractive: true, readyTimeout: 20000 });
