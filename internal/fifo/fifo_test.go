package fifo

import (
	"fmt"
	"testing"
)

const fifoTestingSize = 8

func TestNewFIFO(t *testing.T) {
	for i := range fifoTestingSize + 1 {
		t.Run(fmt.Sprintf("size=%d", i), func(t *testing.T) {
			fifo := NewFIFO(uint(i))
			expectedBufferLength := i
			actualBufferLength := len(fifo.buf)
			if expectedBufferLength != actualBufferLength {
				t.Errorf("expected buffer length %d, got %d", expectedBufferLength, actualBufferLength)
			}
		})
	}
}

func TestFIFOSize(t *testing.T) {
	for i := range fifoTestingSize + 1 {
		t.Run(fmt.Sprintf("size=%d", i), func(t *testing.T) {
			fifo := NewFIFO(uint(i))
			expectedSize := i
			actualSize := fifo.Size()
			if expectedSize != int(actualSize) {
				t.Errorf("expected size %d, got %d", expectedSize, actualSize)
			}
		})
	}
}

func TestFIFOFullEmpty(t *testing.T) {
	t.Run("size=0", func(t *testing.T) {
		fifo0 := NewFIFO(0)
		if !fifo0.IsEmpty() {
			t.Errorf("expected fifo of size 0 to be empty")
		}
		if !fifo0.IsFull() {
			t.Errorf("expected fifo of size 0 to be full")
		}
	})
	for size := 1; size <= fifoTestingSize; size++ {
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			fifo := NewFIFO(uint(size))
			if !fifo.IsEmpty() {
				t.Errorf("expected new fifo to be empty")
			}
			if fifo.IsFull() {
				t.Errorf("expected new fifo to not be full")
			}
			for i := range size {
				if err := fifo.Write(0); err != nil {
					t.Fatalf("unexpected error writing to fifo: %v", err)
				}
				if i < size-1 && fifo.IsFull() {
					t.Errorf("expected fifo to be non-full after writing %d values", i+1)
				}
				if fifo.IsEmpty() {
					t.Errorf("expected fifo to be non-empty after writing %d values", i+1)
				}
			}
			if fifo.IsEmpty() {
				t.Errorf("expected fifo to be non-empty after filling with data")
			}
			if !fifo.IsFull() {
				t.Errorf("expected fifo to be full after filling with data")
			}
		})
	}
}

func TestFIFOReadWrite(t *testing.T) {
	t.Run("size=0", func(t *testing.T) {
		fifo0 := NewFIFO(0)
		err := fifo0.Write(0)
		if err != ErrFIFOFull {
			t.Errorf("expected error ErrFIFOFull, got %v", err)
		}
		_, err = fifo0.Read()
		if err != ErrFIFOEmpty {
			t.Errorf("expected error ErrFIFOEmpty, got %v", err)
		}
	})

	// Randomly chosen.
	testingValues := [fifoTestingSize]uint32{
		0x5e72d6bd, 0x7a705a80, 0xe83e901f, 0xc87d4cb6, 0xe74265dc, 0x7518bb81, 0xc821514a, 0x423fb469,
	}

	for size := 1; size <= fifoTestingSize; size++ {
		fifo := NewFIFO(uint(size))
		for fillAmount := 1; fillAmount <= size; fillAmount++ {
			t.Run(fmt.Sprintf("size=%d,fillAmount=%d", size, fillAmount), func(t *testing.T) {
				for i := range fillAmount {
					err := fifo.Write(testingValues[i])
					if err != nil {
						t.Fatalf("unexpected error writing value %X (index %d) to fifo: %v", testingValues[i], i, err)
					}
				}
				if fillAmount == size {
					err := fifo.Write(0)
					if err != ErrFIFOFull {
						t.Errorf("expected error ErrFIFOFull, got %v", err)
					}
				}
				for i := range fillAmount {
					expectedValue := testingValues[i]
					actualValue, err := fifo.Read()
					if err != nil {
						t.Fatalf("unexpected error reading value from fifo (index %d): %v", i, err)
					}
					if expectedValue != actualValue {
						t.Errorf("expected value %X, got %X (index %d)", expectedValue, actualValue, i)
					}
				}
				_, err := fifo.Read()
				if err != ErrFIFOEmpty {
					t.Errorf("expected error ErrFIFOEmpty, got %v", err)
				}
			})
		}
	}
}
