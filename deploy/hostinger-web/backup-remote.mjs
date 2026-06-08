import { Client } from "ssh2";
import fs from "fs";
import os from "os";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../..");
const envFile = path.join(root, "deploy", ".env");

const host = "145.79.14.155";
const port = 65002;
const username = "u657726332";
const RETENTION_DAYS = 30;
const REMOTE_RETENTION_DAYS = 7;

function loadEnvFile(filePath) {
  if (!fs.existsSync(filePath)) return;
  for (const line of fs.readFileSync(filePath, "utf8").split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const i = trimmed.indexOf("=");
    if (i <= 0) continue;
    const key = trimmed.slice(0, i).trim();
    const val = trimmed.slice(i + 1).trim().replace(/^["']|["']$/g, "");
    if (!process.env[key]) process.env[key] = val;
  }
}

loadEnvFile(envFile);

const password = process.env.SSH_PASSWORD;
if (!password) {
  console.error(`Set SSH_PASSWORD di ${envFile} atau environment variable`);
  process.exit(1);
}

const ts = new Date().toISOString().replace(/[-:]/g, "").slice(0, 15).replace("T", "-");
const backupName = `sipkeu-backup-${ts}.tar.gz`;
const remotePath = `/home/${username}/${backupName}`;

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

function download(sftp, remote, local) {
  return new Promise((resolve, reject) => {
    const write = fs.createWriteStream(local);
    const read = sftp.createReadStream(remote);
    read.on("error", reject);
    write.on("error", reject);
    write.on("close", resolve);
    read.pipe(write);
  });
}

function oneDriveDirs() {
  const home = process.env.USERPROFILE || process.env.HOME || "";
  if (!home) return [];
  const dirs = new Set();
  for (const name of fs.readdirSync(home)) {
    if (!name.startsWith("OneDrive")) continue;
    const full = path.join(home, name);
    try {
      if (fs.statSync(full).isDirectory()) {
        dirs.add(path.join(full, "Backup", "SIPKeu"));
      }
    } catch {
      /* skip */
    }
  }
  return [...dirs];
}

function copyToOneDrive(localPath, name) {
  const copied = [];
  for (const dir of oneDriveDirs()) {
    try {
      fs.mkdirSync(dir, { recursive: true });
      const dest = path.join(dir, name);
      fs.copyFileSync(localPath, dest);
      copied.push(dest);
    } catch {
      /* skip unavailable OneDrive folders */
    }
  }
  return copied;
}

function pruneOldBackups(dir, days) {
  if (!fs.existsSync(dir)) return 0;
  const cutoff = Date.now() - days * 24 * 60 * 60 * 1000;
  let removed = 0;
  for (const name of fs.readdirSync(dir)) {
    if (!name.startsWith("sipkeu-backup-") || !name.endsWith(".tar.gz")) continue;
    const file = path.join(dir, name);
    try {
      if (fs.statSync(file).mtimeMs < cutoff) {
        fs.unlinkSync(file);
        removed++;
        console.log(`    Hapus lama: ${file}`);
      }
    } catch {
      /* skip */
    }
  }
  return removed;
}

const tempPath = path.join(os.tmpdir(), backupName);

const conn = new Client();
conn.on("keyboard-interactive", (_n, _i, _l, prompts, finish) => {
  finish(prompts.map(() => password));
});
conn
  .on("ready", async () => {
    try {
      console.log("==> Connected");
      console.log(`==> Creating backup on server: ${remotePath}`);
      await exec(
        conn,
        `tar -czf ${remotePath} -C /home/${username} sipkeu-data && ls -lh ${remotePath}`
      );
      console.log(`==> Downloading (temp): ${tempPath}`);
      await new Promise((resolve, reject) => {
        conn.sftp((err, sftp) => {
          if (err) return reject(err);
          download(sftp, remotePath, tempPath)
            .then(() => { sftp.end(); resolve(); })
            .catch(reject);
        });
      });

      const stat = fs.statSync(tempPath);
      const oneDrive = copyToOneDrive(tempPath, backupName);
      fs.unlinkSync(tempPath);

      if (!oneDrive.length) {
        throw new Error("Folder OneDrive tidak ditemukan — pastikan OneDrive terpasang dan login");
      }

      let pruned = 0;
      for (const dir of oneDriveDirs()) pruned += pruneOldBackups(dir, RETENTION_DAYS);

      console.log(`==> Membersihkan backup remote > ${REMOTE_RETENTION_DAYS} hari`);
      await exec(
        conn,
        `find /home/${username} -maxdepth 1 -name 'sipkeu-backup-*.tar.gz' -mtime +${REMOTE_RETENTION_DAYS} -delete -print`
      );

      console.log("\n==> Backup selesai (OneDrive saja)");
      console.log(`    Ukuran : ${(stat.size / 1024).toFixed(1)} KB`);
      for (const p of oneDrive) console.log(`    OneDrive: ${p}`);
      if (pruned) console.log(`    Dihapus : ${pruned} backup lama (> ${RETENTION_DAYS} hari)`);
      conn.end();
    } catch (e) {
      try { if (fs.existsSync(tempPath)) fs.unlinkSync(tempPath); } catch { /* ignore */ }
      console.error(e);
      conn.end();
      process.exit(1);
    }
  })
  .on("error", (e) => {
    console.error(e);
    process.exit(1);
  })
  .connect({ host, port, username, password, tryKeyboard: true });
