package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dmundt/stitch/css"
	"github.com/dmundt/stitch/render"
	stitchtpl "github.com/dmundt/stitch/template"
	"github.com/dmundt/stitch/ui"
)

const (
	listenAddr      = "localhost:8080"
	defaultProvider = "stitch"
	defaultMCPPath  = "/mcp"
	defaultSSEPath  = "/mcp/sse"
)

// ListenAddr is the default HTTP bind address for preview and MCP-over-HTTP.
const ListenAddr = listenAddr

// DefaultHTTPEndpoint is the default HTTP server endpoint.
const DefaultHTTPEndpoint = "http://" + listenAddr

// DefaultMCPHTTPEndpoint is the default MCP HTTP endpoint.
const DefaultMCPHTTPEndpoint = DefaultHTTPEndpoint + defaultMCPPath

// DefaultSSEEndpoint is the default MCP SSE endpoint.
const DefaultSSEEndpoint = DefaultHTTPEndpoint + defaultSSEPath

type app struct {
	store sessionStore
}

func newApp() *app {
	return &app{store: newStoreFromEnv()}
}

// newStoreFromEnv resolves a session-file path (via STITCH_SESSION_FILE or
// the OS user-cache directory) and returns a file-backed sessionStore.  On any
// error (unwritable dir, malformed file) it falls back to an in-memory store
// and logs the reason so the operator is never silently surprised.
func newStoreFromEnv() sessionStore {
	path := os.Getenv("STITCH_SESSION_FILE")
	if path == "" {
		if cacheDir, err := os.UserCacheDir(); err == nil {
			path = filepath.Join(cacheDir, "stitch", "sessions.json")
		}
	}
	if path != "" {
		if fs, err := newFileStore(path); err == nil {
			log.Printf("stitch: session store: file %s", path)
			return fs
		} else {
			log.Printf("stitch: session store: file unavailable (%v), using memory", err)
		}
	}
	log.Printf("stitch: session store: memory-only")
	return newStore()
}

// Run starts the Stitch MCP server with stdio + HTTP/SSE transports.
func Run(ctx context.Context) error {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	a := newApp()

	errCh := make(chan error, 2)
	go func() {
		errCh <- a.serveHTTP(ctx)
	}()
	go func() {
		errCh <- a.serveSTDIO()
	}()

	err := <-errCh
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (a *app) dispatchTool(name string, args map[string]any) (map[string]any, error) {
	switch name {
	case "session.create":
		title := asString(args["title"])
		provider := asString(args["provider"])
		session := a.store.createSession(title, provider)
		return map[string]any{
			"session":             session,
			"preview_page_url":    fmt.Sprintf("http://%s/sessions/%s/page", listenAddr, session.ID),
			"preview_header_url":  fmt.Sprintf("http://%s/sessions/%s/blocks/%s", listenAddr, session.ID, stitchtpl.BlockHeader),
			"preview_main_url":    fmt.Sprintf("http://%s/sessions/%s/blocks/%s", listenAddr, session.ID, stitchtpl.BlockMain),
			"preview_footer_url":  fmt.Sprintf("http://%s/sessions/%s/blocks/%s", listenAddr, session.ID, stitchtpl.BlockFooter),
			"mcp_http_endpoint":   DefaultMCPHTTPEndpoint,
			"mcp_sse_endpoint":    DefaultSSEEndpoint,
			"available_providers": listProviders(),
		}, nil
	case "session.get":
		sessionID := requiredString(args, "session_id")
		session, err := a.store.getSession(sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"session": session}, nil
	case "session.reset":
		sessionID := requiredString(args, "session_id")
		session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
			s.Components = map[string]*componentNode{}
			s.Blocks = map[string][]string{stitchtpl.BlockHeader: {}, stitchtpl.BlockMain: {}, stitchtpl.BlockFooter: {}}
			s.Page.HeadSnippets = defaultHeadSnippets()
			return nil
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"session": session}, nil
	case "session.delete":
		sessionID := requiredString(args, "session_id")
		deleted, err := a.store.deleteSession(sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"deleted": deleted}, nil
	case "providers.list":
		return map[string]any{"providers": listProviders()}, nil
	case "page.set_meta":
		sessionID := requiredString(args, "session_id")
		title := asString(args["title"])
		lang := asString(args["lang"])
		session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
			if title != "" {
				s.Page.Title = title
			}
			if lang != "" {
				s.Page.Lang = lang
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"session": session}, nil
	case "page.set_css_provider":
		sessionID := requiredString(args, "session_id")
		provider := requiredString(args, "provider")
		if _, err := css.Get(provider); err != nil {
			return nil, err
		}
		session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
			s.Page.Provider = provider
			return nil
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"session": session}, nil
	case "page.set_head_snippet":
		sessionID := requiredString(args, "session_id")
		snippet := requiredString(args, "snippet")
		if err := validateHeadSnippet(snippet); err != nil {
			return nil, err
		}
		session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
			s.Page.HeadSnippets = append(s.Page.HeadSnippets, snippet)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"session": session}, nil
	case "schema.list_component_types":
		return map[string]any{"types": componentSchemas()}, nil
	case "schema.get_component_fields":
		typeName := canonicalType(requiredString(args, "type"))
		schema, ok := componentSchemas()[typeName]
		if !ok {
			return nil, fmt.Errorf("unknown component type: %s", typeName)
		}
		return map[string]any{"type": typeName, "schema": schema}, nil
	case "ui.create_component":
		return a.toolCreateComponent(args)
	case "ui.update_component":
		return a.toolUpdateComponent(args)
	case "ui.delete_component":
		return a.toolDeleteComponent(args)
	case "ui.move_component":
		return a.toolMoveComponent(args)
	case "ui.get_component":
		return a.toolGetComponent(args)
	case "ui.list_components":
		return a.toolListComponents(args)
	case "render.full":
		sessionID := requiredString(args, "session_id")
		session, err := a.store.getSession(sessionID)
		if err != nil {
			return nil, err
		}
		output, err := a.renderFull(session)
		if err != nil {
			return nil, err
		}
		return map[string]any{"html": output}, nil
	case "render.block":
		sessionID := requiredString(args, "session_id")
		block := requiredString(args, "block")
		session, err := a.store.getSession(sessionID)
		if err != nil {
			return nil, err
		}
		fragment, err := a.renderBlock(session, block)
		if err != nil {
			return nil, err
		}
		return map[string]any{"block": block, "html": fragment}, nil
	case "render.component":
		sessionID := requiredString(args, "session_id")
		componentID := requiredString(args, "component_id")
		session, err := a.store.getSession(sessionID)
		if err != nil {
			return nil, err
		}
		fragment, err := a.renderComponent(session, componentID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"component_id": componentID, "html": fragment}, nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *app) toolCreateComponent(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	typeName := canonicalType(requiredString(args, "type"))
	props, attrs, attrWarnings := sanitizeComponentInput(asMap(args["props"]))
	parentID := asString(args["parent_id"])
	block := asString(args["block"])
	position := asInt(args["position"], -1)

	if !hasComponentSchema(typeName) {
		return nil, fmt.Errorf("unsupported component type: %s", typeName)
	}
	if parentID == "" && block == "" {
		block = stitchtpl.BlockMain
	}

	id := randomID("cmp")
	node := &componentNode{ID: id, Type: typeName, Props: props, Attrs: attrs, Children: []string{}}

	session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
		if parentID != "" {
			parent, ok := s.Components[parentID]
			if !ok {
				return fmt.Errorf("unknown parent_id: %s", parentID)
			}
			insertAt(&parent.Children, id, position)
		} else {
			if err := stitchtpl.ValidateBlocks([]string{block}); err != nil {
				return err
			}
			roots := s.Blocks[block]
			insertAt(&roots, id, position)
			s.Blocks[block] = roots
		}
		s.Components[id] = node
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := map[string]any{"component": node, "session": session}
	if warnings := withComponentWarnings(node.ID, attrWarnings); len(warnings) > 0 {
		result["warnings"] = warnings
	}
	return result, nil
}

func (a *app) toolUpdateComponent(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	componentID := requiredString(args, "component_id")
	props, attrs, attrWarnings := sanitizeComponentInput(asMap(args["props"]))
	session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
		node, ok := s.Components[componentID]
		if !ok {
			return fmt.Errorf("unknown component_id: %s", componentID)
		}
		if t := asString(args["type"]); t != "" {
			normalized := canonicalType(t)
			if !hasComponentSchema(normalized) {
				return fmt.Errorf("unsupported component type: %s", normalized)
			}
			node.Type = normalized
		}
		node.Props = props
		node.Attrs = attrs
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := map[string]any{"component_id": componentID, "session": session}
	if warnings := withComponentWarnings(componentID, attrWarnings); len(warnings) > 0 {
		result["warnings"] = warnings
	}
	return result, nil
}

func (a *app) toolDeleteComponent(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	componentID := requiredString(args, "component_id")
	session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
		if _, ok := s.Components[componentID]; !ok {
			return fmt.Errorf("unknown component_id: %s", componentID)
		}
		removeFromAllParents(s, componentID)
		deleteSubtree(s, componentID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"deleted": true, "session": session}, nil
}

func deleteSubtree(s *sessionState, rootID string) {
	node, ok := s.Components[rootID]
	if !ok {
		return
	}
	for _, childID := range node.Children {
		deleteSubtree(s, childID)
	}
	delete(s.Components, rootID)
}

func removeFromAllParents(s *sessionState, id string) {
	for block, ids := range s.Blocks {
		s.Blocks[block] = removeID(ids, id)
	}
	for _, node := range s.Components {
		node.Children = removeID(node.Children, id)
	}
}

func removeID(list []string, id string) []string {
	out := make([]string, 0, len(list))
	for _, item := range list {
		if item != id {
			out = append(out, item)
		}
	}
	return out
}

func (a *app) toolMoveComponent(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	componentID := requiredString(args, "component_id")
	newParentID := asString(args["new_parent_id"])
	newBlock := asString(args["new_block"])
	position := asInt(args["position"], -1)

	if newParentID == "" && newBlock == "" {
		newBlock = stitchtpl.BlockMain
	}

	session, err := a.store.updateSession(sessionID, func(s *sessionState) error {
		node, ok := s.Components[componentID]
		if !ok || node == nil {
			return fmt.Errorf("unknown component_id: %s", componentID)
		}
		if newParentID != "" {
			if _, ok := s.Components[newParentID]; !ok {
				return fmt.Errorf("unknown new_parent_id: %s", newParentID)
			}
			if createsCycle(s, componentID, newParentID) {
				return errors.New("move would create cycle")
			}
		}
		if newParentID == "" {
			if err := stitchtpl.ValidateBlocks([]string{newBlock}); err != nil {
				return err
			}
		}

		removeFromAllParents(s, componentID)
		if newParentID != "" {
			parent := s.Components[newParentID]
			insertAt(&parent.Children, componentID, position)
		} else {
			roots := s.Blocks[newBlock]
			insertAt(&roots, componentID, position)
			s.Blocks[newBlock] = roots
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"moved": true, "session": session}, nil
}

func createsCycle(s *sessionState, childID, parentID string) bool {
	if childID == parentID {
		return true
	}
	queue := []string{childID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		node := s.Components[id]
		if node == nil {
			continue
		}
		for _, c := range node.Children {
			if c == parentID {
				return true
			}
			queue = append(queue, c)
		}
	}
	return false
}

func (a *app) toolGetComponent(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	componentID := requiredString(args, "component_id")
	session, err := a.store.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	node, ok := session.Components[componentID]
	if !ok {
		return nil, fmt.Errorf("unknown component_id: %s", componentID)
	}
	return map[string]any{"component": node}, nil
}

func (a *app) toolListComponents(args map[string]any) (map[string]any, error) {
	sessionID := requiredString(args, "session_id")
	session, err := a.store.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	components := make([]*componentNode, 0, len(session.Components))
	for _, node := range session.Components {
		components = append(components, node)
	}
	return map[string]any{"components": components, "blocks": session.Blocks}, nil
}

func insertAt(list *[]string, value string, position int) {
	items := *list
	if position < 0 || position >= len(items) {
		*list = append(items, value)
		return
	}
	items = append(items, "")
	copy(items[position+1:], items[position:])
	items[position] = value
	*list = items
}

func (a *app) renderFull(session *sessionState) (string, error) {
	provider, err := css.Get(session.Page.Provider)
	if err != nil {
		provider = css.Stitch()
	}
	page := render.NewWindow(session.Page.Title)
	if session.Page.Lang != "" {
		page.Lang = session.Page.Lang
	}
	for _, snippet := range session.Page.HeadSnippets {
		if err := validateHeadSnippet(snippet); err != nil {
			continue
		}
		page.WithHeadRaw(snippet)
	}

	headerComp, err := a.buildBlockComponent(session, stitchtpl.BlockHeader)
	if err != nil {
		return "", err
	}
	mainComp, err := a.buildBlockComponent(session, stitchtpl.BlockMain)
	if err != nil {
		return "", err
	}
	footerComp, err := a.buildBlockComponent(session, stitchtpl.BlockFooter)
	if err != nil {
		return "", err
	}

	page.TopBarComponent(headerComp)
	page.ContentComponent(mainComp)
	page.StatusBarComponent(footerComp)
	return page.Render(provider)
}

func (a *app) renderBlock(session *sessionState, block string) (string, error) {
	if err := stitchtpl.ValidateBlocks([]string{block}); err != nil {
		return "", err
	}
	comp, err := a.buildBlockComponent(session, block)
	if err != nil {
		return "", err
	}
	return comp.HTML(), nil
}

func (a *app) renderComponent(session *sessionState, componentID string) (string, error) {
	comp, err := a.buildComponent(session, componentID)
	if err != nil {
		return "", err
	}
	return comp.HTML(), nil
}

func (a *app) buildBlockComponent(session *sessionState, block string) (ui.Component, error) {
	ids := session.Blocks[block]
	children := make([]ui.Component, 0, len(ids))
	for _, id := range ids {
		comp, err := a.buildComponent(session, id)
		if err != nil {
			return nil, err
		}
		children = append(children, comp)
	}
	return ui.NewFragment(children), nil
}

func (a *app) buildComponent(session *sessionState, componentID string) (ui.Component, error) {
	node, ok := session.Components[componentID]
	if !ok {
		return nil, fmt.Errorf("unknown component_id: %s", componentID)
	}
	comp, err := a.buildNode(session, node)
	if err != nil {
		return nil, err
	}
	attrs := mergedNodeAttrs(node)
	if id := asString(node.Props["id"]); strings.TrimSpace(id) != "" {
		comp = ui.WithID(id, comp)
		delete(attrs, "id")
		delete(attrs, "ID")
	}
	if len(attrs) > 0 {
		comp = ui.WithAttrs(attrs, comp)
	}
	return comp, nil
}

func mergedNodeAttrs(node *componentNode) map[string]string {
	out := map[string]string{}
	for k, v := range node.Attrs {
		out[k] = v
	}
	if len(out) > 0 {
		return out
	}
	_, legacy, _ := sanitizeComponentInput(map[string]any{"attrs": node.Props["attrs"]})
	for k, v := range legacy {
		out[k] = v
	}
	return out
}

func (a *app) buildNode(session *sessionState, node *componentNode) (ui.Component, error) {
	children, err := a.buildChildren(session, node.Children)
	if err != nil {
		return nil, err
	}
	p := node.Props
	switch node.Type {
	case "action", "button":
		return ui.NewAction(asStringDefault(p, "text", "Action"), asStringDefault(p, "kind", "default")), nil
	case "alert", "status":
		return ui.NewAlert(asStringDefault(p, "text", ""), asStringDefault(p, "tone", "info")), nil
	case "appshell", "workspace":
		var sidebar, content ui.Component
		if len(children) > 0 {
			sidebar = children[0]
		}
		if len(children) > 1 {
			content = children[1]
		}
		return ui.NewAppShell(sidebar, content), nil
	case "article", "panel":
		return ui.NewArticle(asStringDefault(p, "title", ""), children), nil
	case "badge":
		return ui.NewBadge(asStringDefault(p, "text", ""), asStringDefault(p, "tone", "default")), nil
	case "blockquote":
		return ui.NewBlockquote(asStringDefault(p, "text", ""), asStringDefault(p, "cite", "")), nil
	case "breadcrumbs":
		return ui.NewBreadcrumbs(asBreadcrumbItems(p["items"])), nil
	case "card":
		return ui.NewCard(asStringDefault(p, "title", ""), asStringDefault(p, "body", "")), nil
	case "checkbox":
		return ui.NewCheckbox(asStringDefault(p, "name", ""), asStringDefault(p, "value", ""), asStringDefault(p, "label", ""), asBoolDefault(p, "checked", false)), nil
	case "cluster", "toolbar":
		extra := asStringDefault(p, "extra_class", "")
		if node.Type == "toolbar" && extra == "" {
			extra = "toolbar"
		}
		return ui.NewCluster(extra, children), nil
	case "codeblock":
		return ui.NewCodeBlock(asStringDefault(p, "code", "")), nil
	case "column":
		return ui.NewColumn(asStringDefault(p, "size_class", ""), children), nil
	case "container":
		return ui.NewContainer(children), nil
	case "container_fluid":
		return ui.NewContainerFluid(children), nil
	case "descriptionlist":
		return ui.NewDescriptionList(asDescriptionItems(p["items"])), nil
	case "details":
		return ui.NewDetails(asStringDefault(p, "summary", "Details"), children), nil
	case "fieldset":
		return ui.NewFieldset(asStringDefault(p, "legend", ""), children), nil
	case "form":
		return ui.NewForm(asStringDefault(p, "action", ""), asStringDefault(p, "method", "post"), children), nil
	case "fragment":
		return ui.NewFragment(children), nil
	case "grid":
		return ui.NewGrid(asStringDefault(p, "columns_class", ""), children), nil
	case "grid_item":
		return ui.NewGridItem(asStringDefault(p, "span_class", ""), children), nil
	case "heading":
		return ui.NewHeading(asIntDefault(p, "level", 2), asStringDefault(p, "text", "")), nil
	case "hero":
		return ui.NewHero(asStringDefault(p, "title", ""), asStringDefault(p, "subtitle", ""), children), nil
	case "horizontal_rule":
		return ui.NewHorizontalRule(), nil
	case "image":
		return ui.NewImage(asStringDefault(p, "src", ""), asStringDefault(p, "alt", "")), nil
	case "input":
		return ui.NewInput(asStringDefault(p, "label", ""), asStringDefault(p, "name", ""), asStringDefault(p, "placeholder", "")), nil
	case "interactive_action":
		return ui.NewInteractiveAction(asStringDefault(p, "text", "Action"), asStringDefault(p, "kind", "default"), asInteraction(p["interaction"])), nil
	case "interactive_menu":
		return ui.NewInteractiveMenu(asInteractiveMenuLinks(p["links"])), nil
	case "list":
		return ui.NewList(asStringSlice(p["items"])), nil
	case "menu", "nav":
		return ui.NewNav(asNavLinks(p["links"])), nil
	case "ordered_list":
		return ui.NewOrderedList(asStringSlice(p["items"])), nil
	case "pagination":
		return ui.NewPagination(asPageItems(p["items"])), nil
	case "paragraph", "text":
		return ui.NewParagraph(asStringDefault(p, "text", "")), nil
	case "radio":
		return ui.NewRadio(asStringDefault(p, "name", ""), asStringDefault(p, "value", ""), asStringDefault(p, "label", ""), asBoolDefault(p, "checked", false)), nil
	case "row":
		return ui.NewRow(children), nil
	case "section", "view":
		return ui.NewSection(asStringDefault(p, "title", ""), children), nil
	case "select":
		return ui.NewSelect(asStringDefault(p, "label", ""), asStringDefault(p, "name", ""), asSelectOptions(p["options"])), nil
	case "sidebar_layout":
		var sidebar, content ui.Component
		if len(children) > 0 {
			sidebar = children[0]
		}
		if len(children) > 1 {
			content = children[1]
		}
		return ui.NewSidebarLayout(sidebar, content), nil
	case "split":
		var primary, secondary ui.Component
		if len(children) > 0 {
			primary = children[0]
		}
		if len(children) > 1 {
			secondary = children[1]
		}
		return ui.NewSplit(primary, secondary), nil
	case "stack":
		return ui.NewStack(asStringDefault(p, "extra_class", ""), children), nil
	case "table", "datagrid":
		return ui.NewTable(asStringSlice(p["headers"]), asStringMatrix(p["rows"])), nil
	case "textarea":
		return ui.NewTextArea(asStringDefault(p, "label", ""), asStringDefault(p, "name", ""), asStringDefault(p, "placeholder", "")), nil
	case "theme_toggle":
		return ui.NewThemeToggle(), nil
	default:
		return nil, fmt.Errorf("unsupported component type: %s", node.Type)
	}
}

func (a *app) buildChildren(session *sessionState, childIDs []string) ([]ui.Component, error) {
	out := make([]ui.Component, 0, len(childIDs))
	for _, childID := range childIDs {
		comp, err := a.buildComponent(session, childID)
		if err != nil {
			if strings.Contains(err.Error(), "unknown component_id") {
				return nil, fmt.Errorf("missing child component: %s", childID)
			}
			return nil, err
		}
		out = append(out, comp)
	}
	return out, nil
}

func validateHeadSnippet(snippet string) error {
	trimmed := strings.TrimSpace(snippet)
	if trimmed == "" {
		return errors.New("head snippet is empty")
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "<script") {
		closeTag := "</script>"
		closeIdx := strings.LastIndex(lower, closeTag)
		if closeIdx < 0 {
			return errors.New("script head snippet must be a complete <script ...></script> tag")
		}
		openEnd := strings.Index(trimmed, ">")
		if openEnd < 0 || openEnd > closeIdx {
			return errors.New("script head snippet must be a complete <script ...></script> tag")
		}
		openTag := trimmed[:openEnd+1]
		if !strings.Contains(strings.ToLower(openTag), "src=") {
			return errors.New("script head snippet must include a src attribute")
		}
		inlineBody := strings.TrimSpace(trimmed[openEnd+1 : closeIdx])
		if inlineBody != "" {
			return errors.New("inline script bodies are not allowed in head snippets")
		}
		return nil
	}
	if strings.HasPrefix(lower, "<meta") || strings.HasPrefix(lower, "<link") || strings.HasPrefix(lower, "<style") {
		return nil
	}
	return errors.New("head snippet must start with <meta, <link, <style, or <script src=...>")
}

func randomID(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), rand.Intn(100000))
}

func listProviders() []string {
	providers := []string{"none", "stitch", "minstyle", "milligram"}
	out := make([]string, 0, len(providers))
	for _, name := range providers {
		if _, err := css.Get(name); err == nil {
			out = append(out, name)
		}
	}
	return out
}

func requiredString(args map[string]any, key string) string {
	value := strings.TrimSpace(asString(args[key]))
	if value == "" {
		panicf("missing required argument: %s", key)
	}
	return value
}

func panicf(format string, args ...any) {
	panic(errString(fmt.Sprintf(format, args...)))
}

type errString string

func (e errString) Error() string { return string(e) }

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func asStringDefault(m map[string]any, key, def string) string {
	v := asString(m[key])
	if v == "" {
		return def
	}
	return v
}

func asBool(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func asBoolDefault(m map[string]any, key string, def bool) bool {
	if v, ok := m[key]; ok {
		return asBool(v)
	}
	return def
}

func asInt(v any, def int) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		i, err := t.Int64()
		if err == nil {
			return int(i)
		}
	}
	return def
}

func asIntDefault(m map[string]any, key string, def int) int {
	if v, ok := m[key]; ok {
		return asInt(v, def)
	}
	return def
}

func asStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		if s, ok := v.([]string); ok {
			return append([]string{}, s...)
		}
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s := asString(item); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func asStringMatrix(v any) [][]string {
	rows, ok := v.([]any)
	if !ok {
		if matrix, ok := v.([][]string); ok {
			out := make([][]string, len(matrix))
			for i, row := range matrix {
				out[i] = append([]string{}, row...)
			}
			return out
		}
		return [][]string{}
	}
	out := make([][]string, 0, len(rows))
	for _, rowAny := range rows {
		out = append(out, asStringSlice(rowAny))
	}
	return out
}

func asMap(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		if t == nil {
			return map[string]any{}
		}
		return t
	case string:
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			return map[string]any{}
		}
		decoded := map[string]any{}
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			return map[string]any{}
		}
		return decoded
	default:
		return map[string]any{}
	}
}

func sanitizeComponentInput(props map[string]any) (map[string]any, map[string]string, []map[string]any) {
	cleanProps := map[string]any{}
	for k, v := range props {
		if k == "attrs" {
			continue
		}
		cleanProps[k] = v
	}

	attrs := map[string]string{}
	warnings := []map[string]any{}
	rawAttrs := asMap(props["attrs"])
	for key, rawValue := range rawAttrs {
		name := strings.TrimSpace(key)
		if name == "" {
			warnings = append(warnings, map[string]any{"attr": key, "reason": "invalid-name"})
			continue
		}
		value := strings.TrimSpace(anyToString(rawValue))
		if value == "" {
			warnings = append(warnings, map[string]any{"attr": name, "reason": "empty-value"})
			continue
		}
		if !isAllowedAttrName(name) {
			warnings = append(warnings, map[string]any{"attr": name, "reason": "blocked"})
			continue
		}
		if isUnsafeAttrValue(name, value) {
			warnings = append(warnings, map[string]any{"attr": name, "reason": "unsafe-value"})
			continue
		}
		attrs[name] = value
	}

	if len(attrs) > 0 {
		cleanProps["attrs"] = attrs
	}
	return cleanProps, attrs, warnings
}

func isAllowedAttrName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "id" || lower == "class" {
		return true
	}
	if strings.HasPrefix(lower, "on") {
		return false
	}
	if lower == "style" {
		return false
	}
	return strings.HasPrefix(lower, "data-") || strings.HasPrefix(lower, "aria-") || strings.HasPrefix(lower, "hx-")
}

func isUnsafeAttrValue(name, value string) bool {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(lowerValue, "javascript:") {
		return true
	}
	if strings.Contains(lowerValue, "\u0000") {
		return true
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "hx-") && strings.HasPrefix(lowerValue, "data:text/html") {
		return true
	}
	return false
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", t)
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	default:
		return ""
	}
}

func withComponentWarnings(componentID string, warnings []map[string]any) []map[string]any {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(warnings))
	for _, w := range warnings {
		entry := map[string]any{"component_id": componentID}
		for k, v := range w {
			entry[k] = v
		}
		out = append(out, entry)
	}
	return out
}

func asNavLinks(v any) []ui.NavLink {
	items, ok := v.([]any)
	if !ok {
		return []ui.NavLink{}
	}
	out := make([]ui.NavLink, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.NavLink{Label: asString(m["label"]), Href: asString(m["href"])})
	}
	return out
}

func asBreadcrumbItems(v any) []ui.BreadcrumbItem {
	items, ok := v.([]any)
	if !ok {
		return []ui.BreadcrumbItem{}
	}
	out := make([]ui.BreadcrumbItem, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.BreadcrumbItem{Label: asString(m["label"]), Href: asString(m["href"]), Current: asBool(m["current"])})
	}
	return out
}

func asDescriptionItems(v any) []ui.DescriptionItem {
	items, ok := v.([]any)
	if !ok {
		return []ui.DescriptionItem{}
	}
	out := make([]ui.DescriptionItem, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.DescriptionItem{Term: asString(m["term"]), Definition: asString(m["definition"])})
	}
	return out
}

func asPageItems(v any) []ui.PageItem {
	items, ok := v.([]any)
	if !ok {
		return []ui.PageItem{}
	}
	out := make([]ui.PageItem, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.PageItem{
			Label:    asString(m["label"]),
			Href:     asString(m["href"]),
			Current:  asBool(m["current"]),
			Disabled: asBool(m["disabled"]),
		})
	}
	return out
}

func asSelectOptions(v any) []ui.SelectOption {
	items, ok := v.([]any)
	if !ok {
		return []ui.SelectOption{}
	}
	out := make([]ui.SelectOption, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.SelectOption{Value: asString(m["value"]), Label: asString(m["label"]), Selected: asBool(m["selected"])})
	}
	return out
}

func asInteraction(v any) ui.Interaction {
	m := asMap(v)
	return ui.Interaction{
		Boost:   asBool(m["boost"]),
		Delete:  asString(m["delete"]),
		Get:     asString(m["get"]),
		Post:    asString(m["post"]),
		PushURL: asString(m["push_url"]),
		Put:     asString(m["put"]),
		Select:  asString(m["select"]),
		Swap:    asString(m["swap"]),
		Target:  asString(m["target"]),
		Trigger: asString(m["trigger"]),
	}
}

func asInteractiveMenuLinks(v any) []ui.InteractiveMenuLink {
	items, ok := v.([]any)
	if !ok {
		return []ui.InteractiveMenuLink{}
	}
	out := make([]ui.InteractiveMenuLink, 0, len(items))
	for _, item := range items {
		m := asMap(item)
		out = append(out, ui.InteractiveMenuLink{
			Label:       asString(m["label"]),
			Href:        asString(m["href"]),
			Interaction: asInteraction(m["interaction"]),
		})
	}
	return out
}

func canonicalType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	t = strings.ReplaceAll(t, "-", "_")
	t = strings.ReplaceAll(t, " ", "_")
	switch t {
	case "view":
		return "section"
	case "panel":
		return "article"
	case "text":
		return "paragraph"
	case "action":
		return "button"
	case "menu":
		return "nav"
	case "workspace":
		return "appshell"
	case "status":
		return "alert"
	case "toolbar":
		return "cluster"
	case "datagrid":
		return "table"
	default:
		return t
	}
}

func hasComponentSchema(typeName string) bool {
	_, ok := componentSchemas()[typeName]
	return ok
}

func componentSchemas() map[string]map[string]any {
	schemas := map[string]map[string]any{
		"alert":              {"props": []string{"text", "tone"}, "children": false},
		"appshell":           {"props": []string{}, "children": true, "child_slots": []string{"sidebar", "content"}},
		"article":            {"props": []string{"title"}, "children": true},
		"badge":              {"props": []string{"text", "tone"}, "children": false},
		"blockquote":         {"props": []string{"text", "cite"}, "children": false},
		"breadcrumbs":        {"props": []string{"items"}, "children": false},
		"button":             {"props": []string{"text", "kind"}, "children": false},
		"card":               {"props": []string{"title", "body"}, "children": false},
		"checkbox":           {"props": []string{"name", "value", "label", "checked"}, "children": false},
		"cluster":            {"props": []string{"extra_class"}, "children": true},
		"codeblock":          {"props": []string{"code"}, "children": false},
		"column":             {"props": []string{"size_class"}, "children": true},
		"container":          {"props": []string{}, "children": true},
		"container_fluid":    {"props": []string{}, "children": true},
		"descriptionlist":    {"props": []string{"items"}, "children": false},
		"details":            {"props": []string{"summary"}, "children": true},
		"fieldset":           {"props": []string{"legend"}, "children": true},
		"form":               {"props": []string{"action", "method"}, "children": true},
		"fragment":           {"props": []string{}, "children": true},
		"grid":               {"props": []string{"columns_class"}, "children": true},
		"grid_item":          {"props": []string{"span_class"}, "children": true},
		"heading":            {"props": []string{"level", "text"}, "children": false},
		"hero":               {"props": []string{"title", "subtitle"}, "children": true},
		"horizontal_rule":    {"props": []string{}, "children": false},
		"image":              {"props": []string{"src", "alt"}, "children": false},
		"input":              {"props": []string{"label", "name", "placeholder"}, "children": false},
		"interactive_action": {"props": []string{"text", "kind", "interaction"}, "children": false},
		"interactive_menu":   {"props": []string{"links"}, "children": false},
		"list":               {"props": []string{"items"}, "children": false},
		"nav":                {"props": []string{"links"}, "children": false},
		"ordered_list":       {"props": []string{"items"}, "children": false},
		"pagination":         {"props": []string{"items"}, "children": false},
		"paragraph":          {"props": []string{"text"}, "children": false},
		"radio":              {"props": []string{"name", "value", "label", "checked"}, "children": false},
		"row":                {"props": []string{}, "children": true},
		"section":            {"props": []string{"title"}, "children": true},
		"select":             {"props": []string{"label", "name", "options"}, "children": false},
		"sidebar_layout":     {"props": []string{}, "children": true, "child_slots": []string{"sidebar", "content"}},
		"split":              {"props": []string{}, "children": true, "child_slots": []string{"primary", "secondary"}},
		"stack":              {"props": []string{"extra_class"}, "children": true},
		"table":              {"props": []string{"headers", "rows"}, "children": false},
		"textarea":           {"props": []string{"label", "name", "placeholder"}, "children": false},
		"theme_toggle":       {"props": []string{}, "children": false},
	}
	for _, schema := range schemas {
		props, ok := schema["props"].([]string)
		if !ok {
			continue
		}
		hasID := false
		hasAttrs := false
		for _, prop := range props {
			if prop == "id" {
				hasID = true
			}
			if prop == "attrs" {
				hasAttrs = true
			}
		}
		if !hasID {
			props = append(props, "id")
		}
		if !hasAttrs {
			props = append(props, "attrs")
		}
		schema["props"] = props
	}
	return schemas
}

func mcpTools() []map[string]any {
	return []map[string]any{
		toolDef("session_create", "Create a UI authoring session", []string{"title", "provider"}),
		toolDef("session_get", "Get session state", []string{"session_id"}),
		toolDef("session_reset", "Reset session components and blocks", []string{"session_id"}),
		toolDef("session_delete", "Delete a session", []string{"session_id"}),
		toolDef("providers_list", "List available CSS providers", nil),
		toolDef("page_set_meta", "Update page title/lang", []string{"session_id", "title", "lang"}),
		toolDef("page_set_css_provider", "Set CSS provider", []string{"session_id", "provider"}),
		toolDef("page_set_head_snippet", "Append safe head snippet", []string{"session_id", "snippet"}),
		toolDef("schema_list_component_types", "List supported component types", nil),
		toolDef("schema_get_component_fields", "Get fields for one component type", []string{"type"}),
		toolDefWithTypedFields(
			"ui_create_component",
			"Create a component and place in tree",
			map[string]map[string]any{
				"session_id": {"type": "string"},
				"type":       {"type": "string"},
				"props":      {"type": "object", "additionalProperties": true},
				"parent_id":  {"type": "string"},
				"block":      {"type": "string"},
				"position":   {"type": "integer"},
			},
		),
		toolDefWithTypedFields(
			"ui_update_component",
			"Update component props/type",
			map[string]map[string]any{
				"session_id":   {"type": "string"},
				"component_id": {"type": "string"},
				"type":         {"type": "string"},
				"props":        {"type": "object", "additionalProperties": true},
			},
		),
		toolDef("ui_delete_component", "Delete component subtree", []string{"session_id", "component_id"}),
		toolDef("ui_move_component", "Move component within tree", []string{"session_id", "component_id", "new_parent_id", "new_block", "position"}),
		toolDef("ui_get_component", "Get one component", []string{"session_id", "component_id"}),
		toolDef("ui_list_components", "List components and block roots", []string{"session_id"}),
		toolDef("render_full", "Render full HTML page", []string{"session_id"}),
		toolDef("render_block", "Render a single block fragment", []string{"session_id", "block"}),
		toolDef("render_component", "Render one component subtree", []string{"session_id", "component_id"}),
	}
}

func toolDef(name, description string, fields []string) map[string]any {
	props := map[string]any{}
	for _, field := range fields {
		props[field] = map[string]any{"type": "string"}
	}
	return toolDefWithTypedFields(name, description, castSchemaFields(props))
}

func toolDefWithTypedFields(name, description string, fields map[string]map[string]any) map[string]any {
	return map[string]any{
		"name":        name,
		"description": description,
		"inputSchema": map[string]any{
			"type":       "object",
			"properties": fields,
		},
	}
}

func castSchemaFields(fields map[string]any) map[string]map[string]any {
	out := make(map[string]map[string]any, len(fields))
	for key, value := range fields {
		m, ok := value.(map[string]any)
		if !ok {
			continue
		}
		out[key] = m
	}
	return out
}

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetOutput(io.MultiWriter(os.Stderr, io.Discard))
}

func (a *app) safeDispatch(name string, args map[string]any) (_ map[string]any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid arguments: %v", r)
		}
	}()
	return a.dispatchTool(normalizeToolName(name), args)
}

func normalizeToolName(name string) string {
	switch name {
	case "session_create":
		return "session.create"
	case "session_get":
		return "session.get"
	case "session_reset":
		return "session.reset"
	case "session_delete":
		return "session.delete"
	case "providers_list":
		return "providers.list"
	case "page_set_meta":
		return "page.set_meta"
	case "page_set_css_provider":
		return "page.set_css_provider"
	case "page_set_head_snippet":
		return "page.set_head_snippet"
	case "schema_list_component_types":
		return "schema.list_component_types"
	case "schema_get_component_fields":
		return "schema.get_component_fields"
	case "ui_create_component":
		return "ui.create_component"
	case "ui_update_component":
		return "ui.update_component"
	case "ui_delete_component":
		return "ui.delete_component"
	case "ui_move_component":
		return "ui.move_component"
	case "ui_get_component":
		return "ui.get_component"
	case "ui_list_components":
		return "ui.list_components"
	case "render_full":
		return "render.full"
	case "render_block":
		return "render.block"
	case "render_component":
		return "render.component"
	default:
		return name
	}
}

func jsonText(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func escapeForText(s string) string {
	return html.EscapeString(s)
}

func toPrettyBytes(v any) []byte {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
	return bytes.TrimSpace(buf.Bytes())
}
