package teamserver


import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/z3vxo/kronos/internal/database"
	"github.com/z3vxo/kronos/internal/hadesgen"
	"github.com/z3vxo/kronos/internal/httputil"
)


func (ts *TeamServer) HandleNewProfile(w http.ResponseWriter, r *http.Request) {
	var req database.Profile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.SendJSONError(w, "failed decoding json", http.StatusInternalServerError)
		return
	}

	id := uint32(time.Now().UnixNano() >> 1)
	var keyBytes [4]byte
	rand.Read(keyBytes[:])

	if err := ts.db.InsertProfile(req, id, keyBytes[:]); err != nil {
		httputil.SendJSONError(w, "failed inseting profile into db", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (ts *TeamServer) HandleListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := ts.db.ListProfiles()
	if err != nil {
		httputil.SendJSONError(w, "failed getting profiles from db", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profiles)
}

func (ts *TeamServer) HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "profile")

	if err := ts.db.DeleteProfile(name); err != nil {
		httputil.SendJSONError(w, "failed deleting profile from db", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})

}


func (ts *TeamServer) HandleProfileGenerate(w http.ResponseWriter, r *http.Request) {
	profileName := chi.URLParam(r, "profile")

	var req GeneratePayloadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.SendJSONError(w, "failed decoding json", http.StatusBadRequest)
		return
	}

	profile, err := ts.db.GetProfile(profileName)
	if err != nil {
		httputil.SendJSONError(w, "failed retrieving profile from db", http.StatusInternalServerError)
		return
	}

	compiledPath, err := hadesgen.GeneratePayload(profile, req.Name, req.Format, req.Debug)
	if err != nil {
		httputil.SendJSONError(w, "failed generating payload", http.StatusInternalServerError)
		return
	}
	defer os.Remove(compiledPath)

	f, err := os.Open(compiledPath)
	if err != nil {
		httputil.SendJSONError(w, "failed opening compiled binary", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	filename := filepath.Base(compiledPath)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	io.Copy(w, f)
}