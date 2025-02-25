package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	// color palette
	primaryColor   = lipgloss.Color("#7b68ee") 
	secondaryColor = lipgloss.Color("#6c7ae0") 
	bgColor        = lipgloss.Color("#1e1e2e") 
	textColor      = lipgloss.Color("#cdd6f4") 
	mutedTextColor = lipgloss.Color("#a6adc8") 
	accentColor    = lipgloss.Color("#89b4fa") 
	alertColor     = lipgloss.Color("#f38ba8") 
	goodColor      = lipgloss.Color("#a6e3a1") 
	neutralColor   = lipgloss.Color("#f9e2af") 

	// Base styles
	appStyle     = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Background(bgColor)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1e1e2e")).Background(primaryColor).Padding(0, 2).MarginBottom(1)
	headerStyle  = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).MarginBottom(1)
	keyStyle     = lipgloss.NewStyle().Foreground(mutedTextColor)
	valueStyle   = lipgloss.NewStyle().Foreground(textColor).Bold(true)
	spinnerStyle = lipgloss.NewStyle().Foreground(accentColor)
	statusStyle  = lipgloss.NewStyle().Foreground(mutedTextColor).AlignHorizontal(lipgloss.Right)
	sectionStyle = lipgloss.NewStyle().Padding(1).MarginRight(2).Border(lipgloss.RoundedBorder()).BorderForeground(secondaryColor)

	zoneStyle       = lipgloss.NewStyle().Background(secondaryColor).Padding(0, 1).MarginRight(1)
	activeZoneStyle = lipgloss.NewStyle().Background(primaryColor).Foreground(lipgloss.Color("#1e1e2e")).Padding(0, 1).MarginRight(1).Bold(true)
	helpStyle       = lipgloss.NewStyle().Foreground(mutedTextColor).Italic(true).MarginTop(1)
)

type tickMsg time.Time
type fanAnimationTickMsg time.Time
type systemUpdateTickMsg time.Time

type model struct {
	sensorData    string
	zone          int
	color         string
	spinner       spinner.Model
	width         int
	height        int
	lastUpdated   string
	fanFrame      int
	fanSpeed      int
	fanRPM        int
	cpuTemp       int
	sensors       map[string]string
	cpuUsage      float64
	ramUsage      float64
	ramTotal      float64
	ramUsed       float64
	netStrength   int
	netSpeed      float64
	uptime        time.Duration
	startTime     time.Time
	cpuCores      int
	hostInfo      *host.InfoStat
	netStats      []net.IOCountersStat
	prevNetStats  []net.IOCountersStat
	prevStatsTime time.Time
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle

	hostInfo, _ := host.Info()
	cpuCores := runtime.NumCPU()

	return model{
		sensorData:    "Loading...",
		zone:          1,
		color:         "white",
		spinner:       s,
		lastUpdated:   "Initializing...",
		fanFrame:      0,
		fanSpeed:      0,
		fanRPM:        0,
		cpuTemp:       0,
		sensors:       make(map[string]string),
		cpuUsage:      0,
		ramUsage:      0,
		ramTotal:      0,
		ramUsed:       0,
		netStrength:   0,
		netSpeed:      0.0,
		uptime:        0,
		startTime:     time.Now(),
		cpuCores:      cpuCores,
		hostInfo:      hostInfo,
		netStats:      nil,
		prevNetStats:  nil,
		prevStatsTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		fanAnimationTickCmd(),
		systemUpdateTickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "1", "2", "3", "4":
            zone, _ := strconv.Atoi(msg.String())
            m.zone = zone
        case "r":
            m.color = "red"
        case "g":
            m.color = "green"
        case "b":
            m.color = "blue"
        case "w":
            m.color = "white"
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    case tickMsg:
        m.sensorData = getSensorData()
        m.sensors = parseSensorData(m.sensorData)

        if temp, ok := m.sensors["Core 0"]; ok {
            tempVal, err := strconv.ParseFloat(strings.TrimPrefix(temp, "+"), 64)
            if err == nil {
                m.cpuTemp = int(tempVal)
                
                defaultFanSpeed := calculateFanSpeed(m.cpuTemp)
                defaultFanRPM := calculateFanRPM(defaultFanSpeed)
                
                m.fanSpeed = defaultFanSpeed
                m.fanRPM = defaultFanRPM
            }
        }

        fanSpeedFound := false
        for key, value := range m.sensors {
            if strings.Contains(key, "fan") && strings.Contains(value, "RPM") {
                rpmStr := strings.TrimSuffix(value, " RPM")
                rpm, err := strconv.Atoi(rpmStr)
                if err == nil {
                    m.fanRPM = rpm
                    m.fanSpeed = (rpm * 100) / 5000
                    if m.fanSpeed > 100 {
                        m.fanSpeed = 100
                    }
                    fanSpeedFound = true
                    break
                }
            }
        }

        if !fanSpeedFound {
            m.fanRPM = 1000 + rand.Intn(3000) 
            if m.cpuTemp > 70 {
                m.fanRPM = 3000 + rand.Intn(2000) 
            }
            m.fanSpeed = (m.fanRPM * 100) / 5000 
        }

        m.lastUpdated = time.Now().Format("15:04:05")
        cmds = append(cmds, tickCmd())
    case fanAnimationTickMsg:
        m.fanFrame = (m.fanFrame + 1) % 4
        cmds = append(cmds, fanAnimationTickCmd())
    case systemUpdateTickMsg:
        m.updateSystemMetrics()
        cmds = append(cmds, systemUpdateTickCmd())
    }

    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}

func (m model) View() string {
	contentWidth := m.width - 6

	if contentWidth < 20 {
		contentWidth = 20
	}

	spinnerText := m.spinner.View()
	status := statusStyle.Render("Updated: " + m.lastUpdated)

	title := titleStyle.Render(" NitroLit ")

	var titleBar string
	if contentWidth >= 40 {
		titleBar = lipgloss.JoinHorizontal(lipgloss.Center,
			spinnerText+" "+title,
			lipgloss.PlaceHorizontal(contentWidth-lipgloss.Width(spinnerText+title+status), lipgloss.Right, status),
		)
	} else {
		titleBar = lipgloss.JoinVertical(lipgloss.Center,
			spinnerText+" "+title,
			status,
		)
	}

	isNarrow := contentWidth < 60
	isMedium := contentWidth >= 60 && contentWidth < 100
	isWide := contentWidth >= 100

	var barWidth int
	if isNarrow {
		barWidth = max(10, contentWidth-25)
	} else if isMedium {
		barWidth = contentWidth/2 - 20
	} else {
		barWidth = contentWidth/3 - 20
	}

	var sysMetricsDisplay strings.Builder

	cpuBar := renderProgressBar(int(m.cpuUsage), barWidth, m.cpuUsage > 90)
	ramBar := renderProgressBar(int(m.ramUsage), barWidth, m.ramUsage > 90)
	netBar := renderProgressBar(m.netStrength, barWidth, false)

	uptimeStr := formatUptime(m.uptime)

	sysMetricsDisplay.WriteString(headerStyle.Render("System Metrics") + "\n")
	sysMetricsDisplay.WriteString(fmt.Sprintf("%s %s %.1f%% (%d cores)\n", keyStyle.Render("CPU"), cpuBar, m.cpuUsage, m.cpuCores))
	sysMetricsDisplay.WriteString(fmt.Sprintf("%s %s %.1f%% (%.1f/%.1f GB)\n", keyStyle.Render("RAM"), ramBar, m.ramUsage, m.ramUsed, m.ramTotal))
	sysMetricsDisplay.WriteString(fmt.Sprintf("%s %s %d%% (%.1f Mbps)\n", keyStyle.Render("NET"), netBar, m.netStrength, m.netSpeed))

	var sensorDisplay strings.Builder
	sensorDisplay.WriteString(headerStyle.Render("System Info") + "\n")

	sensorDisplay.WriteString(fmt.Sprintf("%s %s\n", keyStyle.Render("Uptime"), valueStyle.Render(uptimeStr)))
	if m.hostInfo != nil {
		sensorDisplay.WriteString(fmt.Sprintf("%s %s %s\n", keyStyle.Render("OS"), valueStyle.Render(m.hostInfo.Platform), valueStyle.Render(m.hostInfo.PlatformVersion)))
		sensorDisplay.WriteString(fmt.Sprintf("%s %s\n", keyStyle.Render("Host"), valueStyle.Render(m.hostInfo.Hostname)))
	}

	prioritySensors := []string{
		"Package id 0", "Core 0", "Core 1", "Core 2",
		"Core 3", "Core 4", "Core 5", "Composite", "temp1", "in0",
	}

	maxSensors := 6
	if isNarrow {
		maxSensors = 3
	} else if isMedium {
		maxSensors = 5
	}

	displayCount := 0

	for _, sensorName := range prioritySensors {
		if displayCount >= maxSensors {
			break
		}

		if value, ok := m.sensors[sensorName]; ok {
			key := keyStyle.Render(sensorName)
			val := valueStyle.Render(value)

			lineText := key + " " + val

			if isNarrow && len(sensorName) > 10 {
				key = keyStyle.Render(sensorName[:7] + "...")
				lineText = key + " " + val
			}

			if lipgloss.Width(lineText) > (contentWidth / 2) {
				lineText = lineText[:(contentWidth/2)-3] + "..."
			}

			sensorDisplay.WriteString(lineText + "\n")
			displayCount++
		}
	}

	fanAnimation := renderFanAnimation(m.fanFrame, m.fanSpeed, m.fanRPM, isNarrow)

	metricsSection := sectionStyle.Copy().Render(sysMetricsDisplay.String())
	sensorsSection := sectionStyle.Copy().Render(sensorDisplay.String())
	fanSection := sectionStyle.Copy().Render(fanAnimation)

	var mainContent string

	if isWide {
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, metricsSection, fanSection)
		mainContent = lipgloss.JoinVertical(lipgloss.Left, topRow, sensorsSection)
	} else if isMedium {
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, metricsSection, fanSection)
		mainContent = lipgloss.JoinVertical(lipgloss.Left, topRow, sensorsSection)
	} else {
		mainContent = lipgloss.JoinVertical(lipgloss.Left, metricsSection, fanSection, sensorsSection)
	}

	var controlSection string

	if contentWidth > 30 {
		zones := make([]string, 4)
		for i := range zones {
			zoneNum := i + 1
			if zoneNum == m.zone {
				zones[i] = activeZoneStyle.Render(fmt.Sprintf("%d", zoneNum))
			} else {
				zones[i] = zoneStyle.Render(fmt.Sprintf("%d", zoneNum))
			}
		}

		var zoneSelector string
		if isNarrow && contentWidth < 40 {
			zoneSelector = lipgloss.JoinVertical(lipgloss.Center, zones...)
		} else {
			zoneSelector = lipgloss.JoinHorizontal(lipgloss.Center, zones...)
		}

		controlSection = lipgloss.JoinVertical(lipgloss.Center,
			headerStyle.Render("Keyboard Zones"),
			zoneSelector)

		controlSection = sectionStyle.Copy().Render(controlSection)
	}

	var helpText string
	if contentWidth > 60 {
		helpText = helpStyle.Render("Press 1-4: zones | r/g/b/w: colors | q: quit")
	} else if contentWidth > 40 {
		helpText = helpStyle.Render("1-4: zones | r/g/b/w: colors | q: quit")
	} else if contentWidth > 20 {
		helpText = helpStyle.Render("1-4: zones | q: quit")
	} else {
		helpText = helpStyle.Render("q: quit")
	}

	var content string
	if controlSection != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			titleBar,
			mainContent,
			controlSection,
			helpText,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			titleBar,
			mainContent,
			helpText,
		)
	}

	return appStyle.Width(contentWidth).Render(content)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fanAnimationTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*600, func(t time.Time) tea.Msg {
		return fanAnimationTickMsg(t)
	})
}

func systemUpdateTickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return systemUpdateTickMsg(t)
	})
}

func getSensorData() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("sensors")
	case "darwin":
		cmd = exec.Command("system_profiler", "SPPowerDataType")
	case "windows":
		return getSampleSensorData()
	default:
		return getSampleSensorData()
	}

	output, err := cmd.Output()
	if err != nil {
		return getSampleSensorData()
	}

	return string(output)
}

func getSampleSensorData() string {
	return `coretemp-isa-0000
Adapter: ISA adapter
Package id 0:  +85.0°C  (high = +100.0°C, crit = +100.0°C)
Core 0:        +80.0°C  (high = +100.0°C, crit = +100.0°C)
Core 1:        +78.0°C  (high = +100.0°C, crit = +100.0°C)
Core 2:        +79.0°C  (high = +100.0°C, crit = +100.0°C)
Core 3:        +85.0°C  (high = +100.0°C, crit = +100.0°C)
Core 4:        +77.0°C  (high = +100.0°C, crit = +100.0°C)
Core 5:        +78.0°C  (high = +100.0°C, crit = +100.0°C)
acpitz-acpi-0
Adapter: ACPI interface
temp1:        +83.0°C
nvme-pci-e100
Adapter: PCI adapter
Composite:    +64.8°C  (low  =  -5.2°C, high = +79.8°C)
                       (crit = +84.8°C)
iwlwifi_1-virtual-0
Adapter: Virtual device
temp1:        +65.0°C
BAT1-acpi-0
Adapter: ACPI interface
in0:          16.65 V
curr1:         0.00 A
fan1:          0 RPM`
}

func parseSensorData(data string) map[string]string {
    result := make(map[string]string)
    lines := strings.Split(data, "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || !strings.Contains(line, ":") {
            continue
        }

        parts := strings.SplitN(line, ":", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])

            if strings.Contains(value, "°C") {
                tempParts := strings.Split(value, " ")
                result[key] = tempParts[0]
            } else if strings.Contains(value, "V") || strings.Contains(value, "A") || strings.Contains(value, "RPM") {
                result[key] = value
            } else {
                result[key] = value
            }
        }
    }

    return result
}

func calculateFanSpeed(cpuTemp int) int {
    if cpuTemp < 40 {
        return 30
    } else if cpuTemp > 80 {
        return 100
    }
    return 30 + (cpuTemp-40)*70/40
}

func calculateFanRPM(fanSpeed int) int {
    return fanSpeed * 50
}


func renderFanAnimation(frame int, speed int, rpm int, isNarrow bool) string {
    frames := []string{
        "◢◤",
        "◣◥",
        "◤◢",
        "◥◣",
    }

    barWidth := 10
    if isNarrow {
        barWidth = 5
    }

    speed = max(0, min(100, speed))
    
    filledBars := barWidth * speed / 100
    speedBar := ""

    for i := 0; i < barWidth; i++ {
        if i < filledBars {
            speedBar += "█"
        } else {
            speedBar += "▒"
        }
    }

    var speedColor lipgloss.Style
    if speed > 80 {
        speedColor = lipgloss.NewStyle().Foreground(alertColor)
    } else if speed < 50 {
        speedColor = lipgloss.NewStyle().Foreground(goodColor)
    } else {
        speedColor = lipgloss.NewStyle().Foreground(neutralColor)
    }

    var fanSymbol string
    fanStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
    
    if rpm < 100 {
        fanSymbol = fanStyle.Render("◯") 
    } else {
        actualFrame := frame
        if speed > 70 {
            actualFrame = frame
        } else if speed > 40 {
            actualFrame = frame / 2
        } else {
            actualFrame = frame / 3
        }
        actualFrame = actualFrame % len(frames)
        fanSymbol = fanStyle.Render(frames[actualFrame])
    }
    
    speedText := speedColor.Render(fmt.Sprintf("%d%% (%d RPM)", speed, rpm))

    var result strings.Builder
    result.WriteString(headerStyle.Render("Fan Speed") + "\n")

    if isNarrow {
        result.WriteString(speedBar + "\n")
        result.WriteString(speedText + "\n")
        result.WriteString(fanSymbol)
    } else {
        result.WriteString(speedBar + " " + speedText + "\n")
        result.WriteString(lipgloss.NewStyle().PaddingTop(1).AlignHorizontal(lipgloss.Center).Width(barWidth + 5).Render(fanSymbol))
    }

    return result.String()
}

func renderProgressBar(value int, width int, warning bool) string {
	if width < 5 {
		width = 5
	}

	barWidth := width - 2
	filledWidth := barWidth * value / 100

	filledWidth = max(0, min(filledWidth, barWidth))
	emptyWidth := max(0, barWidth-filledWidth)

	var barColor lipgloss.Style
	if warning {
		barColor = lipgloss.NewStyle().Foreground(alertColor)
	} else if value < 50 {
		barColor = lipgloss.NewStyle().Foreground(goodColor)
	} else if value < 80 {
		barColor = lipgloss.NewStyle().Foreground(neutralColor)
	} else {
		barColor = lipgloss.NewStyle().Foreground(primaryColor)
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("▒", emptyWidth)

	bracketStyle := lipgloss.NewStyle().Foreground(mutedTextColor)
	return bracketStyle.Render("『") + barColor.Render(filled) + empty + bracketStyle.Render("』")
}

func (m *model) updateSystemMetrics() {
	if m.hostInfo != nil {
		uptimeSec, _ := host.Uptime()
		m.uptime = time.Duration(uptimeSec) * time.Second
	} else {
		m.uptime = time.Since(m.startTime)
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		m.ramTotal = float64(memInfo.Total) / (1024 * 1024 * 1024)
		m.ramUsed = float64(memInfo.Used) / (1024 * 1024 * 1024)
		m.ramUsage = memInfo.UsedPercent
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		m.cpuUsage = cpuPercent[0]
	}

	currentNetStats, err := net.IOCounters(false)
	currentTime := time.Now()

	if err == nil && len(currentNetStats) > 0 && m.prevNetStats != nil && len(m.prevNetStats) > 0 {
		timeDiff := currentTime.Sub(m.prevStatsTime).Seconds()
		if timeDiff > 0 {
			bytesSent := currentNetStats[0].BytesSent - m.prevNetStats[0].BytesSent
			bytesRecv := currentNetStats[0].BytesRecv - m.prevNetStats[0].BytesRecv
			totalBytes := bytesSent + bytesRecv

			m.netSpeed = (float64(totalBytes) * 8 / 1000000) / timeDiff

			if m.netSpeed > 100 {
				m.netStrength = 100
			} else if m.netSpeed < 1 {
				m.netStrength = max(30, int(m.netSpeed*100))
			} else {
				m.netStrength = min(100, max(30, int(m.netSpeed)))
			}
		}
	}

	m.prevNetStats = currentNetStats
	m.prevStatsTime = currentTime
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	rand.Seed(time.Now().UnixNano())

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
