package auth

import (
	"errors"
	"testing"
)

const detaExampleAccessToken = "aaaaaaa_xxxxxxxxxxxxxxxx"
const detaExampleSignature = "v0=aaaaaaa:9369e755d35ea272895cab8546d745107f7b21e6e87e1324e1f85627c04934dd"

func TestCalcSignature_EarlyReturnNotV0(t *testing.T) {
	manager := NewManager()
	exampleToken := CalcSignatureInput{
		AccessToken: jwt,
		HTTPMethod:  "HTTPMethod",
		URI:         "localhost",
		Timestamp:   "today",
		ContentType: "application/json",
		RawBody:     []byte("<h1>Hurr durr</h1>"),
	}
	out, err := manager.CalcSignature(&exampleToken)
	if err != nil {
		t.Errorf("Error calculating signature: %s", err.Error())
	}
	if out == "" {
		t.Logf("Exited early")
	}
}

func TestCalcSignature_InvalidToken(t *testing.T) {
	manager := NewManager()
	detaSignVersion = "v0"
	exampleToken := CalcSignatureInput{
		AccessToken: "",
	}
	out, err := manager.CalcSignature(&exampleToken)
	if err == nil {
		t.Errorf("Failed to throw exception on invalid token")
	} else {
		if errors.Is(err, ErrInvalidAccessToken) {
			if (out == "") {
				t.Logf("Failed as expected!")
			} else {
				t.Errorf("Unexpected output: %s", out)
			}
		} else {
			t.Errorf("Error calculating signature: %s", err.Error())
		}
	}
}

func TestCalcSignature_InvalidVersion(t *testing.T) {
	manager := NewManager()
	detaSignVersion = "v1"
	out, err := manager.CalcSignature(&CalcSignatureInput{})
	if (err != nil) {
		t.Errorf("Error thrown when trying to calculate signature: %s", err.Error())
	}
	if (out != "") {
		t.Errorf("Signature output not empty. Want: %s, got: %s", "", out)
	}
	t.Logf("Invalid version found; exited early")
}

func TestCalcSignature_SigIsCorrect(t *testing.T) {
	manager := NewManager()
	detaSignVersion = "v0"
	exampleToken := CalcSignatureInput{
		AccessToken: detaExampleAccessToken,
		HTTPMethod:  "HTTPMethod",
		URI:         "localhost",
		Timestamp:   "today",
		ContentType: "application/json",
		RawBody:     []byte("<h1>Hurr durr</h1>"),
	}
	out, err := manager.CalcSignature(&exampleToken)
	if (err != nil) {
		t.Errorf("Error calculating signature: %s", err.Error())
	}
	if (out == "") {
		t.Errorf("Signature is empty, expected a signature")
	} else {
		if (out != detaExampleSignature) {
			t.Errorf("Incorrect signature calculated. Got %s, want %s", out, detaExampleSignature)
		} else {
			t.Logf("Correct signature generated")
		}
	}
}


