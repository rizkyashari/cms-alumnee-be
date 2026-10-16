package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSanitizeProfilePhotoAcceptsAndReencodesPNG(t *testing.T) {
	var input bytes.Buffer
	fixture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	fixture.Set(0, 0, color.RGBA{R: 220, G: 90, B: 40, A: 255})
	if err := png.Encode(&input, fixture); err != nil {
		t.Fatalf("encode fixture PNG: %v", err)
	}
	result, contentType, err := sanitizeProfilePhoto(input.Bytes())
	if err != nil {
		t.Fatalf("sanitizeProfilePhoto() error = %v", err)
	}
	if contentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", contentType)
	}
	if _, format, err := image.Decode(bytes.NewReader(result)); err != nil || format != "png" {
		t.Fatalf("sanitized image format = %q, error = %v", format, err)
	}
}

func TestSanitizeProfilePhotoRejectsSVGAndOversizedImage(t *testing.T) {
	if _, _, err := sanitizeProfilePhoto([]byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)); err == nil {
		t.Fatal("sanitizeProfilePhoto() accepted SVG")
	}
	if _, _, err := sanitizeProfilePhoto(bytes.Repeat([]byte{1}, maxProfilePhotoBytes+1)); err == nil {
		t.Fatal("sanitizeProfilePhoto() accepted an oversized upload")
	}
}

func TestValidateRegistrationProfileDomesticAndInternational(t *testing.T) {
	base := registrationProfile{
		fullName: "Alumni Example", batchYear: 2012, countryCode: "ID",
		cityCode: "34.04", city: "Kabupaten Sleman", photo: []byte{1},
	}
	if err := validateRegistrationProfile(base, 2026); err != nil {
		t.Fatalf("valid domestic profile rejected: %v", err)
	}
	base.countryCode = "US"
	base.cityCode = ""
	base.city = "Seattle"
	if err := validateRegistrationProfile(base, 2026); err != nil {
		t.Fatalf("valid international profile rejected: %v", err)
	}
	base.countryCode = "ID"
	base.cityCode = ""
	if err := validateRegistrationProfile(base, 2026); err == nil {
		t.Fatal("domestic profile without a regency code was accepted")
	}
}

func TestProfileCompletionWeightsRequiredFieldsAndSecondContact(t *testing.T) {
	if got := calculateProfileCompletion("Alumni Example", 2012, "Sleman", true, true, false, false); got != 90 {
		t.Fatalf("email registration completion = %d, want 90", got)
	}
	if got := calculateProfileCompletion("Alumni Example", 2012, "Sleman", true, true, true, false); got != 100 {
		t.Fatalf("email and phone completion = %d, want 100", got)
	}
	if got := calculateProfileCompletion("", 0, "", false, false, false, false); got != 0 {
		t.Fatalf("empty profile completion = %d, want 0", got)
	}
	if got := calculateProfileCompletion("Alumni Example", 2012, "Sleman", true, false, false, true); got != 90 {
		t.Fatalf("OAuth profile completion = %d, want 90", got)
	}
}

func TestEmailOTPHashUsesSecretAndChallenge(t *testing.T) {
	service := NewService(nil, false, "")
	service.ConfigureIntegrations(Integrations{ResendAPIKey: "test-key"})
	first := service.emailOTPHash([]byte("challenge-one"), "123456")
	if bytes.Equal(first, service.emailOTPHash([]byte("challenge-two"), "123456")) {
		t.Fatal("OTP hash did not include challenge ID")
	}
	service.integrations.ResendAPIKey = "another-key"
	if bytes.Equal(first, service.emailOTPHash([]byte("challenge-one"), "123456")) {
		t.Fatal("OTP hash did not include configured secret")
	}
}

func TestSendEmailOTPUsesConfiguredResendSender(t *testing.T) {
	service := NewService(nil, false, "")
	service.ConfigureIntegrations(Integrations{
		ResendAPIKey: "re_test_secret", ResendFromEmail: "AlumniHub <otp@example.test>",
	})
	var gotAuthorization string
	var gotBody struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
	}
	service.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host != "api.resend.com" || request.URL.Path != "/emails" {
			t.Fatalf("unexpected email endpoint: %s", request.URL)
		}
		gotAuthorization = request.Header.Get("Authorization")
		if err := json.NewDecoder(request.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode email payload: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id":"email-id"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	if err := service.sendEmailOTP(context.Background(), "alumni@example.test", "123456"); err != nil {
		t.Fatalf("sendEmailOTP() error = %v", err)
	}
	if gotAuthorization != "Bearer re_test_secret" {
		t.Fatalf("authorization = %q, want configured bearer token", gotAuthorization)
	}
	if gotBody.From != "AlumniHub <otp@example.test>" || len(gotBody.To) != 1 || gotBody.To[0] != "alumni@example.test" {
		t.Fatalf("unexpected email recipients or sender: %+v", gotBody)
	}
	if !strings.Contains(gotBody.Text, "123456") || gotBody.Subject == "" {
		t.Fatalf("email payload does not contain the OTP and subject: %+v", gotBody)
	}
}

func TestSendEmailOTPRejectsProviderFailure(t *testing.T) {
	service := NewService(nil, false, "")
	service.ConfigureIntegrations(Integrations{ResendAPIKey: "test-key", ResendFromEmail: "otp@example.test"})
	service.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader("provider details")),
			Header:     make(http.Header),
		}, nil
	})}
	if err := service.sendEmailOTP(context.Background(), "alumni@example.test", "123456"); err == nil {
		t.Fatal("sendEmailOTP() ignored provider failure")
	}
}

func TestEmailOTPChallengeValidityIsBounded(t *testing.T) {
	if emailOTPValidity <= 0 || emailOTPValidity > 15*time.Minute {
		t.Fatalf("email OTP validity = %s, expected a short bounded lifetime", emailOTPValidity)
	}
	if registrationTokenValidity <= emailOTPValidity || registrationTokenValidity > time.Hour {
		t.Fatalf("registration token validity = %s, expected a short bounded lifetime", registrationTokenValidity)
	}
}
