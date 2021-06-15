package auth

import (
	"net/http"
	"testing"
	"time"
)

const jwt = "eyJraWQiOiJzTjVnTk43cWFGVmpPYVwvc3oyVTJvdnNIMTZyNThQb2RQVFpRZWlBQUhNZz0iLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJlMjY1YmNhMS03NGZiLTQ2MjQtOWVlMC0zOGE5ZmQ5YTQ4OTUiLCJldmVudF9pZCI6Ijk4MjUxMzQ3LTFmN2ItNGY0OC1hNTNhLWE3MjZhOWEzOTFiNiIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoiYXdzLmNvZ25pdG8uc2lnbmluLnVzZXIuYWRtaW4iLCJhdXRoX3RpbWUiOjE2MjM0MDYwODQsImlzcyI6Imh0dHBzOlwvXC9jb2duaXRvLWlkcC5ldS1jZW50cmFsLTEuYW1hem9uYXdzLmNvbVwvZXUtY2VudHJhbC0xX1ZhSGgwRW9YNSIsImV4cCI6MTYyMzQyNzY4NCwiaWF0IjoxNjIzNDA2MDg0LCJqdGkiOiIxOWUwMTZmOS05NGU0LTQ1ZTYtYmE4OS0xYjg4Y2ZmMThmN2QiLCJjbGllbnRfaWQiOiI0aW8zYWVxZjFoOTY3dWZhbGs1bjc0MmNmaiIsInVzZXJuYW1lIjoibXdleWEifQ.I5Nn5selXHZVPIJFJ0HSAoqUtQxz9s37e-6YLua2S0M9xcNhx0h-Yr6A3S8JNH4OXsCKYK3r-Y9TOjce4CGRQRTprDIKQnwS3RMOz1jTT6YVXy5n8CEvN2ySC8y-27oLLOjAHLGQjRaj_o8PkrQVSzSxKJbthTS0fNodPTsBt6FpXEZv5ULIJzFnWVxjPD3Rb0h6Ts-iYylvcurWiCKZYQLItuydoCG_99uuIlL1BzDHxBz-QRJYRF_gFJlQbUBfsXHv3VT_ET5vTkXeNILqdq_ON2PrcbXR1uDXm2wvF4TKZsD3tFRFEeH6FK6xM0dnMnogphEowZyU_BIEQlgTsw"

func TestCreateToken(t *testing.T) {
	_ = Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	t.Logf("Token created!")
}

func TestCreateAndWriteToken(t *testing.T) {
	token := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	token.AccessToken = "changed"
	token.IDToken = "changed"
	token.RefreshToken = "changed"
	token.Expires = 999
	token.DetaAccessToken = "changed"
	t.Logf("Fields modified!")
}

func TestCreateAndReadToken(t *testing.T) {
	token := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	if (token.AccessToken != jwt) {
		t.Errorf("token.AccessToken == %s, want: %s", token.AccessToken, jwt)
	}
	if (token.IDToken != "idToken") {
		t.Errorf("token.IDToken == %s, want: %s", token.IDToken, "idToken")
	}
	if (token.RefreshToken != "refreshToken") {
		t.Errorf("token.RefreshToken == %s, want: %s", token.RefreshToken, "refreshToken")
	}
	if (token.Expires != 0) {
		t.Errorf("token.Expires == %d, want: %d", token.Expires, 0)
	}
	if (token.DetaAccessToken != "aaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("token.DetaAccessToken == %s, want: %s", token.DetaAccessToken, "aaaaaaaaaaaaaaaaaaaaaaaaaa")
	}
	t.Logf("Info in token is as expected")
}

func TestCreateManager(t *testing.T) {
	_ = NewManager()
	t.Logf("Manager created!")
}

func TestCreateAndWriteToManager(t *testing.T) {
	manager := NewManager()
	manager.bearerAuth = false
	manager.srv = &http.Server{}
	manager.tokenChan = make(chan *Token)
	manager.errChan = make(chan error, 1)
	t.Logf("Manager fields modified!")
}

func TestManagerStoreTokens(t *testing.T) {
	manager := NewManager()
	token := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	manager.storeTokens(&token)
	t.Logf("Token stored by manager!")
}

func TestManagerGetToken(t *testing.T) {
	manager := NewManager()
	_, err := manager.getTokens()
	if (err != nil) {
		t.Errorf("Error retrieving token: %s", err.Error())
	}
	t.Logf("Token successfully retrieved")
}

func TestManagerGetTokenAndRead(t *testing.T) {
	manager := NewManager()
	token, err := manager.getTokens()
	if (err != nil) {
		t.Errorf("Error retrieving token: %s", err.Error())
	}
	if (token.AccessToken != jwt) {
		t.Errorf("token.AccessToken == %s, want: %s", token.AccessToken, jwt)
	}
	if (token.IDToken != "idToken") {
		t.Errorf("token.IDToken == %s, want: %s", token.IDToken, "idToken")
	}
	if (token.RefreshToken != "refreshToken") {
		t.Errorf("token.RefreshToken == %s, want: %s", token.RefreshToken, "refreshToken")
	}
	if (token.Expires != 1623427684) {
		t.Errorf("token.Expires == %d, want: %d", token.Expires, 1623427684)
	}
	if (token.DetaAccessToken != "aaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("token.DetaAccessToken == %s, want: %s", token.DetaAccessToken, "aaaaaaaaaaaaaaaaaaaaaaaaaa")
	}
	t.Logf("Token successfully retrieved and read")
}

func TestManagerFailGetExpiry(t *testing.T) {
	manager := NewManager()
	_, err :=manager.expiresFromToken("aaaaaaaaaaaaa")
	if (err == nil) {
		t.Errorf("Failed to cause error assigning invalid accessToken")
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

func TestManagerIsTokenExpiredTrue(t *testing.T) {
	manager := NewManager()
	token := Token{
		AccessToken:     jwt,
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         0,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	if (!manager.isTokenExpired(&token)) {
		t.Errorf("Token should have been expired")
	}
}

func TestManagerIsTokenExpiredFalse(t *testing.T) {
	manager := NewManager()
	token := Token{
		AccessToken:     "",
		IDToken:         "idToken",
		RefreshToken:    "refreshToken",
		Expires:         time.Now().Unix()+3000,
		DetaAccessToken: "aaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	if (manager.isTokenExpired(&token)) {
		t.Errorf("Token should not be expired")
	}
}

func TestExpiresFromToken(t * testing.T) {
	manager := NewManager()
	exp, err := manager.expiresFromToken(jwt)
	if (err != nil) {
		t.Errorf("Error getting expiry from token: %s", err.Error())
	}
	if (exp != 1623427684) {
		t.Errorf("Expiry from token incorrect. Want: %d, got: %d", 1623427684, exp)
	}
}

func TestGetTokens(t *testing.T) {
	manager := NewManager()
	token, err := manager.getTokens()
	if (err != nil) {
		t.Errorf("Error getting token from manager: %s", err.Error())
	}
	if (token.AccessToken != jwt) {
		t.Errorf("Incorrect access token: Want: %s, got: %s", jwt, token.AccessToken)
	}
	if (token.RefreshToken != "refreshToken") {
		t.Errorf("Incorrect refresh token. Want: %s, got: %s", "refreshToken", token.RefreshToken)
	}
	if (token.Expires != 1623427684) {
		t.Errorf("Incorrect expiry date. Want: %d, got: %d", 1623427684, token.Expires)
	}
	if (token.DetaAccessToken != "aaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("Incorrect access token. Got: %s, want: %s", token.AccessToken, "aaaaaaaaaaaaaaaaaaaaaaaaaa")
	}
}

func TestIsBearerAuth(t *testing.T) {
	manager := NewManager()
	if (!manager.bearerAuth) {
		t.Errorf("Incorrect bearer auth set. Got %t, Want: %t", manager.bearerAuth, true)
	}
}

func TestIsBearerAuthFalse(t *testing.T) {
	manager := Manager {
		bearerAuth: false,
	}
	if (manager.bearerAuth) {
		t.Errorf("Incorrect bearer auth set. Got %t, Want: %t", manager.bearerAuth, false)
	}
}
