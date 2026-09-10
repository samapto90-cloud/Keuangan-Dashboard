/**
 * Deploy SIPKEU binary via Hostinger API (TUS upload + one-shot cron).
 * Needs HOSTINGER_API_TOKEN (hPanel → Token API).
 *
 * Flow:
 * 1) POST /api/hosting/v1/files/upload-urls
 * 2) TUS upload go-app/keuangan-linux-amd64 → public_html/_sipkeu_deploy/keuangan.new
 * 3) Create * * * * * cron that moves binary + start-remote.sh
 * 4) Poll cron output, then delete cron
 */
import fs from "fs";
import path from "path";
import { spawnSync } from "child_process";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../..");
const goApp = path.join(root, "go-app");

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

const API_BASE = process.env.HOSTINGER_API_BASE || "https://developers.hostinger.com";
const USERNAME = process.env.HOSTINGER_USERNAME || "u657726332";
const DOMAIN = process.env.HOSTINGER_DOMAIN || "sakubijak.com";
const REL_PATH = "_sipkeu_deploy/keuangan.new";

function token() {
  const t = process.env.HOSTINGER_API_TOKEN || process.env.HOSTINGER_TOKEN;
  if (!t) throw new Error("HOSTINGER_API_TOKEN kosong (hPanel → Token API).");
  return t;
}

async function api(method, apiPath, body) {
  const url = `${API_BASE}${apiPath}`;
  const headers = {
    Authorization: `Bearer ${token()}`,
    "User-Agent": "sipkeu-deploy-api/1.0",
    Accept: "application/json",
  };
  let payload;
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }
  const res = await fetch(url, { method, headers, body: payload });
  const text = await res.text();
  let data;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = { raw: text };
  }
  if (!res.ok) {
    throw new Error(`${method} ${apiPath} → ${res.status}: ${text.slice(0, 400)}`);
  }
  return data;
}

function ensureBinary() {
  const bin = path.join(goApp, "keuangan-linux-amd64");
  if (process.env.SKIP_BINARY_REBUILD === "1" && fs.existsSync(bin)) {
    console.log("==> Skip rebuild (SKIP_BINARY_REBUILD=1)");
    return bin;
  }
  const sha =
    process.env.BUILD_SHA ||
    (spawnSync("git", ["rev-parse", "--short", "HEAD"], { cwd: root, encoding: "utf8" }).stdout || "").trim() ||
    "local";
  console.log(`==> Build Linux binary (buildSHA=${sha})...`);
  const r = spawnSync(
    "go",
    ["build", "-ldflags", `-s -w -X main.buildSHA=${sha}`, "-o", "keuangan-linux-amd64", "."],
    {
      cwd: goApp,
      env: { ...process.env, GOOS: "linux", GOARCH: "amd64", CGO_ENABLED: "0" },
      stdio: "inherit",
    },
  );
  if (r.status !== 0) throw new Error("go build gagal");
  if (!fs.existsSync(bin)) throw new Error("binary tidak ada: " + bin);
  return bin;
}

async function tusUpload(uploadMeta, localFile, relativePath) {
  const size = fs.statSync(localFile).size;
  const base = uploadMeta.url.replace(/\/$/, "");
  const target = `${base}/${relativePath}?override=true`;
  const headers = {
    "X-Auth": uploadMeta.auth_key,
    "X-Auth-Rest": uploadMeta.rest_auth_key,
    "Tus-Resumable": "1.0.0",
  };

  console.log(`==> TUS create ${relativePath} (${size} bytes)`);
  const create = await fetch(target, {
    method: "POST",
    headers: {
      ...headers,
      "Upload-Length": String(size),
      "Upload-Offset": "0",
    },
  });
  if (create.status !== 201 && create.status !== 200) {
    const t = await create.text();
    throw new Error(`TUS create ${create.status}: ${t.slice(0, 300)}`);
  }

  // Chunked PATCH (8 MiB) — more reliable on flaky links than one giant body
  const chunkSize = 8 * 1024 * 1024;
  const fd = fs.openSync(localFile, "r");
  let offset = 0;
  try {
    while (offset < size) {
      const len = Math.min(chunkSize, size - offset);
      const buf = Buffer.alloc(len);
      fs.readSync(fd, buf, 0, len, offset);
      process.stdout.write(`==> TUS patch ${offset}/${size}\r`);
      const patch = await fetch(target, {
        method: "PATCH",
        headers: {
          ...headers,
          "Content-Type": "application/offset+octet-stream",
          "Upload-Offset": String(offset),
          "Content-Length": String(len),
        },
        body: buf,
      });
      if (patch.status !== 204 && patch.status !== 200) {
        const t = await patch.text();
        throw new Error(`TUS patch @${offset} → ${patch.status}: ${t.slice(0, 300)}`);
      }
      const next = patch.headers.get("upload-offset");
      offset = next ? Number(next) : offset + len;
    }
  } finally {
    fs.closeSync(fd);
  }
  console.log(`\n==> TUS done ${size} bytes`);
}

function installCronCommand() {
  // Keep under Hostinger 255-char cron command limit
  return [
    'f=$HOME/domains/sakubijak.com/public_html/_sipkeu_deploy/keuangan.new',
    '[ -f "$f" ]&&(pkill -x keuangan||true;sleep 1;mv -f "$f" $HOME/sipkeu/keuangan;chmod +x $HOME/sipkeu/keuangan;rm -rf ${f%/*};bash $HOME/hostinger-web/start-remote.sh)',
  ].join(";");
}

async function sleep(ms) {
  await new Promise((r) => setTimeout(r, ms));
}

async function main() {
  const bin = ensureBinary();

  console.log("==> Request upload URL");
  const uploadMeta = await api("POST", "/api/hosting/v1/files/upload-urls", {
    username: USERNAME,
    domain: DOMAIN,
  });
  if (!uploadMeta?.url || !uploadMeta?.auth_key || !uploadMeta?.rest_auth_key) {
    throw new Error("upload-urls response tidak lengkap: " + JSON.stringify(uploadMeta).slice(0, 300));
  }

  await tusUpload(uploadMeta, bin, REL_PATH);

  const command = installCronCommand();
  if (command.length > 255) throw new Error(`cron command terlalu panjang (${command.length})`);

  console.log("==> Create install cron");
  const created = await api("POST", `/api/hosting/v1/accounts/${USERNAME}/cron-jobs`, {
    time: "* * * * *",
    command,
  });
  const uid = created?.uid || created?.data?.uid;
  if (!uid) throw new Error("cron uid missing: " + JSON.stringify(created).slice(0, 300));
  console.log("==> Cron uid", uid);

  let ok = false;
  for (let i = 0; i < 8; i++) {
    await sleep(20000);
    try {
      const out = await api("GET", `/api/hosting/v1/accounts/${USERNAME}/cron-jobs/${uid}/output`);
      const text = typeof out === "string" ? out : JSON.stringify(out);
      console.log("==> Cron output:", text.slice(0, 500));
      // After first fire, binary should be gone from public_html path; check via list is hard — treat any output as progress
      if (i >= 2) {
        ok = true;
        break;
      }
    } catch (e) {
      console.log("==> Cron output wait:", e.message);
    }
  }

  console.log("==> Delete install cron");
  try {
    await api("DELETE", `/api/hosting/v1/accounts/${USERNAME}/cron-jobs/${uid}`);
  } catch (e) {
    console.warn("==> Hapus cron gagal (hapus manual di hPanel):", e.message);
  }

  console.log("\nDeploy API selesai. Cek: https://sakubijak.com:8888/health");
  if (!ok) console.warn("Peringatan: output cron belum terkonfirmasi; cek keepalive / health.");
}

main().catch((e) => {
  console.error("Deploy API failed:", e.message);
  process.exit(1);
});
