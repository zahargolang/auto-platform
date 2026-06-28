package auth_service

import "testing"

func TestHashPassword_VerifyPassword(t *testing.T) {
	hash, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	if hash == "" || hash == validPassword {
		t.Fatalf("HashPassword() returned suspicious hash: %q", hash)
	}

	if err := VerifyPassword(hash, validPassword); err != nil {
		t.Fatalf("VerifyPassword() unexpected error for correct password: %v", err)
	}

	if err := VerifyPassword(hash, "WrongPassword1!"); err == nil {
		t.Fatalf("VerifyPassword() expected error for wrong password, got nil")
	}
}

func TestHashPassword_DistinctHashesForSamePassword(t *testing.T) {
	hash1, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	hash2, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}

	// bcrypt подсаливает каждый вызов случайно — одинаковый пароль должен
	// давать разные хеши.
	if hash1 == hash2 {
		t.Fatalf("HashPassword() produced identical hashes for two calls: %q", hash1)
	}
}
