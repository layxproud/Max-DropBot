package shortener

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func NewServer(s *Shortener) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/f/", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Path[len("/f/"):]
		if token == "" {
			http.NotFound(w, r)
			return
		}

		target, err := s.Resolve(r.Context(), token)
		if err != nil {
			http.NotFound(w, r)
			log.Warn().
				Str("token", token).
				Msg("short URL not found")
			return
		}

		http.Redirect(w, r, target, http.StatusFound)
	})
	return mux
}
