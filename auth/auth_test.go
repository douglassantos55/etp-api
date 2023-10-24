package auth_test

import (
	"api/auth"
	"strings"
	"testing"
)

func TestAuth(t *testing.T) {
	t.Run("hash password", func(t *testing.T) {
		t.Parallel()

		hashed, err := auth.HashPassword("password")
		if err != nil {
			t.Fatalf("could not hash password: %s", err)
		}

		if hashed == "password" {
			t.Errorf("should hash password, got %s", hashed)
		}
	})

	t.Run("compare password", func(t *testing.T) {
		t.Parallel()

		hash := "$2a$10$OBo6gtRDtR2g8X6S9Qn/Z.1r33jf6QYRSxavEIjG8UfrJ8MLQWRzy"
		err := auth.ComparePassword(hash, "password")

		if err != nil {
			t.Errorf("error comparing passwords: %s", err)
		}
	})

	t.Run("generate token", func(t *testing.T) {
		t.Parallel()

		token, err := auth.GenerateToken(1, "secret")
		if err != nil {
			t.Fatalf("could not generate token: %s", err)
		}

		periods := strings.Count(token, ".")
		if periods != 2 {
			t.Errorf("expected valid token, got %s", token)
		}
	})
}
