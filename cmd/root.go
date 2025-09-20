package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rate-limiter",
	Short: " Ferramente CLI para testes de carga HTTP",
	Long: "Um utilitário de linha de comando para realizar testes de carga em serviços web, com suporte a modo TUI.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar o comando: %v\n", err)
		os.Exit(1)
	}
}