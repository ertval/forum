package domain
import "errors"
var (
    ErrCommentNotFound = errors.New("comment not found")
    ErrUnauthorizedEdit = errors.New("unauthorized to edit comment")
)
