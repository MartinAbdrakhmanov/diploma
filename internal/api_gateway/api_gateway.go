package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/gorilla/mux"
)

type functionRegistry interface {
	Register(ctx context.Context, entry ds.Entry) (id string, err error)
	Get(ctx context.Context, userID, id string) (ds.Function, error)
	Delete(ctx context.Context, userID, id string) error
	FunctionStats(ctx context.Context, userID, id string) (ds.FunctionStats, error)
	List(ctx context.Context, userID string) ([]ds.Function, error)
}

type invoker interface {
	Invoke(ctx context.Context, fn ds.Function, input []byte, timeout time.Duration) (stdout []byte, stderr []byte, err error)
}

type Gateway struct {
	registry functionRegistry
	invoker  invoker
	baseURL  string
}

func New(
	registry functionRegistry,
	invoker invoker,
	baseURL string,
) *Gateway {
	return &Gateway{
		registry: registry,
		invoker:  invoker,
		baseURL:  strings.TrimRight(baseURL, "/"),
	}
}

// POST /functions
// multipart form:
// - name: function name (string)
// - files: repeated file fields (each file's filename is used as relative path)
func (g *Gateway) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Header.Get("X-User-ID") //TODO change me to proper auth
	if userID == "" {
		http.Error(w, "missing user", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.ToLower(r.FormValue("name"))
	runtime := strings.ToLower(r.FormValue("runtime")) // "docker" | "wasm"
	if runtime == "" {
		http.Error(w, "runtime is required", http.StatusBadRequest)
		return
	}

	files := map[string][]byte{}
	for _, fheaders := range r.MultipartForm.File {
		for _, fh := range fheaders {
			f, _ := fh.Open()
			b, _ := io.ReadAll(f)
			_ = f.Close()
			files[fh.Filename] = b
		}
	}

	entry := ds.Entry{
		Name:    name,
		Runtime: runtime,
		Files:   files,
		UserId:  userID,
	}

	id, err := g.registry.Register(ctx, entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"id":         id,
		"invoke_url": fmt.Sprintf("%s/functions/%s/invoke", g.baseURL, id),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /functions/{id}/invoke
// Body: SDKRequest JSON
func (g *Gateway) HandleInvoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Header.Get("X-User-ID") // change me
	if userID == "" {
		http.Error(w, "missing user", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]

	fn, err := g.registry.Get(ctx, userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if fn.ID == "" {
		http.Error(w, "function not found", http.StatusNotFound)
		return
	}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stdout, stderr, err := g.invoker.Invoke(
		ctx,
		fn,
		input,
		time.Duration(fn.Timeout)*time.Second,
	)

	if err != nil {
		http.Error(w, err.Error()+"\n"+string(stderr), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(stdout)
}

// POST /functions/{id}/delete
// Body: function id string (UUID)
func (g *Gateway) HandleDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Header.Get("X-User-ID") // change me
	if userID == "" {
		http.Error(w, "missing user", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]

	err := g.registry.Delete(ctx, userID, id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /functions/{id}/stats
func (g *Gateway) HandleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "missing user", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]

	stats, err := g.registry.FunctionStats(ctx, userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GET /functions
func (g *Gateway) HandleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "missing user", http.StatusUnauthorized)
		return
	}

	functions, err := g.registry.List(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(functions)
}
