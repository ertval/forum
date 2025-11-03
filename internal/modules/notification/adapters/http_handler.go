// INPUT ADAPTER - HTTP Handler
// [OPTIONAL FEATURE: forum-advanced-features]
package adapters
import ("forum/internal/modules/notification/ports"; "net/http")
type HTTPHandler struct { notificationService ports.NotificationService }
func NewHTTPHandler(notificationService ports.NotificationService) *HTTPHandler {
    return &HTTPHandler{notificationService: notificationService}
}
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
    // TODO: GET /notifications, PUT /notifications/{id}/read
}
