package openevec

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/einfo"
	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/eden"
	"github.com/lf-edge/eden/pkg/eve"
	"github.com/lf-edge/eden/pkg/models"
	"github.com/lf-edge/eden/pkg/tests"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/info"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func DownloadEve(cfg *EdenSetupArgs) error {
	model, err := models.GetDevModelByName(cfg.Eve.DevModel)
	if err != nil {
		return fmt.Errorf("GetDevModelByName: %w", err)
	}
	format := model.DiskFormat()
	eveDesc := utils.EVEDescription{
		ConfigPath:  cfg.Adam.Dist,
		Arch:        cfg.Eve.Arch,
		HV:          cfg.Eve.HV,
		Registry:    cfg.Eve.Registry,
		Tag:         cfg.Eve.Tag,
		Format:      format,
		ImageSizeMB: cfg.Eve.ImageSizeMB,
	}
	if err := utils.DownloadEveLive(eveDesc, cfg.Eve.ImageFile); err != nil {
		return err
	}
	if err := utils.DownloadUEFI(eveDesc, filepath.Dir(cfg.Eve.ImageFile)); err != nil {
		return err
	}
	log.Infof(model.DiskReadyMessage(), cfg.Eve.ImageFile)
	fmt.Println(cfg.Eve.ImageFile)
	return nil
}

func OciImage(fileToSave, image, registry string, isLocal bool) error {
	var imageManifest []byte
	var err error
	ref, err := name.ParseReference(image)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", image, err)
	}
	var img v1.Image
	if !isLocal {
		desc, err := remote.Get(ref)
		if err != nil {
			return err
		}
		img, err = desc.Image()
		if err != nil {
			return err
		}
	} else {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return err
		}
		cli.NegotiateAPIVersion(ctx)
		options := daemon.WithClient(cli)
		img, err = daemon.Image(ref, options)
		if err != nil {
			return err
		}
	}
	imageManifest, err = img.RawManifest()
	if err != nil {
		return err
	}
	if err := tarball.WriteToFile(fileToSave, ref, img); err != nil {
		return err
	}

	// add the imageManifest to the tar file
	if err := appendImageManifest(fileToSave, imageManifest); err != nil {
		return fmt.Errorf("unable to append image manifest to tar at %s: %v", fileToSave, err)
	}
	if err := appendImageRepositories(fileToSave, registry, image, imageManifest); err != nil {
		return fmt.Errorf("unable to append image manifest to tar at %s: %v", fileToSave, err)
	}

	return nil
}

// appendImageManifest add the given manifest to the given tar file. Opinionated
// about the name of the file to "imagemanifest-<hash>.json"
func appendImageManifest(tarFile string, manifest []byte) error {
	hash := sha256.Sum256(manifest)
	return appendToTarFile(tarFile, fmt.Sprintf("%s-%x.json", "imagemanifest", hash), manifest)
}

// appendToTarFile add the given bytes to the tar file with the given filename
func appendToTarFile(tarFile, filename string, content []byte) error {
	var (
		f   *os.File
		err error
	)
	// open the existing file
	if f, err = os.OpenFile(tarFile, os.O_RDWR, os.ModePerm); err != nil {
		return err
	}
	defer f.Close()
	// there always is padding at the end of a tar archive, so skip to the end
	// of the actual archive, so it will be read
	if _, err = f.Seek(-2<<9, io.SeekEnd); err != nil {
		return err
	}

	tw := tar.NewWriter(f)

	hdr := &tar.Header{
		Name: filename,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write %s tar header: %v", filename, err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write %s tar body: %v", filename, err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	return nil
}
func appendImageRepositories(tarFile, registry, path string, imageManifest []byte) error {
	// get the top layer for the manifest bytes
	layerHash, err := DockerHashFromManifest(imageManifest)
	if err != nil {
		return fmt.Errorf("unable to get top layer hash: %w", err)
	}
	// need to take out the tag
	parts := strings.Split(path, ":")
	var tag, repo string
	switch len(parts) {
	case 0:
		return fmt.Errorf("malformed repository path %s", path)
	case 1:
		repo = parts[0]
		tag = "latest"
	case 2:
		repo = parts[0]
		tag = parts[1]
	default:
		return fmt.Errorf("malformed repository path has too many ':' %s", path)
	}
	fullRepo := fmt.Sprintf("%s/%s", registry, repo)
	// now build the tag we are after
	var data = map[string]map[string]string{}
	data[fullRepo] = map[string]string{}
	data[fullRepo][tag] = layerHash

	j, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to convert repositories data to json: %w", err)
	}

	return appendToTarFile(tarFile, "repositories", j)
}

// LayersFromManifest get the descriptors for layers from a raw image manifest
func LayersFromManifest(imageManifest []byte) ([]v1.Descriptor, error) {
	manifest, err := v1.ParseManifest(bytes.NewReader(imageManifest))
	if err != nil {
		return nil, fmt.Errorf("unable to parse manifest: %w", err)
	}
	return manifest.Layers, nil
}

// DockerHashFromManifest get the sha256 hash as a string from a raw image
// manifest. The "docker hash" is what is used for the image, i.e. the topmost
// layer.
func DockerHashFromManifest(imageManifest []byte) (string, error) {
	layers, err := LayersFromManifest(imageManifest)
	if err != nil {
		return "", fmt.Errorf("unable to get layers: %w", err)
	}
	if len(layers) < 1 {
		return "", fmt.Errorf("no layers found")
	}
	return layers[len(layers)-1].Digest.Hex, nil
}

func SDInfoEve(devicePath, syslogOutput, eveReleaseOutput string) error {
	eveInfo, err := eden.GetInfoFromSDCard(devicePath)
	if err != nil {
		log.Info("Check is EVE on SD and your access to read SD")
		return fmt.Errorf("problem with access to EVE partitions: %w", err)
	}
	if eveInfo.EVERelease == nil {
		log.Warning("No eve-release found. Probably, no EVE on SD card")
	} else {
		if err = ioutil.WriteFile(eveReleaseOutput, eveInfo.EVERelease, 0666); err != nil {
			return err
		}
		log.Infof("Your eve-release in %s", eveReleaseOutput)
	}
	if eveInfo.Syslog == nil {
		log.Warning("No syslog found, EVE may not have started yet")
	} else {
		if err = ioutil.WriteFile(syslogOutput, eveInfo.Syslog, 0666); err != nil {
			return err
		}
		log.Infof("Your syslog in %s", syslogOutput)
	}
	return nil
}

func UploadGit(absPath, object, branch, directoryToSave string) error {
	commandToRun := fmt.Sprintf("-i /in/%s -o %s -b %s -d %s git",
		filepath.Base(absPath), object, branch, directoryToSave)
	image := fmt.Sprintf("%s:%s", defaults.DefaultProcContainerRef, defaults.DefaultProcTag)
	volumeMap := map[string]string{"/in": filepath.Dir(absPath)}
	result, err := utils.RunDockerCommand(image, commandToRun, volumeMap)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

type TestArgs struct {
	TestArgs     string
	TestOpts     bool
	TestEscript  string
	TestRun      string
	TestTimeout  string
	TestList     string
	TestProg     string
	TestScenario string
	FailScenario string
	CurDir       string
	ConfigFile   string
	Verbosity    string
}

func Test(tstCfg *TestArgs) error {

	switch {
	case tstCfg.TestList != "":
		tests.RunTest(tstCfg.TestProg, []string{"-test.list", tstCfg.TestList}, "", tstCfg.TestTimeout, tstCfg.FailScenario, tstCfg.ConfigFile, tstCfg.Verbosity)
	case tstCfg.TestOpts:
		tests.RunTest(tstCfg.TestProg, []string{"-h"}, "", tstCfg.TestTimeout, tstCfg.FailScenario, tstCfg.ConfigFile, tstCfg.Verbosity)
	case tstCfg.TestEscript != "":
		tests.RunTest("eden.escript.test", []string{"-test.run", "TestEdenScripts/" + tstCfg.TestEscript}, tstCfg.TestArgs, tstCfg.TestTimeout, tstCfg.FailScenario, tstCfg.ConfigFile, tstCfg.Verbosity)
	case tstCfg.TestRun != "":
		tests.RunTest(tstCfg.TestProg, []string{"-test.run", tstCfg.TestRun}, tstCfg.TestArgs, tstCfg.TestTimeout, tstCfg.FailScenario, tstCfg.ConfigFile, tstCfg.Verbosity)
	default:
		tests.RunScenario(tstCfg.TestScenario, tstCfg.TestArgs, tstCfg.TestTimeout, tstCfg.FailScenario, tstCfg.ConfigFile, tstCfg.Verbosity)
	}

	if tstCfg.CurDir != "" {
		err := os.Chdir(tstCfg.CurDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkOutput(input string, shouldHave, shouldNotHave []string) error {
	for _, str := range shouldHave {
		if !strings.Contains(input, str) {
			return fmt.Errorf("Input does not contain %v", str)
		}
	}

	for _, str := range shouldNotHave {
		if strings.Contains(input, str) {
			return fmt.Errorf("Input contains %v", str)
		}
	}

	return nil
}

func checkAppState(ctrl controller.Cloud, devUUID uuid.UUID, appName string, eveState *eve.State, expState string, timeout time.Duration) error {
	startTime := time.Now()

	// Waiting for 15 min maximum to get eclient-mount app in state running
	handleInfo := func(im *info.ZInfoMsg) bool {
		eveState.InfoCallback()(im)
		for _, s := range eveState.Applications() {
			if s.Name == appName {
				if s.EVEState == expState {
					return true
				}
			}
		}
		if time.Now().After(startTime.Add(timeout)) {
			log.Fatal("eclient-mount timeout")
		}
		return false
	}

	if err := ctrl.InfoChecker(devUUID, nil, handleInfo, einfo.InfoNew, 0); err != nil {
		return fmt.Errorf("eclient-mount RUNNING state InfoChecker: %w", err)
	}

	return nil
}

func withCapturingStdout(f func() error) ([]byte, error) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return out, err
}