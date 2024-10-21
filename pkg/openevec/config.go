package openevec

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type EServerConfig struct {
	Port   int          `mapstructure:"port" cobraflag:"eserver-port"`
	Force  bool         `mapstructure:"force" cobraflag:"eserver-force"`
	Tag    string       `mapstructure:"tag" cobraflag:"eserver-tag"`
	IP     string       `mapstructure:"ip"`
	Images ImagesConfig `mapstructure:"images"`
	EVEIP  string       `mapstructure:"eve-ip"`
}

type EClientConfig struct {
	Tag   string `mapstructure:"tag"`
	Image string `mapstructure:"image"`
}

type ImagesConfig struct {
	EServerImageDist string `mapstructure:"dist" cobraflag:"image-dist" resolvepath:""`
}

type EdenConfig struct {
	Download     bool   `mapstructure:"download" cobraflag:"download"`
	BinDir       string `mapstructure:"bin-dist" cobraflag:"bin-dist" resolvepath:""`
	CertsDir     string `mapstructure:"certs-dist" cobraflag:"certs-dist" resolvepath:""`
	Dist         string `mapstructure:"dist"`
	Root         string `mapstructure:"root"`
	SSHKey       string `mapstructure:"ssh-key" cobraflag:"ssh-key" resolvepath:""`
	EdenBin      string `mapstructure:"eden-bin"`
	TestBin      string `mapstructure:"test-bin"`
	TestScenario string `mapstructure:"test-scenario"`
	Tests        string `mapstructure:"tests" resolvepath:""`

	EServer EServerConfig `mapstructure:"eserver"`

	EClient EClientConfig `mapstructure:"eclient"`
	Images  ImagesConfig  `mapstructure:"images"`
}

type RedisConfig struct {
	RemoteURL string `mapstructure:"adam" cobraflag:"adam-redis-url"`
	Tag       string `mapstructure:"tag" cobraflag:"redis-tag"`
	Port      int    `mapstructure:"port" cobraflag:"redis-port"`
	Dist      string `mapstructure:"dist" cobraflag:"redis-dist" resolvepath:""`
	Force     bool   `mapstructure:"force" cobraflag:"redis-force"`
	Eden      string `mapstructure:"eden"`
}

type RemoteConfig struct {
	Redis   bool `mapstructure:"redis" cobraflag:"adam-redis"`
	Enabled bool `mapstructure:"enabled"`
}

// TODO: This config is only used in tests, do we really need it?
type CachingConfig struct {
	Redis   bool   `mapstructure:"redis"`
	Enabled bool   `mapstructure:"enabled"`
	Prefix  string `mapstructure:"prefix"`
}

type AdamConfig struct {
	Tag         string `mapstructure:"tag" cobraflag:"adam-tag"`
	Port        int    `mapstructure:"port" cobraflag:"adam-port"`
	Dist        string `mapstructure:"dist" cobraflag:"adam-dist" resolvepath:""`
	CertsDomain string `mapstructure:"domain" cobraflag:"domain"`
	CertsIP     string `mapstructure:"ip" cobraflag:"ip"`
	CertsEVEIP  string `mapstructure:"eve-ip" cobraflag:"eve-ip"`
	APIv1       bool   `mapstructure:"v1" cobrafalg:"force"`
	Force       bool   `mapstructure:"force" cobraflag:"force"`
	CA          string `mapstructure:"ca"`

	Redis   RedisConfig   `mapstructure:"redis"`
	Remote  RemoteConfig  `mapstructure:"remote"`
	Caching CachingConfig `mapstructure:"caching"`
}

type CustomInstallerConfig struct {
	Path   string `mapstructure:"path" resolvepath:""`
	Format string `mapstructure:"format"`
}

type QemuConfig struct {
	MonitorPort      int `mapstructure:"monitor-port" cobraflag:"qemu-monitor-port"`
	NetDevSocketPort int `mapstructure:"netdev-socket-port" cobraflag:"qemu-netdev-socket-port"`
}

type EveConfig struct {
	CustomInstaller CustomInstallerConfig `mapstructure:"custom-installer"`
	QemuConfig      QemuConfig            `mapstructure:"qemu"`

	QemuFirmware   []string          `mapstructure:"firmware" cobraflag:"eve-firmware"`
	QemuConfigPath string            `mapstructure:"config-part" cobraflag:"config-path"`
	QemuDTBPath    string            `mapstructure:"dtb-part" cobraflag:"dtb-part" resolvepath:""`
	QemuOS         string            `mapstructure:"os" cobraflag:"eve-os"`
	ImageFile      string            `mapstructure:"image-file" cobraflag:"image-file" resolvepath:""`
	CertsUUID      string            `mapstructure:"uuid" cobraflag:"uuid"`
	Dist           string            `mapstructure:"dist" cobraflag:"eve-dist" resolvepath:""`
	Repo           string            `mapstructure:"repo" cobraflag:"eve-repo"`
	Registry       string            `mapstructure:"registry" cobraflag:"eve-registry"`
	Tag            string            `mapstructure:"tag" cobraflag:"eve-tag"`
	UefiTag        string            `mapstructure:"uefi-tag" cobraflag:"eve-uefi-tag"`
	HV             string            `mapstructure:"hv" cobraflag:"eve-hv"`
	Arch           string            `mapstructure:"arch" cobraflag:"eve-arch"`
	Platform       string            `mapstructure:"platform" cobraflag:"eve-platform"`
	HostFwd        map[string]string `mapstructure:"hostfwd" cobraflag:"eve-hostfwd"`
	QemuFileToSave string            `mapstructure:"qemu-config" cobraflag:"qemu-config" resolvepath:""`
	QemuCpus       int               `mapstructure:"cpu" cobraflag:"cpus"`
	QemuMemory     int               `mapstructure:"ram" cobraflag:"memory"`
	ImageSizeMB    int               `mapstructure:"disk" cobraflag:"image-size"`
	DevModel       string            `mapstructure:"devmodel" cobraflag:"devmodel"`
	Ssid           string            `mapstructure:"ssid" cobraflag:"ssid"`
	Password       string            `mapstructure:"password" cobraflag:"password"`
	Serial         string            `mapstructure:"serial" cobraflag:"eve-serial"`
	Accel          bool              `mapstructure:"accel" cobraflag:"eve-accel"`

	Pid            string `mapstructure:"pid" cobraflag:"eve-pid" resolvepath:""`
	Log            string `mapstructure:"log" cobraflag:"eve-log" resolvepath:""`
	TelnetPort     int    `mapstructure:"telnet-port" cobraflag:"eve-telnet-port"`
	Remote         bool   `mapstructure:"remote"`
	RemoteAddr     string `mapstructure:"remote-addr"`
	ModelFile      string `mapstructure:"devmodelfile" cobraflag:"devmodel-file"`
	Cert           string `mapstructure:"cert"`
	DeviceCert     string `mapstructure:"device-cert"`
	Name           string `mapstructure:"name"`
	AdamLogLevel   string `mapstructure:"adam-log-level"`
	LogLevel       string `mapstructure:"log-level"`
	Disks          int    `mapstructure:"disks"`
	BootstrapFile  string `mapstructure:"bootstrap-file" cobraflag:"eve-bootstrap-file"`
	UsbNetConfFile string `mapstructure:"usbnetconf-file" cobraflag:"eve-usbnetconf-file"`
	TPM            bool   `mapstructure:"tpm" cobraflag:"tpm"`
}

type RegistryConfig struct {
	Tag  string `mapstructure:"tag" cobraflag:"registry-flag"`
	Port int    `mapstructure:"port" cobraflag:"registry-port"`
	Dist string `mapstructure:"dist" cobraflag:"registry-dist"`
	IP   string `mapstructure:"ip"`
}

type PacketConfig struct {
	Key string `mapstructure:"key" cobraflag:"key"`
}

type GcpConfig struct {
	Key string `mapstructure:"key" cobraflag:"key"`
}

type SdnConfig struct {
	ImageFile      string `mapstructure:"image-file" cobraflag:"sdn-image-file"`
	SourceDir      string `mapstructure:"source-dir" cobraflag:"sdn-source-dir" resolvepath:""`
	RAM            int    `mapstructure:"ram" cobraflag:"sdn-ram"`
	CPU            int    `mapstructure:"cpu" cobraflag:"sdn-cpu"`
	ConfigDir      string `mapstructure:"config-dir" cobraflag:"sdn-config-dir" resolvepath:""`
	LinuxkitBin    string `mapstructure:"linuxkit-bin" cobraflag:"sdn-linuxkit-bin" resolvepath:""`
	NetModelFile   string `mapstructure:"network-model" cobraflag:"sdn-network-model" resolvepath:""`
	ConsoleLogFile string `mapstructure:"console-log" cobraflag:"sdn-console-log" resolvepath:""`
	Disable        bool   `mapstructure:"disable" cobraflag:"sdn-disable"`
	TelnetPort     int    `mapstructure:"telnet-port" cobraflag:"sdn-telnet-port"`
	MgmtPort       int    `mapstructure:"mgmt-port" cobraflag:"sdn-mgmt-port"`
	PidFile        string `mapstructure:"pid" cobraflag:"sdn-pid" resolvepath:""`
	SSHPort        int    `mapstructure:"ssh-port" cobraflag:"sdn-ssh-port"`
}

type EdenSetupArgs struct {
	Eden     EdenConfig     `mapstructure:"eden"`
	Adam     AdamConfig     `mapstructure:"adam"`
	Eve      EveConfig      `mapstructure:"eve"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Registry RegistryConfig `mapstructure:"registry"`
	Packet   PacketConfig   `mapstructure:"packet"`
	Gcp      GcpConfig      `mapstructure:"gcp"`
	Sdn      SdnConfig      `mapstructure:"sdn"`

	ConfigFile string
	ConfigName string
	EdenDir    string
}

// PodConfig store configuration for Pod deployment
type PodConfig struct {
	Name              string
	Metadata          string
	Registry          string
	Networks          []string
	PortPublish       []string
	ACL               []string
	Vlans             []string
	Mount             []string
	Disks             []string
	Profiles          []string
	AppAdapters       []string
	NoHyper           bool
	VncDisplay        uint32
	VncPassword       string
	DiskSize          string
	VolumeSize        string
	AppMemory         string
	VolumeType        string
	AppCpus           uint32
	StartDelay        uint32
	PinCpus           bool
	ImageFormat       string
	SftpLoad          bool
	DirectLoad        bool
	OpenStackMetadata bool
	DatastoreOverride string
	ACLOnlyHost       bool
}

func Merge(dst, src reflect.Value, flags *pflag.FlagSet) {
	for i := 0; i < dst.NumField(); i++ {
		if dst.Field(i).Kind() == reflect.Struct {
			Merge(dst.Field(i), src.Field(i), flags)
		}
		if dst.Type().Field(i).Tag != "" {
			cobraFlagTag := dst.Type().Field(i).Tag.Get("cobraflag")
			if cobraFlagTag == "" {
				continue
			}
			mapStructureTag := dst.Type().Field(i).Tag.Get("mapstructure")
			// if no mapStructureTag define we are not able to load it from config
			// so set from flag
			if mapStructureTag == "" || flags.Changed(cobraFlagTag) {
				dst.Field(i).Set(src.Field(i))
			}
		}
	}
}

func FromViper(configName, verbosity string) (*EdenSetupArgs, error) {
	var err error
	cfg := &EdenSetupArgs{}
	configNameEnv := os.Getenv(defaults.DefaultConfigEnv)
	if configNameEnv != "" {
		configName = configNameEnv
	}
	cfg.ConfigFile = utils.GetConfig(configName)

	if verbosity == "debug" {
		fmt.Println("configName: ", configName)
		fmt.Println("configFile: ", cfg.ConfigFile)
	}

	cfg, err = LoadConfig(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}

	if err := SetUpLogs(verbosity); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SetUpLogs(level string) error {
	log.SetOutput(os.Stdout)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}

func LoadConfig(configFile string) (*EdenSetupArgs, error) {
	viperLoaded, err := utils.LoadConfigFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	if !viperLoaded {
		return nil, fmt.Errorf("viper cannot be loaded")
	}
	viper.SetDefault("eve.uefi-tag", defaults.DefaultEVETag)

	cfg := &EdenSetupArgs{}

	if err = viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct, %w", err)
	}

	resolvePath(cfg.Eden.Root, reflect.ValueOf(cfg).Elem())

	if configFile == "" {
		configFile, _ = utils.DefaultConfigPath()
	}

	configName := path.Base(configFile)
	if pos := strings.LastIndexByte(configName, '.'); pos != -1 {
		configName = configName[:pos]
	}

	viper.SetConfigName(configName)

	return cfg, nil
}

func resolvePath(path string, v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if _, ok := v.Type().Field(i).Tag.Lookup("resolvepath"); ok {
			if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
				val := f.Interface().(string)
				f.SetString(utils.ResolveAbsPathWithRoot(path, val))
			}
		}
		if f.Kind() == reflect.Struct {
			resolvePath(path, f)
		}
	}
}

func ConfigCheck(configName string) error {
	configFile := utils.GetConfig(configName)
	configSaved := utils.ResolveAbsPath(fmt.Sprintf("%s-%s", configName, defaults.DefaultConfigSaved))

	abs, err := filepath.Abs(configSaved)
	if err != nil {
		return fmt.Errorf("fail in reading filepath: %s\n", err.Error())
	}

	if _, err = os.Lstat(abs); os.IsNotExist(err) {
		if err = utils.CopyFile(configFile, abs); err != nil {
			return fmt.Errorf("copying fail %s\n", err.Error())
		}
	} else {

		viperLoaded, err := utils.LoadConfigFile(abs)
		if err != nil {
			return fmt.Errorf("error reading config %s: %s\n", abs, err.Error())
		}
		if viperLoaded {
			confOld := viper.AllSettings()

			if _, err = utils.LoadConfigFile(configFile); err != nil {
				return fmt.Errorf("error reading config %s: %s", configFile, err.Error())
			}

			confCur := viper.AllSettings()

			if reflect.DeepEqual(confOld, confCur) {
				log.Infof("Config file %s is the same as %s\n", configFile, configSaved)
			} else {
				return fmt.Errorf("the current configuration file %s is different from the saved %s. You can fix this with the commands 'eden config clean' and 'eden config add/set/edit'.\n", configFile, abs)
			}
		} else {
			/* Incorrect saved config -- just rewrite by current */
			if err = utils.CopyFile(configFile, abs); err != nil {
				return fmt.Errorf("copying fail %s\n", err.Error())
			}
		}
	}
	return nil
}

func getValStrRepr(v reflect.Value) string {
	if v.Kind() == reflect.String {
		return fmt.Sprintf("'%v'", v.Interface())
	} else {
		return fmt.Sprintf("%v", v.Interface())
	}
}

func WriteConfig(dst reflect.Value, writer io.Writer, nestLevel int) {
	for i := 0; i < dst.NumField(); i++ {
		if structTag := dst.Type().Field(i).Tag.Get("mapstructure"); structTag != "" {
			io.WriteString(writer, strings.Repeat("  ", nestLevel))
			switch dst.Field(i).Kind() {
			case reflect.Struct:
				io.WriteString(writer, structTag+":\n")
				WriteConfig(dst.Field(i), writer, nestLevel+1)
			case reflect.Map:
				io.WriteString(writer, structTag+":\n")
				iter := dst.Field(i).MapRange()
				for iter.Next() {
					k := iter.Key()
					v := iter.Value()
					io.WriteString(writer, strings.Repeat("  ", nestLevel+1))
					// We assume that map cannot have structure as value
					io.WriteString(writer, fmt.Sprintf("%v: %s\n", k.Interface(), getValStrRepr(v)))
				}
			case reflect.Slice:
				io.WriteString(writer, structTag+":\n")
				for j := 0; j < dst.Field(i).Len(); j++ {
					io.WriteString(writer, strings.Repeat("  ", nestLevel+1))
					elem := dst.Field(i).Index(j)
					io.WriteString(writer, fmt.Sprintf("- %v\n", getValStrRepr(elem)))
				}
			case reflect.String: // we need to wrap string in quotes
				io.WriteString(writer, fmt.Sprintf("%s: '%v'\n", structTag, dst.Field(i)))
			default:
				io.WriteString(writer, fmt.Sprintf("%s: %v\n", structTag, dst.Field(i)))
			}
		}
	}
}

func PrintDifferences(a, b interface{}, parentField string) {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		if valA.Interface() != valB.Interface() {
			fmt.Printf("Field %s differs: %v vs %v\n", parentField, valA.Interface(), valB.Interface())
		}
		return
	}

	typeA := valA.Type()

	for i := 0; i < valA.NumField(); i++ {
		fieldName := typeA.Field(i).Name
		fieldValA := valA.Field(i)
		fieldValB := valB.Field(i)

		fullFieldName := fieldName
		if parentField != "" {
			fullFieldName = parentField + "." + fieldName
		}

		if fieldValA.Kind() == reflect.Struct {
			PrintDifferences(fieldValA.Interface(), fieldValB.Interface(), fullFieldName)
		} else {
			if fieldValA.Interface() != fieldValB.Interface() {
				fmt.Printf("Field %s differs: %v vs %v \n", fullFieldName, fieldValA.Interface(), fieldValB.Interface())
			}
		}
	}
}
