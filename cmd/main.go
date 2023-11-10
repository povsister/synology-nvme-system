package main

import (
	"github.com/povsister/synology-nvme-system/log"
	"github.com/povsister/synology-nvme-system/mdadm"
)

func main() {
	m, err := mdadm.NewMdStat()
	if err != nil {
		panic(err)
	}
	m.Print()

	log.Flush()
}
