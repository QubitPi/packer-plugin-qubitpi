package provisioner

import (
	"fmt"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"os"
	"path/filepath"
	"strings"
)

func ProvisionUpload(ui packersdk.Ui, communicator packersdk.Communicator, dst string, src string) error {
	ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("source should be a file; '%s', however, is a directory", src)
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
