package mdadm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/povsister/synology-nvme-system/blockdev"
	"github.com/povsister/synology-nvme-system/log"
)

/*
# EXAMPLE
# cat /proc/mdstat


/dev/md0:
        Version : 0.90
  Creation Time : Tue Oct 10 01:00:38 2023
     Raid Level : raid1
     Array Size : 8388544 (8.00 GiB 8.59 GB)
  Used Dev Size : 8388544 (8.00 GiB 8.59 GB)
   Raid Devices : 6
  Total Devices : 6
Preferred Minor : 0
    Persistence : Superblock is persistent

    Update Time : Thu Oct 12 17:38:59 2023
          State : clean, degraded, recovering
 Active Devices : 5
Working Devices : 6
 Failed Devices : 0
  Spare Devices : 1

 Rebuild Status : 61% complete

           UUID : 5a5ee445:e64f250e:05d949f7:b0bbaec7
         Events : 0.6738

    Number   Major   Minor   RaidDevice State
       0     259        1        0      active sync   /dev/nvme0n1p1
       1     259        5        1      active sync   /dev/nvme1n1p1
       2       8        1        2      active sync   /dev/sata1p1
       3       8       17        3      active sync   /dev/sata2p1
       4       8       33        4      active sync   /dev/sata3p1
       6       8       49        5      spare rebuilding   /dev/sata4p1

       4       8       33        4      active sync   /dev/sata3p1
       -       0        0        5      removed

*/

type MdStat struct {
	updatedTime time.Time
	allMds      []*Md
}

func NewMdStat() (*MdStat, error) {
	m := &MdStat{}
	return m, m.update()
}

type Md struct {
	rawMdStat    [][]byte
	MdNum        int64
	MdName       string
	MdState      string
	RaidType     string
	RaidDevsCnt  int64
	TotalDevsCnt int64
	Devices      []*MdDevice
}

func (ms *MdStat) Print() {
	log.Info().Int("count", len(ms.allMds)).Msg("Total mds")
	for _, md := range ms.allMds {
		for _, line := range md.rawMdStat {
			log.Info().Str("name", md.MdName).Str("state", md.MdState).Msg(string(line))
		}
		for _, mdv := range md.Devices {
			log.Info().Str("name", md.MdName).Str("device", mdv.PartitionName).
				Str("state", mdv.DeviceState).
				Msg(md.MdName + " " + md.RaidType)
		}
	}
}

type MdDevice struct {
	Md *Md
	blockdev.Partition
	DeviceNumber int64
	DeviceState  string
}

func (ms *MdStat) update() error {
	mdstat, err := os.ReadFile(`/proc/mdstat`)
	if err != nil {
		return fmt.Errorf("err read /proc/mdstat: %w", err)
	}
	ms.allMds = ms.splitMdstat(mdstat)
	for _, md := range ms.allMds {
		md.update()
	}

	return nil
}

func (ms *MdStat) splitMdstat(stat []byte) (ret []*Md) {
	r := bufio.NewReader(bytes.NewReader(stat))
	var (
		mdStarterRgx = regexp.MustCompile(`^(md\d+)\s+:\s+`)
		currentMd    *Md
	)
	endCurrentMd := func() {
		if currentMd != nil {
			ret = append(ret, currentMd)
			currentMd = nil
		}
	}
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		line = bytes.Trim(line, "\n ")
		// empty line
		if len(line) == 0 {
			// end this md processing. try next line
			endCurrentMd()
			continue
		}
		if m := mdStarterRgx.FindSubmatch(line); len(m) > 0 {
			// end this md processing.
			endCurrentMd()
			// new md
			mdNum, _ := strconv.ParseInt(strings.TrimLeft(string(m[1]), "md"), 10, 64)
			log.Debug().Msgf("found new array %s", string(m[1]))
			currentMd = &Md{
				MdNum:  mdNum,
				MdName: string(m[1]),
			}
		}
		if currentMd != nil {
			// read line data
			currentMd.rawMdStat = append(currentMd.rawMdStat, line)
		}
	}
	endCurrentMd()

	return
}

func (md *Md) update() {
	if len(md.MdName) <= 0 {
		return
	}
	mdDev := "/sys/block/" + md.MdName
	allSlaves, err := os.ReadDir(mdDev + "/slaves")
	if err != nil {
		log.Error(err).Msg("can not read md slaves")
		return
	}
	for _, sl := range allSlaves {
		mdv := &MdDevice{
			Md: md,
			Partition: blockdev.Partition{
				PartitionName: sl.Name(),
				PartitionPath: "/dev/" + sl.Name(),
			},
		}
		mdv.update()
		md.Devices = append(md.Devices, mdv)
	}

	mdState, err := os.ReadFile(mdDev + "/md/array_state")
	if err != nil {
		log.Error(err).Msg("can not read md state")
		return
	}
	md.MdState = strings.Trim(string(mdState), "\n")

	mdLevel, err := os.ReadFile(mdDev + "/md/level")
	if err != nil {
		log.Error(err).Msg("can not read md level")
		return
	}
	md.RaidType = strings.Trim(string(mdLevel), "\n")

}

func (mdv *MdDevice) update() {
	mdvPath := "/sys/block/" + mdv.Md.MdName + "/md/" + strings.TrimLeft(strings.Replace(mdv.PartitionPath, "/", "-", -1), "-")

	devState, err := os.ReadFile(mdvPath + "/state")
	if err != nil {
		log.Error(err).Str("mdDevice", mdv.PartitionPath).
			Msg("can not read mdDevice state")
		return
	}
	mdv.DeviceState = strings.Trim(string(devState), "\n")

}
