package ipdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	//"fmt"
)

// Header struct
type Header struct {
	version      uint32
	continentLen uint32
	countryLen   uint32
	areaLen      uint32
	regionLen    uint32
	cityLen      uint32
	ispLen       uint32
	netLen       uint32
}

// Record struct
type Record struct {
	ip          uint32 //ip
	mask        uint8  //mask
	continentID uint32
	countryID   uint32
	areaID      uint32
	regionID    uint32
	cityID      uint32
	ispID       uint32
}

// Hdata struct
type Hdata struct {
	id   uint32
	name string
}

// DB struct
type DB struct {
	Head      Header
	Continent map[uint32]string
	Country   map[uint32]string
	Area      map[uint32]string
	Region    map[uint32]string
	City      map[uint32]string
	Isp       map[uint32]string
	Rstart    int
	Data      []byte
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
		binary.BigEndian.Uint32(data[4 : 4+4]),
		binary.BigEndian.Uint32(data[8 : 8+4]),
		binary.BigEndian.Uint32(data[12 : 12+4]),
		binary.BigEndian.Uint32(data[16 : 16+4]),
		binary.BigEndian.Uint32(data[20 : 20+4]),
		binary.BigEndian.Uint32(data[24 : 24+4]),
		binary.BigEndian.Uint32(data[28 : 28+4]),
	}
	//fmt.Printf("head = %+v", db.Head)
	db.Continent = make(map[uint32]string, db.Head.continentLen)
	for i := 0; i < int(db.Head.continentLen); i++ {
		pos := 32 + i*6
		pack := data[pos : pos+4]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+2]),
		}
		db.Continent[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.Country = make(map[uint32]string, db.Head.countryLen)
	for i := 0; i < int(db.Head.countryLen); i++ {
		pos := 32 + int(db.Head.continentLen)*6 + i*6
		pack := data[pos : pos+4]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+2]),
		}
		db.Country[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.Area = make(map[uint32]string, db.Head.areaLen)
	for i := 0; i < int(db.Head.areaLen); i++ {
		pos := 32 + int(db.Head.continentLen)*6 + int(db.Head.countryLen)*6 + i*68
		pack := data[pos : pos+68]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+64]),
		}
		db.Area[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.Region = make(map[uint32]string, db.Head.regionLen)
	for i := 0; i < int(db.Head.regionLen); i++ {
		pos := 32 + int(db.Head.continentLen)*6 + int(db.Head.countryLen)*6 +
			int(db.Head.areaLen)*68 + i*68
		pack := data[pos : pos+68]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+64]),
		}
		db.Region[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.City = make(map[uint32]string, db.Head.cityLen)
	for i := 0; i < int(db.Head.cityLen); i++ {
		pos := 32 + int(db.Head.continentLen)*6 + int(db.Head.countryLen)*6 +
			int(db.Head.areaLen)*68 + int(db.Head.regionLen)*68 + i*68
		pack := data[pos : pos+68]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+64]),
		}
		db.City[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.Isp = make(map[uint32]string, db.Head.ispLen)
	for i := 0; i < int(db.Head.ispLen); i++ {
		pos := 32 + int(db.Head.continentLen)*6 + int(db.Head.countryLen)*6 +
			int(db.Head.areaLen)*68 + int(db.Head.regionLen)*68 + int(db.Head.cityLen)*68 + i*68
		pack := data[pos : pos+68]
		dc := Hdata{
			binary.BigEndian.Uint32(pack[:4]),
			string(pack[4 : 4+64]),
		}
		db.Isp[dc.id] = dc.name
		//fmt.Printf("dc = %+v", dc)
	}
	db.Rstart = 32 + int(db.Head.continentLen)*6 + int(db.Head.countryLen)*6 +
		int(db.Head.areaLen)*68 + int(db.Head.regionLen)*68 + int(db.Head.cityLen)*68 +
		int(db.Head.ispLen)*68
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

func (db *DB) Lookup(ipv4 string) (*Result, error) {
	return nil, nil
}

func (db *DB) FindByUint(ip32 uint32) (result *Result, err error) {
	f := 0
	n := 0
	l := int(db.Head.netLen) - 1
	for f <= l {
		m := int((f + l) / 2)
		n = n + 1
		p := db.Rstart + m*29
		pack := db.Data[p : p+29]
		r := Record{
			binary.BigEndian.Uint32(pack[:4]),
			uint8(pack[4 : 4+1][0]),
			binary.BigEndian.Uint32(pack[5 : 5+4]),
			binary.BigEndian.Uint32(pack[9 : 9+4]),
			binary.BigEndian.Uint32(pack[13 : 13+4]),
			binary.BigEndian.Uint32(pack[17 : 17+4]),
			binary.BigEndian.Uint32(pack[21 : 21+4]),
			binary.BigEndian.Uint32(pack[25 : 25+4]),
		}
		rs := r.ip
		re := rs + uint32(math.Pow(2, float64(32-r.mask))) - 1
		if ip32 >= rs && ip32 <= re {
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
		if ip32 > re {
			f = m + 1
		}
		if ip32 < rs {
			l = m - 1
		}

		//fmt.Printf(" (ip32)=%d[%s], (rs)=%d[%s], (re)=%d[%s], (mask)=%d\n",
		//	ip32, Long2Ip(ip32).To4(), rs, Long2Ip(rs).To4(), re, Long2Ip(re).To4(), r.mask)

	}
	return nil, errors.New("Unknown error")
}
