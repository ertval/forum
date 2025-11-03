// INPUT ADAPTER - HTTP Handler
package adapters
import ("forum/internal/modules/reaction/ports"; "net/http")
type HTTPHandler struct { reactionService ports.ReactionService }
func NewHTTPHandler(reactionService ports.ReactionService) *HTTPHandler {
    return &HTTPHandler{reactionService: reactionService}
}
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
    // TODO: POST /reactions, DELETE /reactions, GET /reactions
}
