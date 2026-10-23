package luscsi

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	mount "k8s.io/mount-utils"
)

func TestControllerVolumeQuota(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Lustre commands require a Unix shell")
	}

	for _, operation := range []string{"create", "expand"} {
		t.Run(operation, func(t *testing.T) {
			for _, tc := range []struct {
				version   string
				wantQuota bool
			}{
				{version: "2.15.4"},
				{version: "2.16.0"},
				{version: "2.16.1", wantQuota: true},
				{version: "2.16.2", wantQuota: true},
				{version: "2.17.0", wantQuota: true},
			} {
				t.Run(tc.version, func(t *testing.T) {
					root := t.TempDir()
					commands := filepath.Join(root, "bin")
					if err := os.Mkdir(commands, 0o755); err != nil {
						t.Fatal(err)
					}
					// Keep lctl's newline and exercise the real version command pipeline.
					for name, script := range map[string]string{
						"lctl": "#!/bin/sh\nprintf 'target_version: %s\\n' \"$LUSCSI_TEST_SERVER_VERSION\"\n",
						"lfs":  "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$LUSCSI_TEST_QUOTA_LOG\"\n",
					} {
						if err := os.WriteFile(filepath.Join(commands, name), []byte(script), 0o755); err != nil {
							t.Fatal(err)
						}
					}
					quotaLog := filepath.Join(root, "quota.log")
					t.Setenv("PATH", commands+string(os.PathListSeparator)+os.Getenv("PATH"))
					t.Setenv("LUSCSI_TEST_SERVER_VERSION", tc.version)
					t.Setenv("LUSCSI_TEST_QUOTA_LOG", quotaLog)

					server := NewControllerServer(Driver{
						DriverOptions: DriverOptions{
							WorkingMountDir:  filepath.Join(root, "mounts"),
							MountPermissions: 0o755,
						},
						mounter: mount.NewFakeMounter(nil),
					})
					const volumeName = "quota-volume"
					const capacity = int64(2 * 1024 * 1024)
					volumePath := filepath.Join(server.WorkingMountDir, volumeName, volumeName)
					capacityRange := &csi.CapacityRange{RequiredBytes: capacity}
					if operation == "create" {
						response, err := server.CreateVolume(context.Background(), &csi.CreateVolumeRequest{
							Name:          volumeName,
							CapacityRange: capacityRange,
							Parameters: map[string]string{
								StorageParamMgsAddress: "test-mgs",
								StorageParamFsName:     "test-fs",
							},
							VolumeCapabilities: []*csi.VolumeCapability{{
								AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}},
								AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER},
							}},
						})
						if err != nil {
							t.Fatalf("CreateVolume: %v", err)
						}
						if response.GetVolume().GetCapacityBytes() != capacity {
							t.Fatalf("CreateVolume capacity = %d, want %d", response.GetVolume().GetCapacityBytes(), capacity)
						}
					} else {
						if err := os.MkdirAll(volumePath, 0o755); err != nil {
							t.Fatal(err)
						}
						response, err := server.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
							VolumeId:      "test-mgs#test-fs#csi~volume#" + volumeName,
							CapacityRange: capacityRange,
						})
						if err != nil {
							t.Fatalf("ControllerExpandVolume: %v", err)
						}
						if response.GetCapacityBytes() != capacity {
							t.Fatalf("ControllerExpandVolume capacity = %d, want %d", response.GetCapacityBytes(), capacity)
						}
					}

					output, err := os.ReadFile(quotaLog)
					if !tc.wantQuota {
						if !os.IsNotExist(err) {
							t.Fatalf("unexpected quota commands for Lustre %s: %q (error: %v)", tc.version, output, err)
						}
						return
					}
					if err != nil {
						t.Fatalf("volume request succeeded but did not configure quota for Lustre %s: %v", tc.version, err)
					}
					lines := strings.Split(strings.TrimSpace(string(output)), "\n")
					if len(lines) != 2 || !strings.HasPrefix(lines[0], "project ") || !strings.HasPrefix(lines[1], "setquota ") {
						t.Fatalf("expected project and setquota commands, got %q", output)
					}
					if !strings.Contains(lines[1], " -B 2048 ") || !strings.HasSuffix(lines[1], " "+volumePath) {
						t.Fatalf("quota command does not set the requested volume capacity: %q", lines[1])
					}
				})
			}
		})
	}
}
