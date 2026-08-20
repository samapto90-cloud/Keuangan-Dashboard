# Manual QA Checklist — Ular Tangga Nusantara v1.0.0

**Environment:** local staging  
**URL:** http://localhost:5174/cahaya/ (dev) atau http://127.0.0.1:PORT/cahaya/ (prod build)  
**Date:** _______________ **Tester:** _______________

Gunakan 2 browser/profile (Player A & Player B) untuk flow multiplayer.

---

## 0. Pre-flight (otomatis / cepat)

- [ ] `GET /health` → status ok, version `1.0.0`
- [ ] `GET /ready` → ready
- [ ] `GET /cahaya/` → landing/game HTML load
- [ ] Console browser: tidak ada error merah saat load awal
- [ ] Network: assets JS/CSS 200

---

## 1. Landing & First Impression

- [ ] Landing hero: **ULAR TANGGA NUSANTARA**
- [ ] Tagline / CTA: **MAIN SEKARANG**, **LIHAT CARA BERMAIN**
- [ ] Section fitur (multiplayer, belajar, ranked, reward)
- [ ] How to play 5 langkah terlihat
- [ ] CTA bawah: **SIAP MENJADI JUARA?**
- [ ] Tidak ada horizontal overflow di 360 / 390 / 414 / 768 / 1024 / 1440

---

## 2. Onboarding & Tutorial

- [ ] User baru → onboarding muncul
- [ ] Flow: Welcome → Avatar → Username → Board → Dadu → Soal → Ular → Tangga → Mulai
- [ ] Demo dadu berfungsi
- [ ] Demo soal: B=benar, A/C/D=salah feedback
- [ ] **Lewati Tutorial** menyimpan dan tidak muncul lagi
- [ ] Settings → **Reset Tutorial** → muncul lagi setelah reload

---

## 3. Auth

- [ ] Register akun baru (username 3–16, password ≥8)
- [ ] Login sukses
- [ ] Login gagal (password salah) → error jelas, tidak crash
- [ ] Logout → kembali ke home/landing
- [ ] Login lagi → sesi/profile tetap (XP/coins tersimpan)

---

## 4. Profile & Persistence

- [ ] Profil: level, XP, coins, rank label
- [ ] Daily reward status / claim (jika enabled)
- [ ] Achievement list (empty state ramah jika kosong)
- [ ] History match tampil setelah selesai bermain
- [ ] Reload page → data masih ada

---

## 5. Settings / Audio / A11y

- [ ] Master / Music / SFX slider berubah volume
- [ ] Mute menghentikan suara
- [ ] Quality LOW mengurangi particle
- [ ] Reduce Motion mengurangi animasi
- [ ] Font SMALL / NORMAL / LARGE
- [ ] Theme System / Light / Dark
- [ ] Preferensi tersimpan setelah reload

---

## 6. Social

- [ ] Cari pemain
- [ ] Kirim friend request (A→B)
- [ ] Terima request (B)
- [ ] Invite ke room
- [ ] Block / unblock (opsional)
- [ ] Notifikasi unread badge
- [ ] Empty state teman kosong ramah

---

## 7. Multiplayer Casual

- [ ] Create room (2–4 seats)
- [ ] Join via kode
- [ ] Ready semua pemain
- [ ] Host start → countdown → board
- [ ] Giliran jelas (**GILIRANMU!** / waiting)
- [ ] Roll dice → animasi → gerakan
- [ ] Snake / ladder efek
- [ ] Question modal: timer, opsi mudah ditekan, tidak close accidental
- [ ] Benar / salah / timeout feedback
- [ ] Finish match → result (rank, XP, coins)
- [ ] Share result (share API atau copy)

---

## 8. Ranked

- [ ] Queue RANKED
- [ ] Match found
- [ ] Settlement RR berubah
- [ ] Leaderboard ranked update

---

## 9. Questions

- [ ] Soal random dari 4 mapel
- [ ] Jawaban benar tidak bocor sebelum submit
- [ ] Wrong → mundur 10
- [ ] Timeout → mundur 10
- [ ] Final challenge di 100

---

## 10. Reconnect

- [ ] Matikan network singkat saat match
- [ ] Banner **RECONNECTING…**
- [ ] Koneksi kembali → tidak kick langsung
- [ ] State match sync lagi

---

## 11. Offline / PWA

- [ ] Offline screen muncul saat putus total
- [ ] Manifest + icons load
- [ ] Install prompt (Chrome/Edge jika eligible)
- [ ] Setelah install: tidak paksa install lagi
- [ ] SW tidak mengklaim multiplayer offline

---

## 12. Feedback

- [ ] Kirim feedback Bug / Suggestion
- [ ] Question Issue + sub-kategori
- [ ] Rate limit: spam → ditolak (429)

---

## 13. Admin

- [ ] `/admin` login role admin
- [ ] Dashboard status
- [ ] Player list / sanction
- [ ] Question CRUD
- [ ] Report moderation
- [ ] Config view/edit (super admin)
- [ ] Non-admin ditolak (403)

---

## 14. Security spot-checks

- [ ] Client tidak bisa set position/xp/coins via WS payload
- [ ] Admin API tanpa token → 401
- [ ] IDOR: tidak bisa edit profil user lain

---

## 15. Performance / polish

- [ ] Tidak blank screen saat load
- [ ] Tidak layout shift besar
- [ ] Touch targets nyaman di mobile
- [ ] Safe area (notch) tidak menutupi ROLL/ANSWER
- [ ] Console production: tidak ada debug spam

---

## Hasil

| Area | PASS / FAIL | Catatan |
|---|---|---|
| Landing |  |  |
| Onboarding |  |  |
| Auth |  |  |
| Profile |  |  |
| Settings |  |  |
| Social |  |  |
| Casual MP |  |  |
| Ranked |  |  |
| Questions |  |  |
| Reconnect |  |  |
| PWA |  |  |
| Admin |  |  |
| Security |  |  |

**CRITICAL bugs:**  
**HIGH bugs:**  
**MEDIUM / LOW:**  

**Verdict:** READY TO DEPLOY / HOLD
