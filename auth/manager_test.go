package auth

import (
	"testing"
	"time"
)

const jwt = "eyJraWQiOiJzTjVnTk43cWFGVmpPYVwvc3oyVTJvdnNIMTZyNThQb2RQVFpRZWlBQUhNZz0iLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJlMjY1YmNhMS03NGZiLTQ2MjQtOWVlMC0zOGE5ZmQ5YTQ4OTUiLCJldmVudF9pZCI6Ijk4MjUxMzQ3LTFmN2ItNGY0OC1hNTNhLWE3MjZhOWEzOTFiNiIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoiYXdzLmNvZ25pdG8uc2lnbmluLnVzZXIuYWRtaW4iLCJhdXRoX3RpbWUiOjE2MjM0MDYwODQsImlzcyI6Imh0dHBzOlwvXC9jb2duaXRvLWlkcC5ldS1jZW50cmFsLTEuYW1hem9uYXdzLmNvbVwvZXUtY2VudHJhbC0xX1ZhSGgwRW9YNSIsImV4cCI6MTYyMzQyNzY4NCwiaWF0IjoxNjIzNDA2MDg0LCJqdGkiOiIxOWUwMTZmOS05NGU0LTQ1ZTYtYmE4OS0xYjg4Y2ZmMThmN2QiLCJjbGllbnRfaWQiOiI0aW8zYWVxZjFoOTY3dWZhbGs1bjc0MmNmaiIsInVzZXJuYW1lIjoibXdleWEifQ.I5Nn5selXHZVPIJFJ0HSAoqUtQxz9s37e-6YLua2S0M9xcNhx0h-Yr6A3S8JNH4OXsCKYK3r-Y9TOjce4CGRQRTprDIKQnwS3RMOz1jTT6YVXy5n8CEvN2ySC8y-27oLLOjAHLGQjRaj_o8PkrQVSzSxKJbthTS0fNodPTsBt6FpXEZv5ULIJzFnWVxjPD3Rb0h6Ts-iYylvcurWiCKZYQLItuydoCG_99uuIlL1BzDHxBz-QRJYRF_gFJlQbUBfsXHv3VT_ET5vTkXeNILqdq_ON2PrcbXR1uDXm2wvF4TKZsD3tFRFEeH6FK6xM0dnMnogphEowZyU_BIEQlgTsw"

func TestManagerStoreAndRetrieveTokens(t *testing.T) {
	manager := NewManager()
	ttoken := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	manager.storeTokens(&ttoken)
	token, err := manager.getTokens()
	if (err != nil) {
		t.Errorf("Error retrieving token: %s", err.Error())
	}
	if (token.AccessToken != ttoken.AccessToken) {
		t.Errorf("AccessToken not as expected. Got %s, want %s", token.AccessToken, ttoken.AccessToken)
	}
	if (token.IDToken != ttoken.IDToken) {
		t.Errorf("IDToken not as expected. Got %s, want %s", token.IDToken, ttoken.IDToken)
	}
	if (token.RefreshToken != ttoken.RefreshToken) {
		t.Errorf("RefreshToken not as expected. Got %s, want %s", token.RefreshToken, ttoken.RefreshToken)
	}
	if (token.Expires != ttoken.Expires) {
		t.Errorf("Expires not as expected. Want %d, got %d", ttoken.Expires, token.Expires)
	}
	if (token.DetaAccessToken != ttoken.DetaAccessToken) {
		t.Errorf("DetaAccessToken not as expected. Want %s, got %s", ttoken.DetaAccessToken, token.DetaAccessToken)
	}
	if !t.Failed() {
		t.Logf("Successfully stored and read from token")
	}
}

func TestManagerGetExpiry(t *testing.T) {
	manager := NewManager()
	tests := []struct {
		name string
		expectedError error
		expectedOutput int64
		token string 
	} {
		{"invalid_token", ErrAccessTokenFormatInvalid, 0, "aaaaaaaaaaaaa"},
		{"valid_token", nil, 1623427684, jwt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := manager.expiresFromToken(tt.token)
			if (tt.expectedError != err) {
				t.Errorf("Unexpected error: Want %s, Got %s", tt.expectedError.Error(), err.Error())
			}
			if (out != tt.expectedOutput) {
				t.Errorf("Mismatched output. Expected %d, got %d", tt.expectedOutput, out)
			}
			if (!t.Failed()) {
				t.Logf("No fails in getExpiry!")
			}
		})
	}
}

func TestManagerStoreTokenAndGetExpiry(t *testing.T) {
	manager := NewManager()
	token := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	manager.storeTokens(&token)
	exp, err := manager.expiresFromToken(jwt)
	if (err != nil) {
		t.Errorf("Error getting expiry: %s", err.Error())
	}
	if (exp != 1623427684) {
		t.Errorf("Got wrong expiry, want %d, got %d", 1623427684, exp)
	}
}

func TestManagerIsTokenExpired(t *testing.T) {
	manager := NewManager()
	expiredToken := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	validToken := Token{
		AccessToken:     "",
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         time.Now().Unix()+3000,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	var tokentests = []struct {
		name string
		token Token
		valid bool
	} {
		{"token_valid",validToken, true},
		{"token_expired",expiredToken, false},
	}

	for _, tt := range tokentests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid != manager.isTokenExpired(&tt.token) {
				t.Logf("%s passed", tt.name)
			} else {
				t.Errorf("%s failed. Want %t, got %t", tt.name, tt.valid, manager.isTokenExpired(&tt.token))
			}
		})
	}
}
