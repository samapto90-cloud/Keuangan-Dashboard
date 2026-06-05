import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  const cmd = `
echo "=== PHP disable_functions ==="
php -i 2>/dev/null | grep -i "disable_functions"
echo "=== test exec ==="
php -r 'echo function_exists("exec")?"exec_enabled\\n":"exec_DISABLED\\n"; echo function_exists("shell_exec")?"shell_exec_enabled\\n":"shell_exec_DISABLED\\n"; echo function_exists("popen")?"popen_enabled\\n":"popen_DISABLED\\n";'
echo "=== keepalive present ==="
ls -la ~/hostinger-web/keepalive.sh 2>&1
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
