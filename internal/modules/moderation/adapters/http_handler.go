// INPUT ADAPTER - HTTP Handler
// [OPTIONAL FEATURE: forum-moderation]
package adapters
import ("forum/internal/modules/moderation/ports"; "net/http")
type HTTPHandler struct { moderationService ports.ModerationService }
func NewHTTPHandler(moderationService ports.ModerationService) *HTTPHandler {
    return &HTTPHandler{moderationService: moderationService}
}
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
    // TODO: POST /reports, GET /reports, PUT /reports/{id}
}
