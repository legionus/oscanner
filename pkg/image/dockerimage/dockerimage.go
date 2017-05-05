package dockerimage

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/docker/distribution/reference"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/image/factory"
	"github.com/openshift/oscanner/pkg/types"
)

const (
	receiverName    = "docker-image"
	dockerTarPrefix = "rootfs/"
	ownerPermRW     = 0600
)

type DockerImageReceiver struct {
	config *configuration.Configuration
	ref    reference.Reference
	outdir string
	client *docker.Client

	container string
}

func init() {
	factory.Register(receiverName, NewDockerImageReceiver)
}

func NewDockerImageReceiver(config *configuration.Configuration, target, outdir string) (types.TargetReceiver, error) {
	ref, err := reference.Parse(target)
	if err != nil {
		return nil, err
	}
	client, err := docker.NewClient(config.Docker.Addr)
	if err != nil {
		return nil, err
	}
	return &DockerImageReceiver{
		config: config,
		client: client,
		ref:    ref,
		outdir: outdir,
	}, nil
}

func (r *DockerImageReceiver) Fetch() error {
	log.Debugf("fetching docker image %s", r.ref.String())

	imagePullOption := docker.PullImageOptions{
		Repository:    r.ref.String(),
		RawJSONStream: true,
	}

	// TODO add docker config support
	auth := docker.AuthConfiguration{}

	// TODO optional progress

	if err := r.client.PullImage(imagePullOption, auth); err != nil {
		return err
	}

	return nil
}

func (r *DockerImageReceiver) Snapshot() error {
	log.Debugf("creating snapshot of docker image %s", r.ref.String())

	container, err := r.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: r.ref.String(),
			// For security purpose we don't define any entrypoint and command
			Entrypoint: []string{""},
			Cmd:        []string{""},
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create docker container: %v", err)
	}
	r.container = container.ID

	// delete the container when we are done extracting it
	defer func() {
		r.client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
		r.container = ""
	}()

	reader, writer := io.Pipe()
	// handle closing the reader/writer in the method that creates them
	defer writer.Close()
	defer reader.Close()

	// start the copy function first which will block after the first write while waiting for
	// the reader to read.
	errorChannel := make(chan error)
	go func() {
		errorChannel <- r.client.DownloadFromContainer(
			container.ID,
			docker.DownloadFromContainerOptions{
				OutputStream: writer,
				Path:         "/",
			})
	}()

	// block on handling the reads here so we ensure both the write and the reader are finished
	// (read waits until an EOF or error occurs).
	r.readTarStream(reader)

	// capture any error from the copy, ensures both the handleTarStream and DownloadFromContainer
	// are done.
	err = <-errorChannel
	if err != nil {
		return fmt.Errorf("unable to create %s snapshot: %v", r.ref.String(), err)
	}

	return nil
}

func (r *DockerImageReceiver) readTarStream(reader io.ReadCloser) error {
	tr := tar.NewReader(reader)
	if tr == nil {
		return fmt.Errorf("unable to create image tar reader")
	}

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("Unable to extract container: %v\n", err)
		}

		hdrInfo := hdr.FileInfo()

		dstpath := path.Join(r.outdir, strings.TrimPrefix(hdr.Name, dockerTarPrefix))
		// Overriding permissions to allow writing content
		mode := hdrInfo.Mode() | ownerPermRW

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(dstpath, mode); err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("Unable to create directory: %v", err)
				}
				err = os.Chmod(dstpath, mode)
				if err != nil {
					return fmt.Errorf("Unable to update directory mode: %v", err)
				}
			}
		case tar.TypeReg, tar.TypeRegA:
			file, err := os.OpenFile(dstpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
			if err != nil {
				return fmt.Errorf("Unable to create file: %v", err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return fmt.Errorf("Unable to write into file: %v", err)
			}
			file.Close()
		case tar.TypeSymlink:
			if err := os.Symlink(hdr.Linkname, dstpath); err != nil {
				return fmt.Errorf("Unable to create symlink: %v\n", err)
			}
		case tar.TypeLink:
			target := path.Join(r.outdir, strings.TrimPrefix(hdr.Linkname, dockerTarPrefix))
			if err := os.Link(target, dstpath); err != nil {
				return fmt.Errorf("Unable to create link: %v\n", err)
			}
		default:
			// For now we're skipping anything else. Special device files and
			// symlinks are not needed or anyway probably incorrect.
		}

		// maintaining access and modification time in best effort fashion
		os.Chtimes(dstpath, hdr.AccessTime, hdr.ModTime)
	}
}

func (r *DockerImageReceiver) Cleanup() error {
	if len(r.container) > 0 {
		r.client.RemoveContainer(docker.RemoveContainerOptions{
			ID: r.container,
		})
		r.container = ""
	}
	os.RemoveAll(r.outdir)
	return nil
}
