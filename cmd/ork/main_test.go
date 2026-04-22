package main

import (
	"os"
	"strings"
	"testing"
)

func TestReadPasswordFromStdin_Empty(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.WriteString("\n")
		w.Close()
	}()

	_, err = readPasswordFromStdin()
	if err == nil {
		t.Fatal("Expected error for empty password, got nil")
	}
	if !strings.Contains(err.Error(), "password cannot be empty") {
		t.Errorf("Expected 'password cannot be empty' error, got: %v", err)
	}
}

func TestReadPasswordFromStdin_EOF(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.Close()
	}()

	_, err = readPasswordFromStdin()
	if err == nil {
		t.Fatal("Expected error for EOF, got nil")
	}
	if !strings.Contains(err.Error(), "no password provided") {
		t.Errorf("Expected 'no password provided' error, got: %v", err)
	}
}

func TestReadPasswordFromStdin_Valid(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.WriteString("secret123\n")
		w.Close()
	}()

	password, err := readPasswordFromStdin()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if password != "secret123" {
		t.Errorf("Expected 'secret123', got '%s'", password)
	}
}

func TestPromptPassword_Empty(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.WriteString("\n")
		w.Close()
	}()

	_, err = promptPassword("test: ")
	if err == nil {
		t.Fatal("Expected error for empty password, got nil")
	}
	if !strings.Contains(err.Error(), "password cannot be empty") {
		t.Errorf("Expected 'password cannot be empty' error, got: %v", err)
	}
}

func TestPromptPassword_EOF(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.Close()
	}()

	_, err = promptPassword("test: ")
	if err == nil {
		t.Fatal("Expected error for EOF, got nil")
	}
	if !strings.Contains(err.Error(), "no password provided") {
		t.Errorf("Expected 'no password provided' error, got: %v", err)
	}
}

func TestPromptPassword_Valid(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.WriteString("secret123\n")
		w.Close()
	}()

	password, err := promptPassword("test: ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if password != "secret123" {
		t.Errorf("Expected 'secret123', got '%s'", password)
	}
}
