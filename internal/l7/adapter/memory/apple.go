package memory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

type AppleEmbedder struct {
	resolve func(string) (processadapter.Executable, error)
	run     func(context.Context, processadapter.Request) (processadapter.Result, error)
}

func NewAppleEmbedder() AppleEmbedder {
	return AppleEmbedder{resolve: processadapter.Resolve, run: (processadapter.Runner{}).Run}
}
func NewAppleEmbedderWith(resolve func(string) (processadapter.Executable, error), run func(context.Context, processadapter.Request) (processadapter.Result, error)) AppleEmbedder {
	return AppleEmbedder{resolve: resolve, run: run}
}

func (embedder AppleEmbedder) Embed(ctx context.Context, texts []string) (int, int, [][]float32, error) {
	if embedder.resolve == nil || embedder.run == nil || len(texts) == 0 || len(texts) > 100_000 {
		return 0, 0, nil, errors.New("Apple embedding request is invalid")
	}
	request := struct {
		Schema int      `json:"schema"`
		Texts  []string `json:"texts"`
	}{Schema: 1, Texts: texts}
	data, err := json.Marshal(request)
	if err != nil || len(data) > processadapter.MaxInputBytes {
		return 0, 0, nil, errors.New("Apple embedding request is unbounded")
	}
	executable, err := embedder.resolve("l7-embed")
	if err != nil {
		return 0, 0, nil, errors.New("Apple embedding helper is unavailable")
	}
	result, err := embedder.run(ctx, processadapter.Request{Executable: executable.Path, Arguments: []string{}, Input: data, Directory: "/", Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 20, Timeout: 5 * time.Minute})
	if err != nil || result.ExitCode != 0 {
		return 0, 0, nil, errors.New("Apple embedding helper failed")
	}
	var response struct {
		Schema    int         `json:"schema"`
		Revision  int         `json:"revision"`
		Dimension int         `json:"dimension"`
		Vectors   [][]float32 `json:"vectors"`
	}
	if json.Unmarshal(result.Stdout, &response) != nil || response.Schema != 1 || response.Revision < 1 || response.Dimension < 1 || len(response.Vectors) != len(texts) {
		return 0, 0, nil, errors.New("Apple embedding response is invalid")
	}
	for _, vector := range response.Vectors {
		if len(vector) != response.Dimension {
			return 0, 0, nil, errors.New("Apple embedding vector dimension is invalid")
		}
	}
	return response.Revision, response.Dimension, response.Vectors, nil
}
