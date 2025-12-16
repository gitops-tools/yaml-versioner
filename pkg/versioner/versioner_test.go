package versioner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIncrementVersionPatch(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/Chart.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".version", VersionOptions{Patch: true})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []byte(`apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 1.0.2
appVersion: "1.0.0"
`)
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Errorf("failed to update file:\n%s", diff)
	}
}

func TestIncrementVersionMinor(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/Chart.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".version", VersionOptions{Minor: true})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []byte(`apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 1.1.0
appVersion: "1.0.0"
`)
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Errorf("failed to update file:\n%s", diff)
	}
}

func TestIncrementVersionWithNestedFields(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/example.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".cluster.version.", VersionOptions{Patch: true})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := `cluster:
  version: 1.0.3
`
	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Errorf("failed to update file:\n%s", diff)
	}
}

func TestIncrementVersionPatchWithPreRelease(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/Chart.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".version", VersionOptions{Patch: true, PreRelease: "alpha.1"})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []byte(`apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 1.0.2-alpha.1
appVersion: "1.0.0"
`)
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Errorf("failed to update file:\n%s", diff)
	}
}

func TestIncrementVersionWithComments(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/Chart_with_comments.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".version", VersionOptions{Patch: true})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []byte(`apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
# This is the chart version. This version number should be incremented each time you make changes
# to the chart and its templates, including the app version.
# Versions are expected to follow Semantic Versioning (https://semver.org/)
version: 1.0.2
appVersion: "1.0.0"
`)
	if diff := cmp.Diff(want, updated); diff != "" {
		t.Errorf("failed to update file:\n%s", diff)
	}
}

func TestIncrementVersionErrors(t *testing.T) {
	versionTests := []struct {
		name    string
		fname   string
		key     string
		options VersionOptions
		wantErr string
	}{
		{
			name:    "missing key",
			fname:   "testdata/Chart.yaml",
			key:     "test",
			wantErr: `failed to find key "test"`,
		},
		{
			name:    "non-semver key",
			fname:   "testdata/Chart.yaml",
			key:     "description",
			wantErr: `parsing string "A Helm chart for Kubernetes" as SemVer: invalid semantic version`,
		},
	}

	for _, tt := range versionTests {
		t.Run(tt.name, func(t *testing.T) {
			withFixtureCopy(t, tt.fname, func(t *testing.T, fname string) {
				err := IncrementVersion(fname, tt.key, tt.options)

				if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
					t.Errorf("incorrect error returned:\n%s", diff)
				}
			})
		})
	}
}

func TestIncrementVersionNonExistentFile(t *testing.T) {
	err := IncrementVersion("testdata/non-existent.yaml", "test.key", VersionOptions{})

	wantErr := `failed to read YAML file: open testdata/non-existent.yaml: no such file or directory`
	if diff := cmp.Diff(wantErr, err.Error()); diff != "" {
		t.Errorf("incorrect error returned:\n%s", diff)
	}
}

func withFixtureCopy(t *testing.T, src string, f func(t *testing.T, fname string)) []byte {
	tempDir := t.TempDir()
	fixtureName := filepath.Join(tempDir, filepath.Base(src))

	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(fixtureName, b, 0660); err != nil {
		t.Fatal(err)
	}

	f(t, fixtureName)

	b, err = os.ReadFile(fixtureName)
	if err != nil {
		t.Fatal(err)
	}

	return b
}
