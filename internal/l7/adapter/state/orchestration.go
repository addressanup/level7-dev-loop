package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const maxOrchestrationState = 32 << 20

type providerSnapshotFile struct {
	Schema    int                       `json:"schema"`
	Providers []domain.ProviderSnapshot `json:"providers"`
}

func SaveProviderSnapshots(commonDirectory string, snapshots []domain.ProviderSnapshot) error {
	if len(snapshots) > 32 {
		return errors.New("provider snapshot count exceeds 32")
	}
	for _, snapshot := range snapshots {
		if err := validateProviderSnapshot(snapshot); err != nil {
			return err
		}
	}
	return saveOrchestration(commonDirectory, "providers/snapshot.json", providerSnapshotFile{Schema: domain.OrchestrationSchema, Providers: append([]domain.ProviderSnapshot{}, snapshots...)})
}

func LoadProviderSnapshots(commonDirectory string) ([]domain.ProviderSnapshot, bool, error) {
	var file providerSnapshotFile
	found, err := loadOrchestration(commonDirectory, "providers/snapshot.json", &file)
	if err != nil || !found {
		return nil, found, err
	}
	if file.Schema != domain.OrchestrationSchema || len(file.Providers) > 32 {
		return nil, false, errors.New("provider snapshot schema or count is invalid")
	}
	for _, snapshot := range file.Providers {
		if err := validateProviderSnapshot(snapshot); err != nil {
			return nil, false, err
		}
	}
	return append([]domain.ProviderSnapshot{}, file.Providers...), true, nil
}

func SaveRouteDecision(commonDirectory string, decision domain.RouteDecision) error {
	if err := validateRouteDecision(decision); err != nil {
		return err
	}
	return saveOrchestration(commonDirectory, "routes/latest.json", decision)
}

func LoadRouteDecision(commonDirectory string) (domain.RouteDecision, bool, error) {
	var decision domain.RouteDecision
	found, err := loadOrchestration(commonDirectory, "routes/latest.json", &decision)
	if err != nil || !found {
		return decision, found, err
	}
	if err := validateRouteDecision(decision); err != nil {
		return domain.RouteDecision{}, false, err
	}
	return decision, true, nil
}

func saveOrchestration(commonDirectory, relative string, value any) error {
	root, err := orchestrationRoot(commonDirectory)
	if err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > maxOrchestrationState {
		return errors.New("orchestration state is invalid or unbounded")
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	} else if err != nil {
		return err
	}
	return localfile.AtomicReplace(path, data, 0o600)
}

func loadOrchestration(commonDirectory, relative string, target any) (bool, error) {
	root, err := orchestrationRoot(commonDirectory)
	if err != nil {
		return false, err
	}
	data, err := localfile.Read(filepath.Join(root, filepath.FromSlash(relative)), maxOrchestrationState)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := localfile.DecodeJSON(data, target); err != nil {
		return false, err
	}
	return true, nil
}

func orchestrationRoot(commonDirectory string) (string, error) {
	if !filepath.IsAbs(commonDirectory) {
		return "", errors.New("Git common directory must be absolute")
	}
	clean := filepath.Clean(commonDirectory)
	info, err := os.Lstat(clean)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("Git common directory is unsafe")
	}
	return filepath.Join(clean, "l7", "orchestration"), nil
}

func validateProviderSnapshot(snapshot domain.ProviderSnapshot) error {
	if snapshot.Schema != domain.OrchestrationSchema || snapshot.ID == "" || len(snapshot.ID) > 64 || !snapshot.Kind.Valid() || !snapshot.Authentication.Valid() || len(snapshot.Models) > 128 || !safeStateText(snapshot.Diagnostic, 4096) || !safeStateText(snapshot.Next, 4096) {
		return fmt.Errorf("provider snapshot %q is invalid", snapshot.ID)
	}
	seen := make(map[string]bool)
	for _, model := range snapshot.Models {
		if model.ID == "" || len(model.ID) > 160 || strings.ContainsAny(model.ID, "\x00\r\n") || seen[model.ID] || model.ContextWindow < 1024 || model.ContextWindow > 16_000_000 || model.CostClass < 1 || model.CostClass > 5 || model.LatencyClass < 1 || model.LatencyClass > 5 || len(model.Languages) == 0 || len(model.Languages) > 128 || len(model.Efforts) == 0 || len(model.Efforts) > 8 {
			return fmt.Errorf("provider snapshot model %q is invalid", model.ID)
		}
		seen[model.ID] = true
		seenLanguages := make(map[string]bool)
		for _, language := range model.Languages {
			language = strings.ToLower(strings.TrimSpace(language))
			if language == "" || len(language) > 64 || strings.ContainsAny(language, "\x00\r\n") || seenLanguages[language] {
				return fmt.Errorf("provider snapshot model %q has an invalid language", model.ID)
			}
			seenLanguages[language] = true
		}
		seenEfforts := make(map[domain.ReasoningEffort]bool)
		for _, effort := range model.Efforts {
			if !effort.Valid() || seenEfforts[effort] {
				return fmt.Errorf("provider snapshot model %q has invalid effort", model.ID)
			}
			seenEfforts[effort] = true
		}
	}
	return nil
}

func validateRouteDecision(decision domain.RouteDecision) error {
	if decision.Schema != domain.OrchestrationSchema || decision.TaskID == "" || len(decision.TaskID) > 128 || decision.Policy != "balanced" || len(decision.Candidates) > 4096 || len(decision.Fallbacks) > 4096 || len(decision.Escalations) > 128 || !safeStateText(decision.Next, 4096) {
		return errors.New("route decision is invalid")
	}
	if decision.ProviderID != "" && (decision.ModelID == "" || !decision.Effort.Valid()) {
		return errors.New("selected route identity is incomplete")
	}
	return nil
}

func safeStateText(value string, maximum int) bool {
	return len(value) <= maximum && !strings.ContainsAny(value, "\x00\r")
}
