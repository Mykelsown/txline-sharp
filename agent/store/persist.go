package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Mykelsown/txline-sharp/detector"
)

// Persist handles append-only writes to a JSONL signal log file.
// Each line in the file is a self-contained JSON object representing one Signal.
type Persist struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
	writer   *bufio.Writer
}

// NewPersist opens (or creates) the signals JSONL file for appending.
func NewPersist(filePath string) (*Persist, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open signals file %s: %w", filePath, err)
	}
	return &Persist{
		filePath: filePath,
		file:     f,
		writer:   bufio.NewWriter(f),
	}, nil
}

// Append writes a single signal as a JSON line to the log file.
// It is safe for concurrent use.
func (p *Persist) Append(sig detector.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.Marshal(sig)
	if err != nil {
		return fmt.Errorf("marshal signal: %w", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("write signal: %w", err)
	}
	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	// Flush immediately so signals survive a crash.
	return p.writer.Flush()
}

// Close flushes any buffered data and closes the underlying file.
func (p *Persist) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush on close: %w", err)
	}
	return p.file.Close()
}

// LoadAll reads every signal from the JSONL file and returns them in order.
// Returns an empty slice (not an error) if the file does not exist yet.
func LoadAll(filePath string) ([]detector.Signal, error) {
	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open signals file: %w", err)
	}
	defer f.Close()

	var signals []detector.Signal
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var sig detector.Signal
		if err := json.Unmarshal(line, &sig); err != nil {
			return nil, fmt.Errorf("decode signal line: %w", err)
		}
		signals = append(signals, sig)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan signals file: %w", err)
	}

	return signals, nil
}

// WriteAll overwrites the signals file with the provided slice.
// Used by the outcome tracker after resolving signals.
func WriteAll(filePath string, signals []detector.Signal) error {
	// Write to a temp file first, then rename atomically.
	tmpPath := filePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	w := bufio.NewWriter(f)
	for _, sig := range signals {
		data, err := json.Marshal(sig)
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("marshal signal: %w", err)
		}
		w.Write(data)
		w.WriteByte('\n')
	}

	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("flush: %w", err)
	}
	f.Close()

	return os.Rename(tmpPath, filePath)
}