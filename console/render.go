package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"dev-digest/common"
)

func RenderReports(reports []*common.Report) {
	title := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	subtle := color.New(color.FgHiBlack).SprintFunc()
	okc := color.New(color.FgHiGreen).SprintFunc()
	warnc := color.New(color.FgHiYellow).SprintFunc()
	errc := color.New(color.FgHiRed).SprintFunc()

	for i, r := range reports {
		if i > 0 {
			fmt.Println()
		}
		header := fmt.Sprintf("%s %s", title("■"), title(r.Title))
		fmt.Println(header)
		if r.Summary != "" {
			fmt.Println(subtle("→ "), r.Summary)
		}
		for _, w := range r.Warnings {
			fmt.Println(warnc("!"), w)
		}
		for _, e := range r.Errors {
			fmt.Println(errc("x"), e)
		}
		for _, s := range r.Sections {
			if s.Header != "" {
				fmt.Println("  ", color.New(color.Bold).Sprint(s.Header))
			}
			if s.Table != nil {
				tw := tablewriter.NewWriter(os.Stdout)
				tw.SetHeader(s.Table.Headers)
				tw.SetBorder(false)
				tw.SetAutoWrapText(false)
				tw.SetHeaderLine(false)
				tw.SetColumnSeparator(" ")
				tw.SetRowSeparator(" ")
				tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				tw.SetAlignment(tablewriter.ALIGN_LEFT)
				tw.AppendBulk(s.Table.Rows)
				tw.Render()
			}
			if len(s.Lines) > 0 {
				for _, line := range s.Lines {
					fmt.Println("   ", line)
				}
			}
		}
		if r.Meta != nil {
			if d, ok := r.Meta["duration"].(string); ok && d != "" {
				fmt.Println(subtle(fmt.Sprintf("took %s", d)))
			}
		}
	}

	// Footer divider
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(okc("Done."))
}
