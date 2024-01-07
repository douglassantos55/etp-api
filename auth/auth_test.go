package auth_test

import (
	"api/auth"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
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

	t.Run("ParseToken", func(t *testing.T) {
		t.Run("company token", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
				Subject:  "1535",
				Audience: jwt.ClaimStrings{"entrepreneur-client"},
			})

			id, err := auth.ParseToken(token)
			if err != nil {
				t.Fatalf("could not parse token: %s", err)
			}

			if id != 1535 {
				t.Errorf("expected id %d, got %d", 1535, id)
			}
		})

		t.Run("cron token", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
				Audience: jwt.ClaimStrings{"cronjob"},
			})

			id, err := auth.ParseToken(token)
			if err != nil {
				t.Fatalf("could not parse token: %s", err)
			}

			if id != 0 {
				t.Errorf("expected id %d, got %d", 0, id)
			}
		})

		t.Run("invalid token", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
				Subject:  "cron",
				Audience: jwt.ClaimStrings{"entrepreneur-client"},
			})

			_, err := auth.ParseToken(token)
			if err == nil {
				t.Fatal("should not parse token")
			}
		})

		t.Run("invalid cron token", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
				Subject:  "120",
				Audience: jwt.ClaimStrings{"cronjob"},
			})

			id, err := auth.ParseToken(token)
			if err != nil {
				t.Fatalf("could not parse token: %s", err)
			}

			if id != 0 {
				t.Errorf("expected id %d, got %d", 0, id)
			}
		})
	})
}
