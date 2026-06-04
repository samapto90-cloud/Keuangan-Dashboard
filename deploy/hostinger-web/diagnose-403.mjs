import { Client } from "ssh2";

const conn = new Client();
conn.on("ready", () => {
  const cmd = `
echo "=== public_html ==="
ls -la ~/domains/sakubijak.com/public_html/ 2>&1
echo "=== index.php head ==="
head -5 ~/domains/sakubijak.com/public_html/index.php 2>&1
echo "=== .htaccess ==="
cat ~/domains/sakubijak.com/public_html/.htaccess 2>&1
echo "=== Go process ==="
ps aux | grep -E 'keuangan|8888' | grep -v grep
echo "=== health local ==="
curl -s -o /dev/null -w "health:%{http_code}\\n" http://127.0.0.1:8888/health
curl -s -o /dev/null -w "root:%{http_code}\\n" http://127.0.0.1:8888/
echo "=== PHP test ==="
php -r 'echo "php_ok\\n";' 2>&1
echo "=== sipkeu.log tail ==="
tail -20 ~/sipkeu.log 2>&1
echo "=== domain dirs ==="
ls -la ~/domains/ 2>&1
`;
  conn.exec(cmd, (err, stream) => {
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
