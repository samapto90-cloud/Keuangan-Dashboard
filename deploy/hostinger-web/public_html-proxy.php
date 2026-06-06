<?php
/**
 * Proxy SIPKEU → Go backend di localhost:8888
 */
$uri = $_SERVER['REQUEST_URI'] ?? '/';
if (preg_match('#\.\.|//|\x00#', $uri)) {
    http_response_code(400);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode(['error' => 'Permintaan tidak valid']);
    exit;
}

$maxBody = 12 * 1024 * 1024;
$contentLength = (int)($_SERVER['CONTENT_LENGTH'] ?? 0);
if ($contentLength > $maxBody) {
    http_response_code(413);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode(['error' => 'Ukuran data terlalu besar']);
    exit;
}

header('X-Content-Type-Options: nosniff');
header('X-Frame-Options: SAMEORIGIN');
header('Referrer-Policy: strict-origin-when-cross-origin');
if (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') {
    header('Strict-Transport-Security: max-age=31536000; includeSubDomains');
}

$backend = 'http://127.0.0.1:8888';
$url = $backend . $uri;

$method = strtoupper($_SERVER['REQUEST_METHOD'] ?? 'GET');
$headers = [];
if (function_exists('getallheaders')) {
    foreach (getallheaders() as $k => $v) {
        $lk = strtolower($k);
        if (in_array($lk, ['host', 'connection', 'content-length', 'transfer-encoding'], true)) {
            continue;
        }
        $headers[] = $k . ': ' . $v;
    }
}
$headers[] = 'X-Forwarded-Proto: ' . ((!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http');
$headers[] = 'X-Forwarded-Host: ' . ($_SERVER['HTTP_HOST'] ?? 'sakubijak.com');
$clientIp = $_SERVER['REMOTE_ADDR'] ?? '';
if ($clientIp !== '') {
    $headers[] = 'X-Real-IP: ' . $clientIp;
    $headers[] = 'X-Forwarded-For: ' . $clientIp;
}

$body = file_get_contents('php://input');
$hasBody = in_array($method, ['POST', 'PUT', 'PATCH'], true) && $body !== false && $body !== '';

$ch = curl_init($url);
$opts = [
    CURLOPT_CUSTOMREQUEST => $method,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HEADER => true,
    CURLOPT_HTTPHEADER => $headers,
    CURLOPT_TIMEOUT => 120,
    CURLOPT_CONNECTTIMEOUT => 15,
    CURLOPT_FOLLOWLOCATION => false,
    CURLOPT_HTTP_VERSION => CURL_HTTP_VERSION_1_1,
];
if ($hasBody) {
    $opts[CURLOPT_POSTFIELDS] = $body;
}
curl_setopt_array($ch, $opts);

$response = curl_exec($ch);
$curlErr = curl_error($ch);
if ($response === false) {
    http_response_code(502);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode([
        'error' => 'SIPKEU backend tidak merespons',
        'detail' => $curlErr ?: 'Pastikan aplikasi Go berjalan di port 8888',
    ]);
    curl_close($ch);
    exit;
}

$code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
curl_close($ch);

$rawHeaders = substr($response, 0, $headerSize);
$bodyOut = substr($response, $headerSize);
http_response_code($code);

foreach (explode("\r\n", $rawHeaders) as $line) {
    if ($line === '' || stripos($line, 'HTTP/') === 0) {
        continue;
    }
    $p = strpos($line, ':');
    if ($p === false) {
        continue;
    }
    $name = trim(substr($line, 0, $p));
    $value = trim(substr($line, $p + 1));
    if (strcasecmp($name, 'Transfer-Encoding') === 0) {
        continue;
    }
    header($name . ': ' . $value, false);
}

echo $bodyOut;
