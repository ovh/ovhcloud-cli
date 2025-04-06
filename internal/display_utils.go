package internal

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss/tree"
)

func RenderObject(values map[string]any, titleKey string) {
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

func RenderTable(values []map[string]any, columnsToDisplay []string) {
	var rows [][]string

	for _, val := range values {
		var row []string

		for _, col := range columnsToDisplay {
			row = append(row, fmt.Sprint(val[col]))
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
