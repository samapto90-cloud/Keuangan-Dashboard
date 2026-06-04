import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  conn.exec("curl -s http://127.0.0.1:8888/ | sed -n '2924,2930p'; ps aux | grep keuangan | grep -v grep", (err, stream) => {
    stream.on("data", (d) => process.stdout.write(d));
    stream.on("close", () => conn.end());
  });
}).connect({
  host: "145.79.14.155",
  port: 65002,
  username: "u657726332",
  password: process.env.SSH_PASSWORD,
});
