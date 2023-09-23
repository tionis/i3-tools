package certinfo

import (
	"barista.run/bar"
	"barista.run/base/value"
	"barista.run/colors"
	"barista.run/outputs"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/multiplay/go-cticker"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path"
	"time"
)

// Module represents a yubikey barista module that shows an indicator whenever
// the plugged-in yubikey is waiting for user input.
type Module struct {
	certPath       string
	outputFunc     value.Value // of func(bool, bool) bar.Output
	ticker         *cticker.Ticker
	tickerAccuracy int
	cert           *ssh.Certificate
	err            error
	format         string
}

func (m *Module) parseCert() {
	m.cert, m.err = parseCertFile(m.certPath)
}

// ForPath constructs a yubikey module with the given path to the gpg keyring.
func ForPath(certPath, format string) *Module {
	m := &Module{
		certPath:       certPath,
		ticker:         cticker.New(time.Minute, time.Second),
		tickerAccuracy: 60,
		format:         format,
	}

	m.Output(func() bar.Output {
		if m.cert == nil {
			m.parseCert()
		}
		if m.err != nil {
			return outputs.Error(m.err)
		}

		now := uint64(time.Now().Unix())
		var timePassed, timeRemaining uint64
		if now < m.cert.ValidAfter {
			timePassed = 0
		} else {
			timePassed = now - m.cert.ValidAfter
		}
		if now >= m.cert.ValidBefore {
			timeRemaining = 0
		} else {
			timeRemaining = m.cert.ValidBefore - now
		}

		if timeRemaining < 120 || timePassed < 120 {
			if m.tickerAccuracy != 1 {
				m.ticker = cticker.New(time.Second, time.Millisecond*100)
				m.tickerAccuracy = 1
			}
		} else if timeRemaining <= 60*60 || timePassed <= 60*60 {
			if m.tickerAccuracy != 60 {
				m.ticker = cticker.New(time.Minute, time.Second)
				m.tickerAccuracy = 60
			}
		} else if timeRemaining <= 24*60*60 || timePassed <= 24*60*60 {
			if m.tickerAccuracy != 60 {
				m.ticker = cticker.New(time.Hour, time.Minute)
				m.tickerAccuracy = 60 * 60
			}
		} else {
			if m.tickerAccuracy != 60*60 {
				m.ticker = cticker.New(time.Hour, time.Minute)
				m.tickerAccuracy = 60 * 60
			}
		}

		out := outputs.Textf(m.format, fmt.Sprintf("%s/%s", renderTime(timePassed), renderTime(timeRemaining)))
		if timeRemaining == 0 {
			out.Color(colors.Scheme("good"))
			out.Urgent(true)
		} else if timeRemaining < 60*60 {
			out.Color(colors.Scheme("bad"))
		} else if timePassed*100/(timePassed+timeRemaining) > 50 && timeRemaining < 1*24*60*60 {
			out.Color(colors.Scheme("bad"))
		} else if timePassed*100/(timePassed+timeRemaining) > 50 {
			out.Color(colors.Scheme("degraded"))
		}
		return out
	})
	return m
}

// New constructs a new yubikey module using the default paths for the u2f
// pending file and gpg keyring.
func New(format string) *Module {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return ForPath(path.Join(homeDir, ".ssh", "id_ed25519-cert.pub"), format)
}

// Output sets the output format for the module.
func (m *Module) Output(outputFunc func() bar.Output) *Module {
	m.outputFunc.Set(outputFunc)
	return m
}

// Stream starts the module.
func (m *Module) Stream(sink bar.Sink) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(fmt.Errorf("failed to create watcher: %w", err))
		return
	}
	err = watcher.Add(m.certPath)
	if err != nil {
		log.Println("cert path: ", m.certPath)
		log.Println(fmt.Errorf("failed to add file to watcher: %w", err))
		return
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Printf("failed to close watcher: %v", err)
		}
	}(watcher)
	outputFunc := m.outputFunc.Get().(func() bar.Output)
	quit := make(chan struct{})
	sink.Output(outputFunc())

	for {
		select {
		case <-m.ticker.C:
			sink.Output(outputFunc())
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
				m.parseCert()
				sink.Output(outputFunc())
			}
		case <-quit:
			m.ticker.Stop()
			return
		}
	}
}

func renderTime(seconds uint64) string {
	if seconds <= 0 {
		return "expired"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 60*60 {
		return fmt.Sprintf("%dm", seconds/60)
	} else if seconds < 24*60*60 {
		return fmt.Sprintf("%.1fh", float64(seconds)/60/60)
	} else {
		return fmt.Sprintf("%.1fd", float64(seconds)/60/60/24)
	}
}

func parseCertFile(certPath string) (*ssh.Certificate, error) {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	certAsKey, _, _, _, err := ssh.ParseAuthorizedKey(certBytes)
	if err != nil {
		return nil, err
	}
	return certAsKey.(*ssh.Certificate), nil
}
