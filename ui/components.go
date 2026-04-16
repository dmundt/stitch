package ui

import (
	"html/template"
	"strings"

	"github.com/dmundt/stitch/htmx"
)

// Component is the object-oriented rendering contract for stitch UI elements.
// Each component encapsulates its own data and can render itself to HTML.
type Component interface {
	HTML() string
}

// HTML() method implementations.

func (c *AlertComponent) HTML() string {
	tone := c.Tone
	if tone == "" {
		tone = "info"
	}
	return execute("alert", &AlertComponent{Text: c.Text, Tone: tone})
}

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

func (c *ArticleComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *ArticleComponent) HTML() string {
	return execute("article", articleData{
		Title:    c.Title,
		Children: renderChildrenHTML(c.Children),
	})
}

func (c *BadgeComponent) HTML() string {
	tone := c.Tone
	if tone == "" {
		tone = "default"
	}
	return execute("badge", &BadgeComponent{Text: c.Text, Tone: tone})
}

func (c *BlockquoteComponent) HTML() string {
	return execute("blockquote", c)
}

func (c *BreadcrumbsComponent) HTML() string {
	return execute("breadcrumbs", c)
}

func (c *ButtonComponent) HTML() string {
	kind := c.Kind
	if kind == "" {
		kind = "default"
	}
	return execute("button", &ButtonComponent{Text: c.Text, Kind: kind})
}

func (c *CardComponent) HTML() string {
	return execute("card", c)
}

func (c *CheckboxComponent) HTML() string {
	return execute("checkbox", c)
}

func (c *ClusterComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

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

func (c *CodeBlockComponent) HTML() string {
	return execute("codeblock", c)
}

func (c *ColumnComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

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

func (c *ContainerComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *ContainerComponent) HTML() string {
	return execute("container", childrenData{Children: renderChildrenHTML(c.Children)})
}

func (c *ContainerFluidComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *ContainerFluidComponent) HTML() string {
	return execute("containerfluid", childrenData{Children: renderChildrenHTML(c.Children)})
}

func (c *DescriptionListComponent) HTML() string {
	return execute("descriptionlist", c)
}

func (c *DetailsComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *DetailsComponent) HTML() string {
	return execute("details", detailsData{
		Summary:  c.Summary,
		Children: renderChildrenHTML(c.Children),
	})
}

func (c *FieldsetComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *FieldsetComponent) HTML() string {
	return execute("fieldset", fieldsetData{
		Legend:   c.Legend,
		Children: renderChildrenHTML(c.Children),
	})
}

func (c *FormComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *FormComponent) HTML() string {
	return execute("form", formData{
		Action:   c.Action,
		Method:   c.Method,
		Children: renderChildrenHTML(c.Children),
	})
}

func (c *FragmentComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *FragmentComponent) HTML() string {
	return strings.Join(renderChildren(c.Children), "")
}

func (c *GridComponent) Add(items []Component) {
	c.Items = append(c.Items, items...)
}

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

func (c *GridItemComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

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

func (c *HeroComponent) Add(actions []Component) {
	c.Actions = append(c.Actions, actions...)
}

func (c *HeroComponent) HTML() string {
	return execute("hero", heroData{
		Title:    c.Title,
		Subtitle: c.Subtitle,
		Actions:  renderChildrenHTML(c.Actions),
	})
}

func (c *HorizontalRuleComponent) HTML() string {
	return execute("horizontalrule", nil)
}

func (c *ImageComponent) HTML() string {
	return execute("image", c)
}

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

func (c *InteractiveActionComponent) HTML() string {
	return htmx.Button(c.Text, c.Kind, c.Interaction.toHTMXAttrs())
}

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

func (c *ListComponent) HTML() string {
	return execute("list", c)
}

func (c *NavComponent) HTML() string {
	return execute("nav", c)
}

func (c *OrderedListComponent) HTML() string {
	return execute("orderedlist", c)
}

func (c *PaginationComponent) HTML() string {
	return execute("pagination", c)
}

func (c *ParagraphComponent) HTML() string {
	return execute("paragraph", c)
}

func (c *RadioComponent) HTML() string {
	return execute("radio", c)
}

func (c *RowComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *RowComponent) HTML() string {
	return execute("row", childrenData{Children: renderChildrenHTML(c.Children)})
}

func (c *SectionComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

func (c *SectionComponent) HTML() string {
	return execute("section", sectionData{
		Title:    c.Title,
		Children: renderChildrenHTML(c.Children),
	})
}

func (c *SelectComponent) HTML() string {
	return execute("select", c)
}

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

func (c *StackComponent) Add(children []Component) {
	c.Children = append(c.Children, children...)
}

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

func (c *TableComponent) HTML() string {
	return execute("table", c)
}

func (c *TextAreaComponent) HTML() string {
	return execute("textarea", c)
}

func (c *ThemeToggleComponent) HTML() string {
	return execute("themetoggle", nil)
}

// Constructors.

func NewAction(text, kind string) *ButtonComponent {
	return NewButton(text, kind)
}

func NewAlert(text, tone string) *AlertComponent {
	return &AlertComponent{Text: text, Tone: tone}
}

func NewAppShell(sidebar, content Component) *AppShellComponent {
	return &AppShellComponent{Sidebar: sidebar, Content: content}
}

func NewArticle(title string, children []Component) *ArticleComponent {
	return &ArticleComponent{Title: title, Children: cloneComponents(children)}
}

func NewBadge(text, tone string) *BadgeComponent {
	return &BadgeComponent{Text: text, Tone: tone}
}

func NewBlockquote(text, cite string) *BlockquoteComponent {
	return &BlockquoteComponent{Text: text, Cite: cite}
}

func NewBreadcrumbs(items []BreadcrumbItem) *BreadcrumbsComponent {
	return &BreadcrumbsComponent{Items: items}
}

func NewButton(text, kind string) *ButtonComponent {
	return &ButtonComponent{Text: text, Kind: kind}
}

func NewCard(title, body string) *CardComponent {
	return &CardComponent{Title: title, Body: body}
}

func NewCheckbox(name, value, label string, checked bool) *CheckboxComponent {
	return &CheckboxComponent{Name: name, Value: value, Label: label, Checked: checked}
}

func NewCluster(extraClass string, children []Component) *ClusterComponent {
	return &ClusterComponent{ExtraClass: extraClass, Children: cloneComponents(children)}
}

func NewCodeBlock(code string) *CodeBlockComponent {
	return &CodeBlockComponent{Code: code}
}

func NewColumn(sizeClass string, children []Component) *ColumnComponent {
	return &ColumnComponent{SizeClass: sizeClass, Children: cloneComponents(children)}
}

func NewContainer(children []Component) *ContainerComponent {
	return &ContainerComponent{Children: cloneComponents(children)}
}

func NewContainerFluid(children []Component) *ContainerFluidComponent {
	return &ContainerFluidComponent{Children: cloneComponents(children)}
}

func NewDataGrid(headers []string, rows [][]string) *TableComponent {
	return NewTable(headers, rows)
}

func NewDataGridWithClass(className string, headers []string, rows [][]string) *TableComponent {
	return NewTableWithClass(className, headers, rows)
}

func NewDescriptionList(items []DescriptionItem) *DescriptionListComponent {
	return &DescriptionListComponent{Items: items}
}

func NewDetails(summary string, children []Component) *DetailsComponent {
	return &DetailsComponent{Summary: summary, Children: cloneComponents(children)}
}

func NewFieldset(legend string, children []Component) *FieldsetComponent {
	return &FieldsetComponent{Legend: legend, Children: cloneComponents(children)}
}

func NewForm(action, method string, children []Component) *FormComponent {
	return &FormComponent{Action: action, Method: method, Children: cloneComponents(children)}
}

func NewFragment(children []Component) *FragmentComponent {
	return &FragmentComponent{Children: cloneComponents(children)}
}

func NewGrid(columnsClass string, items []Component) *GridComponent {
	return &GridComponent{ColumnsClass: columnsClass, Items: cloneComponents(items)}
}

func NewGridItem(spanClass string, children []Component) *GridItemComponent {
	return &GridItemComponent{SpanClass: spanClass, Children: cloneComponents(children)}
}

func NewHeading(level int, text string) *HeadingComponent {
	return &HeadingComponent{Level: level, Text: text}
}

func NewHero(title, subtitle string, actions []Component) *HeroComponent {
	return &HeroComponent{Title: title, Subtitle: subtitle, Actions: cloneComponents(actions)}
}

func NewHorizontalRule() *HorizontalRuleComponent {
	return &HorizontalRuleComponent{}
}

func NewImage(src, alt string) *ImageComponent {
	return &ImageComponent{Src: src, Alt: alt}
}

func NewInput(label, name, placeholder string) *InputComponent {
	return &InputComponent{Label: label, Name: name, Placeholder: placeholder}
}

func NewInteractiveAction(text, kind string, interaction Interaction) *InteractiveActionComponent {
	return &InteractiveActionComponent{Text: text, Kind: kind, Interaction: interaction}
}

func NewInteractiveMenu(links []InteractiveMenuLink) *InteractiveMenuComponent {
	out := make([]InteractiveMenuLink, len(links))
	copy(out, links)
	return &InteractiveMenuComponent{Links: out}
}

func NewList(items []string) *ListComponent {
	return &ListComponent{Items: items}
}

func NewMenu(links []NavLink) *NavComponent {
	return NewNav(links)
}

func NewNav(links []NavLink) *NavComponent {
	return &NavComponent{Links: links}
}

func NewOrderedList(items []string) *OrderedListComponent {
	return &OrderedListComponent{Items: items}
}

func NewPagination(items []PageItem) *PaginationComponent {
	return &PaginationComponent{Items: items}
}

func NewPanel(title string, children []Component) *ArticleComponent {
	return NewArticle(title, children)
}

func NewParagraph(text string) *ParagraphComponent {
	return &ParagraphComponent{Text: text}
}

func NewRadio(name, value, label string, checked bool) *RadioComponent {
	return &RadioComponent{Name: name, Value: value, Label: label, Checked: checked}
}

func NewRow(children []Component) *RowComponent {
	return &RowComponent{Children: cloneComponents(children)}
}

func NewSection(title string, children []Component) *SectionComponent {
	return &SectionComponent{Title: title, Children: cloneComponents(children)}
}

func NewSelect(label, name string, options []SelectOption) *SelectComponent {
	return &SelectComponent{Label: label, Name: name, Options: options}
}

func NewSidebarLayout(sidebar, content Component) *SidebarLayoutComponent {
	return &SidebarLayoutComponent{Sidebar: sidebar, Content: content}
}

func NewSplit(primary, secondary Component) *SplitComponent {
	return &SplitComponent{Primary: primary, Secondary: secondary}
}

func NewStack(extraClass string, children []Component) *StackComponent {
	return &StackComponent{ExtraClass: extraClass, Children: cloneComponents(children)}
}

func NewStatus(text, tone string) *AlertComponent {
	return NewAlert(text, tone)
}

func NewTable(headers []string, rows [][]string) *TableComponent {
	return &TableComponent{Headers: headers, Rows: rows}
}

func NewTableWithClass(className string, headers []string, rows [][]string) *TableComponent {
	return &TableComponent{ClassName: className, Headers: headers, Rows: rows}
}

func NewText(text string) *ParagraphComponent {
	return NewParagraph(text)
}
func NewTextArea(label, name, placeholder string) *TextAreaComponent {
	return &TextAreaComponent{Label: label, Name: name, Placeholder: placeholder}
}
func NewThemeToggle() *ThemeToggleComponent {
	return &ThemeToggleComponent{}
}

func NewToolbar(children []Component) *ClusterComponent {
	return NewCluster("toolbar", children)
}

func NewView(title string, children []Component) *SectionComponent {
	return NewSection(title, children)
}

func NewWorkspace(sidebar, content Component) *AppShellComponent {
	return NewAppShell(sidebar, content)
}

