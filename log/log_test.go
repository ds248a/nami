package log

import (
	"errors"
	"os"
	"testing"
	"time"
)

type Q struct {
	A int `json:"a"`
	B int `json:"b"`
}

func fileConfig(fname string) *Config {
	return &Config{
		Debug:   false,
		Format:  "file",
		LogFile: fname,
	}
}

func logClear() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	lg.file.Close()
	os.Remove(lg.fname)
	lg.open()
}

func TestConfig(t *testing.T) {
	logClear()

	t.Run("Config: file format", func(t *testing.T) {
		Msg("message 1").Query(&Q{A: 1}).Save()

		cfg := fileConfig("test.log")
		if err := NewLog(cfg); err != nil {
			t.Fatal("error config", err)
		}

		Msg("message 2").Query(&Q{A: 2}).Save()
		s, err := lg.Stat()
		if err != nil {
			t.Fatal("error file info", err)
		}

		if s.Name() != cfg.LogFile {
			t.Errorf("expect %v, want %v", s.Name(), cfg.LogFile)
		}

		time.Sleep(time.Millisecond * 200)
	})

	t.Run("Re-Config: file format", func(t *testing.T) {
		Msg("message 3").Query(&Q{A: 3}).Save()

		cfg := fileConfig("nami.log")
		if err := NewLog(cfg); err != nil {
			t.Fatal("error config", err)
		}

		Msg("message 4").Query(&Q{A: 4}).Save()
		s, err := lg.Stat()
		if err != nil {
			t.Fatal("error file info", err)
		}

		if s.Name() != cfg.LogFile {
			t.Errorf("expect %v, want %v", s.Name(), cfg.LogFile)
		}

		time.Sleep(time.Millisecond * 200)
	})

	t.Run("File format", func(t *testing.T) {
		s, err := lg.Stat()
		if err != nil {
			t.Fatal("error file info", err)
		}

		if s.Size() == 0 {
			t.Errorf("empty log file")
		}

		logClear()
	})
}

func TestWrite(t *testing.T) {
	t.Run("File write", func(t *testing.T) {
		Msg("test message").Query(&Q{A: 1}).Save()
		Err(errors.New("test err")).Query(&Q{B: 2}).Save()

		time.Sleep(time.Millisecond * 200)

		s, err := lg.Stat()
		if err != nil {
			t.Fatal("error file info", err)
		}

		if s.Size() == 0 {
			t.Errorf("empty log file")
		}

		// logClear()
	})
}
