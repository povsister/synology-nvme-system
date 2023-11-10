package mdadm

import (
	"os"
	"testing"
)

func TestMdSplit(t *testing.T) {
	mds := &MdStat{}
	data, err := os.ReadFile("testdata/procMdStatRebuildingWriteMostly")
	if err != nil {
		t.Fatal(err)
	}
	mds.allMds = mds.splitMdstat(data)

	for _, md := range mds.allMds {
		t.Logf("%s - Num:%d rawData:", md.MdName, md.MdNum)
		for _, line := range md.rawMdStat {
			t.Log(string(line))
		}
	}
}
