package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dracory/ork/vault"
)

func TestApiResponse_String(t *testing.T) {
	r := apiResponse{
		Status:  "success",
		Message: "test message",
	}

	result := r.String()

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", parsed["status"])
	}
	if parsed["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", parsed["message"])
	}
}

func TestApiSuccess(t *testing.T) {
	result := apiSuccess("operation completed")

	if result.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
	if result.Message != "operation completed" {
		t.Errorf("Expected message 'operation completed', got '%s'", result.Message)
	}
}

func TestApiSuccessWithData(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := apiSuccessWithData("data loaded", data)

	if result.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
	if result.Message != "data loaded" {
		t.Errorf("Expected message 'data loaded', got '%s'", result.Message)
	}
	if result.Data == nil {
		t.Error("Expected data to be non-nil")
	}
}

func TestApiError(t *testing.T) {
	result := apiError("something went wrong")

	if result.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", result.Status)
	}
	if result.Message != "something went wrong" {
		t.Errorf("Expected message 'something went wrong', got '%s'", result.Message)
	}
}

func TestVaultUI_getCSS(t *testing.T) {
	ui := &vaultUI{vaultPath: "test.vault"}
	css := ui.getCSS()

	if css == "" {
		t.Error("Expected non-empty CSS")
	}

	// Check for Notiflix CSS link
	if !strings.Contains(css, "notiflix") {
		t.Error("Expected CSS to contain notiflix reference")
	}

	// Check for expected style rules
	expectedStyles := []string{
		".card {",
		".btn-primary {",
		".modal {",
		".login-container {",
	}

	for _, style := range expectedStyles {
		if !strings.Contains(css, style) {
			t.Errorf("Expected CSS to contain '%s'", style)
		}
	}
}

func TestVaultUI_getJS(t *testing.T) {
	ui := &vaultUI{vaultPath: "test.vault"}
	js := ui.getJS("\"test.vault\"")

	if js == "" {
		t.Error("Expected non-empty JS")
	}

	// Check for Vue.js usage
	if !strings.Contains(js, "Vue.createApp") {
		t.Error("Expected JS to contain Vue.createApp")
	}

	// Check for Notiflix usage
	if !strings.Contains(js, "Notiflix.Notify") {
		t.Error("Expected JS to contain Notiflix.Notify")
	}

	// Check for vault path
	if !strings.Contains(js, "test.vault") {
		t.Error("Expected JS to contain vault path")
	}

	// Check for expected methods
	expectedMethods := []string{
		"login()",
		"keysList()",
		"keyAdd()",
		"keyUpdate()",
		"keyRemove()",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(js, method) {
			t.Errorf("Expected JS to contain method '%s'", method)
		}
	}
}

func TestVaultUI_getHTML(t *testing.T) {
	ui := &vaultUI{vaultPath: "test.vault"}
	html := ui.getHTML()

	if html == "" {
		t.Error("Expected non-empty HTML")
	}

	// Check for DOCTYPE
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}

	// Check for Vue.js CDN
	if !strings.Contains(html, "vue@3") {
		t.Error("Expected HTML to contain Vue.js CDN")
	}

	// Check for Notiflix CDN
	if !strings.Contains(html, "notiflix") {
		t.Error("Expected HTML to contain Notiflix CDN")
	}

	// Check for app mount point
	if !strings.Contains(html, `id="app"`) {
		t.Error("Expected HTML to contain app mount point")
	}

	// Check for login page
	if !strings.Contains(html, "pageLoginShow") {
		t.Error("Expected HTML to contain login page reference")
	}

	// Check for keys page
	if !strings.Contains(html, "pageKeysShow") {
		t.Error("Expected HTML to contain keys page reference")
	}

	// Check for modal references
	expectedModals := []string{
		"keyAddModalVisible",
		"keyUpdateModalVisible",
		"keyRemoveModalVisible",
	}

	for _, modal := range expectedModals {
		if !strings.Contains(html, modal) {
			t.Errorf("Expected HTML to contain modal '%s'", modal)
		}
	}
}

func TestApiResponse_EmptyData(t *testing.T) {
	r := apiResponse{
		Status:  "success",
		Message: "test",
	}

	result := r.String()

	// Empty data should not be included due to omitempty
	if strings.Contains(result, "data") {
		t.Error("Expected empty data field to be omitted from JSON")
	}
}

func TestApiResponse_String_MarshalError(t *testing.T) {
	// Create a response with unmarshalable data (channel cannot be JSON-encoded)
	r := apiResponse{
		Status:  "success",
		Message: "test",
		Data:    make(chan int),
	}

	result := r.String()
	if !strings.Contains(result, "internal server error") {
		t.Errorf("Expected fallback error message, got: %s", result)
	}
}

func TestHandleRequest_PostOnly(t *testing.T) {
	ui := &vaultUI{vaultPath: "test.vault"}

	actions := []string{"login", "keys", "key-add", "key-update", "key-remove"}
	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/?a="+action, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			ui.handleRequest(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
			}
			body := rr.Body.String()
			if !strings.Contains(body, "Method not allowed") {
				t.Errorf("Expected 'Method not allowed' in body, got: %s", body)
			}
		})
	}
}

func TestHandleLogin_Success(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create vault with data
	v, err := vault.Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := v.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	ui := &vaultUI{vaultPath: vaultPath}

	form := url.Values{}
	form.Set("password", password)
	req, err := http.NewRequest("POST", "/?a=login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	ui.handleRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Login successful") {
		t.Errorf("Expected successful login, got: %s", body)
	}
	if !strings.Contains(body, "value1") {
		t.Errorf("Expected response to contain vault data, got: %s", body)
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create vault and save it
	v, err := vault.Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := v.KeySet("KEY1", "value1"); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	ui := &vaultUI{vaultPath: vaultPath}

	form := url.Values{}
	form.Set("password", "wrong-password")
	req, err := http.NewRequest("POST", "/?a=login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	ui.handleRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid password") {
		t.Errorf("Expected 'Invalid password' error, got: %s", body)
	}
}

func TestHandleLogin_MissingPassword(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	password := "TestPass123"

	// Create vault
	v, err := vault.Create(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("Failed to close vault: %v", err)
	}

	ui := &vaultUI{vaultPath: vaultPath}

	form := url.Values{}
	req, err := http.NewRequest("POST", "/?a=login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	ui.handleRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Password is required") {
		t.Errorf("Expected 'Password is required' error, got: %s", body)
	}
}
