package main

import "testing"

func TestMigratePortalOperatorsFromLegacy(t *testing.T) {
	cfg := PortalAuthConfig{
		AdminUsername:    "admin",
		OperatorUsername: "op1",
		OperatorPassword: "secret",
		OperatorName:     "Operator Satu",
	}
	if !migratePortalOperators(&cfg) {
		t.Fatal("expected migration to change config")
	}
	if len(cfg.Operators) != 1 {
		t.Fatalf("expected 1 operator, got %d", len(cfg.Operators))
	}
	if cfg.Operators[0].Username != "op1" || !cfg.Operators[0].Enabled {
		t.Fatalf("unexpected operator: %+v", cfg.Operators[0])
	}
	if cfg.Operators[0].ID == "" {
		t.Fatal("operator id empty")
	}
}

func TestFindOperatorByUsername(t *testing.T) {
	cfg := PortalAuthConfig{
		Operators: []PortalOperator{
			{ID: "a", Username: "OpUser", Password: "hash", Name: "A", Enabled: true},
			{ID: "b", Username: "disabled", Password: "hash", Name: "B", Enabled: false},
		},
	}
	op, ok := findOperatorByUsername(cfg, "opuser")
	if !ok || op.ID != "a" {
		t.Fatalf("expected enabled operator a, got %+v ok=%v", op, ok)
	}
	if _, ok := findOperatorByUsername(cfg, "disabled"); ok {
		t.Fatal("disabled operator should not authenticate")
	}
}

func TestOperatorUsernameTaken(t *testing.T) {
	cfg := PortalAuthConfig{
		AdminUsername: "admin",
		Operators: []PortalOperator{
			{ID: "1", Username: "op1", Enabled: true},
		},
	}
	if !operatorUsernameTaken(cfg, "admin", "") {
		t.Fatal("admin username should be taken")
	}
	if !operatorUsernameTaken(cfg, "OP1", "") {
		t.Fatal("existing operator username should be taken")
	}
	if operatorUsernameTaken(cfg, "OP1", "1") {
		t.Fatal("same operator id should be excluded")
	}
}
