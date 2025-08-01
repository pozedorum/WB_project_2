package parser

import (
	"fmt"
	"os"
	"testing"

	"task15/internal/core"
)

func TestTokenizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		// Базовые случаи
		{
			name:     "simple command",
			input:    "ls -l",
			expected: []string{"ls", "-l"},
		},
		{
			name:     "command with quotes",
			input:    `echo "hello world"`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "single quoted command",
			input:    `echo 'hello $USER'`,
			expected: []string{"echo", "hello $USER"},
		},

		// Тесты с экранированием
		{
			name:     "escaped space",
			input:    `echo hello\ world`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "escaped quote",
			input:    `echo \"hello\"`,
			expected: []string{"echo", `"hello"`},
		},

		// Тесты с операторами
		{
			name:     "pipe operator",
			input:    "ls | grep go",
			expected: []string{"ls", "|", "grep", "go"},
		},
		{
			name:     "double pipe no spaces",
			input:    "ls||grep",
			expected: []string{"ls", "||", "grep"},
		},
		{
			name:     "redirect with append",
			input:    "echo hello >> file.txt",
			expected: []string{"echo", "hello", ">>", "file.txt"},
		},

		// Комбинированные случаи
		{
			name:     "combined operators",
			input:    "ls -l | grep test && echo found || echo not found",
			expected: []string{"ls", "-l", "|", "grep", "test", "&&", "echo", "found", "||", "echo", "not", "found"},
		},
		{
			name:     "complex quotes and operators",
			input:    `echo "hello" > file.txt && cat << EOF`,
			expected: []string{"echo", "hello", ">", "file.txt", "&&", "cat", "<<", "EOF"},
		},

		// Ошибочные случаи
		{
			name:    "unclosed double quote",
			input:   `echo "hello`,
			wantErr: true,
		},
		{
			name:    "unclosed single quote",
			input:   `echo 'hello`,
			wantErr: true,
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "multiple spaces",
			input:    "   ls    -l   ",
			expected: []string{"ls", "-l"},
		},
		{
			name:     "special chars in quotes",
			input:    `echo "| && > <"`,
			expected: []string{"echo", "| && > <"},
		},
		{
			name:     "mixed quotes",
			input:    `echo 'literal $USER' "expanded $USER"`,
			expected: []string{"echo", "literal $USER", "expanded " + os.Getenv("USER")},
		},
		{
			name:     "escaped quotes",
			input:    `echo \$USER '"$USER"' "'$USER'"`,
			expected: []string{"echo", "$USER", `"$USER"`, "'" + os.Getenv("USER") + "'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizeString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenizeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareStringSlices(got, tt.expected) {
				t.Errorf("TokenizeString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Вспомогательная функция для сравнения слайсов строк
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseTokens(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected *core.Command
		wantErr  bool
	}{
		// Простые команды
		{
			name:   "simple command",
			tokens: []string{"ls", "-l"},
			expected: &core.Command{
				Name: "ls",
				Args: []string{"-l"},
			},
		},
		{
			name:   "command with redirect",
			tokens: []string{"echo", "hello", ">", "file.txt"},
			expected: &core.Command{
				Name: "echo",
				Args: []string{"hello"},
				Redirects: []core.Redirect{
					{Type: ">", File: "file.txt"},
				},
			},
		},

		// Конвейеры
		{
			name:   "single pipe",
			tokens: []string{"ls", "|", "grep", "test"},
			expected: &core.Command{
				Name: "ls",
				PipeTo: &core.Command{
					Name: "grep",
					Args: []string{"test"},
				},
			},
		},
		{
			name:   "multiple pipes",
			tokens: []string{"cat", "file.txt", "|", "grep", "error", "|", "wc", "-l"},
			expected: &core.Command{
				Name: "cat",
				Args: []string{"file.txt"},
				PipeTo: &core.Command{
					Name: "grep",
					Args: []string{"error"},
					PipeTo: &core.Command{
						Name: "wc",
						Args: []string{"-l"},
					},
				},
			},
		},

		// Условные операторы
		{
			name:   "AND operator",
			tokens: []string{"make", "&&", "./app"},
			expected: &core.Command{
				Name: "make",
				AndNext: &core.Command{
					Name: "./app",
				},
			},
		},
		{
			name:   "OR operator",
			tokens: []string{"test", "-f", "file", "||", "touch", "file"},
			expected: &core.Command{
				Name: "test",
				Args: []string{"-f", "file"},
				OrNext: &core.Command{
					Name: "touch",
					Args: []string{"file"},
				},
			},
		},

		// Комбинированные случаи
		{
			name:   "pipe with AND",
			tokens: []string{"ls", "|", "grep", "txt", "&&", "wc", "-l"},
			expected: &core.Command{
				Name: "ls",
				PipeTo: &core.Command{
					Name: "grep",
					Args: []string{"txt"},
					AndNext: &core.Command{
						Name: "wc",
						Args: []string{"-l"},
					},
				},
			},
		},
		// Ошибочные случаи
		{
			name:    "empty command",
			tokens:  []string{},
			wantErr: true,
		},
		{
			name:    "unexpected operator",
			tokens:  []string{"&&", "ls"},
			wantErr: true,
		},
		{
			name:    "missing file for redirect",
			tokens:  []string{"echo", ">"},
			wantErr: true,
		},
		{
			name:    "multiple operators",
			tokens:  []string{"ls", "&&", "&&", "wc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTokens(tt.tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareCommands(got, tt.expected) {
				t.Errorf("parseTokens() = %v, want %v", commandToString(got), commandToString(tt.expected))
			}
		})
	}
}

// Вспомогательные функции для сравнения команд
func compareCommands(a, b *core.Command) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Name != b.Name {
		return false
	}
	if !compareStringSlices(a.Args, b.Args) {
		return false
	}
	if !compareRedirects(a.Redirects, b.Redirects) {
		return false
	}
	if !compareCommands(a.PipeTo, b.PipeTo) {
		return false
	}
	if !compareCommands(a.AndNext, b.AndNext) {
		return false
	}
	if !compareCommands(a.OrNext, b.OrNext) {
		return false
	}
	return true
}

func compareRedirects(a, b []core.Redirect) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type || a[i].File != b[i].File {
			return false
		}
	}
	return true
}

func commandToString(cmd *core.Command) string {
	if cmd == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Name:%s Args:%v Redirects:%v PipeTo:%s AndNext:%s OrNext:%s}",
		cmd.Name, cmd.Args, cmd.Redirects,
		commandToString(cmd.PipeTo),
		commandToString(cmd.AndNext),
		commandToString(cmd.OrNext))
}
