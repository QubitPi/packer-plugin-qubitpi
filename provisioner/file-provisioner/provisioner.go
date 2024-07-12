// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

// This package implements a provisioner for Packer that uploads a local file onto the remote machine
package file

import (
	"fmt"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"os"
	"path/filepath"
	"strings"
)

func Provision(ctx interpolate.Context, ui packersdk.Ui, communicator packersdk.Communicator, source string, destination string) error {
	src, err := interpolate.Render(source, &ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	dst, err := interpolate.Render(destination, &ctx)
	if err != nil {
		return fmt.Errorf("error interpolating destination: %s", err)
	}

	ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err = communicator.UploadDir(dst, src, nil); err != nil {
			ui.Error(fmt.Sprintf("Upload failed: %s", err))
			return err
		}
		return nil
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	filedst := dst
	if strings.HasSuffix(dst, "/") {
		filedst = dst + filepath.Base(src)
	}

	pf := ui.TrackProgress(filepath.Base(src), 0, info.Size(), f)
	defer pf.Close()

	// Upload the file
	if err = communicator.Upload(filedst, pf, &fi); err != nil {
		if strings.Contains(err.Error(), "Error restoring file") {
			ui.Error(fmt.Sprintf("Upload failed: %s; this can occur when "+
				"your file destination is a folder without a trailing "+
				"slash.", err))
		}
		ui.Error(fmt.Sprintf("Upload failed: %s", err))
		return err
	}

	return nil
}
