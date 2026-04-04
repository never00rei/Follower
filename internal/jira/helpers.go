package jira

import "strings"

func plainTextToADF(text string) ADFDocument {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	paragraphs := strings.Split(text, "\n\n")
	content := make([]ADFNode, 0, len(paragraphs))

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		content = append(content, ADFNode{
			Type: "paragraph",
			Content: []ADFNode{
				{
					Type: "text",
					Text: paragraph,
				},
			},
		})
	}

	if len(content) == 0 {
		content = append(content, ADFNode{
			Type: "paragraph",
			Content: []ADFNode{
				{
					Type: "text",
					Text: "",
				},
			},
		})
	}

	return ADFDocument{
		Type:    "doc",
		Version: 1,
		Content: content,
	}
}
