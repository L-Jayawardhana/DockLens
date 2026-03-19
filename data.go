package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// ─── Docker Data Types ────────────────────────────────────────────────────────

type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusStopped ContainerStatus = "exited"
	StatusPaused  ContainerStatus = "paused"
	StatusCreated ContainerStatus = "created"
)

type Container struct {
	ID       string
	Name     string
	Image    string
	Status   ContainerStatus
	Ports    string
	Created  time.Time
	CPU      float64
	Memory   float64
	MemMax   float64
	Network  string
	Logs     []LogLine
	RawStats *container.StatsResponse
}

type LogLine struct {
	Timestamp string
	Text      string
}

type Image struct {
	ID      string
	Name    string
	Tag     string
	Size    int64
	Created time.Time
}

type Volume struct {
	Name       string
	Driver     string
	MountPoint string
	Size       string
	Created    time.Time
}

type Network struct {
	ID         string
	Name       string
	Driver     string
	Scope      string
	Subnet     string
	Containers []string
}

type SystemInfo struct {
	DockerVersion   string
	APIVersion      string
	OS              string
	Arch            string
	Kernel          string
	TotalMemory     string
	CPUs            int
	Containers      int
	ContainersUp    int
	Images          int
	StorageDriver   string
	DiskTotal       string
	DiskUsed        string
	ImagesSize      string
	VolumesSize     string
	BuildCacheSize  string
	ReclaimableSize string
}

// ─── Real Data Fetching ───────────────────────────────────────────────────────

func getContainers(cli *client.Client, prev []Container) ([]Container, error) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Index previous stats for fast lookup
	prevMap := make(map[string]*container.StatsResponse)
	for _, p := range prev {
		prevMap[p.ID] = p.RawStats
	}

	var results []Container
	for _, c := range containers {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var ports []string
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		id12 := c.ID[:12]
		res := Container{
			ID:      id12,
			Name:    name,
			Image:   c.Image,
			Status:  ContainerStatus(c.State),
			Ports:   strings.Join(ports, ", "),
			Created: time.Unix(c.Created, 0),
			Network: "bridge",
		}

		// Fetch real stats if running
		if res.Status == StatusRunning {
			stats, err := cli.ContainerStatsOneShot(ctx, c.ID)
			if err == nil {
				var v container.StatsResponse
				if err := json.NewDecoder(stats.Body).Decode(&v); err == nil {
					res.RawStats = &v

					// Use previous stats for better CPU calculation if available
					prevStats := prevMap[id12]
					if prevStats == nil {
						// If no previous stats are available, use the current stats as the "previous"
						// This will result in zero CPU delta for the first tick, which is expected.
						prevStats = &v
					}
					res.CPU = calculateCPUPercent(res.RawStats, prevStats)
					res.Memory = calculateMemUsageUnixNoCache(&v) / 1024 / 1024
					res.MemMax = float64(v.MemoryStats.Limit) / 1024 / 1024
				}
				stats.Body.Close()
			}
		}

		results = append(results, res)
	}
	return results, nil
}

func calculateCPUPercent(current, previous *container.StatsResponse) float64 {
	if current == nil || previous == nil {
		return 0.0
	}

	var (
		cpuPercent     = 0.0
		systemDelta    = float64(current.CPUStats.SystemUsage) - float64(previous.CPUStats.SystemUsage)
		containerDelta = float64(current.CPUStats.CPUUsage.TotalUsage) - float64(previous.CPUStats.CPUUsage.TotalUsage)
		onlineCPUs     = float64(current.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(current.CPUStats.CPUUsage.PercpuUsage))
	}

	if systemDelta > 0.0 && containerDelta > 0.0 {
		cpuPercent = (containerDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateMemUsageUnixNoCache(v *container.StatsResponse) float64 {
	return float64(v.MemoryStats.Usage - v.MemoryStats.Stats["cache"])
}

func getImages(cli *client.Client) ([]Image, error) {
	ctx := context.Background()
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []Image
	for _, img := range images {
		name := "<none>"
		tag := "<none>"
		if len(img.RepoTags) > 0 {
			parts := strings.Split(img.RepoTags[0], ":")
			name = parts[0]
			if len(parts) > 1 {
				tag = parts[1]
			}
		}

		results = append(results, Image{
			ID:      img.ID[7:19],
			Name:    name,
			Tag:     tag,
			Size:    img.Size,
			Created: time.Unix(img.Created, 0),
		})
	}
	return results, nil
}

func getVolumes(cli *client.Client) ([]Volume, error) {
	ctx := context.Background()
	volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []Volume
	for _, v := range volumes.Volumes {
		results = append(results, Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			MountPoint: v.Mountpoint,
			Size:       "N/A",
			Created:    time.Now(),
		})
	}
	return results, nil
}

func getNetworks(cli *client.Client) ([]Network, error) {
	ctx := context.Background()
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []Network
	for _, n := range networks {
		subnet := "N/A"
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
		}

		var containers []string
		for _, c := range n.Containers {
			containers = append(containers, c.Name)
		}

		results = append(results, Network{
			ID:         n.ID[:12],
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Subnet:     subnet,
			Containers: containers,
		})
	}
	return results, nil
}

func getSystemInfo(cli *client.Client) (SystemInfo, error) {
	ctx := context.Background()
	info, err := cli.Info(ctx)
	if err != nil {
		return SystemInfo{}, err
	}

	version, _ := cli.ServerVersion(ctx)

	return SystemInfo{
		DockerVersion: version.Version,
		APIVersion:    version.APIVersion,
		OS:            info.OperatingSystem,
		Arch:          info.Architecture,
		Kernel:        info.KernelVersion,
		TotalMemory:   fmt.Sprintf("%.2f GiB", float64(info.MemTotal)/1024/1024/1024),
		CPUs:          info.NCPU,
		Containers:    info.Containers,
		ContainersUp:  info.ContainersRunning,
		Images:        info.Images,
		StorageDriver: info.Driver,
	}, nil
}
