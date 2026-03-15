package components

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func render(component templ.Component) (string, error) {
	var sb strings.Builder
	err := component.Render(context.Background(), &sb)
	return sb.String(), err
}

func TestButton(t *testing.T) {
	t.Parallel()
	t.Run("renders with label", func(t *testing.T) {
		t.Parallel()
		component := Button(ButtonProps{Label: "Click Me"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "Click Me")
		require.Contains(t, html, "bg-brand-600")
	})

	t.Run("renders loading state", func(t *testing.T) {
		t.Parallel()
		component := Button(ButtonProps{Label: "Loading", Loading: true})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "aria-busy=\"true\"")
		require.Contains(t, html, "animate-spin")
	})
}

func TestBadge(t *testing.T) {
	t.Parallel()
	t.Run("renders variants correctly", func(t *testing.T) {
		t.Parallel()
		component := Badge(BadgeProps{Label: "Success", Variant: BadgeSuccess})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "bg-success-50")
		require.Contains(t, html, "Success")
	})
}

func TestInput(t *testing.T) {
	t.Parallel()
	t.Run("renders with error", func(t *testing.T) {
		t.Parallel()
		component := Input(InputProps{ID: "test-input", Error: "Field is required"})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)
		input := doc.Find("input")
		require.Equal(t, "true", input.AttrOr("aria-invalid", ""))
		require.Equal(t, "test-input-error", input.AttrOr("aria-describedby", ""))
	})
}

func TestCurrencyAmount(t *testing.T) {
	t.Parallel()
	t.Run("formats positive BRL correctly", func(t *testing.T) {
		t.Parallel()
		component := CurrencyAmount(CurrencyAmountProps{Cents: 1250, CurrencyCode: "BRL"})
		html, err := render(component)
		require.NoError(t, err)

		// Accept both dot or comma based on runner locale
		require.True(t, strings.Contains(html, "12.50") || strings.Contains(html, "12,50"), "HTML should contain formatted amount: %s", html)
		require.Contains(t, html, "R$")
		require.Contains(t, html, "text-success-600")
	})

	t.Run("formats negative USD with sign", func(t *testing.T) {
		t.Parallel()
		component := CurrencyAmount(CurrencyAmountProps{Cents: -500, CurrencyCode: "USD", ShowSign: true})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "- $ 5.00")
		require.Contains(t, html, "text-danger-600")
	})
}

func TestSelect(t *testing.T) {
	t.Parallel()
	t.Run("renders options correctly", func(t *testing.T) {
		t.Parallel()
		options := []SelectOption{
			{Label: "Option 1", Value: "1"},
			{Label: "Option 2", Value: "2", Selected: true},
		}
		component := Select(SelectProps{ID: "test-select", Options: options})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		selectEl := doc.Find("select")
		require.Equal(t, "test-select", selectEl.AttrOr("id", ""))
		require.Equal(t, 2, selectEl.Find("option").Length())
		require.Equal(t, "Option 2", selectEl.Find("option[selected]").Text())
	})
}

func TestTextarea(t *testing.T) {
	t.Parallel()
	t.Run("renders with value and rows", func(t *testing.T) {
		t.Parallel()
		component := Textarea(TextareaProps{ID: "test-area", Value: "Hello World", Rows: 5})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		textarea := doc.Find("textarea")
		require.Equal(t, "test-area", textarea.AttrOr("id", ""))
		require.Equal(t, "5", textarea.AttrOr("rows", ""))
		require.Equal(t, "Hello World", textarea.Text())
	})
}

func TestCheckbox(t *testing.T) {
	t.Parallel()
	t.Run("renders label and checked state", func(t *testing.T) {
		t.Parallel()
		component := Checkbox(CheckboxProps{ID: "test-check", Label: "Accept terms", Checked: true})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		input := doc.Find("input[type=\"checkbox\"]")
		require.Equal(t, "test-check", input.AttrOr("id", ""))
		require.True(t, input.Is("[checked]"))
		require.Contains(t, doc.Find("label").Text(), "Accept terms")
	})
}

func TestModal(t *testing.T) {
	t.Parallel()
	t.Run("renders accessibility attributes", func(t *testing.T) {
		t.Parallel()
		component := Modal(ModalProps{
			ID:      "test-modal",
			Title:   "Modal Title",
			Content: templ.Raw("<div>Content</div>"),
		})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		modal := doc.Find("#test-modal")
		require.Equal(t, "dialog", modal.AttrOr("role", ""))
		require.Equal(t, "true", modal.AttrOr("aria-modal", ""))
		require.Contains(t, modal.AttrOr("aria-labelledby", ""), "test-modal-title")
		require.Contains(t, doc.Find("h3").Text(), "Modal Title")
	})
}

func TestDrawer(t *testing.T) {
	t.Parallel()
	t.Run("renders with title and size", func(t *testing.T) {
		t.Parallel()
		component := Drawer(DrawerProps{
			ID:      "test-drawer",
			Title:   "Drawer Title",
			Size:    DrawerSizeLg,
			Content: templ.Raw("<div>Content</div>"),
		})
		html, err := render(component)
		require.NoError(t, err)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		drawer := doc.Find("#test-drawer")
		require.Equal(t, "dialog", drawer.AttrOr("role", ""))
		require.Contains(t, doc.Find("h2").Text(), "Drawer Title")
		require.Contains(t, html, "max-w-lg")
	})
}

func TestToast(t *testing.T) {
	t.Parallel()
	t.Run("renders error toast with alert role", func(t *testing.T) {
		t.Parallel()
		component := Toast(ToastProps{ID: "t1", Variant: ToastError, Title: "Error", Message: "Something failed"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "role=\"alert\"")
		require.Contains(t, html, "Error")
		require.Contains(t, html, "Something failed")
	})

	t.Run("renders info toast with status role", func(t *testing.T) {
		t.Parallel()
		component := Toast(ToastProps{ID: "t2", Variant: ToastInfo, Title: "Info"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "role=\"status\"")
	})
}

func TestSpinner(t *testing.T) {
	t.Parallel()
	t.Run("renders with status role", func(t *testing.T) {
		t.Parallel()
		component := Spinner(SpinnerProps{Size: "lg"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "role=\"status\"")
		require.Contains(t, html, "Loading")
		require.Contains(t, html, "h-8 w-8")
	})
}

func TestSkeleton(t *testing.T) {
	t.Parallel()
	t.Run("renders as decorative", func(t *testing.T) {
		t.Parallel()
		component := Skeleton(SkeletonProps{Shape: SkeletonCircle, Class: "h-12 w-12"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "aria-hidden=\"true\"")
		require.Contains(t, html, "rounded-full")
	})
}

func TestEmptyState(t *testing.T) {
	t.Parallel()
	t.Run("renders title and description", func(t *testing.T) {
		t.Parallel()
		component := EmptyState(EmptyStateProps{Title: "No data", Description: "Try again later"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "No data")
		require.Contains(t, html, "Try again later")
	})
}

func TestPageHeader(t *testing.T) {
	t.Parallel()
	t.Run("renders title and subtitle", func(t *testing.T) {
		t.Parallel()
		component := PageHeader(PageHeaderProps{Title: "Accounts", Subtitle: "Manage your accounts"})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "Accounts")
		require.Contains(t, html, "Manage your accounts")
	})
}

func TestCard(t *testing.T) {
	t.Parallel()
	t.Run("renders header and padding", func(t *testing.T) {
		t.Parallel()
		header := templ.Raw("<h3>Header</h3>")
		component := Card(CardProps{Header: header, Padding: CardPaddingSm})
		html, err := render(component)
		require.NoError(t, err)

		require.Contains(t, html, "<h3>Header</h3>")
		require.Contains(t, html, "px-4 py-4")
	})
}
