package log

import (
	"context"
	"errors"
	"sync"
)

var (
	gLogFormat   = map[string]uint8{"std": 1, "net": 1, "file": 1}
	errLogFormat = errors.New("Error log format upload")
	errLogFile   = errors.New("Error log file")
)

// --------------------------------
//    Log Init
// --------------------------------

// Регистрация настроек обработчика сообщений.
func NewLog(cfg *Config) error {
	Debug("-- NewLog")
	if _, ok := gLogFormat[cfg.Format]; !ok {
		return errLogFormat
	}

	switch cfg.Format {
	case "file":
		if err := newFile(cfg); err != nil {
			return err
		}

	case "net":
		if err := newNet(cfg); err != nil {
			return err
		}
	default:
		if err := newStd(cfg); err != nil {
			return err
		}
	}

	lg.mu.Lock()
	lg.Debug = cfg.Debug
	lg.format = cfg.Format
	lg.mu.Unlock()

	return nil
}

func newFile(cfg *Config) error {
	if len(cfg.LogFile) == 0 {
		return errLogFile
	}

	lg.mu.Lock()
	defer lg.mu.Unlock()

	if lg.file == nil {
		if err := lg.open(cfg.LogFile); err != nil {
			return err
		}
		return nil
	}

	if err := lg.rename(cfg.LogFile); err != nil {
		return err
	}

	return nil
}

func newNet(cfg *Config) error {
	return nil
}

func newStd(cfg *Config) error {
	return nil
}

// Обработка завершения работы приложения.
func Close(ct context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	Debug("-- Log Close")
	lg.Closer = true

	if len(lg.ChMsg) > 0 {
	loop:
		for {
			select {
			// ожидание таймера уровня приложения
			case <-ct.Done():
				break loop
			// ожидание освобождения канала
			case n := <-lg.ChLen:
				if n == 0 {
					break loop
				}
			}
		}
	}

	lg.Cancel()
}
