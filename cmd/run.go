package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
	"rate-limiter/internal/loadtest"
	"rate-limiter/internal/tui"

	"github.com/spf13/cobra"
)

var(
	flagURL string
	flagRequests int
	flagConcurrency int
	flagTimeout time.Duration
	flagTUI bool
)

func init() {
	runCmd := &cobra.Command{
		Use: "run",
		Short: "Executa o teste de carga",
		RunE: run,
	}

	runCmd.Flags().StringVar(&flagURL, "url", "", "URL do serviço web a ser testado (obrigatório)")
	runCmd.Flags().IntVar(&flagRequests, "requests", 100, "Número total de requests a serem enviadas(obrigatório)")
	runCmd.Flags().IntVar(&flagConcurrency, "concurrency", 10, "Número de chamadas simultaneas(obrigatório)")
	runCmd.Flags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "Timeout para requisição (ex: 30s)")
	runCmd.Flags().BoolVar(&flagTUI, "tui", false, "Ativa o modo TUI(bubble tea)")

	_=runCmd.MarkFlagRequired("url")
	_=runCmd.MarkFlagRequired("requests")
	_=runCmd.MarkFlagRequired("concurrency")
	
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	if flagRequests <= 0 {
		return fmt.Errorf("O número de requests deve ser maior que zero")
	}
	if flagConcurrency <= 0 {
		return fmt.Errorf("O número de concorrência deve ser maior que zero")
	}
	cfg := loadtest.Config{
		URL:         flagURL,
		Total: 	flagRequests,
		Concurrency: flagConcurrency,
		Timeout:     flagTimeout,
		Method:      "GET",
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if flagTUI {
		return tui.Run(ctx, cfg)
	}

	start := time.Now()
	summary,err := loadtest.Run(ctx, cfg)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro durante execução: %v\n", err)
	}
	
	printSummary(cfg, summary, elapsed)

	return nil
}

func printSummary(cfg loadtest.Config, s loadtest.Summary, elapsed time.Duration) {
	fmt.Println("===== Relatório de Teste de Carga =====")
    fmt.Printf("URL: %s\n", cfg.URL)
    fmt.Printf("Total de requests configurados: %d\n", cfg.Total)
    fmt.Printf("Concorrência: %d\n", cfg.Concurrency)
    fmt.Printf("Tempo total: %v\n", elapsed)
    fmt.Printf("Requests executados: %d\n", s.TotalDone)
    fmt.Printf("HTTP 200: %d\n", s.Status200)
    if s.Errors > 0 {
        fmt.Printf("Erros (sem status HTTP): %d\n", s.Errors)
    }
    fmt.Println("Distribuição de status HTTP:")
    for _, kv := range s.StatusDitributionSorted() {
        if kv.Code == 0 {
            // 0 representa erro sem resposta HTTP
            continue
        }
        fmt.Printf("  %d: %d\n", kv.Code, kv.Count)
    }
    rps := float64(s.TotalDone) / elapsed.Seconds()
    if !mathIsValid(rps) {
        rps = 0
    }
    fmt.Printf("RPS médio aproximado: %.2f req/s\n", rps)
}

func mathIsValid(f float64) bool {
    return !((f != f) || (f > 1e308) || (f < -1e308))
}
