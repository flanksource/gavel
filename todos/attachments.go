package todos

import "strings"

// AttachmentURLPrefix is the dashboard path stored attachments are served from.
// A todo body embeds `<prefix><id>` as a markdown link or image target, which
// only resolves against a running gavel server.
const AttachmentURLPrefix = "/api/todos/attachments/"

// attachmentLinkMarker matches an attachment used as a markdown link or image
// target, which is the only form todo bodies write.
const attachmentLinkMarker = "](" + AttachmentURLPrefix

// HasAttachmentURLs reports whether a body still carries server-relative
// attachment links. Anything outside the dashboard — an agent prompt, a GitHub
// issue — cannot fetch those, so callers use this to decide whether an absolute
// origin is required.
func HasAttachmentURLs(body string) bool {
	return strings.Contains(body, attachmentLinkMarker)
}

// AbsolutizeAttachmentURLs rewrites the relative attachment links stored in a
// todo body to absolute URLs rooted at origin. It matches only markdown
// link/image targets and is idempotent: once rewritten the link no longer
// carries the relative marker, so a second pass leaves it untouched.
func AbsolutizeAttachmentURLs(body, origin string) string {
	if origin == "" || !HasAttachmentURLs(body) {
		return body
	}
	return strings.ReplaceAll(body, attachmentLinkMarker, "]("+origin+AttachmentURLPrefix)
}
