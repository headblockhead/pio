package pio

import (
	"fmt"
	"math/rand"
	"testing"
)

const fifoTestingSize = 8

func TestNewFIFO(t *testing.T) {
	for i := range fifoTestingSize + 1 {
		t.Run(fmt.Sprintf("size=%d", i), func(t *testing.T) {
			fifo := newFIFO(uint(i))
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
			fifo := newFIFO(uint(i))
			expectedSize := i
			actualSize := fifo.size()
			if expectedSize != int(actualSize) {
				t.Errorf("expected size %d, got %d", expectedSize, actualSize)
			}
		})
	}
}

func TestFIFOFullEmpty(t *testing.T) {
	t.Run("size=0", func(t *testing.T) {
		fifo0 := newFIFO(0)
		if !fifo0.isEmpty() {
			t.Errorf("expected fifo of size 0 to be empty")
		}
		if !fifo0.isFull() {
			t.Errorf("expected fifo of size 0 to be full")
		}
	})
	for size := 1; size <= fifoTestingSize; size++ {
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			fifo := newFIFO(uint(size))
			if !fifo.isEmpty() {
				t.Errorf("expected new fifo to be empty")
			}
			if fifo.isFull() {
				t.Errorf("expected new fifo to not be full")
			}
			for i := range size {
				if err := fifo.write(0); err != nil {
					t.Fatalf("unexpected error writing to fifo: %v", err)
				}
				if i < size-1 && fifo.isFull() {
					t.Errorf("expected fifo to be non-full after writing %d values", i+1)
				}
				if fifo.isEmpty() {
					t.Errorf("expected fifo to be non-empty after writing %d values", i+1)
				}
			}
			if fifo.isEmpty() {
				t.Errorf("expected fifo to be non-empty after filling with data")
			}
			if !fifo.isFull() {
				t.Errorf("expected fifo to be full after filling with data")
			}
		})
	}
}

func TestFIFOReadWrite(t *testing.T) {
	t.Run("size=0", func(t *testing.T) {
		fifo0 := newFIFO(0)
		err := fifo0.write(0)
		if err != ErrFIFOFull {
			t.Errorf("expected error ErrFIFOFull, got %v", err)
		}
		_, err = fifo0.read()
		if err != ErrFIFOEmpty {
			t.Errorf("expected error ErrFIFOEmpty, got %v", err)
		}
	})

	var testingValues [fifoTestingSize]uint32
	for i := range fifoTestingSize {
		testingValues[i] = rand.Uint32()
	}

	for size := 1; size <= fifoTestingSize; size++ {
		fifo := newFIFO(uint(size))
		for fillAmount := 1; fillAmount <= size; fillAmount++ {
			t.Run(fmt.Sprintf("size=%d,fillAmount=%d", size, fillAmount), func(t *testing.T) {
				for i := range fillAmount {
					err := fifo.write(testingValues[i])
					if err != nil {
						t.Fatalf("unexpected error writing value %X to fifo (fillAmount %d): %v", testingValues[i], fillAmount, err)
					}
				}
				if fillAmount == size {
					err := fifo.write(0)
					if err != ErrFIFOFull {
						t.Errorf("expected error ErrFIFOFull, got %v (fillAmount %d)", err, fillAmount)
					}
				}
				for i := range fillAmount {
					expectedValue := testingValues[i]
					actualValue, err := fifo.read()
					if err != nil {
						t.Fatalf("unexpected error reading value from fifo (fillAmount %d): %v", fillAmount, err)
					}
					if expectedValue != actualValue {
						t.Errorf("expected value %X, got %X (fillAmount %d)", expectedValue, actualValue, fillAmount)
					}
				}
				_, err := fifo.read()
				if err != ErrFIFOEmpty {
					t.Errorf("expected error ErrFIFOEmpty, got %v (fillAmount %d)", err, fillAmount)
				}
			})
		}
	}
}
