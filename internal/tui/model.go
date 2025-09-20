package tui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"rate-limiter/internal/loadtest"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	ctx context.Context
	cfg loadtest.Config
	summary loadtest.Summary
	events <- chan loadtest.Event
	progress progress.Model
	startedAt time.Time
	elapsed time.Duration
	totalDone int
	total int
	status200 int
	errors int
	statusMap map[int]int
	done bool
	err error
	width int
	lastUpdate time.Time
}

func Run(ctx context.Context, cfg loadtest.Config) error {
	m := initialModel(ctx, cfg)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

func initialModel (ctx context.Context, cfg loadtest.Config) model {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 20
	p.SetPercent(0)
	ev := loadtest.StartAsync(ctx, cfg)

	return model {
		ctx: ctx,
		cfg: cfg,
		events: ev,
		progress: p,
		startedAt: time.Now(),
		total: cfg.Total,
		statusMap: map[int]int{},
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(listenEvents(m.events), tickCmd())
}

func listenEvents(ch <- chan loadtest.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <- ch
		if !ok {
			return nil
		}
		return ev
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tickMsg:
		m.elapsed = time.Since(m.startedAt)
		if m.done {
			return m, nil
		}
		return m, tea.Batch(listenEvents(m.events), tickCmd())
	case loadtest.Started:
		return m, listenEvents(m.events)
	case loadtest.RequestDone:
		m.totalDone++
		if msg.Err != nil {
			m.errors++
		}else{
			if msg.Status == 200 {
				m.status200++
			}
			m.statusMap[msg.Status] = m.statusMap[msg.Status] + 1
		}
		var cmds []tea.Cmd
		if m.total > 0 {
			p := float64(m.totalDone) / float64(m.total)
			cmds = append(cmds, m.progress.SetPercent(p))
		}
		cmds = append(cmds, listenEvents(m.events))
		
		return m, tea.Batch(cmds...)
	case loadtest.Finished:
		m.done = true
		sum := msg.Summary
		m.totalDone = sum.TotalDone
		m.status200 = sum.Status200
		m.errors = sum.Errors
		m.statusMap = sum.StatusMap
		m.elapsed = sum.EndedAt.Sub(sum.StartedAt)
		return m, nil
	
	case tea.KeyMsg:
		switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
    title := lipgloss.NewStyle().Bold(true).Render("Loadtester (TUI)")
    header := fmt.Sprintf("URL: %s | Total: %d | Concorrência: %d", m.cfg.URL, m.total, m.cfg.Concurrency)

    percent := 0.0
    if m.total > 0 {
        percent = float64(m.totalDone) / float64(m.total)
        if percent < 0 {
            percent = 0
        }
        if percent > 1 {
            percent = 1
        }
    }
    bar := m.progress.ViewAs(percent)

    rps := 0.0
    if m.elapsed > 0 {
        rps = float64(m.totalDone) / m.elapsed.Seconds()
    }

    statusLines := ""
    if len(m.statusMap) == 0 && m.errors == 0 {
        statusLines = "Aguardando respostas..."
    } else {
        // Ordena por código
        keys := make([]int, 0, len(m.statusMap))
        for k := range m.statusMap {
            keys = append(keys, k)
        }
        sort.Ints(keys)
        for _, code := range keys {
            if code == 0 {
                continue
            }
            statusLines += fmt.Sprintf("  %d: %d\n", code, m.statusMap[code])
        }
        if m.errors > 0 {
            statusLines += fmt.Sprintf("  erros (sem status HTTP): %d\n", m.errors)
        }
    }

    body := fmt.Sprintf(
        "%s\n%s\n\n%s\n\nProgresso: %d/%d\nHTTP 200: %d\nTempo: %v | RPS: %.2f\n\nDistribuição de status:\n%s",
        title,
        header,
        bar,
        m.totalDone, m.total,
        m.status200,
        m.elapsed.Truncate(100*time.Millisecond),
        rps,
        statusLines,
    )

    if m.done {
        body += "\nConcluído. Pressione q para sair."
    } else {
        body += "\nPressione q para sair."
    }

    return body
}