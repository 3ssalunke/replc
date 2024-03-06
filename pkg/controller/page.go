package controller

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/3ssalunke/replc/pkg/htmx"
	"github.com/3ssalunke/replc/templates"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

// Page consists of all data that will be used to render a page response for a given controller.
// While it's not required for a controller to render a Page on a route, this is the common data
// object that will be passed to the templates, making it easy for all controllers to share
// functionality both on the back and frontend. The Page can be expanded to include anything else
// your app wants to support.
// Methods on this page also then become available in the templates, which can be more useful than
// the funcmap if your methods require data stored in the page, such as the context.
type Page struct {
	// AppName stores the name of the application.
	// If omitted, the configuration value will be used.
	AppName string

	// Title stores the title of the page
	Title string

	// Context stores the request context
	Context echo.Context

	// ToURL is a function to convert a route name and optional route parameters to a URL
	ToURL func(name string, params ...any) string

	// Path stores the path of the current request
	Path string

	// URL stores the URL of the current request
	URL string

	// Data stores whatever additional data that needs to be passed to the templates.
	// This is what the controller uses to pass the content of the page.
	Data any

	// Form stores a struct that represents a form on the page.
	// This should be a struct with fields for each form field, using both "form" and "validate" tags
	// It should also contain a Submission field of type FormSubmission if you wish to have validation
	// messagesa and markup presented to the user
	Form any

	// Layout stores the name of the layout base template file which will be used when the page is rendered.
	// This should match a template file located within the layouts directory inside the templates directory.
	// The template extension should not be included in this value.
	Layout templates.Layout

	// Name stores the name of the page as well as the name of the template file which will be used to render
	// the content portion of the layout template.
	// This should match a template file located within the pages directory inside the templates directory.
	// The template extension should not be included in this value.
	Name templates.Page

	// IsHome stores whether the requested page is the home page or not
	IsHome bool

	// StatusCode stores the HTTP status code that will be returned
	StatusCode int

	// Metatags stores metatag values
	Metatags struct {
		// Description stores the description metatag value
		Description string

		// Keywords stores the keywords metatag values
		Keywords []string
	}

	// CSRF stores the CSRF token for the given request.
	// This will only be populated if the CSRF middleware is in effect for the given request.
	// If this is populated, all forms must include this value otherwise the requests will be rejected.
	CSRF string

	// Headers stores a list of HTTP headers and values to be set on the response
	Headers map[string]string

	// RequestID stores the ID of the given request.
	// This will only be populated if the request ID middleware is in effect for the given request.
	RequestID string

	HTMX struct {
		Request  htmx.Request
		Response *htmx.Response
	}
}

// NewPage creates and initiatizes a new Page for a given request context
func NewPage(ctx echo.Context) Page {
	p := Page{
		Context:    ctx,
		ToURL:      ctx.Echo().Reverse,
		Path:       ctx.Request().URL.Path,
		URL:        ctx.Request().URL.String(),
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		RequestID:  ctx.Response().Header().Get(echo.HeaderXRequestID),
	}

	p.IsHome = p.Path == "/"

	if csrf := ctx.Get(echomw.DefaultCSRFConfig.ContextKey); csrf != nil {
		p.CSRF = csrf.(string)
	}

	p.HTMX.Request = htmx.GetRequest(ctx)

	return p
}

// RenderPage renders a Page as an HTTP response
func (c *Controller) RenderPage(ctx echo.Context, page Page) error {
	var buf *bytes.Buffer
	var err error
	templateGroup := "page"

	// Page name is required
	if page.Name == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "page render failed due to missing name")
	}

	// Use the app name in configuration if a value was not set
	if page.AppName == "" {
		page.AppName = c.Container.Config.App.Name
	}

	// Check if this is an HTMX non-boosted request which indicates that only partial
	// content should be rendered
	if page.HTMX.Request.Enabled && !page.HTMX.Request.Boosted {
		// Switch the layout which will only render the page content
		page.Layout = templates.LayoutHTMX

		// Alter the template group so this is cached separately
		templateGroup = "page:htmx"
	}

	// Parse and execute the templates for the Page
	// As mentioned in the documentation for the Page struct, the templates used for the page will be:
	// 1. The layout/base template specified in Page.Layout
	// 2. The content template specified in Page.Name
	// 3. All templates within the components directory
	// Also included is the function map provided by the funcmap package
	buf, err = c.Container.TemplateRenderer.
		Parse().
		Group(templateGroup).
		Key(string(page.Name)).
		Base(string(page.Layout)).
		Files(
			fmt.Sprintf("layouts/%s", page.Layout),
			fmt.Sprintf("pages/%s", page.Name),
		).
		Directories("components").
		Execute(page)

	if err != nil {
		return c.Fail(err, "failed to parse and execute templates")
	}

	// Set the status code
	ctx.Response().Status = page.StatusCode

	// Set any headers
	for k, v := range page.Headers {
		ctx.Response().Header().Set(k, v)
	}

	// Apply the HTMX response, if one
	if page.HTMX.Response != nil {
		page.HTMX.Response.Apply(ctx)
	}

	return ctx.HTMLBlob(ctx.Response().Status, buf.Bytes())
}

// Redirect redirects to a given route name with optional route parameters
func (c *Controller) Redirect(ctx echo.Context, route string, routeParams ...any) error {
	url := ctx.Echo().Reverse(route, "abc?def")

	log.Println(route, routeParams, url)
	if htmx.GetRequest(ctx).Boosted {
		htmx.Response{
			Redirect: url,
		}.Apply(ctx)

		return nil
	} else {
		return ctx.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// Fail is a helper to fail a request by returning a 500 error and logging the error
func (c *Controller) Fail(err error, log string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s: %v", log, err))
}
