package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/constants"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/utils"
)

var Config = &ConfigData{}

type Disk struct {
	Ignores []string `json:"ignores"`
}

type ConfigData struct {
	loaded          bool   `json:"-"`
	Id              string `json:"id"`
	RemoteHost      string `json:"remote_host"`
	Secret          string `json:"secret"`
	PublicKey       string `json:"public_key"`
	PrivateKey      string `json:"private_key"`
	ServerPublicKey string `json:"server_public_key"`
	Disk            Disk   `json:"disk"`
}

func (c *ConfigData) Save() (err error) {
	if !c.loaded {
		err = &errortypes.WriteError{
			errors.New("config: Config file has not been loaded"),
		}
		return
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File marshal error"),
		}
		return
	}

	err = ioutil.WriteFile(constants.ConfPath, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File write error"),
		}
		return
	}

	return
}

func Load() (err error) {
	data := &ConfigData{}

	exists, err := utils.Exists(constants.ConfPath)
	if err != nil {
		return
	}

	if !exists {
		data.loaded = true
		Config = data
		return
	}

	file, err := ioutil.ReadFile(constants.ConfPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File read error"),
		}
		return
	}

	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File unmarshal error"),
		}
		return
	}

	data.loaded = true

	Config = data

	return
}

func Save() (err error) {
	err = Config.Save()
	if err != nil {
		return
	}

	return
}

func Init() (err error) {
	err = utils.ExistsMkdir(constants.VarDir, 0755)
	if err != nil {
		return
	}

	err = Load()
	if err != nil {
		return
	}

	exists, err := utils.Exists(constants.ConfPath)
	if err != nil {
		panic(err)
	}

	if !exists {
		err = Save()
		if err != nil {
			return
		}
	}

	return
}
