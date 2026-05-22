package errs

import "testing"

func TestMarkMatchesCategoryAndOriginal(t *testing.T) {
	original := New("todo not found")
	err := Mark(original, ErrNotFound)

	if !Is(err, ErrNotFound) {
		t.Fatalf("Is(%v, %v) = false, want true", err, ErrNotFound)
	}
	if !Is(err, original) {
		t.Fatalf("Is(%v, %v) = false, want true", err, original)
	}
}

func TestValidationErrorsMatchInvalidInput(t *testing.T) {
	err := NewValidationErrors([]ValidationError{
		{Field: "title", Message: "is required"},
	})

	if !IsInvalidInput(err) {
		t.Fatalf("IsInvalidInput(%v) = false, want true", err)
	}
}
