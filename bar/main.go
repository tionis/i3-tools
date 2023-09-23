package bar

import (
	barista "barista.run"
	"barista.run/bar"
	"barista.run/colors"
	"barista.run/format"
	"barista.run/modules/battery"
	"barista.run/modules/clock"
	"barista.run/modules/diskspace"
	"barista.run/modules/meminfo"
	"barista.run/modules/netinfo"
	"barista.run/modules/sysinfo"
	"barista.run/modules/volume"
	"barista.run/modules/volume/alsa"
	"barista.run/modules/wlan"
	"barista.run/outputs"
	"fmt"
	"github.com/martinlindhe/unit"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"tasadar.net/tionis/i3-tools/bar/certinfo"
	"tasadar.net/tionis/i3-tools/bar/yubikey"
	"time"
)

const (
	storageSymbol  = " "
	wifiSymbol     = " "
	ethernetSymbol = " "
	certSymbol     = " "
	volumeSymbol   = " "
	//warnSymbol     = " "
	//errorSymbol    = " "
	//infoSymbol     = " "
	//ramSymbol      = " "
	//timeSymbol     = " "
	//loadSymbol     = " "
	//yubikeySymbol  = " "

)

type Config struct {
	Ethernet         bool
	Wifi             bool
	Battery          bool
	IPv6             bool
	WifiIPs          bool
	TerminalEmulator string
	ColorGood        string
	ColorBad         string
	ColorDegraded    string
}

func Status(c Config) error {
	log.SetOutput(log.Writer())
	colors.LoadFromMap(map[string]string{
		"good":     c.ColorGood,
		"bad":      c.ColorBad,
		"degraded": c.ColorDegraded,
	})

	// Display information about ssh certificate
	barista.Add(certinfo.New(certSymbol + "[%s]"))

	// Display system load
	loadWarnLimit := float64(runtime.NumCPU()) * 0.8
	barista.Add(sysinfo.New().Output(func(i sysinfo.Info) bar.Output {
		out := outputs.Textf("%.2f/%.2f/%.2f", i.Loads[0], i.Loads[1], i.Loads[2])
		if i.Loads[0] > loadWarnLimit {
			out.Color(colors.Scheme("bad"))
		}
		return out
	}))

	// storage
	barista.Add(diskspace.New("/").Output(func(i diskspace.Info) bar.Output {
		out := outputs.Pango(storageSymbol + format.IBytesize(i.Available))
		switch {
		case i.AvailFrac() < 0.05:
			out.Color(colors.Scheme("bad"))
		case i.Available < 3*unit.Gigabyte:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.1:
			out.Color(colors.Scheme("degraded"))
		}
		out.OnClick(func(e bar.Event) {
			if e.Button == bar.ButtonLeft {
				_ = exec.Command(c.TerminalEmulator, "-e", "gdu").Run()
			}
		})
		return out
	}))

	// volume
	barista.Add(volume.New(alsa.DefaultMixer()).Output(func(v volume.Volume) bar.Output {
		if v.Mute {
			return outputs.Textf("%s[MUT]", volumeSymbol).Color(colors.Scheme("degraded"))
		}
		return outputs.Textf("%s[%02d%%]", volumeSymbol, v.Pct())
	}))

	// Display yubikey touch prompt
	barista.Add(yubikey.New().Output(func(gpg, u2f bool) bar.Output {
		var reason []string
		if gpg {
			reason = append(reason, "GPG")
		}
		if u2f {
			reason = append(reason, "U2F")
		}
		if len(reason) == 0 {
			return nil
		}
		out := outputs.Textf("[YK: %s]", strings.Join(reason, ","))
		out.Urgent(true)
		return out
	}))

	// network
	if c.IPv6 {
		barista.Add(netinfo.New().Output(func(s netinfo.State) bar.Output {
			if !s.Enabled() {
				return nil
			}
			for _, ip := range s.IPs {
				if ip.To4() == nil && ip.IsGlobalUnicast() {
					return outputs.Text(ip.String()).Color(colors.Scheme("good"))
				}
			}
			return outputs.Text("no IPv6").Color(colors.Scheme("bad"))
		}))
	}

	barista.Add(wlan.Any().Output(func(w wlan.Info) bar.Output {
		switch {
		case w.Connected():
			out := fmt.Sprintf(wifiSymbol+"[%s]", w.SSID)
			if c.WifiIPs {
				if len(w.IPs) > 0 {
					out += fmt.Sprintf(" %s", w.IPs[0])
				}
			}
			return outputs.Text(out).Color(colors.Scheme("good")).OnClick(func(e bar.Event) {
				if e.Button == bar.ButtonLeft {
					_ = exec.Command(c.TerminalEmulator, "-e", "nmtui").Run()
				}
			})
		case w.Connecting():
			return outputs.Text(wifiSymbol + "[connecting...]").Color(colors.Scheme("degraded"))
		case w.Enabled():
			return outputs.Text(wifiSymbol + "[down]").Color(colors.Scheme("bad"))
		default:
			return nil
		}
	}))

	if c.Ethernet {
		barista.Add(netinfo.Prefix("e").Output(func(s netinfo.State) bar.Output {
			switch {
			case s.Connected():
				ip := "<no ip>"
				if len(s.IPs) > 0 {
					ip = s.IPs[0].String()
				}
				return outputs.Textf(ethernetSymbol+"[%s]", ip).Color(colors.Scheme("good"))
			case s.Connecting():
				return outputs.Text(ethernetSymbol + "[connecting...]").Color(colors.Scheme("degraded"))
			case s.Enabled():
				return outputs.Text(ethernetSymbol + "[down]").Color(colors.Scheme("bad"))
			default:
				return nil
			}
		}))
	}

	// battery
	statusName := map[battery.Status]string{
		battery.Charging:    " ",
		battery.Discharging: " ",
		battery.NotCharging: " ",
		battery.Unknown:     "",
	}
	barista.Add(battery.All().Output(func(b battery.Info) bar.Output {
		if b.Status == battery.Disconnected {
			return outputs.Text("NO BATTERY").Color(colors.Scheme("bad"))
		}
		if b.Status == battery.Full {
			return outputs.Text("FULL")
		}
		out := outputs.Textf("%s %d%% %s",
			statusName[b.Status],
			b.RemainingPct(),
			b.RemainingTime())
		if b.Discharging() {
			if b.RemainingPct() < 10 || b.RemainingTime() < 10*time.Minute {
				out.Color(colors.Scheme("bad")).Urgent(true)
			} else if b.RemainingPct() < 20 || (b.RemainingTime() != 0 && b.RemainingTime() < 30*time.Minute) {
				out.Color(colors.Scheme("bad"))
			}
		}
		return out
	}))

	// ram
	barista.Add(meminfo.New().Output(func(i meminfo.Info) bar.Output {
		if i.Available() < 0.7*unit.Gigabyte {
			return outputs.Textf(`MEMORY < %s`,
				format.IBytesize(i.Available())).
				Color(colors.Scheme("bad"))
		}
		out := outputs.Textf(`%s/%s`,
			format.IBytesize(i["MemTotal"]-i.Available()),
			format.IBytesize(i.Available()))
		switch {
		case i.AvailFrac() < 0.05:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.1:
			out.Color(colors.Scheme("degraded"))
		}
		return out
	}))

	// time
	barista.Add(clock.Local().OutputFormat("2006-01-02 15:04:05"))

	// if crash on screen locking and
	// using `status_command exec /path/to/i3-tools bar render`
	// in i3 config does not help, try uncommenting this:
	// barista.SuppressSignals(true)
	return barista.Run()
}
