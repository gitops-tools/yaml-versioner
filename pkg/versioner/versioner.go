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

	// Update the Major version i.e. x.0.0
	Major bool

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

	// Validate that only one version component is being incremented
	setCount := 0
	if options.Patch {
		setCount++
	}
	if options.Minor {
		setCount++
	}
	if options.Major {
		setCount++
	}

	if setCount == 0 {
		return fmt.Errorf("at least one of --major, --minor, or --patch must be specified")
	}
	if setCount > 1 {
		return fmt.Errorf("only one of --major, --minor, or --patch can be specified")
	}

	if options.Patch {
		preRelease := v.Prerelease()
		if options.PreRelease != "" {
			preRelease = options.PreRelease
		}
		v = semver.New(v.Major(), v.Minor(), v.Patch()+1, preRelease, v.Metadata())
	}

	if options.Minor {
		preRelease := v.Prerelease()
		if options.PreRelease != "" {
			preRelease = options.PreRelease
		}
		v = semver.New(v.Major(), v.Minor()+1, 0, preRelease, v.Metadata())
	}

	if options.Major {
		preRelease := v.Prerelease()
		if options.PreRelease != "" {
			preRelease = options.PreRelease
		}
		v = semver.New(v.Major()+1, 0, 0, preRelease, v.Metadata())
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
	if err != nil || value == nil {
		return "", fmt.Errorf("failed to find key %q in %s", key, filename)
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

	// Preserve original file permissions
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("reading file info for %s: %w", filename, err)
	}
	originalMode := fileInfo.Mode()

	// Temporarily make the file writable if it's read-only
	if originalMode&0200 == 0 {
		if err := os.Chmod(filename, originalMode|0200); err != nil {
			return fmt.Errorf("temporarily setting write permission on %s: %w", filename, err)
		}
	}

	if err := os.WriteFile(filename, []byte(updated), originalMode); err != nil {
		return fmt.Errorf("writing updated file to %s: %w", filename, err)
	}

	// Restore original permissions if we modified them
	if originalMode&0200 == 0 {
		if err := os.Chmod(filename, originalMode); err != nil {
			return fmt.Errorf("restoring original permissions on %s: %w", filename, err)
		}
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
