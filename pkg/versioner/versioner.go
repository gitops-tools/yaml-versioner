package versioner

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// VersionOptions configures how the version is updated.
type VersionOptions struct {
	// Update the Patch version i.e. 0.0.x
	Patch bool

	// Update the Minor version i.e. 0.x.0
	Minor bool

	// Added to the Patch version
	PreRelease string
}

// IncrementVersion updates a key within a YAML file to increment a version.
func IncrementVersion(filename, key string, options VersionOptions) error {
	s, err := readCurrentVersion(filename, key)
	if err != nil {
		return err
	}

	v, err := semver.NewVersion(s)
	if err != nil {
		return fmt.Errorf("parsing string %q as SemVer: %s", s, err)
	}

	if options.Patch {
		preRelease := v.Prerelease()
		if options.PreRelease != "" {
			preRelease = options.PreRelease
		}
		v = semver.New(v.Major(), v.Minor(), v.Patch()+1, preRelease, v.Metadata())
	}

	// TODO: Error when Minor and Patch are set.

	if options.Minor {
		v = semver.New(v.Major(), v.Minor()+1, 0, v.Prerelease(), v.Metadata())
	}

	return setNewVersion(filename, key, v)
}

func readCurrentVersion(filename, key string) (string, error) {
	rn, err := kyaml.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read YAML file: %w", err)
	}

	keyElements := keyToSlice(key)
	value, err := rn.Pipe(kyaml.Lookup(keyElements...))
	if err != nil {
		return "", fmt.Errorf("failed to find key %q", key)
	}

	if value == nil {
		return "", fmt.Errorf("failed to find key %q", key)
	}

	s, err := value.String()
	if err != nil {
		return "", fmt.Errorf("converting key to string: %s", err)
	}

	return strings.TrimSpace(s), nil
}

func setNewVersion(filename, key string, newVersion *semver.Version) error {
	rn, err := kyaml.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	keyElements := keyToSlice(key)
	var lookupKey []string
	var field string

	if len(keyElements) == 1 {
		field = keyElements[0]
	}
	if len(keyElements) > 1 {
		field = keyElements[len(keyElements)-1]
		lookupKey = keyElements[0 : len(keyElements)-1]
	}

	_, err = rn.Pipe(
		kyaml.Lookup(lookupKey...),
		kyaml.SetField(field, kyaml.NewStringRNode(newVersion.String())))
	if err != nil {
		return fmt.Errorf("failed to update YAML: %s", err)
	}

	updated, err := rn.String()
	if err != nil {
		return fmt.Errorf("converting updated document to string: %s", err)
	}

	// TODO Truncate and reset and overwrite to preserve permissions
	if err := os.WriteFile(filename, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing updated file to %s: %w", filename, err)
	}

	return nil
}

func keyToSlice(s string) []string {
	elements := strings.Split(s, ".")
	var result []string
	for _, v := range elements {
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}
