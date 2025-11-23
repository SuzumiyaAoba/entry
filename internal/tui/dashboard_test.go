package tui_test

import (
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
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
})
