import { Client } from "ssh2";
import fs from "fs";
import path from "path";
import readline from "readline";
import { spawnSync } from "child_process";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../..");
const goApp = path.join(root, "go-app");

const host = "145.79.14.155";
const port = 65002;
const username = "u657726332";

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

function askPassword() {
  if (process.env.SSH_PASSWORD) return Promise.resolve(process.env.SSH_PASSWORD);
  return new Promise((resolve) => {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    rl.question("Password SSH Hostinger (hPanel > SSH Access): ", (answer) => {
      rl.close();
      resolve(answer.trim());
    });
  });
}

function ensureBinary() {
  const bin = path.join(goApp, "keuangan-linux-amd64");
  if (fs.existsSync(bin)) return bin;
  console.log("==> Binary belum ada, build Linux...");
  const r = spawnSync("go", ["build", "-ldflags=-s -w -X main.buildSHA=manual", "-o", "keuangan-linux-amd64", "."], {
    cwd: goApp,
    env: { ...process.env, GOOS: "linux", GOARCH: "amd64", CGO_ENABLED: "0" },
    stdio: "inherit",
    shell: true,
  });
  if (r.status !== 0) {
    console.error("Build gagal. Pastikan Go terinstall.");
    process.exit(1);
  }
  return bin;
}

async function deploy(password) {
  if (!password) {
    console.error("Password SSH kosong.");
    process.exit(1);
  }
  ensureBinary();

  const conn = new Client();
  conn.on("keyboard-interactive", (_name, _instructions, _lang, prompts, finish) => {
    finish(prompts.map(() => password));
  });

  await new Promise((resolve, reject) => {
    conn
      .on("ready", async () => {
        try {
          console.log("==> Connected");
          await exec(conn, "mkdir -p ~/sipkeu ~/sipkeu-data ~/hostinger-web ~/domains/sakubijak.com/public_html");
          await exec(conn, "pkill -x keuangan || pkill -f sipkeu/keuangan || true");
          await new Promise((r) => setTimeout(r, 2000));
          await new Promise((res, rej) => {
            conn.sftp((err, sftp) => {
              if (err) return rej(err);
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
                res();
              })().catch(rej);
            });
          });
          await exec(conn, "mv -f ~/sipkeu/keuangan.new ~/sipkeu/keuangan && chmod +x ~/sipkeu/keuangan");
          await exec(conn, "for f in ~/hostinger-web/start-remote.sh ~/hostinger-web/keepalive.sh ~/hostinger-web/.env; do [ -f \"$f\" ] && perl -pi -e 's/\\r\\n/\\n/g' \"$f\" 2>/dev/null || sed -i 's/\\r$//' \"$f\" 2>/dev/null || true; done");
          console.log("==> Start SIPKEU");
          await exec(conn, "bash ~/hostinger-web/start-remote.sh");
          console.log("\nDeploy OK: https://sakubijak.com:8888");
          console.log("Cek versi: https://sakubijak.com:8888/health");
          conn.end();
          resolve();
        } catch (e) {
          console.error("Deploy failed:", e.message);
          conn.end();
          reject(e);
        }
      })
      .on("error", (e) => reject(new Error(`SSH error: ${e.message}`)))
      .connect({ host, port, username, password, tryKeyboard: true, tryKeyboardInteractive: true, readyTimeout: 30000 });
  });
}

const password = await askPassword();
await deploy(password);
