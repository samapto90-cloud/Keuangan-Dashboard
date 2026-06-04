import { Client } from "ssh2";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const uploads = [
  { local: path.join(__dirname, "public_html-proxy.php"), remote: "/home/u657726332/domains/sakubijak.com/public_html/index.php" },
  { local: path.join(__dirname, "public_html-htaccess"), remote: "/home/u657726332/domains/sakubijak.com/public_html/.htaccess" },
];

function upload(sftp, local, remote) {
  return new Promise((resolve, reject) => {
    const read = fs.createReadStream(local);
    const write = sftp.createWriteStream(remote, { mode: 0o644 });
    write.on("close", resolve);
    write.on("error", (e) => reject(e));
    read.pipe(write);
  });
}

const conn = new Client();
conn.on("ready", () => {
  conn.sftp(async (err, sftp) => {
    if (err) throw err;
    try {
      for (const u of uploads) {
        console.log("Upload", path.basename(u.local));
        await upload(sftp, u.local, u.remote);
      }
      console.log("Proxy OK");
      conn.end();
    } catch (e) {
      console.error(e.message);
      conn.end();
      process.exit(1);
    }
  });
}).connect({
  host: "145.79.14.155",
  port: 65002,
  username: "u657726332",
  password: process.env.SSH_PASSWORD,
});
