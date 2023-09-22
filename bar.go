package main

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
	"barista.run/modules/wlan"
	"barista.run/outputs"
	"fmt"
	"github.com/martinlindhe/unit"
	"strings"
	"tasadar.net/tionis/i3-tools/certinfo"
	"tasadar.net/tionis/i3-tools/yubikey"
	"time"
)

type barConfig struct {
	ethernet bool
	wifi     bool
	battery  bool
	ipv6     bool
	wifiIPs  bool
}

func defaultBar() barConfig {
	return barConfig{
		ethernet: false,
		wifi:     true,
		battery:  true,
		ipv6:     false,
		wifiIPs:  false,
	}
}

func i3status(c barConfig) error {
	colors.LoadFromMap(map[string]string{
		"good":     "#0f0",
		"bad":      "#f00",
		"degraded": "#ff0",
	})

	// Display information about ssh certificate
	barista.Add(certinfo.New())

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

	// Display system load
	barista.Add(sysinfo.New().Output(func(i sysinfo.Info) bar.Output {
		out := outputs.Textf("%.2f/%.2f/%.2f", i.Loads[0], i.Loads[1], i.Loads[2])
		if i.Loads[0] > 5.0 {
			out.Color(colors.Scheme("bad"))
		}
		return out
	}))

	// storage
	barista.Add(diskspace.New("/").Output(func(i diskspace.Info) bar.Output {
		out := outputs.Text(format.IBytesize(i.Available))
		switch {
		case i.AvailFrac() < 0.2:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.33:
			out.Color(colors.Scheme("degraded"))
		}
		return out
	}))

	// network
	if c.ipv6 {
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
			out := fmt.Sprintf("W: (%s)", w.SSID)
			if c.wifiIPs {
				if len(w.IPs) > 0 {
					out += fmt.Sprintf(" %s", w.IPs[0])
				}
			}
			return outputs.Text(out).Color(colors.Scheme("good"))
		case w.Connecting():
			return outputs.Text("W: connecting...").Color(colors.Scheme("degraded"))
		case w.Enabled():
			return outputs.Text("W: down").Color(colors.Scheme("bad"))
		default:
			return nil
		}
	}))

	if c.ethernet {
		barista.Add(netinfo.Prefix("e").Output(func(s netinfo.State) bar.Output {
			switch {
			case s.Connected():
				ip := "<no ip>"
				if len(s.IPs) > 0 {
					ip = s.IPs[0].String()
				}
				return outputs.Textf("E: %s", ip).Color(colors.Scheme("good"))
			case s.Connecting():
				return outputs.Text("E: connecting...").Color(colors.Scheme("degraded"))
			case s.Enabled():
				return outputs.Text("E: down").Color(colors.Scheme("bad"))
			default:
				return nil
			}
		}))
	}

	// battery
	statusName := map[battery.Status]string{
		battery.Charging:    "CHR",
		battery.Discharging: "BAT",
		battery.NotCharging: "NOT",
		battery.Unknown:     "UNK",
	}
	barista.Add(battery.All().Output(func(b battery.Info) bar.Output {
		if b.Status == battery.Disconnected {
			return nil
		}
		if b.Status == battery.Full {
			return outputs.Text("FULL")
		}
		out := outputs.Textf("%s %d%% %s",
			statusName[b.Status],
			b.RemainingPct(),
			b.RemainingTime())
		if b.Discharging() {
			if b.RemainingPct() < 20 || (b.RemainingTime() != 0 && b.RemainingTime() < 30*time.Minute) {
				out.Color(colors.Scheme("bad"))
			}
		}
		return out
	}))

	// ram
	barista.Add(meminfo.New().Output(func(i meminfo.Info) bar.Output {
		if i.Available() < unit.Gigabyte {
			return outputs.Textf(`MEMORY < %s`,
				format.IBytesize(i.Available())).
				Color(colors.Scheme("bad"))
		}
		out := outputs.Textf(`%s/%s`,
			format.IBytesize(i["MemTotal"]-i.Available()),
			format.IBytesize(i.Available()))
		switch {
		case i.AvailFrac() < 0.2:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.33:
			out.Color(colors.Scheme("degraded"))
		}
		return out
	}))

	// time
	barista.Add(clock.Local().OutputFormat("2006-01-02 15:04:05"))

	// if crash on locking and using `status_command exec /path/to/i3-tools bar render`
	// in i3 config does not help, try uncommenting this:
	// barista.SuppressSignals(true)
	return barista.Run()
}
