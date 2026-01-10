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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type functionRegistry interface {
	Register(ctx context.Context, entry ds.Entry) (id string, err error)
	Get(ctx context.Context, id string) (ds.Function, error)
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

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	runtime := r.FormValue("runtime") // "docker" | "wasm"
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

	userID := uuid.New()
	entry := ds.Entry{
		Name:    name,
		Runtime: runtime,
		Files:   files,
		UserId:  userID.String(), // потом из auth
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
	id := mux.Vars(r)["id"]

	fn, err := g.registry.Get(ctx, id)
	if err != nil {
		http.Error(w, "function not found", http.StatusNotFound)
		return
	}

	input, _ := io.ReadAll(r.Body)

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
