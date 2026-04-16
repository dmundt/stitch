package ui

import "html/template"

// Supporting types used as template data fields and public API types.

type BreadcrumbItem struct {
Label   string
Href    string
Current bool
}

type DescriptionItem struct {
Term       string
Definition string
}

type NavLink struct {
Label string
Href  string
}

type PageItem struct {
Label    string
Href     string
Current  bool
Disabled bool
}

type SelectOption struct {
Value    string
Label    string
Selected bool
}

// Component struct definitions.

type AlertComponent struct {
Text string
Tone string
}

type AppShellComponent struct {
Sidebar Component
Content Component
}

type ArticleComponent struct {
Title    string
Children []Component
}

type BadgeComponent struct {
Text string
Tone string
}

type BlockquoteComponent struct {
Text string
Cite string
}

type BreadcrumbsComponent struct {
Items []BreadcrumbItem
}

type ButtonComponent struct {
Text string
Kind string
}

type CardComponent struct {
Title string
Body  string
}

type CheckboxComponent struct {
Name    string
Value   string
Label   string
Checked bool
}

type ClusterComponent struct {
ExtraClass string
Children   []Component
}

type CodeBlockComponent struct {
Code string
}

type ColumnComponent struct {
SizeClass string
Children  []Component
}

type ContainerComponent struct {
Children []Component
}

type ContainerFluidComponent struct {
Children []Component
}

type DescriptionListComponent struct {
Items []DescriptionItem
}

type DetailsComponent struct {
Summary  string
Children []Component
}

type FieldsetComponent struct {
Legend   string
Children []Component
}

type FormComponent struct {
Action   string
Method   string
Children []Component
}

// FragmentComponent renders multiple child components with no wrapper element.
type FragmentComponent struct {
Children []Component
}

type GridComponent struct {
ColumnsClass string
Items        []Component
}

type GridItemComponent struct {
SpanClass string
Children  []Component
}

type HeadingComponent struct {
Level int
Text  string
}

type HeroComponent struct {
Title    string
Subtitle string
Actions  []Component
}

type HorizontalRuleComponent struct{}

type ImageComponent struct {
Src string
Alt string
}

type InputComponent struct {
Label       string
Name        string
Placeholder string
}

// Interaction represents HTMX behavior in GUI-semantic form.
// It keeps hx-* details out of application code while still enabling
// partial updates, target swaps, and history updates.
type Interaction struct {
Boost   bool
Delete  string
Get     string
Post    string
PushURL string
Put     string
Select  string
Swap    string
Target  string
Trigger string
}

type InteractiveActionComponent struct {
Text        string
Kind        string
Interaction Interaction
}

type InteractiveMenuComponent struct {
Links []InteractiveMenuLink
}

type InteractiveMenuLink struct {
Label       string
Href        string
Interaction Interaction
}

type ListComponent struct {
Items []string
}

type NavComponent struct {
Links []NavLink
}

type OrderedListComponent struct {
Items []string
}

type PaginationComponent struct {
Items []PageItem
}

type ParagraphComponent struct {
Text string
}

type RadioComponent struct {
Name    string
Value   string
Label   string
Checked bool
}

type RowComponent struct {
Children []Component
}

type SectionComponent struct {
Title    string
Children []Component
}

type SelectComponent struct {
Label   string
Name    string
Options []SelectOption
}

type SidebarLayoutComponent struct {
Sidebar Component
Content Component
}

type SplitComponent struct {
Primary   Component
Secondary Component
}

type StackComponent struct {
ExtraClass string
Children   []Component
}

type TableComponent struct {
ClassName string
Headers   []string
Rows      [][]string
}

type TextAreaComponent struct {
Label       string
Name        string
Placeholder string
}

type ThemeToggleComponent struct{}

// Template render data types.
// These structs are passed to template.Execute for components that compose
// pre-rendered child HTML. They are separate from component structs so that
// template data shapes are decoupled from the public component API.

type appShellData struct {
HasSidebar bool
Sidebar    template.HTML
Content    template.HTML
}

type articleData struct {
Title    string
Children []template.HTML
}

type childrenData struct {
Children []template.HTML
}

type classChildrenData struct {
Class    string
Children []template.HTML
}

type detailsData struct {
Summary  string
Children []template.HTML
}

type fieldsetData struct {
Legend   string
Children []template.HTML
}

type formData struct {
Action   string
Method   string
Children []template.HTML
}

type heroData struct {
Title    string
Subtitle string
Actions  []template.HTML
}

type sectionData struct {
Title    string
Children []template.HTML
}

type sidebarData struct {
Sidebar template.HTML
Content template.HTML
}

type splitData struct {
Primary   template.HTML
Secondary template.HTML
}
