package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrUnauthorized() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		StatusText:     "Authentication failed.",
		ErrorText:      "Invalid email or password.",
	}
}

func ErrConflict(message string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusConflict,
		StatusText:     "Request conflict.",
		ErrorText:      message,
	}
}

func ErrInternal(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error.",
		ErrorText:      err.Error(),
	}
}
