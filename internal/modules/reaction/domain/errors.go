package domain
import "errors"
var (
    ErrReactionNotFound = errors.New("reaction not found")
    ErrDuplicateReaction = errors.New("reaction already exists")
)
