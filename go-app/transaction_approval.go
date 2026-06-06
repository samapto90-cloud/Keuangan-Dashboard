package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	trxStatusPending  = "pending"
	trxStatusApproved = "approved"
	trxStatusRejected = "rejected"
)

func effectiveTrxStatus(t Transaction) string {
	s := strings.TrimSpace(strings.ToLower(t.Status))
	if s == "" {
		return trxStatusApproved
	}
	return s
}

func trxCountsTowardRealisasi(t Transaction) bool {
	s := effectiveTrxStatus(t)
	return s == trxStatusApproved || s == trxStatusPending
}

func trxIsApproved(t Transaction) bool {
	return effectiveTrxStatus(t) == trxStatusApproved
}

func stampTransactionOnCreate(sess *Session, t *Transaction) {
	now := time.Now().Format(time.RFC3339)
	if strings.TrimSpace(t.SubmittedAt) == "" {
		t.SubmittedAt = now
	}
	if sess != nil && strings.TrimSpace(t.CreatedBy) == "" {
		t.CreatedBy = sess.Username
	}
	if sess != nil && sess.Role == "admin" {
		t.Status = trxStatusApproved
		t.ReviewedBy = sess.Username
		t.ReviewedAt = now
		return
	}
	t.Status = trxStatusPending
	t.ReviewedBy = ""
	t.ReviewedAt = ""
	t.ReviewNote = ""
}

func operatorMayModifyTransaction(sess *Session, t Transaction) bool {
	if sess == nil || sess.Role == "admin" {
		return true
	}
	if sess.Role != "operator" {
		return false
	}
	st := effectiveTrxStatus(t)
	// Operator hanya boleh ubah/hapus setelah admin menolak; pending = sudah diajukan (kunci).
	if st != trxStatusRejected {
		return false
	}
	if cb := strings.TrimSpace(t.CreatedBy); cb != "" && cb != sess.Username {
		return false
	}
	return true
}

func mergeTransactionUpdate(sess *Session, existing, updated Transaction) (Transaction, error) {
	if sess == nil {
		return Transaction{}, fmt.Errorf("unauthorized")
	}
	if sess.Role == "operator" && !operatorMayModifyTransaction(sess, existing) {
		st := effectiveTrxStatus(existing)
		if st == trxStatusPending {
			return Transaction{}, fmt.Errorf("Transaksi sudah diajukan dan tidak dapat diubah. Tunggu persetujuan atau penolakan Admin.")
		}
		if st == trxStatusApproved {
			return Transaction{}, fmt.Errorf("Transaksi yang sudah disetujui tidak dapat diubah operator.")
		}
		return Transaction{}, fmt.Errorf("Transaksi tidak dapat diubah oleh operator.")
	}

	updated.ID = existing.ID
	updated.CreatedBy = existing.CreatedBy
	if updated.CreatedBy == "" {
		updated.CreatedBy = sess.Username
	}
	if strings.TrimSpace(updated.SubmittedAt) == "" {
		updated.SubmittedAt = existing.SubmittedAt
	}

	if sess.Role == "admin" {
		incoming := strings.TrimSpace(strings.ToLower(updated.Status))
		switch incoming {
		case trxStatusPending, trxStatusApproved, trxStatusRejected:
			updated.Status = incoming
		default:
			updated.Status = existing.Status
		}
		if strings.TrimSpace(updated.Status) == "" {
			updated.Status = trxStatusApproved
		}
		if effectiveTrxStatus(updated) == trxStatusApproved && effectiveTrxStatus(existing) != trxStatusApproved {
			updated.ReviewedBy = sess.Username
			updated.ReviewedAt = time.Now().Format(time.RFC3339)
		}
		return updated, nil
	}

	updated.Status = trxStatusPending
	updated.SubmittedAt = time.Now().Format(time.RFC3339)
	updated.ReviewedBy = ""
	updated.ReviewedAt = ""
	updated.ReviewNote = ""
	return updated, nil
}

func findModuleTransaction(mod *SipkeuModule, id int) (Transaction, bool) {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	for _, t := range mod.txs {
		if t.ID == id {
			return t, true
		}
	}
	return Transaction{}, false
}

func handleTransactionReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	sess := getSession(r)
	if sess == nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid, silakan login"})
		return
	}
	if sess.Role != "admin" {
		jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Hanya Admin yang dapat menyetujui atau menolak transaksi"})
		return
	}

	var body struct {
		ID     int    `json:"id"`
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if body.ID <= 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "ID transaksi tidak valid"})
		return
	}

	action := strings.TrimSpace(strings.ToLower(body.Action))
	var newStatus string
	switch action {
	case "approve", "setuju":
		newStatus = trxStatusApproved
	case "reject", "tolak":
		newStatus = trxStatusRejected
	default:
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Aksi harus approve atau reject"})
		return
	}

	mod := moduleFromRequest(r)
	existing, ok := findModuleTransaction(mod, body.ID)
	if !ok {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Transaksi tidak ditemukan"})
		return
	}
	if effectiveTrxStatus(existing) != trxStatusPending {
		jsonResponse(w, http.StatusConflict, map[string]string{"error": "Transaksi ini sudah diproses sebelumnya"})
		return
	}

	now := time.Now().Format(time.RFC3339)
	updated := existing
	updated.Status = newStatus
	updated.ReviewedBy = sess.Username
	updated.ReviewedAt = now
	updated.ReviewNote = strings.TrimSpace(body.Note)

	mod.mu.Lock()
	found := false
	for i, t := range mod.txs {
		if t.ID == body.ID {
			mod.txs[i] = updated
			found = true
			break
		}
	}
	mod.mu.Unlock()
	if !found {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Transaksi tidak ditemukan"})
		return
	}
	persistModule(mod)

	verb := "disetujui"
	if newStatus == trxStatusRejected {
		verb = "ditolak"
	}
	recordAudit(sess.Username, "review_transaksi", mod.ID,
		fmt.Sprintf("Transaksi #%d %s (BPK %s)", body.ID, verb, strings.TrimSpace(existing.NoBPK)), clientIP(r))

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"transaction": updated,
		"message":     fmt.Sprintf("Transaksi #%d berhasil %s.", body.ID, verb),
	})
}
