package wallet

import (
	"testing"
)

func TestNew(t *testing.T) {
	w := New(100.0)
	bal, history := w.Debit(0, Entry{})
	if bal != 100.0 {
		t.Errorf("expected initial balance 100.0, got %f", bal)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 entry after debit, got %d", len(history))
	}
}

func TestDebit(t *testing.T) {
	w := New(100.0)
	bal, history := w.Debit(10.5, Entry{UserMsg: "hello"})
	if bal != 89.5 {
		t.Errorf("expected balance 89.5, got %f", bal)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
	if history[0].UserMsg != "hello" {
		t.Errorf("expected history entry UserMsg 'hello', got '%s'", history[0].UserMsg)
	}
}

func TestDebitMultiple(t *testing.T) {
	w := New(100.0)
	w.Debit(10.0, Entry{})
	w.Debit(20.0, Entry{})
	bal, history := w.Debit(5.0, Entry{})
	if bal != 65.0 {
		t.Errorf("expected balance 65.0, got %f", bal)
	}
	if len(history) != 3 {
		t.Errorf("expected 3 history entries, got %d", len(history))
	}
}

func TestReset(t *testing.T) {
	w := New(100.0)
	w.Debit(50.0, Entry{})
	w.Reset(100.0)
	bal, history := w.Debit(0, Entry{})
	if bal != 100.0 {
		t.Errorf("expected balance 100.0 after reset, got %f", bal)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 entry after reset + debit, got %d", len(history))
	}
}

func TestDebitReturnsCopy(t *testing.T) {
	w := New(100.0)
	_, history := w.Debit(1.0, Entry{UserMsg: "a"})
	history[0].UserMsg = "mutated"
	_, history2 := w.Debit(1.0, Entry{UserMsg: "b"})
	if history2[0].UserMsg == "mutated" {
		t.Error("Debit should return a copy, not a reference to internal history")
	}
}
