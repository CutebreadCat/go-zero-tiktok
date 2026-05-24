package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const testSecret = "test-secret-key"

func TestGenerateAndParseToken(t *testing.T) {
	token, err := GenerateToken(testSecret, "u1", AccessTokenType, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != "u1" {
		t.Errorf("UserID = %s, want u1", claims.UserID)
	}
	if claims.TokenType != AccessTokenType {
		t.Errorf("TokenType = %s, want %s", claims.TokenType, AccessTokenType)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken(testSecret, "u1", AccessTokenType, time.Hour)
	_, err := ParseToken("wrong-secret", token)
	if err == nil {
		t.Error("expected error with wrong secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	token, _ := GenerateToken(testSecret, "u1", AccessTokenType, -time.Hour)
	_, err := ParseToken(testSecret, token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	claims := JwtClaims{
		Claims: Claims{UserID: "u1", TokenType: AccessTokenType},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := ParseToken(testSecret, tokenString)
	if err == nil {
		t.Error("expected error for wrong signing method")
	}
}

func TestGenerateAccessToken_RefreshTokenType(t *testing.T) {
	tests := []struct {
		name   string
		genFn  func(string, string) (string, error)
		expect string
	}{
		{"access", GenerateAccessToken, AccessTokenType},
		{"refresh", GenerateRefreshToken, RefreshTokenType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.genFn(testSecret, "u1")
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			claims, err := ParseToken(testSecret, token)
			if err != nil {
				t.Fatalf("ParseToken: %v", err)
			}
			if claims.TokenType != tt.expect {
				t.Errorf("TokenType = %s, want %s", claims.TokenType, tt.expect)
			}
		})
	}
}
