package main

import (
	"path/filepath"
	"os"
	"flag"
	"fmt"
	"log"

	"github.com/kivisade/tabfmt"
	"github.com/dustin/go-humanize"
)

func p2(n uint64) (p uint) {
	if n <= 1 {
		return
	}
	for n != 0 {
		n >>= 1
		p++
	}
	p--
	return
}

func unitconv(p2 uint) (uint, string) {
	switch {
	case p2 < 10:
		return 1 << p2, "B"
	case p2 < 20:
		return 1 << (p2 - 10), "kB"
	case p2 < 30:
		return 1 << (p2 - 20), "MB"
	}
	return 1 << (p2 - 30), "GB"
}

func sizeRange(p2 uint) string {
	if p2 == 0 {
		return "0 - 1 B"
	}
	n1, s1 := unitconv(p2)
	n2, s2 := unitconv(p2 + 1)
	switch {
	case s2 == "B":
		return fmt.Sprintf("%d - %d %s", n1, n2-1, s2)
	case s1 == s2:
		return fmt.Sprintf("%d - %d %s", n1, n2, s2)
	}
	return fmt.Sprintf("%d %s - %d %s", n1, s1, n2, s2)
}

func alloc(fileSz, unitSz uint64) (nBlocks, overhead uint64) {
	if unitSz != 0 {
		nBlocks, overhead = fileSz/unitSz+1, unitSz-fileSz%unitSz
	}
	return
}

type StatCounter struct {
	blockSz       uint64     // disk block (allocation unit) size
	totalCount    uint64     // total number of scanned files
	totalSize     uint64     // total size of scanned files
	totalBlocks   uint64     // total number of blocks occupied by scanned files
	totalOverhead uint64     // total overhead on scanned files
	maxP          uint       // last range index (power of 2), for which files of size [2^i; 2^(i+1)) exist
	count         [40]uint64 // number of files with size in range [2^i; 2^(i+1)), where i is the array index (0..39)
	size          [40]uint64 // total size occupied by files with size in range [2^i; 2^(i+1))
	nBlocks       [40]uint64 // total number of blocks (allocation units) occupied by files with size in range [2^i; 2^(i+1))
	overhead      [40]uint64 // total disk usage overhead (in bytes) on files with size in range [2^i; 2^(i+1))
	overCount     uint64     // number of files with size over 512 GiB (2^39 bytes)
	overSize      uint64     // total size occupied by files with size over 512 GiB (2^39 bytes)
}

func NewStatCounter(blockSz uint64) *StatCounter {
	return &StatCounter{blockSz: blockSz}
}

func (s *StatCounter) addFile(size uint64) {
	var nBlocks, overhead uint64 = alloc(size, s.blockSz)

	s.totalCount++
	s.totalSize += size
	s.totalBlocks += nBlocks
	s.totalOverhead += overhead

	if p := p2(size); p < 40 {
		s.count[p]++
		s.size[p] += size
		s.nBlocks[p] += nBlocks
		s.overhead[p] += overhead
		if p > s.maxP {
			s.maxP = p
		}
	} else {
		s.overCount++
		s.overSize += size
	}
}

func (s *StatCounter) Walk(path string, f os.FileInfo, err error) (abort error) {
	if err != nil {
		log.Println(err)
		return
	}
	if f.IsDir() {
		return
	}
	s.addFile(uint64(f.Size()))
	return
}

func (s *StatCounter) Print() {
	var (
		tab                     = new(tabfmt.Table)
		avgNBlocks, avgOverhead float64
	)

	tab.AddRow("#", "File size", "Files count", "Occupied (bytes)", "Occupied", "Avg. AU", "Total AU", "Avg. OHD", "Total OHD")
	tab.AddRow("=", "=========", "===========", "================", "========", "=======", "========", "========", "=========")

	var i uint
	for i = 0; i <= s.maxP; i++ {
		if s.count[i] > 0 {
			avgNBlocks, avgOverhead = float64(s.nBlocks[i])/float64(s.count[i]), float64(s.overhead[i])/float64(s.count[i])
			tab.AddRow(
				fmt.Sprintf("%d", i),
				sizeRange(i),
				fmt.Sprintf("%d", s.count[i]),
				fmt.Sprintf("%d", s.size[i]),
				humanize.Bytes(s.size[i]),
				fmt.Sprintf("%.2f", avgNBlocks),
				fmt.Sprintf("%d", s.nBlocks[i]),
				fmt.Sprintf("%.2f", avgOverhead),
				humanize.Bytes(s.overhead[i]),
			)
		} else {
			avgNBlocks, avgOverhead = 0, 0
			tab.AddRow(fmt.Sprintf("%d", i), sizeRange(i), "0", "--", "--", "--", "--", "--", "--")
		}
	}

	if s.totalCount > 0 {
		avgNBlocks, avgOverhead = float64(s.totalBlocks)/float64(s.totalCount), float64(s.totalOverhead)/float64(s.totalCount)
	} else {
		avgNBlocks, avgOverhead = 0, 0
	}

	tab.AddRow("-", "---------", "-----------", "----------------", "--------", "-------", "--------", "--------", "---------")
	tab.AddRow(
		"**",
		"TOTAL",
		fmt.Sprintf("%d", s.totalCount),
		fmt.Sprintf("%d", s.totalSize),
		humanize.Bytes(s.totalSize),
		fmt.Sprintf("%.2f", avgNBlocks),
		fmt.Sprintf("%d", s.totalBlocks),
		fmt.Sprintf("%.2f", avgOverhead),
		humanize.Bytes(s.totalOverhead),
	)

	tab.Print("\t")
}

func (s *StatCounter) PrintSimple() {
	var i uint
	for i = 0; i < s.maxP; i++ {
		fmt.Printf("%d\t%s\t%d\t%d\t%s\n", i, sizeRange(i), s.count[i], s.size[i], humanize.Bytes(s.size[i]))
	}
}

func main() {
	var (
		stats   *StatCounter
		blockSz uint64
		out     string
	)

	flag.StringVar(&out, "out", "formatted", "Output format ('formatted' for pretty-printed table or 'tab' for Excel-friendly tabbed format)")
	flag.Uint64Var(&blockSz, "block", 4096, "Disk block (allocation unit) size in bytes")

	flag.Parse()

	stats = NewStatCounter(blockSz)

	root := flag.Arg(0)

	if err := filepath.Walk(root, stats.Walk); err != nil {
		log.Printf("Error while recursively walking %s: %s", root, err)
	}

	switch out {
	case "formatted":
		stats.Print()
	default:
		stats.PrintSimple()
	}
}
