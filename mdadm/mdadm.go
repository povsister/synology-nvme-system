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
)

/*
# EXAMPLE
# cat /proc/mdstat

Personalities : [raid1] [raid0]
md5 : active raid1 sata4p3[0]
      17567603712 blocks super 1.2 [1/1] [U]

md3 : active raid1 sata1p3[0] sata2p3[1]
      17567603712 blocks super 1.2 [2/2] [UU]

md4 : active raid1 sata3p3[0]
      17567603712 blocks super 1.2 [1/1] [U]

md2 : active raid0 nvme0n1p3[0] nvme1n1p3[1]
      3979348608 blocks super 1.2 64k chunks [2/2] [UU]

md1 : active raid1 nvme1n1p2[1] nvme0n1p2[0] sata1p2[2] sata4p2[5] sata3p2[4] sata2p2[3]
      2097088 blocks [6/6] [UUUUUU]

md0 : active raid1 nvme1n1p1[1] nvme0n1p1[0] sata1p1[2] sata4p1[5] sata3p1[4] sata2p1[3]
      8388544 blocks [6/6] [UUUUUU]
md0 : active raid1 nvme1n1p1[1] nvme0n1p1[0] sata1p1[2] sata4p1[6](F) sata3p1[4] sata2p1[3]
      8388544 blocks [6/5] [UUUUU_]
md0 : active raid1 sata4p1[6] nvme1n1p1[1] nvme0n1p1[0] sata1p1[2] sata3p1[4] sata2p1[3]
      8388544 blocks [6/5] [UUUUU_]
      [==>..................]  recovery = 11.7% (985024/8388544) finish=0.3min speed=328341K/sec

unused devices: <none>

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

type Md struct {
	rawMdStat    [][]byte
	MdNum        int64
	MdName       string
	MdStatus     string
	RaidType     string
	RaidDevsCnt  int64
	TotalDevsCnt int64
	Devices      []*MdDevice
}

type MdDevice struct {
	blockdev.Partition
	DeviceNumber int64
	DeviceStatus string
}

func (ms *MdStat) update() error {
	mdstat, err := os.ReadFile(`/proc/mdstat`)
	if err != nil {
		return fmt.Errorf("err read /proc/mdstat: %w", err)
	}
	detectedMd := ms.splitMdstat(mdstat)
}

func (ms *MdStat) splitMdstat(stat []byte) (ret []*Md) {
	r := bufio.NewReader(bytes.NewReader(stat))
	var (
		isSameMdProcessing = false
		mdStarterRgx       = regexp.MustCompile(`^(md\d+)\s+:\s+`)
		currentMd          *Md
	)
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		// empty line
		if len(line) == 0 {
			// end this md processing. try next line
			isSameMdProcessing = false
			continue
		}
		if !isSameMdProcessing {
			if m := mdStarterRgx.FindSubmatch(line); len(m) > 0 {
				isSameMdProcessing = true
				mdNum, _ := strconv.ParseInt(strings.TrimLeft(string(m[1]), "md"), 10, 64)
				currentMd = &Md{
					MdNum:  mdNum,
					MdName: string(m[1]),
				}
			} else {
				// md pattern not matching. try next line
				continue
			}
		}
		// read line data

	}

	return
}
