import { storedSessionToken } from "../auth/session";
import { toast } from "./chrome";

const CATEGORIES = [
  { id: "BUG", label: "Bug" },
  { id: "SUGGESTION", label: "Saran" },
  { id: "QUESTION_ISSUE", label: "Masalah Soal" },
  { id: "OTHER", label: "Lainnya" },
] as const;

const QUESTION_SUB = [
  { id: "WRONG_ANSWER", label: "Jawaban Salah" },
  { id: "TYPO", label: "Typo" },
  { id: "UNCLEAR", label: "Soal Tidak Jelas" },
  { id: "OTHER", label: "Lainnya" },
];

export function openFeedbackModal(context?: { page?: string; matchId?: string; questionId?: string }): void {
  const layer = document.createElement("div");
  layer.className = "ui-modal-layer";
  layer.dataset.modal = "feedback";
  layer.innerHTML = `
    <div class="ui-modal nt-card" role="dialog" aria-modal="true">
      <h2>Kirim Feedback</h2>
      <form class="nt-form" id="fb-form">
        <label>Kategori
          <select name="category" required>${CATEGORIES.map((c) => `<option value="${c.id}">${c.label}</option>`).join("")}</select>
        </label>
        <label class="fb-q-sub" hidden>Sub-kategori Soal
          <select name="questionSub">${QUESTION_SUB.map((c) => `<option value="${c.id}">${c.label}</option>`).join("")}</select>
        </label>
        <label>Pesan<textarea name="message" required rows="4" maxlength="1000" placeholder="Jelaskan masalah atau saran…"></textarea></label>
        <button type="submit" class="nt-btn nt-btn-primary">KIRIM</button>
        <button type="button" class="nt-btn nt-btn-ghost" data-close>Batal</button>
      </form>
    </div>`;
  document.body.appendChild(layer);

  const form = layer.querySelector<HTMLFormElement>("#fb-form");
  const cat = form?.querySelector<HTMLSelectElement>('select[name="category"]');
  const qSub = layer.querySelector<HTMLElement>(".fb-q-sub");
  cat?.addEventListener("change", () => {
    if (qSub) qSub.hidden = cat.value !== "QUESTION_ISSUE";
  });

  layer.querySelector("[data-close]")?.addEventListener("click", () => layer.remove());
  layer.addEventListener("click", (e) => {
    if (e.target === layer) layer.remove();
  });

  form?.addEventListener("submit", (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const category = String(fd.get("category") || "OTHER");
    const message = String(fd.get("message") || "").trim();
    if (message.length < 5) {
      toast("Pesan terlalu pendek.", "warning");
      return;
    }
    const token = storedSessionToken();
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (token) headers.Authorization = `Bearer ${token}`;
    void fetch("/cahaya/api/feedback", {
      method: "POST",
      headers,
      body: JSON.stringify({
        category,
        message,
        questionSub: category === "QUESTION_ISSUE" ? String(fd.get("questionSub") || "") : "",
        page: context?.page || location.pathname,
        matchId: context?.matchId || "",
        questionId: context?.questionId || "",
      }),
    })
      .then(async (res) => {
        if (!res.ok) {
          const j = (await res.json().catch(() => ({}))) as { error?: string };
          throw new Error(j.error || "Gagal mengirim");
        }
        toast("Terima kasih! Feedback diterima.", "success");
        layer.remove();
      })
      .catch((err: Error) => toast(err.message || "Gagal mengirim", "error"));
  });
}
