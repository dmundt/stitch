package ui

import "html/template"

// BreadcrumbItem describes one step in a breadcrumb trail.
//
// It is consumed by BreadcrumbsComponent to render hierarchical navigation,
// and Current marks the active location.
type BreadcrumbItem struct {
	Label   string
	Href    string
	ID      string
	Class   string
	Attrs   map[string]string
	Current bool
}

// DescriptionItem describes one term-definition pair for description lists.
//
// It is rendered by DescriptionListComponent as semantic <dt>/<dd> content.
type DescriptionItem struct {
	Term       string
	Definition string
}

// NavLink describes one navigation item rendered by NavComponent.
//
// Label is user-facing text and Href is the navigation destination.
type NavLink struct {
	Label string
	Href  string
	ID    string
	Class string
	Attrs map[string]string
}

// PageItem describes one item in a pagination control.
//
// Current marks the active page, and Disabled renders a non-interactive item.
type PageItem struct {
	Label    string
	Href     string
	ID       string
	Class    string
	Attrs    map[string]string
	Current  bool
	Disabled bool
}

// SelectOption describes one option in a SelectComponent.
//
// Selected marks the initial selected option for rendered form controls.
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
}

// AlertComponent models inline status and feedback messages.
//
// Tone selects a style variant and defaults to "info" at render time when
// left empty. Valid tone names are provider-defined.
type AlertComponent struct {
	Text string
	Tone string
}

// AppShellComponent models a page-level shell with sidebar and content areas.
//
// Sidebar is optional; when nil, only the content region is rendered.
type AppShellComponent struct {
	Sidebar Component
	Content Component
}

// ArticleComponent models a titled, self-contained content region.
//
// It is suited for cards, panels, and grouped content sections.
type ArticleComponent struct {
	Title    string
	Children []Component
}

// BadgeComponent models compact status or metadata labels.
//
// Tone defaults to "default" at render time when left empty. Valid tone names
// are provider-defined.
type BadgeComponent struct {
	Text string
	Tone string
}

// BlockquoteComponent models quoted content and optional attribution.
//
// Cite is optional and typically represents the source or author.
type BlockquoteComponent struct {
	Text string
	Cite string
}

// BreadcrumbsComponent models hierarchical breadcrumb navigation.
//
// It renders ordered location context from BreadcrumbItem entries.
type BreadcrumbsComponent struct {
	Items []BreadcrumbItem
}

// ButtonComponent models a clickable action element.
//
// Kind selects a style variant and defaults to "default" at render time when
// left empty. Valid kind names are provider-defined.
type ButtonComponent struct {
	Text string
	Kind string
}

// CardComponent models concise, titled content grouped in a card container.
//
// It is intended for compact summary content and dashboard tiles.
type CardComponent struct {
	Title string
	Body  string
}

// CheckboxComponent models a checkbox form field.
//
// Checked controls the initial checked state in rendered markup.
type CheckboxComponent struct {
	Name    string
	Value   string
	Label   string
	Checked bool
}

// ClusterComponent models a wrapping horizontal cluster of child components.
//
// ExtraClass appends additional layout classes.
type ClusterComponent struct {
	ExtraClass string
	Children   []Component
}

// CodeBlockComponent models preformatted code or command snippets.
//
// Use it for examples where whitespace and line breaks should be preserved.
type CodeBlockComponent struct {
	Code string
}

// ColumnComponent models a grid column region.
//
// SizeClass can be used to control span behavior in theme CSS.
type ColumnComponent struct {
	SizeClass string
	Children  []Component
}

// ContainerComponent models a centered constrained-width layout wrapper.
//
// It is commonly used as a page content boundary.
type ContainerComponent struct {
	Children []Component
}

// ContainerFluidComponent models a full-width layout wrapper.
//
// It is useful for edge-to-edge sections and wide app shells.
type ContainerFluidComponent struct {
	Children []Component
}

// DescriptionListComponent models semantic term-definition content.
//
// It consumes DescriptionItem entries as glossary-like UI content.
type DescriptionListComponent struct {
	Items []DescriptionItem
}

// DetailsComponent models disclosure content with a summary label.
//
// It represents expandable content where Summary is always visible.
type DetailsComponent struct {
	Summary  string
	Children []Component
}

// FieldsetComponent models grouped form controls under a legend.
//
// It adds semantic grouping for related inputs.
type FieldsetComponent struct {
	Legend   string
	Children []Component
}

// FormComponent models an HTML form and its child controls.
//
// Action and Method map directly to form submission attributes.
type FormComponent struct {
	Action   string
	Method   string
	Children []Component
}

// FragmentComponent groups child components without adding a wrapper element.
//
// It is useful when semantic wrappers would be redundant or undesirable.
type FragmentComponent struct {
	Children []Component
}

// GridComponent models a grid layout for child items.
//
// ColumnsClass can be used to select column template variants in theme CSS.
type GridComponent struct {
	ColumnsClass string
	Items        []Component
}

// GridItemComponent models a single item inside GridComponent.
//
// SpanClass can be used to control item spanning in theme CSS.
type GridItemComponent struct {
	SpanClass string
	Children  []Component
}

// HeadingComponent models semantic heading content.
//
// Level is normalized to the inclusive range 1..6 during rendering.
type HeadingComponent struct {
	Level int
	Text  string
}

// HeroComponent models a prominent introductory section with optional actions.
//
// It is commonly used for page headers and call-to-action areas.
type HeroComponent struct {
	Title    string
	Subtitle string
	Actions  []Component
}

// HorizontalRuleComponent models a thematic section break.
//
// It renders a semantic boundary between related content groups.
type HorizontalRuleComponent struct{}

// ImageComponent models an image element with alternative text.
//
// Alt should describe the image for accessibility.
type ImageComponent struct {
	Src string
	Alt string
}

// InputComponent models a labeled text input field.
//
// Placeholder provides optional hint text for expected input.
type InputComponent struct {
	Label       string
	Name        string
	Placeholder string
}

// Interaction represents HTMX behavior in GUI-semantic form.
// It keeps hx-* details out of application code while still enabling
// partial updates, target swaps, and history updates.
//
// Get, Post, Put, and Delete map directly to the corresponding hx-* request
// attributes. Target, Swap, Select, Trigger, and PushURL map to HTMX behavior
// attributes. Values are written as attributes, so treat URL and selector input
// as trusted application-controlled strings.
//
// Use at most one of Get, Post, Put, or Delete on a single Interaction value
// to avoid ambiguous request intent.
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

// InteractiveActionComponent models a button action with HTMX interaction.
//
// It combines action presentation with request/target behavior in one value.
type InteractiveActionComponent struct {
	Text        string
	Kind        string
	Interaction Interaction
}

// InteractiveMenuComponent models navigation links with HTMX behavior.
//
// Each entry can trigger partial-page updates instead of full navigation.
type InteractiveMenuComponent struct {
	Links []InteractiveMenuLink
}

// InteractiveMenuLink describes one HTMX-enabled navigation entry.
//
// It combines traditional Href semantics with optional HTMX interaction.
type InteractiveMenuLink struct {
	Label       string
	Href        string
	ID          string
	Class       string
	Attrs       map[string]string
	Interaction Interaction
}

// ListComponent models an unordered list of text items.
//
// It is intended for simple bullet-style content.
type ListComponent struct {
	Items []string
}

// NavComponent models a semantic navigation region.
//
// It consumes NavLink entries to render menu-style navigation.
type NavComponent struct {
	Links []NavLink
}

// OrderedListComponent models an ordered sequence of text items.
//
// It is suited for instructions, steps, and ranked content.
type OrderedListComponent struct {
	Items []string
}

// PaginationComponent models page-to-page navigation controls.
//
// It consumes PageItem entries and supports current/disabled states.
type PaginationComponent struct {
	Items []PageItem
}

// ParagraphComponent models body copy content.
//
// It is the default text primitive for prose in composed layouts.
type ParagraphComponent struct {
	Text string
}

// RadioComponent models a radio-button form field.
//
// Fields with matching Name participate in a shared selection group.
type RadioComponent struct {
	Name    string
	Value   string
	Label   string
	Checked bool
}

// RowComponent models a row-level layout container.
//
// It is typically paired with ColumnComponent for grid-like composition.
type RowComponent struct {
	Children []Component
}

// SectionComponent models a titled section of related UI content.
//
// It provides a semantic boundary for grouped page content.
type SectionComponent struct {
	Title    string
	Children []Component
}

// SelectComponent models a labeled select field and options.
//
// Options provides the set of selectable values in rendered markup.
type SelectComponent struct {
	Label   string
	Name    string
	Options []SelectOption
}

// SidebarLayoutComponent models a two-region layout with sidebar and content.
//
// It is useful for documentation and dashboard-style page structures.
type SidebarLayoutComponent struct {
	Sidebar Component
	Content Component
}

// SplitComponent models a primary-secondary two-pane layout.
//
// It is suited for detail panes and complementary side content.
type SplitComponent struct {
	Primary   Component
	Secondary Component
}

// StackComponent models a vertical layout stack.
//
// ExtraClass appends additional layout classes.
type StackComponent struct {
	ExtraClass string
	Children   []Component
}

// TableComponent models tabular data with optional table class.
//
// Headers define column names and Rows provide per-row cell values.
type TableComponent struct {
	ClassName string
	Headers   []string
	Rows      [][]string
}

// TextAreaComponent models a labeled multiline text input.
//
// It is intended for free-form input longer than a single line.
type TextAreaComponent struct {
	Label       string
	Name        string
	Placeholder string
}

// ThemeToggleComponent models a theme toggle control.
//
// It provides a compact control point for switching visual themes.
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
