package versioner

import (
	"os"
	"path/filepath"
	"strings"
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

func TestIncrementVersionMajor(t *testing.T) {
	updated := withFixtureCopy(t, "testdata/Chart.yaml", func(t *testing.T, fname string) {
		err := IncrementVersion(fname, ".version", VersionOptions{Major: true})
		if err != nil {
			t.Fatal(err)
		}
	})

	want := []byte(`apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 2.0.0
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
			wantErr: `failed to find key "test" in testdata/Chart.yaml`,
		},
		{
			name:    "non-semver key",
			fname:   "testdata/Chart.yaml",
			key:     "description",
			wantErr: `parsing string "A Helm chart for Kubernetes" as SemVer: invalid semantic version`,
		},
		{
			name:    "no version component specified",
			fname:   "testdata/Chart.yaml",
			key:     "version",
			options: VersionOptions{},
			wantErr: `at least one of --major, --minor, or --patch must be specified`,
		},
		{
			name:    "multiple version components specified",
			fname:   "testdata/Chart.yaml",
			key:     "version",
			options: VersionOptions{Patch: true, Minor: true},
			wantErr: `only one of --major, --minor, or --patch can be specified`,
		},
		{
			name:    "all version components specified",
			fname:   "testdata/Chart.yaml",
			key:     "version",
			options: VersionOptions{Patch: true, Minor: true, Major: true},
			wantErr: `only one of --major, --minor, or --patch can be specified`,
		},
	}

	for _, tt := range versionTests {
		t.Run(tt.name, func(t *testing.T) {
			withFixtureCopy(t, tt.fname, func(t *testing.T, fname string) {
				err := IncrementVersion(fname, tt.key, tt.options)

				if err == nil {
					t.Fatal("expected error but got nil")
				}

				// For the missing key error, we check if the error contains the expected substring
				// since the full path will vary in tests
				if tt.name == "missing key" {
					wantSubstr := `failed to find key "test"`
					if errMsg := err.Error(); !strings.Contains(errMsg, wantSubstr) {
						t.Errorf("expected error containing %q, got %q", wantSubstr, errMsg)
					}
				} else if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
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

func TestIncrementVersionPreservesPermissions(t *testing.T) {
	testCases := []struct {
		name        string
		permissions os.FileMode
	}{
		{
			name:        "read-only permissions",
			permissions: 0444,
		},
		{
			name:        "read-write owner only",
			permissions: 0600,
		},
		{
			name:        "read-write owner, read-only group and others",
			permissions: 0644,
		},
		{
			name:        "full permissions owner, read-execute group and others",
			permissions: 0755,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "Chart.yaml")

			// Create test file with specific permissions
			b, err := os.ReadFile("testdata/Chart.yaml")
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(testFile, b, tc.permissions); err != nil {
				t.Fatal(err)
			}

			// Verify initial permissions
			infoBefore, err := os.Stat(testFile)
			if err != nil {
				t.Fatal(err)
			}

			if infoBefore.Mode() != tc.permissions {
				t.Fatalf("initial permissions mismatch: got %v, want %v", infoBefore.Mode(), tc.permissions)
			}

			// Increment version
			err = IncrementVersion(testFile, ".version", VersionOptions{Patch: true})
			if err != nil {
				t.Fatal(err)
			}

			// Verify permissions are preserved
			infoAfter, err := os.Stat(testFile)
			if err != nil {
				t.Fatal(err)
			}

			if infoAfter.Mode() != tc.permissions {
				t.Errorf("permissions not preserved: got %v, want %v", infoAfter.Mode(), tc.permissions)
			}

			// Verify the content was actually updated
			updated, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(updated), "version: 1.0.2") {
				t.Errorf("version was not updated correctly: %s", string(updated))
			}
		})
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
