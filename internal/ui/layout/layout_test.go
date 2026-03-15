package layout

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderToDoc(t *testing.T, component templ.Component) *goquery.Document {
	t.Helper()
	r, w := io.Pipe()
	go func() {
		err := component.Render(context.Background(), w)
		if err != nil {
			t.Errorf("failed to render component: %v", err)
		}

		if err := w.Close(); err != nil {
			t.Errorf("failed to close pipe writer: %v", err)
		}
	}()

	doc, err := goquery.NewDocumentFromReader(r)
	require.NoError(t, err)
	return doc
}

func TestBaseLayout(t *testing.T) {
	t.Parallel()

	props := BaseProps{
		Title:       "Test Dashboard",
		CurrentPath: "/dashboard",
		User: &domain.User{
			Name:  "Test User",
			Email: "test@example.com",
			Role:  "user",
		},
		Tenant: &domain.Tenant{
			Name: "Test Household",
		},
		Content: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := io.WriteString(w, `<div id="test-content">Welcome to Moolah</div>`)
			if err != nil {
				return fmt.Errorf("failed to write string: %w", err)
			}
			return nil
		}),
	}

	t.Run("renders basic structure", func(t *testing.T) {
		t.Parallel()
		doc := renderToDoc(t, base(props))

		// Check title
		assert.Equal(t, "Test Dashboard — Moolah", doc.Find("title").Text())

		// Check main content slot
		assert.Equal(t, "Welcome to Moolah", doc.Find("#main-content").Text())
		assert.Positive(t, doc.Find("#test-content").Length())

		// Check sidebar links
		dashboardLink := doc.Find("#sidebar section a[href='/dashboard']").First()
		assert.Positive(t, dashboardLink.Length())
		class, _ := dashboardLink.Attr("class")
		assert.Contains(t, class, "bg-brand-50") // active state
	})

	t.Run("renders admin links only for sysadmin", func(t *testing.T) {
		t.Parallel()
		// Non-admin props
		docUser := renderToDoc(t, base(props))
		assert.Zero(t, docUser.Find("a[href='/admin/tenants']").Length())

		// Admin props
		adminProps := props
		adminProps.User = &domain.User{Name: "Admin User", Role: "sysadmin", Email: "admin@example.com"}
		docAdmin := renderToDoc(t, base(adminProps))
		assert.Positive(t, docAdmin.Find("a[href='/admin/tenants']").Length())
	})

	t.Run("includes required scripts", func(t *testing.T) {
		t.Parallel()
		doc := renderToDoc(t, base(props))

		scripts := doc.Find("script")
		var alpine, htmx bool
		scripts.Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			if strings.Contains(src, "alpine.min.js") {
				alpine = true
			}
			if strings.Contains(src, "htmx.min.js") {
				htmx = true
			}
		})

		assert.True(t, alpine, "Alpine.js should be included")
		assert.True(t, htmx, "HTMX should be included")
	})

	t.Run("includes theme toggle", func(t *testing.T) {
		t.Parallel()
		doc := renderToDoc(t, base(props))
		assert.Equal(t, 1, doc.Find("button[title='Toggle theme']").Length(), "Theme toggle button should be present")
	})

	t.Run("renders user initials correctly", func(t *testing.T) {
		t.Parallel()
		customProps := props
		customProps.User = &domain.User{Name: "Garnizé", Email: "g@example.com"}
		doc := renderToDoc(t, base(customProps))

		// Check topbar initial
		topbarInitial := doc.Find("header .rounded-full").First().Text()
		assert.Equal(t, "G", strings.TrimSpace(topbarInitial))

		// Check sidebar initial
		sidebarInitial := doc.Find("aside .rounded-full").First().Text()
		assert.Equal(t, "G", strings.TrimSpace(sidebarInitial))
	})

	t.Run("handles nil user and tenant gracefully", func(t *testing.T) {
		t.Parallel()
		nilProps := BaseProps{
			Title:   "Public Page",
			Content: props.Content,
		}
		doc := renderToDoc(t, base(nilProps))

		// Should not find user-specific elements
		assert.Zero(t, doc.Find("header .rounded-full").Length())
		assert.Zero(t, doc.Find("aside .p-4.border-t").Length())
	})

	t.Run("handles empty user name gracefully", func(t *testing.T) {
		t.Parallel()
		emptyNameProps := props
		emptyNameProps.User = &domain.User{Name: "", Email: "empty@example.com"}
		doc := renderToDoc(t, base(emptyNameProps))

		// Should not render avatar if name is empty
		assert.Zero(t, doc.Find("header .rounded-full").Length())
		assert.Zero(t, doc.Find("aside .p-4.border-t").Length())
	})

	t.Run("renders logout link", func(t *testing.T) {
		t.Parallel()
		doc := renderToDoc(t, base(props))
		logoutLink := doc.Find("a[href='/auth/logout']").First()
		assert.Positive(t, logoutLink.Length())
	})

	t.Run("renders all sidebar links including icons", func(t *testing.T) {
		t.Parallel()
		adminProps := props
		adminProps.User = &domain.User{Name: "Admin User", Role: "sysadmin", Email: "admin@example.com"}
		doc := renderToDoc(t, base(adminProps))

		links := []string{
			"/dashboard", "/accounts", "/transactions",
			"/categories", "/master-purchases",
			"/investments/portfolio", "/investments/positions", "/investments/income",
			"/admin/tenants", "/admin/users", "/admin/audit-logs",
		}

		for _, href := range links {
			link := doc.Find(fmt.Sprintf("a[href='%s']", href))
			assert.Positive(t, link.Length(), "Link %s should be present", href)
			// Icons are rendered inside the link, checking for presence of svg
			assert.Positive(t, link.Find("svg").Length(), "Icon for link %s should be present", href)
		}
	})
}
