/*
// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ciao-project/ciao/networking/libsnnet"
	"github.com/ciao-project/ciao/payloads"
	"github.com/golang/glog"
)

type startTimes struct {
	startStamp        time.Time
	backingImageCheck time.Time
	networkStamp      time.Time
	creationStamp     time.Time
	runStamp          time.Time
}

func createInstance(vm virtualizer, instanceDir string, cfg *vmConfig,
	bridge, gatewayIP string, userData, metaData []byte) (err error) {
	err = os.MkdirAll(instanceDir, 0775)
	if err != nil {
		glog.Errorf("Cannot create instance directory: %v", err)
		return
	}
	err = os.Chmod(instanceDir, 0775)
	if err != nil {
		glog.Errorf("Unable to set permissions for instance directory: %v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			_ = os.RemoveAll(instanceDir)
		}
	}()

	err = vm.createImage(bridge, gatewayIP, userData, metaData)
	if err != nil {
		glog.Errorf("Unable to create image %v", err)
		panic(err)
	}

	err = cfg.save(instanceDir)
	if err != nil {
		glog.Errorf("Failed to store state information %v", err)
		panic(err)
	}

	return
}

func processStart(cmd *insStartCmd, instanceDir string, vm virtualizer, conn serverConn) (*startTimes, *startError) {
	var err error
	var vnicName string
	var bridge string
	var gatewayIP string
	var vnicCfg *libsnnet.VnicConfig
	var st startTimes
	var fds []*os.File

	st.startStamp = time.Now()

	cfg := cmd.cfg

	/*
		Need to check to see if the instance exists first.  Otherwise
		if it does exist but we fail for another reason first, the instance would be
		deleted.
	*/

	_, err = os.Stat(instanceDir)
	if err == nil {
		err = fmt.Errorf("Instance %s has already been created", cfg.Instance)
		return nil, &startError{err, payloads.InstanceExists, cmd.cfg.Restart}
	}

	err = vm.ensureBackingImage()
	if err != nil {
		return nil, &startError{err, payloads.ImageFailure, cmd.cfg.Restart}
	}

	st.backingImageCheck = time.Now()

	if networking {
		vnicCfg, err = createVnicCfg(cfg)
		if err != nil {
			glog.Errorf("Could not create VnicCFG: %s", err)
			return nil, &startError{err, payloads.InvalidData, cmd.cfg.Restart}
		}
	}

	if vnicCfg != nil {
		vnicName, bridge, gatewayIP, fds, err = createVnic(conn, vnicCfg)
		if err != nil {
			return nil, &startError{err, payloads.NetworkFailure, cmd.cfg.Restart}
		}
		defer func() {
			for _, f := range fds {
				_ = f.Close()
			}
		}()
	}

	st.networkStamp = time.Now()

	err = createInstance(vm, instanceDir, cfg, bridge, gatewayIP, cmd.userData,
		cmd.metaData)
	if err != nil {
		if vnicCfg != nil {
			destroyVnic(conn, vnicCfg)
		}
		return nil, &startError{err, payloads.ImageFailure, cmd.cfg.Restart}
	}

	st.creationStamp = time.Now()

	err = vm.startVM(vnicName, getNodeIPAddress(), cephID, fds)
	if err != nil {
		if vnicCfg != nil {
			destroyVnic(conn, vnicCfg)
		}
		return nil, &startError{err, payloads.LaunchFailure, cmd.cfg.Restart}
	}

	st.runStamp = time.Now()

	return &st, nil
}
