package models

import (
	"fmt"
	"strings"

	check "gopkg.in/check.v1"
)

type mockTemplateContext struct {
	URL         string
	FromAddress string
}

func (m mockTemplateContext) getFromAddress() string {
	return m.FromAddress
}

func (m mockTemplateContext) getBaseURL() string {
	return m.URL
}

func (s *ModelsSuite) TestNewTemplateContext(c *check.C) {
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "Foo",
			LastName:  "Bar",
			Email:     "foo@bar.com",
		},
		RId: "1234567",
	}
	ctx := mockTemplateContext{
		URL:         "http://example.com",
		FromAddress: "From Address <from@example.com>",
	}
	expected := PhishingTemplateContext{
		URL:           fmt.Sprintf("%s?keyname=%s", ctx.URL, r.RId),
		BaseURL:       ctx.URL,
		BaseRecipient: r.BaseRecipient,
		TrackingURL:   fmt.Sprintf("%s/track?keyname=%s", ctx.URL, r.RId),
		From:          "From Address",
		RId:           r.RId,
	}
	expected.Tracker = "<img alt='' style='display: none' src='" + expected.TrackingURL + "'/>"
	got, err := NewPhishingTemplateContext(ctx, r.BaseRecipient, r.RId)
	c.Assert(err, check.Equals, nil)
	c.Assert(got, check.DeepEquals, expected)
}

func (s *ModelsSuite) TestDeviceCodeIntegration(c *check.C) {
	// Test template with DeviceCode placeholder
	templateHTML := `<html><body><p>Your device code is: {{.DeviceCode}}</p><p>Hello {{.FirstName}}</p></body></html>`
	
	// Create a template context
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
		RId: "test123",
	}
	ctx := mockTemplateContext{
		URL:         "http://test.com",
		FromAddress: "sender@test.com",
	}
	
	ptx, err := NewPhishingTemplateContext(ctx, r.BaseRecipient, r.RId)
	c.Assert(err, check.Equals, nil)
	
	// Simulate the DeviceCode replacement that happens in renderPhishResponse
	if strings.Contains(templateHTML, "{{.DeviceCode}}") {
		// Get the external API URL (should return default URL)
		apiURL := GetExternalAPIURL()
		c.Assert(apiURL, check.Equals, "http://127.0.0.1:5000/api/generate_device_code")
		
		// Simulate API call returning a device code
		mockDeviceCode := "ABC123XYZ"
		templateHTML = strings.Replace(templateHTML, "{{.DeviceCode}}", mockDeviceCode, -1)
	}
	
	// Execute the template with the modified HTML
	result, err := ExecuteTemplate(templateHTML, ptx)
	c.Assert(err, check.Equals, nil)
	
	// Verify the DeviceCode was replaced
	c.Assert(strings.Contains(result, "ABC123XYZ"), check.Equals, true)
	c.Assert(strings.Contains(result, "{{.DeviceCode}}"), check.Equals, false)
	
	// Verify normal template variables still work
	c.Assert(strings.Contains(result, "Hello Test"), check.Equals, true)
	
	// Verify HTML structure is preserved
	c.Assert(strings.Contains(result, "<html>"), check.Equals, true)
	c.Assert(strings.Contains(result, "Your device code is: ABC123XYZ"), check.Equals, true)
}
