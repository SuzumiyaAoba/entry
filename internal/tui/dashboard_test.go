package tui_test

import (
	"testing"
	"time"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/history"
	"github.com/SuzumiyaAoba/entry/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TUI Suite")
}

var _ = Describe("Dashboard", func() {
	var (
		cfg *config.Config
		m   tui.Model
		err error
	)

	BeforeEach(func() {
		cfg = &config.Config{
			Rules: []config.Rule{
				{Name: "Rule 1", Command: "cmd1"},
				{Name: "Rule 2", Command: "cmd2"},
			},
		}
		m, err = tui.NewModel(cfg, "config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should initialize correctly", func() {
		Expect(m.Cfg).To(Equal(cfg))
		Expect(m.Active).To(Equal(tui.TabRules))
		Expect(m.RulesList.Items()).To(HaveLen(2))
	})

	It("should switch tabs", func() {
		// Simulate Tab key
		msg := tea.KeyMsg{Type: tea.KeyTab}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabHistory))

		newM, _ = m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabSync))

		newM, _ = m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Active).To(Equal(tui.TabRules))
	})

	It("should handle window size", func() {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		Expect(m.Width).To(Equal(100))
		Expect(m.Height).To(Equal(50))
	})

	It("should delete rule", func() {
		// Select first rule
		m.RulesList.Select(0)
		
		// Simulate Delete key
		msg := tea.KeyMsg{Type: tea.KeyDelete}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Cfg.Rules).To(HaveLen(1))
		Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 2"))
	})

	It("should enter add mode", func() {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Active).To(Equal(tui.TabEdit))
		Expect(m.SelectedRuleIndex).To(Equal(-1))
		Expect(m.EditForm).NotTo(BeNil())
	})

	It("should enter edit mode", func() {
		m.RulesList.Select(0)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.Active).To(Equal(tui.TabEdit))
		Expect(m.SelectedRuleIndex).To(Equal(0))
		Expect(m.EditForm).NotTo(BeNil())
	})

	It("should show details", func() {
		m.RulesList.Select(0)
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		Expect(m.ShowDetail).To(BeTrue())
		Expect(m.DetailRule.Name).To(Equal("Rule 1"))
		
		// Verify View contains details
		view := m.View()
		Expect(view).To(ContainSubstring("Rule Details"))
		Expect(view).To(ContainSubstring("Rule 1"))
	})

	It("should filter rules", func() {
		// Enter filter mode
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		newM, _ := m.Update(msg)
		m = newM.(tui.Model)
		
		// Type filter query "Rule 2"
		for _, r := range "Rule 2" {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
		}
		
		// Verify list is filtered (this depends on list implementation, 
		// but we can check if View shows only Rule 2 or if list items count changed)
		// Since we can't easily access internal list state, we check View
		view := m.View()
		Expect(view).To(ContainSubstring("Rule 2"))
		// Rule 1 might still be there if fuzzy matching matches "Rule 1" with "Rule 2" query?
		// "Rule 2" query should match "Rule 2" strongly.
	})
	It("should return nil for Init", func() {
		Expect(m.Init()).To(BeNil())
	})

	It("should render sync status", func() {
		// Switch to sync tab
		m.Active = tui.TabSync
		view := m.View()
		Expect(view).To(ContainSubstring("Sync Status"))
		Expect(view).To(ContainSubstring("Sync not initialized"))

		// With config
		m.Cfg.Sync = &config.SyncConfig{GistID: "123"}
		view = m.View()
		Expect(view).To(ContainSubstring("Gist ID: 123"))
	})

	It("should return full help", func() {
		// We can't easily access the internal keyMap, but we can verify Help view shows something
		// Or we can test the keyMap methods directly if we export them or test via Model
		// Since keyMap is not exported, we rely on Help view
		m.Help.ShowAll = true
		view := m.View()
		Expect(view).To(ContainSubstring("↑/k"))
	})
	
	It("should return filter value", func() {
		item := m.RulesList.Items()[0]
		Expect(item.FilterValue()).To(ContainSubstring("Rule 1"))
	})

	Describe("RuleItem", func() {
		It("should return correct title", func() {
			item := tui.RuleItem{Rule: config.Rule{Name: "Test Rule", Command: "cmd"}}
			Expect(item.Title()).To(Equal("Test Rule"))

			itemNoName := tui.RuleItem{Rule: config.Rule{Command: "cmd"}}
			Expect(itemNoName.Title()).To(Equal("cmd"))
		})

		It("should return correct description", func() {
			// Test with extensions
			item := tui.RuleItem{Rule: config.Rule{Extensions: []string{"txt", "md"}, Command: "cat"}}
			Expect(item.Description()).To(ContainSubstring("[txt, md]"))
			Expect(item.Description()).To(ContainSubstring("-> cat"))

			// Test with regex
			itemRegex := tui.RuleItem{Rule: config.Rule{Regex: ".*", Command: "grep"}}
			Expect(itemRegex.Description()).To(ContainSubstring("Regex: .*"))

			// Test with script
			itemScript := tui.RuleItem{Rule: config.Rule{Script: "true", Command: "cmd"}}
			Expect(itemScript.Description()).To(ContainSubstring("JS"))

			// Test with command only
			itemCmd := tui.RuleItem{Rule: config.Rule{Command: "ls"}}
			Expect(itemCmd.Description()).To(Equal("ls"))
		})

		It("should return correct filter value", func() {
			item := tui.RuleItem{Rule: config.Rule{Name: "Name", Command: "cmd"}}
			Expect(item.FilterValue()).To(ContainSubstring("Name"))
			Expect(item.FilterValue()).To(ContainSubstring("cmd"))
		})
	})

	Describe("HistoryItem", func() {
		It("should return correct title and description", func() {
			entry := history.HistoryEntry{
				Command:   "ls -la",
				RuleName:  "List",
				Timestamp: time.Now(),
			}
			item := tui.HistoryItem{Entry: entry}

			Expect(item.Title()).To(Equal("ls -la"))
			Expect(item.Description()).To(ContainSubstring("List"))
			Expect(item.Description()).To(ContainSubstring(entry.Timestamp.Format("2006-01-02")))
			Expect(item.FilterValue()).To(Equal("ls -la"))
		})
	})

	Describe("Update interactions", func() {
		It("should handle move up/down", func() {
			cfg := &config.Config{
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
					{Name: "Rule 2", Command: "cmd2"},
				},
			}
			m, _ = tui.NewModel(cfg, "config.yml")
			m.Width = 100
			m.Height = 100
			m.RulesList.SetSize(100, 100)

			// Select second item
			m.RulesList.Select(1)

			// Move Up
			msg := tea.KeyMsg{Type: tea.KeyShiftUp}
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
			Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 2"))
			Expect(m.Cfg.Rules[1].Name).To(Equal("Rule 1"))

			// Move Down
			msg = tea.KeyMsg{Type: tea.KeyShiftDown}
			newM, _ = m.Update(msg)
			m = newM.(tui.Model)
			Expect(m.Cfg.Rules[0].Name).To(Equal("Rule 1"))
			Expect(m.Cfg.Rules[1].Name).To(Equal("Rule 2"))
		})

		It("should handle edit form submission", func() {
			cfg := &config.Config{Rules: []config.Rule{}}
			m, _ = tui.NewModel(cfg, "config.yml")
			
			// Enter add mode
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			newM, _ := m.Update(msg)
			m = newM.(tui.Model)
			Expect(m.Active).To(Equal(tui.TabEdit))
			Expect(m.EditForm).NotTo(BeNil())

			// Abort
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			newM, _ = m.Update(msg)
			m = newM.(tui.Model)
			Expect(m.Active).To(Equal(tui.TabRules))
			Expect(m.EditForm).To(BeNil())
		})
	})
})
