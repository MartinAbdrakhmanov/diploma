package ds

type BuildRequest struct {
	Name    string
	Runtime string // go | wasm
	Source  []byte // zip / tar / raw
}

type BuildResult struct {
	Image string // mini-faas/hello:abc123
}
