import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  conn.exec("df -h ~; ls -la ~/sipkeu/; ls -la ~/domains/sakubijak.com/public_html/ | head -20", (err, stream) => {
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