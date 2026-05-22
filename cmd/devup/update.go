package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/RajarshiParmar/devup/internal/ui"
	"github.com/RajarshiParmar/devup/internal/updater"
)

func newUpdateCmd(f *globalFlags) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update devup to the latest version",
		Long:  `Checks GitHub Releases for a newer version and downloads the appropriate binary.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUpdate(cmd.Context(), f, yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

func runUpdate(ctx context.Context, f *globalFlags, yes bool) error {
	u := ui.New(f.noColor)
	u.Header("devup update")

	up := updater.New("RajarshiParmar", "devup", version)

	// Check for latest release.
	var release *updater.Release
	_, err := u.SpinStep("Checking for updates", func() error {
		var checkErr error
		release, checkErr = up.CheckLatest(ctx)
		return checkErr
	})
	if err != nil {
		return err
	}

	if !up.IsNewer(release) {
		fmt.Println()
		u.Success("Already up to date (%s)", version)
		return nil
	}

	// Show what's available.
	fmt.Println()
	u.Step("Current version: %s", version)
	u.Step("Latest version:  %s", release.TagName)

	// Show changelog if available.
	if release.Body != "" {
		fmt.Println()
		u.Header("Changelog")
		// Trim and limit output.
		body := strings.TrimSpace(release.Body)
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		fmt.Println(body)
	}
	fmt.Println()

	// Confirm update.
	if !yes {
		var confirm bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Update to %s?", release.TagName)).
			Value(&confirm).
			Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		if !confirm {
			u.Warn("Update cancelled")
			return nil
		}
	}

	// Find the correct asset for this platform.
	asset, err := updater.FindAsset(release)
	if err != nil {
		u.Error("%v", err)
		return err
	}

	// Download and replace.
	_, err = u.SpinStep(fmt.Sprintf("Downloading %s", asset.Name), func() error {
		body, dlErr := up.Download(ctx, asset)
		if dlErr != nil {
			return dlErr
		}
		defer body.Close()

		// Extract binary from tarball.
		binary, extractErr := extractBinaryFromTarGz(body, "devup")
		if extractErr != nil {
			return extractErr
		}
		defer binary.Close()

		return updater.ReplaceBinary(binary)
	})
	if err != nil {
		return err
	}

	fmt.Println()
	u.Success("Updated to %s", release.TagName)
	return nil
}

// extractBinaryFromTarGz reads a .tar.gz stream and returns a reader for the
// named file within it.
func extractBinaryFromTarGz(r io.Reader, name string) (io.ReadCloser, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("decompressing archive: %w", err)
	}

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			gz.Close()
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		// Match the binary name (may be prefixed with directory).
		if strings.HasSuffix(hdr.Name, "/"+name) || hdr.Name == name {
			// Return a composite closer that closes both the tar reader stream
			// and the gzip reader when done.
			return &gzTarReader{Reader: tr, gz: gz}, nil
		}
	}

	gz.Close()
	return nil, fmt.Errorf("binary %q not found in archive", name)
}

// gzTarReader wraps a tar.Reader to also close the gzip reader.
type gzTarReader struct {
	io.Reader
	gz io.Closer
}

func (r *gzTarReader) Close() error {
	if err := r.gz.Close(); err != nil {
		return fmt.Errorf("closing gzip reader: %w", err)
	}
	return nil
}
