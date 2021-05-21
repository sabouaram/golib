/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nutsdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
)

const (
	_DefaultFolderData   = "data"
	_DefaultFolderBackup = "backup"
	_DefaultFolderWal    = "wal"
	_DefaultFolderHost   = "host"
)

type NutsDBFolder struct {
	// Working represents the main working folder witch will include sub directories : data, backup, temp...
	// If the base directory is empty, all the sub directory will be absolute directories.
	Base string `mapstructure:"base" json:"base" yaml:"base" toml:"base"`

	// Data represents the sub-dir for the opening database.
	// By default, it will use `data` as sub folder
	Data string `mapstructure:"sub_data" json:"sub_data" yaml:"sub_data" toml:"sub_data"`

	// Backup represents the sub-dir with all backup sub-folder.
	// By default, it will use `backup` as sub folder
	Backup string `mapstructure:"sub_backup" json:"sub_backup" yaml:"sub_backup" toml:"sub_backup"`

	// Temp represents the sub-dir for cluster negotiation.
	// By default, it will use the system temporary folder
	Temp string `mapstructure:"sub_temp" json:"sub_temp" yaml:"sub_temp" toml:"sub_temp"`

	// Temp represents the sub-dir for cluster negotiation.
	// By default, it will use the system temporary folder
	WalDir string `mapstructure:"wal_dir" json:"wal_dir" yaml:"wal_dir" toml:"wal_dir"`

	// Temp represents the sub-dir for cluster negotiation.
	// By default, it will use the system temporary folder
	HostDir string `mapstructure:"host_dir" json:"host_dir" yaml:"host_dir" toml:"host_dir"`

	// LimitNumberBackup represents how many backup will be keep.
	LimitNumberBackup uint8 `mapstructure:"limit_number_backup" json:"limit_number_backup" yaml:"limit_number_backup" toml:"limit_number_backup"`

	// Permission represents the perission apply to folder created.
	Permission os.FileMode `mapstructure:"permission" json:"permission" yaml:"permission" toml:"permission"`
}

func (f NutsDBFolder) Validate() liberr.Error {
	val := validator.New()
	err := val.Struct(f)

	if e, ok := err.(*validator.InvalidValidationError); ok {
		return ErrorValidateConfig.ErrorParent(e)
	}

	out := ErrorValidateConfig.Error(nil)

	for _, e := range err.(validator.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}

func (f NutsDBFolder) getDirectory(base, dir string) (string, liberr.Error) {
	if f.Permission == 0 {
		f.Permission = 0770
	}

	var (
		abs string
		err error
	)

	if len(dir) < 1 {
		return "", nil
	}

	if len(base) > 0 {
		dir = filepath.Join(base, dir)
	}

	if abs, err = filepath.Abs(dir); err != nil {
		return "", ErrorFolderCheck.ErrorParent(err)
	}

	if _, err = os.Stat(abs); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", ErrorFolderCheck.ErrorParent(err)
	} else if err != nil {
		if err = os.MkdirAll(abs, f.Permission); err != nil {
			return "", ErrorFolderCreate.ErrorParent(err)
		}
	}

	return abs, nil
}

func (f NutsDBFolder) GetDirectoryBase() (string, liberr.Error) {
	return f.getDirectory("", f.Base)
}

func (f NutsDBFolder) GetDirectoryData() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Data); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderData)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryBackup() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Backup); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderBackup)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryWal() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.WalDir); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderWal)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryHost() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.HostDir); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory(base, _DefaultFolderHost)
	} else {
		return fs, nil
	}
}

func (f NutsDBFolder) GetDirectoryTemp() (string, liberr.Error) {
	if base, err := f.GetDirectoryBase(); err != nil {
		return "", err
	} else if fs, err := f.getDirectory(base, f.Temp); err != nil {
		return "", err
	} else if fs == "" {
		return f.getDirectory("", os.TempDir())
	} else {
		return fs, nil
	}
}