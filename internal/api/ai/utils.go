package ai

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func fetchWebpageBody(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch webpage: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	content := string(body)

	// Extract body content
	bodyStart := strings.Index(strings.ToLower(content), "<body")
	if bodyStart == -1 {
		return content, nil // Return the whole content if no body tag found
	}

	bodyEnd := strings.LastIndex(strings.ToLower(content), "</body>") + 7
	if bodyEnd == -1+7 {
		return content[bodyStart:], nil
	}

	bodyContent := content[bodyStart:bodyEnd]

	// Remove non-essential elements
	bodyContent = removeTagsAndContent(bodyContent, "script")
	bodyContent = removeTagsAndContent(bodyContent, "style")
	bodyContent = removeTagsAndContent(bodyContent, "svg")
	bodyContent = removeHTMLComments(bodyContent)
	bodyContent = removeTagsAndContent(bodyContent, "iframe")
	bodyContent = removeTagsAndContent(bodyContent, "nav")
	bodyContent = removeTagsAndContent(bodyContent, "footer")
	bodyContent = removeTagsAndContent(bodyContent, "header")
	bodyContent = removeTagsAndContent(bodyContent, "button")
	bodyContent = removeTagsAndContent(bodyContent, "noscript")

	// Remove form elements
	bodyContent = removeTagsAndContent(bodyContent, "input")
	bodyContent = removeTagsAndContent(bodyContent, "textarea")
	bodyContent = removeTagsAndContent(bodyContent, "select")

	// Remove attributes
	bodyContent = removeDataAttributes(bodyContent)
	bodyContent = removeAriaAttributes(bodyContent)

	return bodyContent, nil
}

// Helper function to remove specified HTML tags and their content
func removeTagsAndContent(html, tag string) string {
	lowercaseHTML := strings.ToLower(html)
	result := html

	for {
		tagStart := strings.Index(lowercaseHTML, "<"+tag)
		if tagStart == -1 {
			break
		}

		tagEnd := strings.Index(lowercaseHTML[tagStart:], "</"+tag+">")
		if tagEnd == -1 {
			break
		}
		tagEnd += tagStart + len(tag) + 3 // Add length of "</tag>"

		// Update both the result and the lowercase version for next iteration
		result = result[:tagStart] + result[tagEnd:]
		lowercaseHTML = strings.ToLower(result)
	}

	return result
}

// Helper function to remove HTML comments
func removeHTMLComments(html string) string {
	result := html
	for {
		commentStart := strings.Index(result, "<!--")
		if commentStart == -1 {
			break
		}

		commentEnd := strings.Index(result[commentStart:], "-->")
		if commentEnd == -1 {
			break
		}
		commentEnd += commentStart + 3 // Add length of "-->"

		result = result[:commentStart] + result[commentEnd:]
	}

	return result
}

// Helper function to remove data-* attributes
func removeDataAttributes(html string) string {
	result := html
	for {
		dataAttrStart := strings.Index(strings.ToLower(result), " data-")
		if dataAttrStart == -1 {
			break
		}

		// Find the end of the attribute (next space or >)
		attrEndSpace := strings.Index(result[dataAttrStart+1:], " ")
		attrEndBracket := strings.Index(result[dataAttrStart+1:], ">")

		var attrEnd int
		if attrEndSpace == -1 && attrEndBracket == -1 {
			break
		} else if attrEndSpace == -1 {
			attrEnd = dataAttrStart + 1 + attrEndBracket
		} else if attrEndBracket == -1 {
			attrEnd = dataAttrStart + 1 + attrEndSpace
		} else {
			attrEnd = dataAttrStart + 1 + min(attrEndSpace, attrEndBracket)
		}

		// Remove the attribute
		result = result[:dataAttrStart] + result[attrEnd:]
	}

	return result
}

// Helper function to remove aria-* attributes
func removeAriaAttributes(html string) string {
	result := html
	for {
		ariaAttrStart := strings.Index(strings.ToLower(result), " aria-")
		if ariaAttrStart == -1 {
			break
		}

		// Find the end of the attribute (next space or >)
		attrEndSpace := strings.Index(result[ariaAttrStart+1:], " ")
		attrEndBracket := strings.Index(result[ariaAttrStart+1:], ">")

		var attrEnd int
		if attrEndSpace == -1 && attrEndBracket == -1 {
			break
		} else if attrEndSpace == -1 {
			attrEnd = ariaAttrStart + 1 + attrEndBracket
		} else if attrEndBracket == -1 {
			attrEnd = ariaAttrStart + 1 + attrEndSpace
		} else {
			attrEnd = ariaAttrStart + 1 + min(attrEndSpace, attrEndBracket)
		}

		// Remove the attribute
		result = result[:ariaAttrStart] + result[attrEnd:]
	}

	return result
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
