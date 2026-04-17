package ui

import (
	"html"
	"html/template"
	"strings"

	"github.com/dmundt/stitch/htmx"
)

// Component is the object-oriented rendering contract for Stitch UI elements.
//
// Implementations should return valid HTML fragments suitable for insertion
// into html/template output. Stitch treats component output as trusted fragment
// HTML when composing nested structures.
type Component interface {
	HTML() string
}

// WithID wraps any component and injects an id attribute on the root element.
// If id is blank, the original component is returned unchanged.
func WithID(id string, child Component) Component {
	id = strings.TrimSpace(id)
	if id == "" || child == nil {
		return child
	}
	return &idComponent{ID: id, Child: child}
}

type idComponent struct {
	ID    string
	Child Component
}

func (c *idComponent) HTML() string {
	return injectIDAttribute(c.Child.HTML(), c.ID)
}

func injectIDAttribute(fragment, id string) string {
	if fragment == "" || strings.TrimSpace(id) == "" {
		return fragment
	}
	search := 0
	for search < len(fragment) {
		idx := strings.Index(fragment[search:], "<")
		if idx < 0 {
			return fragment
		}
		start := search + idx
		if start+1 >= len(fragment) {
			return fragment
		}
		next := fragment[start+1]
		if next == '/' || next == '!' || next == '?' {
			search = start + 1
			continue
		}

		end := findTagEnd(fragment, start+1)
		if end < 0 {
			return fragment
		}

		tag := fragment[start:end]
		if hasIDAttribute(tag) {
			return fragment
		}

		insertAt := end
		if end > start && fragment[end-1] == '/' {
			insertAt = end - 1
		}
		attr := ` id="` + html.EscapeString(id) + `"`
		return fragment[:insertAt] + attr + fragment[insertAt:]
	}
	return fragment
}

func findTagEnd(s string, from int) int {
	quote := byte(0)
	for i := from; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == '>' {
			return i
		}
	}
	return -1
}

func hasIDAttribute(openTag string) bool {
	lower := strings.ToLower(openTag)
	for i := 0; i < len(lower); i++ {
		if lower[i] != 'i' {
			continue
		}
		if i+1 >= len(lower) || lower[i+1] != 'd' {
			continue
		}
		if i > 0 {
			prev := lower[i-1]
			if (prev >= 'a' && prev <= 'z') || prev == '-' || prev == '_' {
				continue
			}
		}
		j := i + 2
		for j < len(lower) && (lower[j] == ' ' || lower[j] == '\t' || lower[j] == '\n' || lower[j] == '\r') {
			j++
		}
		if j < len(lower) && lower[j] == '=' {
			return true
		}
	}
	return false
}

// HTML renders alert markup with a default "info" tone when unset.
func (c *AlertComponent) HTML() string {
	tone := c.Tone
	if tone == "" {
		tone = "info"
	}
	return execute("alert", &AlertComponent{Text: c.Text, Tone: tone})
}

// HTML renders the app shell layout.
//
// Child component output is treated as trusted fragment HTML.
func (c *AppShellComponent) HTML() string {
	var sidebar template.HTML
	if c.Sidebar != nil {
		sidebar = template.HTML(c.Sidebar.HTML())
	}
	var content template.HTML
	if c.Content != nil {
		content = template.HTML(c.Content.HTML())
	}
	return execute("appshell", appShellData{
		HasSidebar: c.Sidebar != nil,
		Sidebar:    sidebar,
		Content:    content,
	})
}

// Add appends child components in order.
func (c *ArticleComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders article markup with its child components.
func (c *ArticleComponent) HTML() string {
	return execute("article", articleData{
		Title:    c.Title,
		Children: renderChildrenHTML(c.Children),
	})
}

// HTML renders badge markup with a default "default" tone when unset.
func (c *BadgeComponent) HTML() string {
	tone := c.Tone
	if tone == "" {
		tone = "default"
	}
	return execute("badge", &BadgeComponent{Text: c.Text, Tone: tone})
}

// HTML renders blockquote markup.
func (c *BlockquoteComponent) HTML() string {
	return execute("blockquote", c)
}

// HTML renders breadcrumbs markup.
func (c *BreadcrumbsComponent) HTML() string {
	return execute("breadcrumbs", c)
}

// HTML renders button markup with a default "default" kind when unset.
func (c *ButtonComponent) HTML() string {
	kind := c.Kind
	if kind == "" {
		kind = "default"
	}
	return execute("button", &ButtonComponent{Text: c.Text, Kind: kind})
}

// HTML renders card markup.
func (c *CardComponent) HTML() string {
	return execute("card", c)
}

// HTML renders checkbox markup.
func (c *CheckboxComponent) HTML() string {
	return execute("checkbox", c)
}

// Add appends child components in order.
func (c *ClusterComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders cluster layout markup.
func (c *ClusterComponent) HTML() string {
	class := "cluster"
	if c.ExtraClass != "" {
		class += " " + c.ExtraClass
	}
	return execute("cluster", classChildrenData{
		Class:    class,
		Children: renderChildrenHTML(c.Children),
	})
}

// HTML renders code block markup.
func (c *CodeBlockComponent) HTML() string {
	return execute("codeblock", c)
}

// Add appends child components in order.
func (c *ColumnComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders column layout markup.
func (c *ColumnComponent) HTML() string {
	class := "column"
	if c.SizeClass != "" {
		class += " " + c.SizeClass
	}
	return execute("column", classChildrenData{
		Class:    class,
		Children: renderChildrenHTML(c.Children),
	})
}

// Add appends child components in order.
func (c *ContainerComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders container markup.
func (c *ContainerComponent) HTML() string {
	return execute("container", childrenData{Children: renderChildrenHTML(c.Children)})
}

// Add appends child components in order.
func (c *ContainerFluidComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders fluid container markup.
func (c *ContainerFluidComponent) HTML() string {
	return execute("containerfluid", childrenData{Children: renderChildrenHTML(c.Children)})
}

// HTML renders description list markup.
func (c *DescriptionListComponent) HTML() string {
	return execute("descriptionlist", c)
}

// Add appends child components in order.
func (c *DetailsComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders details markup.
func (c *DetailsComponent) HTML() string {
	return execute("details", detailsData{
		Summary:  c.Summary,
		Children: renderChildrenHTML(c.Children),
	})
}

// Add appends child components in order.
func (c *FieldsetComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders fieldset markup.
func (c *FieldsetComponent) HTML() string {
	return execute("fieldset", fieldsetData{
		Legend:   c.Legend,
		Children: renderChildrenHTML(c.Children),
	})
}

// Add appends child components in order.
func (c *FormComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders form markup.
func (c *FormComponent) HTML() string {
	return execute("form", formData{
		Action:   c.Action,
		Method:   c.Method,
		Children: renderChildrenHTML(c.Children),
	})
}

// Add appends child components in order.
func (c *FragmentComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders all child component HTML with no wrapper element.
func (c *FragmentComponent) HTML() string {
	return strings.Join(renderChildren(c.Children), "")
}

// Add appends grid items in order.
func (c *GridComponent) Add(items []Component) {
	c.Items = append(c.Items, items...)
}

// HTML renders grid markup.
func (c *GridComponent) HTML() string {
	class := "grid"
	if c.ColumnsClass != "" {
		class += " " + c.ColumnsClass
	}
	return execute("grid", classChildrenData{
		Class:    class,
		Children: renderChildrenHTML(c.Items),
	})
}

// Add appends child components in order.
func (c *GridItemComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders grid item markup.
func (c *GridItemComponent) HTML() string {
	class := "grid-item"
	if c.SpanClass != "" {
		class += " " + c.SpanClass
	}
	return execute("griditem", classChildrenData{
		Class:    class,
		Children: renderChildrenHTML(c.Children),
	})
}

// HTML renders heading markup.
//
// Level is clamped to the inclusive range 1..6.
func (c *HeadingComponent) HTML() string {
	level := c.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return execute("heading", &HeadingComponent{Level: level, Text: c.Text})
}

// Add appends action components in order.
func (c *HeroComponent) Add(actions []Component) {
	c.Actions = append(c.Actions, actions...)
}

// HTML renders hero markup.
func (c *HeroComponent) HTML() string {
	return execute("hero", heroData{
		Title:    c.Title,
		Subtitle: c.Subtitle,
		Actions:  renderChildrenHTML(c.Actions),
	})
}

// HTML renders a horizontal rule.
func (c *HorizontalRuleComponent) HTML() string {
	return execute("horizontalrule", nil)
}

// HTML renders image markup.
func (c *ImageComponent) HTML() string {
	return execute("image", c)
}

// HTML renders input markup.
func (c *InputComponent) HTML() string {
	return execute("input", c)
}

func (i Interaction) toHTMXAttrs() htmx.Attrs {
	return htmx.Attrs{
		Boost:   i.Boost,
		Delete:  i.Delete,
		Get:     i.Get,
		Post:    i.Post,
		PushURL: i.PushURL,
		Put:     i.Put,
		Select:  i.Select,
		Swap:    i.Swap,
		Target:  i.Target,
		Trigger: i.Trigger,
	}
}

// HTML renders an HTMX-enabled action button.
//
// Interaction fields are mapped to hx-* attributes.
func (c *InteractiveActionComponent) HTML() string {
	return htmx.Button(c.Text, c.Kind, c.Interaction.toHTMXAttrs())
}

// HTML renders an HTMX-enabled navigation menu.
//
// Each link interaction is mapped to hx-* attributes.
func (c *InteractiveMenuComponent) HTML() string {
	links := make([]htmx.NavLink, 0, len(c.Links))
	for _, link := range c.Links {
		links = append(links, htmx.NavLink{
			Label: link.Label,
			Href:  link.Href,
			HX:    link.Interaction.toHTMXAttrs(),
		})
	}
	return htmx.Nav(links)
}

// HTML renders list markup.
func (c *ListComponent) HTML() string {
	return execute("list", c)
}

// HTML renders navigation markup.
func (c *NavComponent) HTML() string {
	return execute("nav", c)
}

// HTML renders ordered list markup.
func (c *OrderedListComponent) HTML() string {
	return execute("orderedlist", c)
}

// HTML renders pagination markup.
func (c *PaginationComponent) HTML() string {
	return execute("pagination", c)
}

// HTML renders paragraph markup.
func (c *ParagraphComponent) HTML() string {
	return execute("paragraph", c)
}

// HTML renders radio input markup.
func (c *RadioComponent) HTML() string {
	return execute("radio", c)
}

// Add appends child components in order.
func (c *RowComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders row markup.
func (c *RowComponent) HTML() string {
	return execute("row", childrenData{Children: renderChildrenHTML(c.Children)})
}

// Add appends child components in order.
func (c *SectionComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders section markup.
func (c *SectionComponent) HTML() string {
	return execute("section", sectionData{
		Title:    c.Title,
		Children: renderChildrenHTML(c.Children),
	})
}

// HTML renders select markup.
func (c *SelectComponent) HTML() string {
	return execute("select", c)
}

// HTML renders sidebar layout markup.
//
// Child component output is treated as trusted fragment HTML.
func (c *SidebarLayoutComponent) HTML() string {
	var sidebar, content template.HTML
	if c.Sidebar != nil {
		sidebar = template.HTML(c.Sidebar.HTML())
	}
	if c.Content != nil {
		content = template.HTML(c.Content.HTML())
	}
	return execute("sidebarLayout", sidebarData{Sidebar: sidebar, Content: content})
}

// HTML renders split layout markup.
//
// Child component output is treated as trusted fragment HTML.
func (c *SplitComponent) HTML() string {
	var primary, secondary template.HTML
	if c.Primary != nil {
		primary = template.HTML(c.Primary.HTML())
	}
	if c.Secondary != nil {
		secondary = template.HTML(c.Secondary.HTML())
	}
	return execute("split", splitData{Primary: primary, Secondary: secondary})
}

// Add appends child components in order.
func (c *StackComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

// HTML renders stack layout markup.
func (c *StackComponent) HTML() string {
	class := "stack"
	if c.ExtraClass != "" {
		class += " " + c.ExtraClass
	}
	return execute("stack", classChildrenData{
		Class:    class,
		Children: renderChildrenHTML(c.Children),
	})
}

// HTML renders table markup.
func (c *TableComponent) HTML() string {
	return execute("table", c)
}

// HTML renders textarea markup.
func (c *TextAreaComponent) HTML() string {
	return execute("textarea", c)
}

// HTML renders a theme toggle control.
func (c *ThemeToggleComponent) HTML() string {
	return execute("themetoggle", nil)
}

// NewAction is a GUI-style alias for NewButton.
//
// It exists for action-oriented naming and has identical behavior.
func NewAction(text, kind string) *ButtonComponent {
	return NewButton(text, kind)
}

// NewAlert creates an AlertComponent for inline feedback messages.
func NewAlert(text, tone string) *AlertComponent {
	return &AlertComponent{Text: text, Tone: tone}
}

// NewAppShell creates an AppShellComponent with optional sidebar and content.
func NewAppShell(sidebar, content Component) *AppShellComponent {
	return &AppShellComponent{Sidebar: sidebar, Content: content}
}

// NewArticle creates an ArticleComponent.
//
// The children slice is shallow-copied.
func NewArticle(title string, children []Component) *ArticleComponent {
	return &ArticleComponent{Title: title, Children: cloneComponents(children)}
}

// NewBadge creates a BadgeComponent for concise status labels.
func NewBadge(text, tone string) *BadgeComponent {
	return &BadgeComponent{Text: text, Tone: tone}
}

// NewBlockquote creates a BlockquoteComponent with optional citation.
func NewBlockquote(text, cite string) *BlockquoteComponent {
	return &BlockquoteComponent{Text: text, Cite: cite}
}

// NewBreadcrumbs creates a BreadcrumbsComponent from breadcrumb items.
func NewBreadcrumbs(items []BreadcrumbItem) *BreadcrumbsComponent {
	return &BreadcrumbsComponent{Items: items}
}

// NewButton creates a ButtonComponent action.
func NewButton(text, kind string) *ButtonComponent {
	return &ButtonComponent{Text: text, Kind: kind}
}

// NewCard creates a CardComponent for concise titled content.
func NewCard(title, body string) *CardComponent {
	return &CardComponent{Title: title, Body: body}
}

// NewCheckbox creates a CheckboxComponent form field.
func NewCheckbox(name, value, label string, checked bool) *CheckboxComponent {
	return &CheckboxComponent{Name: name, Value: value, Label: label, Checked: checked}
}

// NewCluster creates a ClusterComponent.
//
// The children slice is shallow-copied.
func NewCluster(extraClass string, children []Component) *ClusterComponent {
	return &ClusterComponent{ExtraClass: extraClass, Children: cloneComponents(children)}
}

// NewCodeBlock creates a CodeBlockComponent.
func NewCodeBlock(code string) *CodeBlockComponent {
	return &CodeBlockComponent{Code: code}
}

// NewColumn creates a ColumnComponent.
//
// The children slice is shallow-copied.
func NewColumn(sizeClass string, children []Component) *ColumnComponent {
	return &ColumnComponent{SizeClass: sizeClass, Children: cloneComponents(children)}
}

// NewContainer creates a centered ContainerComponent.
//
// The children slice is shallow-copied.
func NewContainer(children []Component) *ContainerComponent {
	return &ContainerComponent{Children: cloneComponents(children)}
}

// NewContainerFluid creates a full-width ContainerFluidComponent.
//
// The children slice is shallow-copied.
func NewContainerFluid(children []Component) *ContainerFluidComponent {
	return &ContainerFluidComponent{Children: cloneComponents(children)}
}

// NewDataGrid is a GUI-style alias for NewTable.
func NewDataGrid(headers []string, rows [][]string) *TableComponent {
	return NewTable(headers, rows)
}

// NewDataGridWithClass is a GUI-style alias for NewTableWithClass.
func NewDataGridWithClass(className string, headers []string, rows [][]string) *TableComponent {
	return NewTableWithClass(className, headers, rows)
}

// NewDescriptionList creates a DescriptionListComponent from term-definition items.
func NewDescriptionList(items []DescriptionItem) *DescriptionListComponent {
	return &DescriptionListComponent{Items: items}
}

// NewDetails creates a DetailsComponent.
//
// The children slice is shallow-copied.
func NewDetails(summary string, children []Component) *DetailsComponent {
	return &DetailsComponent{Summary: summary, Children: cloneComponents(children)}
}

// NewFieldset creates a FieldsetComponent.
//
// The children slice is shallow-copied.
func NewFieldset(legend string, children []Component) *FieldsetComponent {
	return &FieldsetComponent{Legend: legend, Children: cloneComponents(children)}
}

// NewForm creates a FormComponent with action and method attributes.
//
// The children slice is shallow-copied.
func NewForm(action, method string, children []Component) *FormComponent {
	return &FormComponent{Action: action, Method: method, Children: cloneComponents(children)}
}

// NewFragment creates a FragmentComponent.
//
// The children slice is shallow-copied.
func NewFragment(children []Component) *FragmentComponent {
	return &FragmentComponent{Children: cloneComponents(children)}
}

// NewGrid creates a GridComponent.
//
// The items slice is shallow-copied.
func NewGrid(columnsClass string, items []Component) *GridComponent {
	return &GridComponent{ColumnsClass: columnsClass, Items: cloneComponents(items)}
}

// NewGridItem creates a GridItemComponent.
//
// The children slice is shallow-copied.
func NewGridItem(spanClass string, children []Component) *GridItemComponent {
	return &GridItemComponent{SpanClass: spanClass, Children: cloneComponents(children)}
}

// NewHeading creates a HeadingComponent.
//
// Level is normalized during rendering, not during construction.
func NewHeading(level int, text string) *HeadingComponent {
	return &HeadingComponent{Level: level, Text: text}
}

// NewHero creates a HeroComponent with optional action components.
//
// The actions slice is shallow-copied.
func NewHero(title, subtitle string, actions []Component) *HeroComponent {
	return &HeroComponent{Title: title, Subtitle: subtitle, Actions: cloneComponents(actions)}
}

// NewHorizontalRule creates a HorizontalRuleComponent.
func NewHorizontalRule() *HorizontalRuleComponent {
	return &HorizontalRuleComponent{}
}

// NewImage creates an ImageComponent.
func NewImage(src, alt string) *ImageComponent {
	return &ImageComponent{Src: src, Alt: alt}
}

// NewInput creates an InputComponent.
func NewInput(label, name, placeholder string) *InputComponent {
	return &InputComponent{Label: label, Name: name, Placeholder: placeholder}
}

// NewInteractiveAction creates an InteractiveActionComponent.
func NewInteractiveAction(text, kind string, interaction Interaction) *InteractiveActionComponent {
	return &InteractiveActionComponent{Text: text, Kind: kind, Interaction: interaction}
}

// NewInteractiveMenu creates an InteractiveMenuComponent.
//
// The links slice is copied to avoid caller slice reordering side effects.
func NewInteractiveMenu(links []InteractiveMenuLink) *InteractiveMenuComponent {
	out := make([]InteractiveMenuLink, len(links))
	copy(out, links)
	return &InteractiveMenuComponent{Links: out}
}

// NewList creates a ListComponent.
func NewList(items []string) *ListComponent {
	return &ListComponent{Items: items}
}

// NewMenu is a GUI-style alias for NewNav.
//
// It exists for menu-oriented naming and has identical behavior.
func NewMenu(links []NavLink) *NavComponent {
	return NewNav(links)
}

// NewNav creates a NavComponent.
//
// The links slice is retained as provided.
func NewNav(links []NavLink) *NavComponent {
	return &NavComponent{Links: links}
}

// NewOrderedList creates an OrderedListComponent.
func NewOrderedList(items []string) *OrderedListComponent {
	return &OrderedListComponent{Items: items}
}

// NewPagination creates a PaginationComponent.
func NewPagination(items []PageItem) *PaginationComponent {
	return &PaginationComponent{Items: items}
}

// NewPanel is a GUI-style alias for NewArticle.
//
// It exists for panel-oriented naming and has identical behavior.
func NewPanel(title string, children []Component) *ArticleComponent {
	return NewArticle(title, children)
}

// NewParagraph creates a ParagraphComponent.
func NewParagraph(text string) *ParagraphComponent {
	return &ParagraphComponent{Text: text}
}

// NewRadio creates a RadioComponent form field.
func NewRadio(name, value, label string, checked bool) *RadioComponent {
	return &RadioComponent{Name: name, Value: value, Label: label, Checked: checked}
}

// NewRow creates a RowComponent.
//
// The children slice is shallow-copied.
func NewRow(children []Component) *RowComponent {
	return &RowComponent{Children: cloneComponents(children)}
}

// NewSection creates a SectionComponent.
//
// The children slice is shallow-copied.
func NewSection(title string, children []Component) *SectionComponent {
	return &SectionComponent{Title: title, Children: cloneComponents(children)}
}

// NewSelect creates a SelectComponent.
//
// The options slice is retained as provided.
func NewSelect(label, name string, options []SelectOption) *SelectComponent {
	return &SelectComponent{Label: label, Name: name, Options: options}
}

// NewSidebarLayout creates a SidebarLayoutComponent.
func NewSidebarLayout(sidebar, content Component) *SidebarLayoutComponent {
	return &SidebarLayoutComponent{Sidebar: sidebar, Content: content}
}

// NewSplit creates a SplitComponent with primary and secondary regions.
func NewSplit(primary, secondary Component) *SplitComponent {
	return &SplitComponent{Primary: primary, Secondary: secondary}
}

// NewStack creates a StackComponent.
//
// The children slice is shallow-copied.
func NewStack(extraClass string, children []Component) *StackComponent {
	return &StackComponent{ExtraClass: extraClass, Children: cloneComponents(children)}
}

// NewStatus is a GUI-style alias for NewAlert.
//
// It exists for status-oriented naming and has identical behavior.
func NewStatus(text, tone string) *AlertComponent {
	return NewAlert(text, tone)
}

// NewTable creates a TableComponent.
//
// Headers and rows slices are retained as provided.
func NewTable(headers []string, rows [][]string) *TableComponent {
	return &TableComponent{Headers: headers, Rows: rows}
}

// NewTableWithClass creates a TableComponent with an explicit class name.
//
// Headers and rows slices are retained as provided.
func NewTableWithClass(className string, headers []string, rows [][]string) *TableComponent {
	return &TableComponent{ClassName: className, Headers: headers, Rows: rows}
}

// NewText is a GUI-style alias for NewParagraph.
//
// It exists for text-oriented naming and has identical behavior.
func NewText(text string) *ParagraphComponent {
	return NewParagraph(text)
}

// NewTextArea creates a TextAreaComponent.
func NewTextArea(label, name, placeholder string) *TextAreaComponent {
	return &TextAreaComponent{Label: label, Name: name, Placeholder: placeholder}
}

// NewThemeToggle creates a ThemeToggleComponent.
func NewThemeToggle() *ThemeToggleComponent {
	return &ThemeToggleComponent{}
}

// NewToolbar is a GUI-style alias for NewCluster with toolbar class.
//
// It exists for toolbar-oriented naming and has identical behavior.
func NewToolbar(children []Component) *ClusterComponent {
	return NewCluster("toolbar", children)
}

// NewView is a GUI-style alias for NewSection.
//
// It exists for view-oriented naming and has identical behavior.
func NewView(title string, children []Component) *SectionComponent {
	return NewSection(title, children)
}

// NewWorkspace is a GUI-style alias for NewAppShell.
//
// It exists for workspace-oriented naming and has identical behavior.
func NewWorkspace(sidebar, content Component) *AppShellComponent {
	return NewAppShell(sidebar, content)
}
