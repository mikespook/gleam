package node

var (
    ErrHandler ErrorHandlerFunc
)

type ErrorHandlerFunc func(error)


func _err(err error) {
    if ErrHandler != nil {
        ErrHandler(err)
    }
}
