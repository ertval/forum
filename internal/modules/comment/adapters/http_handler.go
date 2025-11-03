// INPUT ADAPTER - HTTP Handler
package adapters
import ("forum/internal/modules/comment/ports"; "net/http")
type HTTPHandler struct { commentService ports.CommentService }
func NewHTTPHandler(commentService ports.CommentService) *HTTPHandler {
    return &HTTPHandler{commentService: commentService}
}
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
    // TODO: POST /comments, GET /comments/{id}, PUT /comments/{id}, DELETE /comments/{id}
}
