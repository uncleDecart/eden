package openevec

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/utils"
	uuid "github.com/satori/go.uuid"
)

func GetDefaultConfig(currentPath string) *EdenSetupArgs {

	ip, err := utils.GetIPForDockerAccess()
	if err != nil {
		return nil
	}

	edenDir, err := utils.DefaultEdenDir()
	if err != nil {
		return nil
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil
	}

	imageDist := fmt.Sprintf("%s-%s", defaults.DefaultContext, defaults.DefaultImageDist)
	certsDist := fmt.Sprintf("%s-%s", defaults.DefaultContext, defaults.DefaultCertsDist)

	firmware := []string{filepath.Join(imageDist, "eve", "OVMF.fd")}
	if runtime.GOARCH == "amd64" {
		firmware = []string{
			filepath.Join(imageDist, "eve", "OVMF_CODE.fd"),
			filepath.Join(imageDist, "eve", "OVMF_VARS.fd")}
	}

	defaultEdenConfig := &EdenSetupArgs{
		Eden: EdenConfig{
			Root:         filepath.Join(currentPath, defaults.DefaultDist),
			Tests:        filepath.Join(currentPath, defaults.DefaultDist, "tests"),
			Download:     true,
			BinDir:       filepath.Join(defaults.DefaultDist, defaults.DefaultBinDist),
			SSHKey:       fmt.Sprintf("%s-%s", defaults.DefaultContext, defaults.DefaultSSHKey),
			CertsDir:     filepath.Join(fmt.Sprintf("%s-%s", defaults.DefaultContext, defaults.DefaultCertsDist)),
			TestBin:      defaults.DefaultBinDist,
			EdenBin:      "eden.escript.test",
			TestScenario: defaults.DefaultTestScenario,

			Images: ImagesConfig{
				EServerImageDist: defaults.DefaultEserverDist,
			},

			EServer: EServerConfig{
				IP:    ip,
				EVEIP: defaults.DefaultDomain,

				Port:  defaults.DefaultEserverPort,
				Force: true,
				Tag:   defaults.DefaultEServerTag,
			},

			EClient: EClientConfig{
				Tag:   defaults.DefaultEClientTag,
				Image: defaults.DefaultEClientContainerRef,
			},
		},

		Adam: AdamConfig{
			Tag:         defaults.DefaultAdamTag,
			Port:        defaults.DefaultAdamPort,
			Dist:        defaults.DefaultAdamDist,
			CertsDomain: defaults.DefaultDomain,
			CertsIP:     ip,
			CertsEVEIP:  ip,
			Force:       true,
			CA:          filepath.Join(certsDist, "root-certificate.pem"),
			APIv1:       false,

			Redis: RedisConfig{
				RemoteURL: fmt.Sprintf("%s:%d", defaults.DefaultRedisContainerName, defaults.DefaultRedisPort),
				Tag:       defaults.DefaultRedisTag,
				Port:      defaults.DefaultRedisPort,
				Eden:      fmt.Sprintf("%s:%d", ip, defaults.DefaultRedisPort),
			},

			Remote: RemoteConfig{
				Enabled: true,
				Redis:   true,
			},

			Caching: CachingConfig{
				Enabled: false,
				Redis:   false,
				Prefix:  "cache",
			},
		},

		Eve: EveConfig{
			Name:         strings.ToLower(defaults.DefaultContext),
			DevModel:     defaults.DefaultQemuModel,
			ModelFile:    "",
			Arch:         runtime.GOARCH,
			QemuOS:       runtime.GOOS,
			Accel:        true,
			HV:           defaults.DefaultEVEHV,
			CertsUUID:    id.String(),
			Cert:         filepath.Join(certsDist, "onboard.cert.pem"),
			DeviceCert:   filepath.Join(certsDist, "device.cert.pem"),
			QemuFirmware: firmware,
			Dist:         fmt.Sprintf("%s-%s", defaults.DefaultContext, defaults.DefaultEVEDist),
			Repo:         defaults.DefaultEveRepo,
			Registry:     defaults.DefaultEveRegistry,
			Tag:          defaults.DefaultEVETag,
			UefiTag:      defaults.DefaultEVETag,
			HostFwd: map[string]string{
				strconv.Itoa(defaults.DefaultSSHPort): "22",
				"5911":                                "5901",
				"5912":                                "5902",
				"8027":                                "8027",
				"8028":                                "8028"},
			QemuFileToSave: defaults.DefaultQemuFileToSave,
			QemuCpus:       defaults.DefaultCpus,
			QemuMemory:     defaults.DefaultMemory,
			ImageSizeMB:    defaults.DefaultEVEImageSize,
			Serial:         defaults.DefaultEVESerial,
			Pid:            fmt.Sprintf("%s-eve.pid", strings.ToLower(defaults.DefaultContext)),
			Log:            fmt.Sprintf("%s-eve.log", strings.ToLower(defaults.DefaultContext)),
			TelnetPort:     defaults.DefaultTelnetPort,
			TPM:            defaults.DefaultTPMEnabled,
			ImageFile:      filepath.Join(imageDist, "eve", "live.img"),
			QemuDTBPath:    "",
			QemuConfigPath: certsDist,
			Remote:         defaults.DefaultEVERemote,
			RemoteAddr:     defaults.DefaultEVEHost,
			LogLevel:       defaults.DefaultEveLogLevel,
			AdamLogLevel:   defaults.DefaultAdamLogLevel,
			Ssid:           "",
			Disks:          defaults.DefaultAdditionalDisks,
			BootstrapFile:  "",
			UsbNetConfFile: "",
			Platform:       "none",

			CustomInstaller: CustomInstallerConfig{
				Path:   "",
				Format: "",
			},

			QemuConfig: QemuConfig{
				MonitorPort:      defaults.DefaultQemuMonitorPort,
				NetDevSocketPort: defaults.DefaultQemuNetdevSocketPort,
			},
		},

		Redis: RedisConfig{
			Tag:  defaults.DefaultRedisTag,
			Port: defaults.DefaultRedisPort,
			Dist: defaults.DefaultRedisDist,
		},

		Registry: RegistryConfig{
			Tag:  defaults.DefaultRegistryTag,
			Port: defaults.DefaultRegistryPort,
			IP:   ip,
			Dist: defaults.DefaultRegistryDist,
		},

		Sdn: SdnConfig{
			RAM:            defaults.DefaultSdnMemory,
			CPU:            defaults.DefaultSdnCpus,
			ConsoleLogFile: filepath.Join(currentPath, defaults.DefaultDist, "sdn-console.log"),
			Disable:        true,
			TelnetPort:     defaults.DefaultSdnTelnetPort,
			MgmtPort:       defaults.DefaultSdnMgmtPort,
			PidFile:        filepath.Join(currentPath, defaults.DefaultDist, "sdn.pid"),
			SSHPort:        defaults.DefaultSdnSSHPort,
			SourceDir:      filepath.Join(currentPath, "sdn"),
			ConfigDir:      filepath.Join(edenDir, fmt.Sprintf("%s-sdn", "default")),
			ImageFile:      filepath.Join(imageDist, "eden", "sdn-efi.qcow2"),
			LinuxkitBin:    filepath.Join(currentPath, defaults.DefaultBuildtoolsDir, "linuxkit"),
			NetModelFile:   "",
		},

		Gcp: GcpConfig{
			Key: "",
		},

		Packet: PacketConfig{
			Key: "",
		},

		ConfigName: defaults.DefaultContext,
		ConfigFile: utils.GetConfig(defaults.DefaultContext),
		EdenDir:    edenDir,
	}

	resolvePath(reflect.ValueOf(defaultEdenConfig).Elem(), defaultEdenConfig.Eden.Root)

	return defaultEdenConfig
}

func GetDefaultPodConfig() *PodConfig {
	dpc := &PodConfig{
		AppMemory:         humanize.Bytes(defaults.DefaultAppMem * 1024),
		DiskSize:          humanize.Bytes(0),
		VolumeType:        "qcow2",
		AppCpus:           defaults.DefaultAppCPU,
		ACLOnlyHost:       false,
		NoHyper:           false,
		Registry:          "remote",
		DirectLoad:        true,
		SftpLoad:          false,
		VolumeSize:        humanize.IBytes(defaults.DefaultVolumeSize),
		OpenStackMetadata: false,
		PinCpus:           false,
	}

	return dpc
}
