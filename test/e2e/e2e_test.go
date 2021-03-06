// Copyright (c) 2017 Huawei Technologies Co., Ltd. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	_ "reflect"
	"runtime"
	"testing"

	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
)

var (
	c = client.NewClient(&client.Config{"http://localhost:50040", nil})

	localIqn = "iqn.2017-10.io.opensds:volume:00000001"
)

func init() {
	fmt.Println("Start creating profile...")
	var body = &model.ProfileSpec{
		Name:        "default",
		Description: "default policy",
		Extras:      model.ExtraSpec{},
	}
	prf, err := c.CreateProfile(body)
	if err != nil {
		fmt.Printf("create profile failed: %v\n", err)
		return
	}
	prfBody, _ := json.MarshalIndent(prf, "", "	")
	fmt.Println("create profile success, got:", string(prfBody))
}

func TestListDocks(t *testing.T) {
	t.Log("Start listing docks...")
	dcks, err := c.ListDocks()
	if err != nil {
		t.Error("list docks failed:", err)
		return
	}
	dcksBody, _ := json.MarshalIndent(dcks, "", "	")
	t.Log("list docks success, got:", string(dcksBody))
}

func TestListPools(t *testing.T) {
	t.Log("Start listing pools...")
	pols, err := c.ListPools()
	if err != nil {
		t.Error("list pools failed:", err)
		return
	}
	polsBody, _ := json.MarshalIndent(pols, "", "	")
	t.Log("list pools success, got:", string(polsBody))
}

func TestCreateVolume(t *testing.T) {
	t.Log("Start creating volume...")
	var body = &model.VolumeSpec{
		Name:        "test",
		Description: "This is a test",
		Size:        int64(1),
	}
	vol, err := c.CreateVolume(body)
	if err != nil {
		t.Error("create volume failed:", err)
		return
	}
	volBody, _ := json.MarshalIndent(vol, "", "	")
	t.Log("Create volume success, got:", string(volBody))

	cleanVolumeIfFailedOrFinished(t, vol.Id)
}

func TestGetVolume(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, vol.Id)

	t.Log("Start checking volume...")
	result, err := c.GetVolume(vol.Id)
	if err != nil {
		t.Error("Check volume failed:", err)
		return
	}
	// TODO Test the return value.
	// if !reflect.DeepEqual(vol, result) {
	// 		t.Errorf("Expected %+v, got %+v", vol, result)
	// 		return
	// }

	volBody, _ := json.MarshalIndent(result, "", "	")
	t.Log("Check volume success, got:", string(volBody))
}

func TestListVolumes(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, vol.Id)

	t.Log("Start checking all volumes...")
	vols, err := c.ListVolumes()
	if err != nil {
		t.Error("Check all volumes failed:", err)
		return
	}
	volsBody, _ := json.MarshalIndent(vols, "", "	")
	t.Log("Check all volumes success, got", string(volsBody))
}

func TestUpdateVolume(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}

	t.Log("Start updating volume...")
	var body = &model.VolumeSpec{
		Name:        "Update Volume Name",
		Description: "Update Volume Description",
	}

	newVol, err := c.UpdateVolume(vol.Id, body)
	if err != nil {
		t.Error("update volume failed:", err)
		return
	}

	newVolBody, _ := json.MarshalIndent(newVol, "", "	")
	cleanVolumeIfFailedOrFinished(t, newVol.Id)
	t.Log("Update volume success, got:", string(newVolBody))
}

func TestExtendVolume(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}

	t.Log("Start extending volume...")
	body := &model.ExtendVolumeSpec{
		Extend: model.ExtendSpec{NewSize: int64(vol.Size + 1)},
	}

	newVol, err := c.ExtendVolume(vol.Id, body)
	if err != nil {
		t.Error("extend volume failed:", err)
		return
	}

	newVolBody, _ := json.MarshalIndent(newVol, "", "	")
	cleanVolumeIfFailedOrFinished(t, newVol.Id)
	t.Log("Extend volume success, got:", string(newVolBody))
}

func TestDeleteVolume(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}

	t.Log("Start deleting volume...")
	if err := c.DeleteVolume(vol.Id, nil); err != nil {
		t.Error("delete volume failed:", err)
		return
	}
	t.Log("Delete volume success!")
}

/*
func TestCreateVolumeAttachment(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, vol.Id)

	t.Log("Start creating volume attachment...")
	host, _ := os.Hostname()
	var body = &model.VolumeAttachmentSpec{
		VolumeId: vol.Id,
		HostInfo: model.HostInfo{
			Host:      host,
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Ip:        getHostIp(),
			Initiator: localIqn,
		},
	}
	atc, err := c.CreateVolumeAttachment(body)
	if err != nil {
		t.Error("create volume attachment failed:", err)
		return
	}
	atcBody, _ := json.MarshalIndent(atc, "", "	")
	t.Log("create volume attachment success, got", string(atcBody))

	t.Log("Start cleaning volume attachment...")
	if err := c.DeleteVolumeAttachment(atc.Id, nil); err != nil {
		t.Error("Clean volume attachment failed:", err)
		return
	}
	t.Log("End cleaning volume attachment...")
}

func TestGetVolumeAttachment(t *testing.T) {
	atc, err := prepareVolumeAttachment(t)
	if err != nil {
		t.Error("Failed to run volume attachment prepare function:", err)
		return
	}
	defer cleanVolumeAndAttachmentIfFailedOrFinished(t, atc.VolumeId, atc.Id)

	t.Log("Start checking volume attachment...")
	result, err := c.GetVolumeAttachment(atc.Id)
	if err != nil {
		t.Error("Check volume attachment failed:", err)
		return
	}
	// TODO Test the return value.
	// if !reflect.DeepEqual(atc, result) {
	// 	t.Errorf("Expected %+v, got %+v", atc, result)
	// 	return
	// }

	atcBody, _ := json.MarshalIndent(result, "", "	")
	t.Log("Check volume attachment success, got:", string(atcBody))
}

func TestListVolumeAttachments(t *testing.T) {
	atc, err := prepareVolumeAttachment(t)
	if err != nil {
		t.Error("Failed to run volume attachment prepare function:", err)
		return
	}
	defer cleanVolumeAndAttachmentIfFailedOrFinished(t, atc.VolumeId, atc.Id)

	t.Log("Start checking all volume attachments...")
	atcs, err := c.ListVolumeAttachments()
	if err != nil {
		t.Error("Check all volume attachments failed:", err)
		return
	}
	atcsBody, _ := json.MarshalIndent(atcs, "", "	")
	t.Log("list volume attachments success, got:", string(atcsBody))
}

func TestDeleteVolumeAttachment(t *testing.T) {
	atc, err := prepareVolumeAttachment(t)
	if err != nil {
		t.Error("Failed to run volume attachment prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, atc.VolumeId)

	t.Log("Start deleting volume attachment...")
	if err := c.DeleteVolumeAttachment(atc.Id, nil); err != nil {
		t.Error("delete volume attachment failed:", err)
		return
	}
	t.Log("Delete volume attachment success!")
}
*/

func TestCreateVolumeSnapshot(t *testing.T) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, vol.Id)

	t.Log("Start creating volume snapshot...")
	var body = &model.VolumeSnapshotSpec{
		Name:        "test-snapshot",
		Description: "This is a snapshot test",
		VolumeId:    vol.Id,
	}
	snp, err := c.CreateVolumeSnapshot(body)
	if err != nil {
		t.Error("create volume snapshot failed:", err)
		return
	}
	snpBody, _ := json.MarshalIndent(snp, "", "	")
	t.Log("create volume snapshot success, got:", string(snpBody))

	t.Log("Start cleaing volume snapshot...")
	if err := c.DeleteVolumeSnapshot(snp.Id, nil); err != nil {
		t.Error("Clean volume snapshot failed:", err)
		return
	}
	t.Log("End cleaing volume snapshot...")
}

func TestGetVolumeSnapshot(t *testing.T) {
	snp, err := prepareVolumeSnapshot(t)
	if err != nil {
		t.Error("Failed to run volume snapshot prepare function:", err)
		return
	}
	defer cleanVolumeAndSnapshotIfFailedOrFinished(t, snp.VolumeId, snp.Id)

	t.Log("Start checking volume snapshot...")
	result, err := c.GetVolumeSnapshot(snp.Id)
	if err != nil {
		t.Error("Check volume snapshot failed:", err)
		return
	}
	// TODO Test the return value.
	//
	//	if !reflect.DeepEqual(snp, result) {
	//		t.Errorf("Expected %+v, got %+v", snp, result)
	//		return
	//	}

	snpBody, _ := json.MarshalIndent(result, "", "	")
	t.Log("Check volume snapshot success, got:", string(snpBody))
}

func TestListVolumeSnapshots(t *testing.T) {
	snp, err := prepareVolumeSnapshot(t)
	if err != nil {
		t.Error("Failed to run volume snapshot prepare function:", err)
		return
	}
	defer cleanVolumeAndSnapshotIfFailedOrFinished(t, snp.VolumeId, snp.Id)

	t.Log("Start checking all volume snapshots...")
	snps, err := c.ListVolumeSnapshots()
	if err != nil {
		t.Error("list volume snapshots failed:", err)
		return
	}
	snpsBody, _ := json.MarshalIndent(snps, "", "	")
	t.Log("Check all volume snapshots success, got:", string(snpsBody))
}

func TestDeleteVolumeSnapshot(t *testing.T) {
	snp, err := prepareVolumeSnapshot(t)
	if err != nil {
		t.Error("Failed to run volume snapshot prepare function:", err)
		return
	}
	defer cleanVolumeIfFailedOrFinished(t, snp.VolumeId)

	t.Log("Start deleting volume snapshot...")
	if err := c.DeleteVolumeSnapshot(snp.Id, nil); err != nil {
		t.Error("delete volume snapshot failed:", err)
		return
	}
	t.Log("Delete volume snapshot success!")
}

func TestUpdateVolumeSnapshot(t *testing.T) {
	snp, err := prepareVolumeSnapshot(t)
	if err != nil {
		t.Error("Failed to run volume snapshot prepare function:", err)
		return
	}
	defer cleanVolumeAndSnapshotIfFailedOrFinished(t, snp.VolumeId, snp.Id)

	t.Log("Start updating volume snapshot...")
	var body = &model.VolumeSnapshotSpec{
		Name:        "Update Volume Snapshot Name",
		Description: "Update Volume Snapshot Description",
	}

	newSnp, err := c.UpdateVolumeSnapshot(snp.Id, body)
	if err != nil {
		t.Error("update volume snapshot failed:", err)
		return
	}

	newSnpBody, _ := json.MarshalIndent(newSnp, "", "	")
	t.Log("Update volume snapshot success, got:", string(newSnpBody))
}

func prepareVolume(t *testing.T) (*model.VolumeSpec, error) {
	t.Log("Start preparing volume...")
	var body = &model.VolumeSpec{
		Name:        "test",
		Description: "This is a test",
		Size:        int64(1),
	}
	vol, err := c.CreateVolume(body)
	if err != nil {
		t.Error("prepare volume failed:", err)
		return nil, err
	}
	t.Log("End preparing volume...")
	return vol, nil
}

func prepareVolumeAttachment(t *testing.T) (*model.VolumeAttachmentSpec, error) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return nil, err
	}

	t.Log("Start preparing volume attachment...")
	host, _ := os.Hostname()
	var body = &model.VolumeAttachmentSpec{
		VolumeId: vol.Id,
		HostInfo: model.HostInfo{
			Host:      host,
			Platform:  runtime.GOARCH,
			OsType:    runtime.GOOS,
			Ip:        getHostIp(),
			Initiator: localIqn,
		},
	}
	atc, err := c.CreateVolumeAttachment(body)
	if err != nil {
		t.Error("prepare volume attachment failed:", err)
		// Run volume clean function if failed to prepare volume attachment.
		cleanVolumeIfFailedOrFinished(t, atc.VolumeId)
		return nil, err
	}
	t.Log("End preparing volume attachment...")
	return atc, nil
}

func prepareVolumeSnapshot(t *testing.T) (*model.VolumeSnapshotSpec, error) {
	vol, err := prepareVolume(t)
	if err != nil {
		t.Error("Failed to run volume prepare function:", err)
		return nil, err
	}

	t.Log("Start preparing volume snapshot...")
	var body = &model.VolumeSnapshotSpec{
		Name:        "test-snapshot",
		Description: "This is a snapshot test",
		VolumeId:    vol.Id,
	}
	snp, err := c.CreateVolumeSnapshot(body)
	if err != nil {
		t.Error("prepare volume snapshot failed:", err)
		// Run volume clean function if failed to prepare volume snapshot.
		cleanVolumeIfFailedOrFinished(t, snp.VolumeId)
		return nil, err
	}
	t.Log("End preparing volume snapshot...")
	return snp, nil
}

func cleanVolumeIfFailedOrFinished(t *testing.T, volID string) error {
	t.Log("Start cleaning volume...")
	if err := c.DeleteVolume(volID, nil); err != nil {
		t.Error("Clean volume failed:", err)
		return err
	}
	t.Log("End cleaning volume...")
	return nil
}

func cleanVolumeAndAttachmentIfFailedOrFinished(t *testing.T, volID, atcID string) error {
	t.Log("Start cleaning volume attachment...")
	if err := c.DeleteVolumeAttachment(atcID, nil); err != nil {
		t.Error("Clean volume attachment failed:", err)
		return err
	}
	t.Log("End cleaning volume attachment...")

	t.Log("Start cleaning volume...")
	if err := c.DeleteVolume(volID, nil); err != nil {
		t.Error("Clean volume failed:", err)
		return err
	}
	t.Log("End cleaning volume...")
	return nil
}

func cleanVolumeAndSnapshotIfFailedOrFinished(t *testing.T, volID, snpID string) error {
	t.Log("Start cleaing volume snapshot...")
	if err := c.DeleteVolumeSnapshot(snpID, nil); err != nil {
		t.Error("Clean volume snapshot failed:", err)
		return err
	}
	t.Log("End cleaning volume snapshot...")

	t.Log("Start cleaning volume...")
	if err := c.DeleteVolume(volID, nil); err != nil {
		t.Error("Clean volume failed:", err)
		return err
	}
	t.Log("End cleaning volume...")
	return nil
}

// getHostIp return Host IP
func getHostIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String()
		}
	}

	return "127.0.0.1"
}
