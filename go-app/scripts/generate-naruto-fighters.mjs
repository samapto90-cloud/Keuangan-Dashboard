/**
 * Generate / enhance Naruto & Sasuke portal assets via Gemini 3 Pro Image.
 *
 * Usage (from repo root):
 *   set GEMINI_API_KEY=your_key          (Windows)
 *   export GEMINI_API_KEY=your_key       (Linux/macOS)
 *   node go-app/scripts/generate-naruto-fighters.mjs
 *
 * Optional: GEMINI_MODEL=gemini-3-pro-image (default)
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const assetsDir = path.join(__dirname, "../assets/naruto-smp-runners");
const model = process.env.GEMINI_MODEL || "gemini-3-pro-image";
const apiKey = process.env.GEMINI_API_KEY;

const prompts = {
  naruto: `Create a hyper-realistic cinematic 3D render of Naruto Uzumaki (Shippuden orange/black tracksuit, Konoha headband) in dynamic ninja sprint pose, full body, transparent background PNG style, dramatic rim lighting, photoreal skin and fabric texture, no text, no watermark, game-quality character art.`,
  sasuke: `Create a hyper-realistic cinematic 3D render of Sasuke Uchiha (Shippuden dark blue outfit, purple rope belt, sword on back) in dynamic battle lunge pose, full body, transparent background PNG style, dramatic purple rim lighting, photoreal texture, no text, no watermark, game-quality character art.`,
};

async function generateImage(name, prompt) {
  const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent`;
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-goog-api-key": apiKey,
    },
    body: JSON.stringify({
      contents: [{ role: "user", parts: [{ text: prompt }] }],
      generationConfig: {
        responseModalities: ["IMAGE", "TEXT"],
      },
    }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data?.error?.message || JSON.stringify(data).slice(0, 400));
  }
  const parts = data?.candidates?.[0]?.content?.parts || [];
  for (const part of parts) {
    const inline = part.inlineData || part.inline_data;
    if (inline?.data) {
      const out = path.join(assetsDir, `${name}.png`);
      fs.writeFileSync(out, Buffer.from(inline.data, "base64"));
      console.log(`OK: ${out}`);
      return;
    }
  }
  throw new Error(`No image in response for ${name}`);
}

if (!apiKey) {
  console.error("GEMINI_API_KEY belum diset. Tambahkan ke deploy/.env lalu jalankan ulang.");
  process.exit(1);
}

console.log(`Model: ${model}`);
for (const [name, prompt] of Object.entries(prompts)) {
  console.log(`Generating ${name}…`);
  await generateImage(name, prompt);
}
console.log("Selesai. Rebuild & deploy agar asset baru tampil di portal.");
