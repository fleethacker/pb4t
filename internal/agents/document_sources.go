package agents

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/fleethacker/pb4t/internal/planner"
	"rsc.io/pdf"
)

const sourceReviewMissing = "Source excerpt/page could not be matched locally; verify against the original PDF"

// CheckDocumentSources locates literal excerpts on physical PDF pages. This
// checks the source quotation, not the model's interpretation of the policy.
func CheckDocumentSources(doc *planner.TrainingDocument, original []byte) {
	pages, err := documentPages(original)
	for i := range doc.Suggestions {
		s := &doc.Suggestions[i]
		s.Missing = slices.DeleteFunc(s.Missing, func(note string) bool { return note == sourceReviewMissing })
		s.CitationStatus = "needs_review"
		parts := []string{}
		for _, part := range strings.Split(strings.ReplaceAll(s.Excerpt, "…", "..."), "...") {
			if normalized := sourceText(part); normalized != "" {
				parts = append(parts, normalized)
			}
		}
		page := 0
		if err == nil && len(strings.Join(parts, "")) >= 20 {
			for j := range pages {
				if sourcePartsMatch(pages, j, parts) {
					if page == 0 || j+1 == s.Page {
						page = j + 1
					}
				}
			}
		}
		if page > 0 {
			s.Page = page
			s.CitationStatus = "matched"
		} else {
			s.Missing = append(s.Missing, sourceReviewMissing)
		}
	}
}

// Ellipses may omit text, but every quoted segment must appear literally and
// in order. A quote can continue onto the next page; its link starts on the
// page containing the first segment. Paraphrases are never fuzzy-matched.
func sourcePartsMatch(pages []string, start int, parts []string) bool {
	text := pages[start]
	if start+1 < len(pages) {
		text += pages[start+1]
	}
	position := 0
	for i, part := range parts {
		offset := strings.Index(text[position:], part)
		if offset < 0 || (i == 0 && position+offset >= len(pages[start])) {
			return false
		}
		position += offset + len(part)
	}
	return len(parts) > 0
}

func sourceText(text string) string {
	text = strings.NewReplacer("ﬁ", "fi", "ﬂ", "fl").Replace(text)
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, text)
}

func documentPages(original []byte) (pages []string, err error) {
	// The PDF reader may panic on unsupported content; extraction can still
	// return reviewable drafts with an explicit unverified-source notice.
	defer func() {
		if failure := recover(); failure != nil {
			pages = nil
			err = fmt.Errorf("read PDF text: %v", failure)
		}
	}()
	r, err := pdf.NewReader(bytes.NewReader(original), int64(len(original)))
	if err != nil {
		return nil, err
	}
	for n := 1; n <= r.NumPage(); n++ {
		var text strings.Builder
		for _, t := range r.Page(n).Content().Text {
			text.WriteString(t.S)
		}
		pages = append(pages, sourceText(text.String()))
	}
	return pages, nil
}
