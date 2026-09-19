package httpapi

import "net/http"

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	session, err := s.auth.Register(r.Context(), in.Email, in.Name, in.Password)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	session, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r))
}
