#!/usr/bin/env python3
"""Parse Kumpulan_Lengkap_390_Soal_SMA_dan_SD.pdf → sma-edu-questions.json"""
from __future__ import annotations

import json
import re
import time
from pathlib import Path

from pypdf import PdfReader

PDF = Path(r"c:\Users\DIK_u\OneDrive\Documents\Kumpulan_Lengkap_390_Soal_SMA_dan_SD.pdf")
OUT = Path(__file__).resolve().parents[1] / "mmo" / "data" / "sma-edu-questions.json"

SUBJECT_MAP = {
    "PAI": "PAI",
    "PENDIDIKAN AGAMA ISLAM": "PAI",
    "PENDIDIKAN AGAMA ISLAM (PAI)": "PAI",
    "BAHASA INGGRIS": "BAHASA_INGGRIS",
    "B. INGGRIS": "BAHASA_INGGRIS",
    "MATEMATIKA": "MATEMATIKA",
    "BAHASA JAWA": "BAHASA_JAWA",
    "B. JAWA": "BAHASA_JAWA",
    "IPA": "IPA",
    "IPS": "IPS",
    "EKONOMI": "IPS",
    "GEOGRAFI": "IPS",
    "SEJARAH": "IPS",
}


def norm_subject(raw: str) -> str:
    s = re.sub(r"\s+", " ", raw.strip().upper())
    s = s.split("(")[0].strip()
    for k, v in SUBJECT_MAP.items():
        if k in s or s in k:
            return v
    return "UMUM"


def clean_opt(t: str) -> str:
    t = t.replace("\u2218", "").replace("◦", "").replace("•", "")
    t = re.sub(r"\s+", " ", t).strip(" .\t\n\r-–—")
    return t


def parse_answers(text: str) -> dict[int, str]:
    ans: dict[int, str] = {}
    # patterns like 1 (B), 301.B, 181.B
    for m in re.finditer(r"(\d+)\s*[\.(]\s*([A-E])\s*[\)]?", text, re.I):
        ans[int(m.group(1))] = m.group(2).upper()
    for m in re.finditer(r"(\d+)\.([A-E])\b", text, re.I):
        ans[int(m.group(1))] = m.group(2).upper()
    return ans


def extract_blocks(text: str) -> list[tuple[int, str]]:
    """Return [(num, body_including_options)]."""
    # Normalize letter labels that appear on separate lines after options
    text = re.sub(r"\n\s*([A-E])\.\s*\n", "\n", text)
    parts = re.split(r"(?m)^\s*(\d{1,3})\.\s+", text)
    # parts[0]=preamble, then num, body, num, body...
    out: list[tuple[int, str]] = []
    i = 1
    while i + 1 < len(parts):
        num = int(parts[i])
        body = parts[i + 1]
        # stop body at next kunci / section header noise somewhat
        body = re.split(r"(?m)^Kunci Jawaban", body)[0]
        out.append((num, body.strip()))
        i += 2
    return out


def parse_options_inline(body: str) -> tuple[str, list[str]] | None:
    """SD style: question... A. x · B. y · C. z · D. w"""
    # Replace middle dots
    body = body.replace("·", " ").replace("�", " ")
    m = re.search(
        r"^(.*?)\s*A\.\s*(.*?)\s*B\.\s*(.*?)\s*C\.\s*(.*?)\s*D\.\s*(.*?)(?:\s*E\.\s*(.*?))?$",
        body,
        re.S | re.I,
    )
    if not m:
        return None
    q = clean_opt(m.group(1))
    opts = [clean_opt(m.group(i) or "") for i in range(2, 7)]
    opts = [o for o in opts if o]
    if len(opts) < 4 or not q:
        return None
    return q, opts


def parse_options_multiline(body: str) -> tuple[str, list[str]] | None:
    """SMA style: question then 5 lines of options without A. labels (labels dumped later)."""
    lines = [clean_opt(x) for x in body.splitlines()]
    lines = [x for x in lines if x and not re.fullmatch(r"[A-E]\.", x, re.I)]
    if len(lines) < 5:
        return None
    # Last 4 or 5 lines are options; question is the rest
    # Heuristic: if body has subject tags like "(Ekonomi)" keep in question
    # Find where options start: typically after blank conceptually — take last 5 non-empty as opts if 5 opts
    for n_opts in (5, 4):
        if len(lines) < n_opts + 1:
            continue
        opts = lines[-n_opts:]
        q = " ".join(lines[:-n_opts]).strip()
        # Options shouldn't be too long paragraphs
        if q and all(1 <= len(o) <= 180 for o in opts):
            return q, opts
    return None


def to_abcd(opts: list[str], correct: str) -> tuple[str, str, str, str, str] | None:
    letters = ["A", "B", "C", "D", "E"]
    correct = correct.upper()
    if correct not in letters[: len(opts)]:
        return None
    idx = letters.index(correct)
    if len(opts) >= 5 and correct == "E":
        # Keep A,B,C and move E→D
        a, b, c, d = opts[0], opts[1], opts[2], opts[4]
        return a, b, c, d, "D"
    # Drop E if present
    four = opts[:4]
    if idx >= 4:
        return None
    while len(four) < 4:
        four.append("-")
    return four[0], four[1], four[2], four[3], correct


def main() -> None:
    reader = PdfReader(str(PDF))
    full = "\n".join((p.extract_text() or "") for p in reader.pages)
    answers = parse_answers(full)
    blocks = extract_blocks(full)

    # Track current subject heading
    subject = "UMUM"
    grade = "SMA"
    items = []
    now = int(time.time() * 1000)
    seen = set()

    # Pre-scan headings in full text for subject switches by line order — apply per block via searching preceding text
    # Simpler: walk text and assign subject when we see subject headers
    # Rebuild with positions
    for num, body in blocks:
        if num < 1 or num > 390:
            continue
        if num in seen:
            continue
        if num >= 221:
            grade = "SD"
            subject = "UMUM"
        else:
            grade = "SMA"

        # Detect subject tag in body like "(Ekonomi)" or leading subject name
        tag = re.search(r"^\(([^)]+)\)\s*", body)
        if tag:
            subject = norm_subject(tag.group(1))
            body = body[tag.end() :]
        else:
            head = re.match(
                r"^(Pendidikan Agama Islam(?:\s*\(PAI\))?|Bahasa Inggris|Matematika|Bahasa Jawa|IPA|IPS)\s*",
                body,
                re.I,
            )
            if head:
                subject = norm_subject(head.group(1))
                body = body[head.end() :]

        parsed = parse_options_inline(body) or parse_options_multiline(body)
        if not parsed:
            continue
        qtext, opts = parsed
        # Skip "dilewati"
        if "dilewati" in qtext.lower():
            continue
        ans = answers.get(num)
        if not ans:
            # For obvious SD math we can leave unset — skip
            continue
        mapped = to_abcd(opts, ans)
        if not mapped:
            continue
        a, b, c, d, corr = mapped
        # difficulty rough
        if grade == "SD":
            diff = "EASY" if num < 300 else "MEDIUM"
        else:
            diff = "EASY" if num <= 60 else "MEDIUM" if num <= 150 else "HARD"

        # Better subject from nearby for SMA: use answer-key section subjects periodically
        # Keep UMUM for SD mixed
        qid = f"{grade.lower()}-{num:03d}"
        items.append(
            {
                "id": qid,
                "subject": subject if grade == "SMA" else "UMUM",
                "category": subject.lower() if grade == "SMA" else "sd",
                "grade": grade,
                "difficulty": diff,
                "question": qtext,
                "optionA": a,
                "optionB": b,
                "optionC": c,
                "optionD": d,
                "correctAnswer": corr,
                "explanation": f"Jawaban yang tepat adalah {corr}.",
                "active": True,
                "createdAt": now,
                "updatedAt": now,
            }
        )
        seen.add(num)

    # Improve SMA subjects using answer key session ranges
    # Sessions of 30: PAI 1-5, Eng 6-10, Math 11-15, Jawa 16-20, IPA 21-25, IPS 26-30
    cycle = [
        (1, 5, "PAI"),
        (6, 10, "BAHASA_INGGRIS"),
        (11, 15, "MATEMATIKA"),
        (16, 20, "BAHASA_JAWA"),
        (21, 25, "IPA"),
        (26, 30, "IPS"),
    ]
    for it in items:
        if it["grade"] != "SMA":
            continue
        n = int(it["id"].split("-")[1])
        if n > 180:
            # mixed PNS etc → UMUM/IPS
            if it["subject"] == "UMUM":
                it["subject"] = "IPS"
            continue
        pos = ((n - 1) % 30) + 1
        for a, b, sub in cycle:
            if a <= pos <= b:
                it["subject"] = sub
                it["category"] = sub.lower()
                break

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(items, ensure_ascii=False, indent=2), encoding="utf-8")
    sma = sum(1 for x in items if x["grade"] == "SMA")
    sd = sum(1 for x in items if x["grade"] == "SD")
    print(f"wrote {len(items)} questions → {OUT} (SMA={sma} SD={sd})")
    missing = [i for i in range(1, 391) if i not in seen and i != 299]
    print("missing count", len(missing), "sample", missing[:20])


if __name__ == "__main__":
    main()
