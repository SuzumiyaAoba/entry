package tui

import (
	"fmt"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/history"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type Tab int

const (
	TabRules Tab = iota
	TabHistory
	TabSync
	TabEdit
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	tabStyle  = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	activeTabStyle = tabStyle.Copy().Border(activeTabBorder, true)

	windowStyle = lipgloss.NewStyle().
			BorderForeground(highlight).
			Padding(1, 0).
			Border(lipgloss.NormalBorder()).
			UnsetBorderTop()
)

// RuleItem adapts config.Rule to list.Item
type RuleItem struct {
	Rule config.Rule
}

func (i RuleItem) Title() string {
	if i.Rule.Name != "" {
		return i.Rule.Name
	}
	return i.Rule.Command
}

func (i RuleItem) Description() string {
	desc := ""
	if len(i.Rule.Extensions) > 0 {
		desc += fmt.Sprintf("[%s] ", strings.Join(i.Rule.Extensions, ", "))
	}
	if i.Rule.Regex != "" {
		desc += fmt.Sprintf("Regex: %s ", i.Rule.Regex)
	}
	if i.Rule.Script != "" {
		desc += "JS "
	}
	if desc == "" {
		desc = i.Rule.Command
	} else {
		desc += "-> " + i.Rule.Command
	}
	return desc
}

func (i RuleItem) FilterValue() string { return i.Title() + " " + i.Description() }

// HistoryItem adapts history.HistoryEntry to list.Item
type HistoryItem struct {
	Entry history.HistoryEntry
}

func (i HistoryItem) Title() string { return i.Entry.Command }
func (i HistoryItem) Description() string {
	return fmt.Sprintf("%s - %s", i.Entry.Timestamp.Format("2006-01-02 15:04:05"), i.Entry.RuleName)
}
func (i HistoryItem) FilterValue() string { return i.Title() }

type keyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Delete   key.Binding
	Edit     key.Binding
	Quit     key.Binding
	Up       key.Binding
	Down     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Edit, k.Delete, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.Quit},
		{k.Up, k.Down, k.Edit, k.Delete},
	}
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "delete rule"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit rule"),
	),
}

type Model struct {
	Cfg         *config.Config
	ConfigPath  string
	History     []history.HistoryEntry
	Active      Tab
	Width       int
	Height      int
	
	RulesList   list.Model
	HistoryList list.Model
	Help        help.Model
	
	// Edit State
	EditForm          *huh.Form
	SelectedRuleIndex int
	EditExtensionsStr string
}

func NewModel(cfg *config.Config, configPath string) (Model, error) {
	hist, _ := history.LoadHistory()

	// Setup Rules List
	ruleItems := lo.Map(cfg.Rules, func(r config.Rule, _ int) list.Item {
		return RuleItem{Rule: r}
	})
	
	rulesList := list.New(ruleItems, list.NewDefaultDelegate(), 0, 0)
	rulesList.Title = "Rules"
	rulesList.SetShowHelp(false)

	// Setup History List
	histItems := lo.Map(hist, func(h history.HistoryEntry, _ int) list.Item {
		return HistoryItem{Entry: h}
	})
	
	historyList := list.New(histItems, list.NewDefaultDelegate(), 0, 0)
	historyList.Title = "History"
	historyList.SetShowHelp(false)

	return Model{
		Cfg:         cfg,
		ConfigPath:  configPath,
		History:     hist,
		Active:      TabRules,
		RulesList:   rulesList,
		HistoryList: historyList,
		Help:        help.New(),
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		// Adjust list sizes
		h, v := windowStyle.GetFrameSize()
		listHeight := msg.Height - v - 5 // 3 for tabs, 2 for help
		listWidth := msg.Width - h

		m.RulesList.SetSize(listWidth, listHeight)
		m.HistoryList.SetSize(listWidth, listHeight)
		m.Help.Width = msg.Width

	case tea.KeyMsg:
		// Global keys (except when editing)
		if m.Active != TabEdit {
			switch {
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, keys.Tab):
				m.Active = (m.Active + 1) % 3 // Cycle through main tabs only
				return m, nil
			case key.Matches(msg, keys.ShiftTab):
				if m.Active == 0 {
					m.Active = TabSync
				} else {
					m.Active--
				}
				return m, nil
			}
		}

		// Context specific keys
		if m.Active == TabRules {
			switch {
			case key.Matches(msg, keys.Delete):
				if len(m.Cfg.Rules) > 0 {
					index := m.RulesList.Index()
					if index >= 0 && index < len(m.Cfg.Rules) {
						// Remove from slice
						m.Cfg.Rules = append(m.Cfg.Rules[:index], m.Cfg.Rules[index+1:]...)
						
						// Save config
						if err := config.SaveConfig(m.ConfigPath, m.Cfg); err != nil {
							// Handle error
						}

						// Remove from list
						m.RulesList.RemoveItem(index)
					}
				}
			case key.Matches(msg, keys.Edit):
				if len(m.Cfg.Rules) > 0 {
					index := m.RulesList.Index()
					if index >= 0 && index < len(m.Cfg.Rules) {
						m.SelectedRuleIndex = index
						// Bind directly to the slice element
						rule := &m.Cfg.Rules[index]
						
						// Prepare temporary variable for extensions
						m.EditExtensionsStr = strings.Join(rule.Extensions, ", ")
						
						// Create Form
						m.EditForm = huh.NewForm(
							huh.NewGroup(
								huh.NewInput().
									Title("Name").
									Value(&rule.Name),
								huh.NewInput().
									Title("Command").
									Value(&rule.Command),
								huh.NewInput().
									Title("Extensions (comma separated)").
									Value(&m.EditExtensionsStr), // Bind to the model's temporary string
								huh.NewInput().
									Title("Regex").
									Value(&rule.Regex),
								huh.NewConfirm().
									Title("Terminal").
									Value(&rule.Terminal),
								huh.NewConfirm().
									Title("Background").
									Value(&rule.Background),
							),
						).WithTheme(huh.ThemeCharm())
						
						m.EditForm.Init()
						m.Active = TabEdit
					}
				}
			}
		}
	}

	// Delegate to active component
	switch m.Active {
	case TabRules:
		m.RulesList, cmd = m.RulesList.Update(msg)
		cmds = append(cmds, cmd)
	case TabHistory:
		m.HistoryList, cmd = m.HistoryList.Update(msg)
		cmds = append(cmds, cmd)
	case TabEdit:
		if m.EditForm != nil {
			form, cmd := m.EditForm.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.EditForm = f
			}
			cmds = append(cmds, cmd)

			if m.EditForm.State == huh.StateCompleted {
				// Parse extensions back to slice from the temporary string
				rule := &m.Cfg.Rules[m.SelectedRuleIndex]
				if m.EditExtensionsStr == "" {
					rule.Extensions = []string{}
				} else {
					rule.Extensions = lo.Map(strings.Split(m.EditExtensionsStr, ","), func(item string, _ int) string {
						return strings.TrimSpace(item)
					})
				}

				// Save config
				if err := config.SaveConfig(m.ConfigPath, m.Cfg); err != nil {
					// Handle error
				}
				
				// Update list item
				m.RulesList.SetItem(m.SelectedRuleIndex, RuleItem{Rule: *rule})

				m.Active = TabRules
				m.EditForm = nil
				m.EditExtensionsStr = "" // Clear temporary string
			} else if m.EditForm.State == huh.StateAborted {
				m.Active = TabRules
				m.EditForm = nil
				m.EditExtensionsStr = "" // Clear temporary string
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.Active == TabEdit {
		return windowStyle.Width(m.Width - windowStyle.GetHorizontalFrameSize()).Render(m.EditForm.View())
	}

	doc := strings.Builder{}

	// Tabs
	var tabs []string
	for i, t := range []string{"Rules", "History", "Sync"} {
		if m.Active == Tab(i) {
			tabs = append(tabs, activeTabStyle.Render(t))
		} else {
			tabs = append(tabs, tabStyle.Render(t))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	// Content
	var content string
	switch m.Active {
	case TabRules:
		content = m.RulesList.View()
	case TabHistory:
		content = m.HistoryList.View()
	case TabSync:
		content = m.renderSync()
	}

	doc.WriteString(windowStyle.Width(m.Width - windowStyle.GetHorizontalFrameSize()).Render(content))
	doc.WriteString("\n")
	
	// Help
	doc.WriteString(m.Help.View(keys))

	return doc.String()
}

func (m Model) renderSync() string {
	s := strings.Builder{}
	s.WriteString(titleStyle.Render("Sync Status"))
	s.WriteString("\n\n")
	
	if m.Cfg.Sync != nil && m.Cfg.Sync.GistID != "" {
		s.WriteString(fmt.Sprintf("Gist ID: %s\n", m.Cfg.Sync.GistID))
		if m.Cfg.Sync.Token != "" {
			s.WriteString("Token: (Stored)\n")
		} else {
			s.WriteString("Token: (Not stored)\n")
		}
	} else {
		s.WriteString("Sync not initialized.\nRun ':config sync init' to setup.")
	}
	
	return s.String()
}
