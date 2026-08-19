package mmo

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	accountBcryptCost = 10
	sessionTTL        = 12 * time.Hour
	minPasswordLen    = 8
)

var usernameRE = regexp.MustCompile(`^[A-Za-z0-9_]{3,18}$`)

type GameAccount struct {
	PlayerID     string `json:"playerId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	CreatedAt    int64  `json:"createdAt"`
}

type GameSession struct {
	Token    string    `json:"token"`
	PlayerID string    `json:"playerId"`
	Username string    `json:"username"`
	Expires  time.Time `json:"expires"`
}

type accountBlob struct {
	Users    []GameAccount `json:"users"`
	Sessions []GameSession `json:"sessions"`
}

type AccountStore struct {
	mu       sync.Mutex
	path     string
	byUser   map[string]*GameAccount
	byEmail  map[string]*GameAccount
	byID     map[string]*GameAccount
	sessions map[string]*GameSession
}

func accountStorePath() string {
	if p := strings.TrimSpace(os.Getenv("CAHAYA_ACCOUNT_STORE")); p != "" {
		return p
	}
	return filepath.Join("data", "cahaya-accounts.json")
}

func OpenAccountStore(path string) *AccountStore {
	s := &AccountStore{
		path:     path,
		byUser:   map[string]*GameAccount{},
		byEmail:  map[string]*GameAccount{},
		byID:     map[string]*GameAccount{},
		sessions: map[string]*GameSession{},
	}
	s.load()
	return s
}

func (s *AccountStore) load() {
	raw, err := os.ReadFile(s.path)
	if err != nil || len(raw) == 0 {
		return
	}
	var blob accountBlob
	if json.Unmarshal(raw, &blob) != nil {
		return
	}
	now := time.Now()
	for i := range blob.Users {
		u := blob.Users[i]
		if u.PlayerID == "" || u.Username == "" {
			continue
		}
		cp := u
		s.index(&cp)
	}
	for i := range blob.Sessions {
		ss := blob.Sessions[i]
		if ss.Token == "" || now.After(ss.Expires) {
			continue
		}
		cp := ss
		s.sessions[cp.Token] = &cp
	}
}

func (s *AccountStore) index(u *GameAccount) {
	s.byUser[strings.ToLower(u.Username)] = u
	s.byEmail[strings.ToLower(u.Email)] = u
	s.byID[u.PlayerID] = u
}

func (s *AccountStore) flushLocked() error {
	blob := accountBlob{Users: make([]GameAccount, 0, len(s.byID)), Sessions: make([]GameSession, 0, len(s.sessions))}
	for _, u := range s.byID {
		if u != nil {
			blob.Users = append(blob.Users, *u)
		}
	}
	now := time.Now()
	for _, ss := range s.sessions {
		if ss != nil && now.Before(ss.Expires) {
			blob.Sessions = append(blob.Sessions, *ss)
		}
	}
	return writeAtomicJSON(s.path, blob)
}

func hashGamePassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), accountBcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func checkGamePassword(hash, plain string) bool {
	if hash == "" || plain == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func validateUsername(name string) string {
	name = strings.TrimSpace(name)
	if !usernameRE.MatchString(name) {
		return ""
	}
	return name
}

func validateEmail(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return ""
	}
	if !strings.Contains(email, ".") {
		return ""
	}
	return email
}

func validatePassword(pw string) string {
	if len(pw) < minPasswordLen {
		return "password terlalu pendek"
	}
	hasLetter, hasDigit := false, false
	for _, r := range pw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "password harus berisi huruf dan angka"
	}
	return ""
}

func newSessionToken() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return "tok_" + hex.EncodeToString(b[:])
}

func (s *AccountStore) Register(username, email, password, confirm string) (*GameSession, string) {
	username = validateUsername(username)
	email = validateEmail(email)
	if username == "" {
		return nil, "username tidak valid"
	}
	if email == "" {
		return nil, "email tidak valid"
	}
	if password != confirm {
		return nil, "konfirmasi password tidak sama"
	}
	if msg := validatePassword(password); msg != "" {
		return nil, msg
	}
	hash, err := hashGamePassword(password)
	if err != nil {
		return nil, "gagal mengamankan password"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byUser[strings.ToLower(username)] != nil {
		return nil, "username sudah dipakai"
	}
	if s.byEmail[email] != nil {
		return nil, "email sudah dipakai"
	}
	acc := &GameAccount{
		PlayerID:     randomID("p_"),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UnixMilli(),
	}
	s.index(acc)
	sess := s.issueLocked(acc)
	if err := s.flushLocked(); err != nil {
		return nil, "gagal menyimpan akun"
	}
	return sess, ""
}

func (s *AccountStore) Login(userOrEmail, password string) (*GameSession, string) {
	userOrEmail = strings.TrimSpace(userOrEmail)
	if userOrEmail == "" || password == "" {
		return nil, "akun atau password kosong"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.byUser[strings.ToLower(userOrEmail)]
	if acc == nil {
		acc = s.byEmail[strings.ToLower(userOrEmail)]
	}
	if acc == nil || !checkGamePassword(acc.PasswordHash, password) {
		return nil, "akun atau password salah"
	}
	sess := s.issueLocked(acc)
	_ = s.flushLocked()
	return sess, ""
}

func (s *AccountStore) ResetPassword(username, email, password, confirm string) string {
	username = validateUsername(username)
	email = validateEmail(email)
	if username == "" || email == "" {
		return "username dan email wajib"
	}
	if password != confirm {
		return "konfirmasi password tidak sama"
	}
	if msg := validatePassword(password); msg != "" {
		return msg
	}
	hash, err := hashGamePassword(password)
	if err != nil {
		return "gagal mengamankan password"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.byUser[strings.ToLower(username)]
	if acc == nil || strings.ToLower(acc.Email) != email {
		return "akun tidak ditemukan"
	}
	acc.PasswordHash = hash
	for tok, ss := range s.sessions {
		if ss != nil && ss.PlayerID == acc.PlayerID {
			delete(s.sessions, tok)
		}
	}
	_ = s.flushLocked()
	return ""
}

func (s *AccountStore) issueLocked(acc *GameAccount) *GameSession {
	sess := &GameSession{
		Token:    newSessionToken(),
		PlayerID: acc.PlayerID,
		Username: acc.Username,
		Expires:  time.Now().Add(sessionTTL),
	}
	s.sessions[sess.Token] = sess
	return sess
}

func (s *AccountStore) Lookup(token string) *GameSession {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.sessions[token]
	if ss == nil || time.Now().After(ss.Expires) {
		if ss != nil {
			delete(s.sessions, token)
		}
		return nil
	}
	return ss
}

func (s *AccountStore) Logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, strings.TrimSpace(token))
	_ = s.flushLocked()
}

func (s *AccountStore) AccountByID(playerID string) *GameAccount {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.byID[playerID]
	if acc == nil {
		return nil
	}
	cp := *acc
	cp.PasswordHash = ""
	return &cp
}

func (s *AccountStore) AuthenticateWS(in AuthIn) (*GameSession, string) {
	if sess := s.Lookup(in.Token); sess != nil {
		return sess, ""
	}
	user := strings.TrimSpace(in.Username)
	if user == "" {
		user = strings.TrimSpace(in.Name)
	}
	if user != "" && strings.TrimSpace(in.Password) != "" {
		return s.Login(user, in.Password)
	}
	return nil, "login diperlukan"
}
