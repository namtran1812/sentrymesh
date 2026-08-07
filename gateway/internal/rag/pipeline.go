package rag

import (
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

type PipelineResult struct {
	Context []Document   `json:"context"`
	Trace   ContextTrace `json:"trace"`
}

func BuildContext(
	requestID string,
	principal identity.Identity,
	documents []Document,
) PipelineResult {
	contextDocs := make([]Document, 0)
	entries := make([]ContextEntry, 0, len(documents))

	for _, document := range documents {
		auth := Authorize(document, principal)

		if !auth.Allowed {
			entries = append(
				entries,
				ContextEntry{
					DocumentID:     document.ID,
					Source:         document.Source,
					TrustLevel:     document.TrustLevel,
					Classification: document.Classification,
					OwnerTeam:      document.OwnerTeam,
					Decision:       Block,
					Included:       false,
					Reason:         auth.Reason,
				},
			)

			continue
		}

		result := Inspect(document)

		entry := ContextEntry{
			DocumentID:     document.ID,
			Source:         document.Source,
			TrustLevel:     document.TrustLevel,
			Classification: document.Classification,
			OwnerTeam:      document.OwnerTeam,
			Decision:       result.Decision,
			Included:       false,
			Reason:         result.Reason,
		}

		if result.Decision == Block {
			entries = append(entries, entry)
			continue
		}

		safeDocument := document
		safeDocument.Content = result.SanitizedContent

		entry.Included = true

		contextDocs = append(
			contextDocs,
			safeDocument,
		)

		entries = append(
			entries,
			entry,
		)
	}

	return PipelineResult{
		Context: contextDocs,
		Trace: ContextTrace{
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
			Entries:   entries,
		},
	}
}
