package openevec

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eden/pkg/projects"
	"github.com/lf-edge/eve/api/go/config"
	log "github.com/sirupsen/logrus"
)

type configChanger interface {
	getControllerAndDevFromConfig(cfg *EdenSetupArgs) (controller.Cloud, *device.Ctx, error)
	setControllerAndDev(controller.Cloud, *device.Ctx) error
}

type fileChanger struct {
	fileConfig string
	oldHash    [32]byte
}

func changerByControllerMode(controllerMode string) (configChanger, error) {
	if controllerMode == "" {
		return &adamChanger{}, nil
	}
	modeType, modeURL, err := projects.GetControllerMode(controllerMode)
	if err != nil {
		return nil, err
	}
	log.Debugf("Mode type: %s", modeType)
	log.Debugf("Mode url: %s", modeURL)
	var changer configChanger
	switch modeType {
	case "file":
		changer = &fileChanger{fileConfig: modeURL}
	case "adam":
		changer = &adamChanger{adamURL: modeURL}

	default:
		return nil, fmt.Errorf("not implemented type: %s", modeType)
	}
	return changer, nil
}

func (ctx *fileChanger) setControllerAndDev(ctrl controller.Cloud, dev *device.Ctx) error {
	res, err := ctrl.GetConfigBytes(dev, false)
	if err != nil {
		return fmt.Errorf("GetConfigBytes error: %w", err)
	}
	if ctx.oldHash == sha256.Sum256(res) {
		log.Debug("config not modified")
		return nil
	}
	if res, err = controller.VersionIncrement(res); err != nil {
		return fmt.Errorf("VersionIncrement error: %w", err)
	}
	if err = os.WriteFile(ctx.fileConfig, res, 0755); err != nil {
		return fmt.Errorf("WriteFile error: %w", err)
	}
	log.Debug("config modification done")
	return nil
}

func (ctx *fileChanger) getControllerAndDevFromConfig(cfg *EdenSetupArgs) (controller.Cloud, *device.Ctx, error) {
	if ctx.fileConfig == "" {
		return nil, nil, fmt.Errorf("cannot use empty url for file")
	}
	if _, err := os.Lstat(ctx.fileConfig); os.IsNotExist(err) {
		return nil, nil, err
	}
	vars, err := InitVarsFromConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("InitVarsFromConfig error: %w", err)
	}
	ctrl, err := controller.CloudPrepare(vars)
	if err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(ctx.fileConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("file reading error: %w", err)
	}
	var deviceConfig config.EdgeDevConfig
	err = json.Unmarshal(data, &deviceConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal error: %w", err)
	}
	dev, err := ctrl.ConfigParse(&deviceConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("configParse error: %w", err)
	}
	res, err := ctrl.GetConfigBytes(dev, false)
	if err != nil {
		return nil, nil, fmt.Errorf("GetConfigBytes error: %w", err)
	}
	ctx.oldHash = sha256.Sum256(res)
	return ctrl, dev, nil
}

type adamChanger struct {
	adamURL string
}

func (ctx *adamChanger) getController(cfg *EdenSetupArgs) (controller.Cloud, error) {
	vars, err := InitVarsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("InitVarsFromConfig error: %w", err)
	}
	ctrl, err := controller.CloudPrepare(vars)
	if err != nil {
		return nil, fmt.Errorf("CloudPrepare error: %w", err)
	}
	return ctrl, nil
}

func (ctx *adamChanger) getControllerAndDevFromConfig(cfg *EdenSetupArgs) (controller.Cloud, *device.Ctx, error) {
	ctrl, err := ctx.getController(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("getController error: %w", err)
	}
	vars, err := InitVarsFromConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("InitVarsFromConfig error: %w", err)
	}
	ctrl.SetVars(vars)
	devFirst, err := ctrl.GetDeviceCurrent()
	if err != nil {
		return nil, nil, fmt.Errorf("GetDeviceCurrent error: %w", err)
	}
	return ctrl, devFirst, nil
}

func (ctx *adamChanger) setControllerAndDev(ctrl controller.Cloud, dev *device.Ctx) error {
	if err := ctrl.ConfigSync(dev); err != nil {
		return fmt.Errorf("configSync error: %w", err)
	}
	return nil
}
