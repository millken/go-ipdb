package ipdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"strings"
	//"fmt"
)

const METALEN = 20

// Header struct
type Header struct {
	version                                                       uint32
	continentLen, countryLen, areaLen, regionLen, cityLen, ispLen uint16
	netLen                                                        uint32
}

// Record struct
type Record struct {
	ip                                                      uint32 //ip
	mask                                                    uint8  //mask
	continentID, countryID, areaID, regionID, cityID, ispID uint16
}

// Hdata struct
type Hdata struct {
	id   uint16
	name string
}

// DB struct
type DB struct {
	Head                                        Header
	Continent, Country, Area, Region, City, Isp map[uint16]string
	Idx                                         map[int]uint32
	Rstart                                      int
	Data                                        []byte
}

// Result structs
type Result struct {
	Cidr, Continent, Country, Area, Region, City, Isp string
}

func Load(dataFile string) (db *DB, err error) {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return
	}
	db = new(DB)
	db.Init(data)
	return
}

func IP2long(ipv4 string) (ret uint32, err error) {
	ret = 0
	ip := net.ParseIP(ipv4)
	if ip == nil {
		err = errors.New("Invalid ip format")
		return
	}
	return binary.BigEndian.Uint32(ip.To4()), nil
}

func Long2Ip(a uint32) net.IP {
	return net.IPv4(byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}

func (db *DB) Init(data []byte) {
	db.Head = Header{
		binary.BigEndian.Uint32(data[:4]),
		binary.BigEndian.Uint16(data[4 : 4+2]),
		binary.BigEndian.Uint16(data[6 : 6+2]),
		binary.BigEndian.Uint16(data[8 : 8+2]),
		binary.BigEndian.Uint16(data[10 : 10+2]),
		binary.BigEndian.Uint16(data[12 : 12+2]),
		binary.BigEndian.Uint16(data[14 : 14+2]),
		binary.BigEndian.Uint32(data[16 : 16+4]),
	}
	//fmt.Printf("head = %+v", db.Head)
	db.Continent = make(map[uint16]string, db.Head.continentLen)
	for i := 0; i < int(db.Head.continentLen); i++ {
		pos := METALEN + i*4
		pack := data[pos : pos+4]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+2]),
		}
		db.Continent[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.Country = make(map[uint16]string, db.Head.countryLen)
	for i := 0; i < int(db.Head.countryLen); i++ {
		pos := METALEN + int(db.Head.continentLen)*4 + i*4
		pack := data[pos : pos+4]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+2]),
		}
		db.Country[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.Area = make(map[uint16]string, db.Head.areaLen)
	for i := 0; i < int(db.Head.areaLen); i++ {
		pos := METALEN + int(db.Head.continentLen)*4 + int(db.Head.countryLen)*4 + i*66
		pack := data[pos : pos+66]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+64]),
		}
		db.Area[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.Region = make(map[uint16]string, db.Head.regionLen)
	for i := 0; i < int(db.Head.regionLen); i++ {
		pos := METALEN + int(db.Head.continentLen)*4 + int(db.Head.countryLen)*4 +
			int(db.Head.areaLen)*66 + i*66
		pack := data[pos : pos+66]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+64]),
		}
		db.Region[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.City = make(map[uint16]string, db.Head.cityLen)
	for i := 0; i < int(db.Head.cityLen); i++ {
		pos := METALEN + int(db.Head.continentLen)*4 + int(db.Head.countryLen)*4 +
			int(db.Head.areaLen)*66 + int(db.Head.regionLen)*66 + i*66
		pack := data[pos : pos+66]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+64]),
		}
		db.City[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.Isp = make(map[uint16]string, db.Head.ispLen)
	for i := 0; i < int(db.Head.ispLen); i++ {
		pos := METALEN + int(db.Head.continentLen)*4 + int(db.Head.countryLen)*4 +
			int(db.Head.areaLen)*66 + int(db.Head.regionLen)*66 + int(db.Head.cityLen)*66 + i*66
		pack := data[pos : pos+66]
		dc := Hdata{
			binary.BigEndian.Uint16(pack[:2]),
			string(pack[2 : 2+64]),
		}
		db.Isp[dc.id] = strings.Trim(dc.name, "\u0000")
		//fmt.Printf("dc = %+v", dc)
	}
	db.Rstart = METALEN + int(db.Head.continentLen)*4 + int(db.Head.countryLen)*4 +
		int(db.Head.areaLen)*66 + int(db.Head.regionLen)*66 + int(db.Head.cityLen)*66 +
		int(db.Head.ispLen)*66

	db.Idx = make(map[int]uint32, 256)
	for i := 0; i < 256; i++ {
		off := db.Rstart + int(db.Head.netLen)*17 + i*4
		db.Idx[i] = binary.BigEndian.Uint32(data[off : off+4])
	}
	//fmt.Printf("idx = %+v", db.Idx)
	db.Data = data
	return
}

func (db *DB) Find(ipv4 string) (result *Result, err error) {
	var ip32 uint32
	if ip32, err = IP2long(ipv4); err != nil {
		return nil, err
	}
	return db.FindByUint(ip32)
}

func (db *DB) findStartIdxOffset(ip uint32) (start int) {
	ipFirst := int(ip >> 24)
	start = 0
	for i := ipFirst; i > 0; i-- {
		if db.Idx[i] != 0 {
			start = int(db.Idx[i])
			break
		}
	}
	return
}

func (db *DB) findEndIdxOffset(ip uint32) (end int) {
	ipFirst := int(ip >> 24)
	end = int(db.Head.netLen - 1)
	for i := ipFirst + 1; i < 256; i++ {
		if db.Idx[i] != 0 {
			end = int(db.Idx[i])
			break
		}
	}
	return
}

func (db *DB) FindByUint(ip uint32) (result *Result, err error) {
	f := db.findStartIdxOffset(ip)
	n := 0
	l := db.findEndIdxOffset(ip)
	for f <= l {
		m := int((f + l) / 2)
		n = n + 1
		p := db.Rstart + m*17
		pack := db.Data[p : p+17]
		r := Record{
			binary.BigEndian.Uint32(pack[:4]),
			uint8(pack[4 : 4+1][0]),
			binary.BigEndian.Uint16(pack[5 : 5+2]),
			binary.BigEndian.Uint16(pack[7 : 7+2]),
			binary.BigEndian.Uint16(pack[9 : 9+2]),
			binary.BigEndian.Uint16(pack[11 : 11+2]),
			binary.BigEndian.Uint16(pack[13 : 13+2]),
			binary.BigEndian.Uint16(pack[15 : 15+2]),
		}
		rs := r.ip
		re := rs + uint32(math.Pow(2, float64(32-r.mask))) - 1
		if ip >= rs && ip <= re {
			return &Result{
				Cidr:      fmt.Sprintf("%s/%d", Long2Ip(rs).To4(), r.mask),
				Continent: db.Continent[r.continentID],
				Country:   db.Country[r.countryID],
				Area:      db.Area[r.areaID],
				Region:    db.Region[r.regionID],
				City:      db.City[r.cityID],
				Isp:       db.Isp[r.ispID],
			}, nil
		}
		if ip > re {
			f = m + 1
		}
		if ip < rs {
			l = m - 1
		}
	}
	return nil, errors.New("Not Found")
}
