package display

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/PaesslerAG/gval"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/ghodss/yaml"
	"github.com/tidwall/gjson"
	"gopkg.in/ini.v1"
)

func renderObject(values map[string]any, titleKey string) {
	enumeratorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).MarginRight(1)
	rootStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35")).Bold(true).Underline(true).MarginBottom(1)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("140"))

	t := tree.Root(values[titleKey]).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(enumeratorStyle).
		RootStyle(rootStyle).
		ItemStyle(itemStyle)

	for key, value := range values {
		if key == titleKey {
			continue
		}

		t.Child(key)

		switch v := value.(type) {
		case map[string]any:
			subTree := tree.New()
			for k, val := range v {
				subTree.Child(k)
				subTree.Child(generateChild(val))
			}
			t.Child(subTree)
		case []any:
			if len(v) == 0 {
				t.Child(tree.New().Child("[]"))
			} else {
				for _, val := range v {
					t.Child(generateChild(val))
				}
			}
		case string:
			if len(v) > 80 {
				t.Child(tree.New().Child(v[:50] + "..."))
			} else {
				t.Child(tree.New().Child(v))
			}
		default:
			if value == nil {
				t.Child(tree.New().Child("null"))
			} else if value == "" {
				t.Child(tree.New().Child(`""`))
			} else {
				t.Child(tree.New().Child(value))
			}
		}
	}

	fmt.Println(table.New().Border(lipgloss.NormalBorder()).Row(fmt.Sprint(t)))
}

func generateChild(value any) *tree.Tree {
	child := tree.New()

	switch v := value.(type) {
	case map[string]any:
		subTree := tree.New()
		for k, val := range v {
			subTree.Child(k)
			subTree.Child(generateChild(val))
		}
		child.Child(subTree)
	case []any:
		if len(v) == 0 {
			child.Child("[]")
		} else {
			for _, val := range v {
				child.Child(generateChild(val))
			}
		}
	case string:
		if len(v) > 80 {
			child.Child(v[:50] + "...")
		} else {
			child.Child(v)
		}
	default:
		if value == nil {
			child.Child("null")
		} else if value == "" {
			child.Child(`""`)
		} else {
			child.Child(value)
		}
	}

	return child
}

func RenderTable(values []byte, columnsToDisplay []string) {
	var rows [][]string

	lines := gjson.ParseBytes(values)
	for _, line := range lines.Array() {
		var row []string

		for _, col := range columnsToDisplay {
			v := line.Get(col)
			row = append(row, fmt.Sprint(v.Value()))
		}

		rows = append(rows, row)
	}

	var (
		purple = lipgloss.Color("99")
		gray   = lipgloss.Color("245")

		headerStyle = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle   = lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle = cellStyle.Foreground(gray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(columnsToDisplay...).
		Rows(rows...)

	fmt.Println(t)
}

func RenderTableFiltered(values []map[string]any, columnsToDisplay []string, jsonOutput, yamlOutput bool) {
	switch {
	case yamlOutput:
		if err := PrettyPrintYAML(values); err != nil {
			log.Fatalf("error displaying YAML results: %s", err)
		}
		return
	case jsonOutput:
		if err := PrettyPrintJSON(values); err != nil {
			log.Fatalf("error displaying JSON results: %s", err)
		}
		return
	}

	var (
		rows      [][]string
		selectors gval.Evaluables
	)

	for _, col := range columnsToDisplay {
		evaluator, err := gval.Full().NewEvaluable(col)
		if err != nil {
			log.Fatalf("invalid column to display %q: %s", col, err)
		}
		selectors = append(selectors, evaluator)
	}

	// Extract values to display
	for _, value := range values {
		values, err := selectors.EvalStrings(context.Background(), value)
		if err != nil {
			log.Fatalf("failed to select row fields: %s", err)
		}

		rows = append(rows, values)
	}

	var (
		purple = lipgloss.Color("99")
		gray   = lipgloss.Color("245")

		headerStyle = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle   = lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle = cellStyle.Foreground(gray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(columnsToDisplay...).
		Rows(rows...)

	fmt.Println(t)
}

func RenderConfigTable(cfg *ini.File) {
	var (
		rows    [][]string
		columns = []string{"section", "key", "value"}
	)

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		rows = append(rows, []string{section.Name()})
		for _, key := range section.Keys() {
			rows = append(rows, []string{"", key.Name(), key.Value()})
		}
	}

	var (
		purple    = lipgloss.Color("99")
		gray      = lipgloss.Color("245")
		lightGray = lipgloss.Color("241")

		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case len(rows[row]) == 1:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(columns...).
		Rows(rows...)

	fmt.Println(t)
}

func PrettyPrintJSON(value any) error {
	bytesOut, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytesOut))

	return nil
}

func PrettyPrintYAML(value any) error {
	bytesOut, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Println(string(bytesOut))

	return nil
}

func RenderTableRaw(value any, jsonOutput, yamlOutput bool) {
	switch {
	case yamlOutput:
		if err := PrettyPrintYAML(value); err != nil {
			log.Fatalf("error displaying YAML results: %s", err)
		}
	case jsonOutput:
		if err := PrettyPrintJSON(value); err != nil {
			log.Fatalf("error displaying JSON results: %s", err)
		}
	}
}

func OutputObject(value map[string]any, idKey string, jsonOutput, yamlOutput bool) {
	switch {
	case yamlOutput:
		if err := PrettyPrintYAML(value); err != nil {
			log.Fatalf("error displaying YAML results: %s", err)
		}
	case jsonOutput:
		if err := PrettyPrintJSON(value); err != nil {
			log.Fatalf("error displaying JSON results: %s", err)
		}
	default:
		renderObject(value, idKey)
	}
}
