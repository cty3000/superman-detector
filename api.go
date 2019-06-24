package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/ardielle/ardielle-go/rdl"
	"gitlab.com/cty3000/superman-detector/supermandetector"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"github.com/umahmood/haversine"
)

type SupermanDetectorImpl struct {
	baseUrl        string
	ipaccessdb     *sql.DB
	geodb          *geoip2.Reader
	speedThreshold int32
}

// NewSupermanDetectorImpl is an implementation to initialize a SupermanDetectorImpl
func NewSupermanDetectorImpl(baseUrl string) (*SupermanDetectorImpl, error) {
	var err error

	impl := new(SupermanDetectorImpl)
	impl.geodb, err = impl.InitGeoDB()
	if err != nil {
		return nil, err
	}
	impl.ipaccessdb, err = impl.InitIPAccessDB()
	if err != nil {
		return nil, err
	}

	impl.baseUrl = baseUrl
	impl.speedThreshold = 500

	return impl, nil
}

// InitGeoDB is an implementation to make a connection with GeoLite2 City database
func (impl *SupermanDetectorImpl) InitGeoDB() (*geoip2.Reader, error) {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		defer db.Close()
		return nil, err
	}

	return db, nil
}

// InitIPAccessDB is an implementation to initialize sqlite database for ip access record
func (impl *SupermanDetectorImpl) InitIPAccessDB() (*sql.DB, error) {
	os.Remove("./ipaccess.db")

	db, err := sql.Open("sqlite3", "./ipaccess.db")
	if err != nil {
		defer db.Close()
		return nil, err
	}

	sqlStmt := `
	create table ipaccess (username text not null, unix_timestamp integer not null, event_uuid text not null primary key, ip_address text not null, lat real not null, lon real not null, radius not null);
	delete from ipaccess;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	return db, nil
}

// IpAccessRequest2CurrentGeo is an implementation to obtain a current geolocation from the request information
func (impl *SupermanDetectorImpl) IpAccessRequest2CurrentGeo(request *supermandetector.IpAccessRequest) (*supermandetector.CurrentGeo, error) {
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(string(request.Ip_address))
	city, err := impl.geodb.City(ip)
	if err != nil {
		return nil, err
	}

	return supermandetector.NewCurrentGeo(&supermandetector.CurrentGeo{
		Lat:    float64(city.Location.Latitude),
		Lon:    float64(city.Location.Longitude),
		Radius: int32(city.Location.AccuracyRadius),
	}), nil
}

// GenerateIpAccessRecord is an implementation to generate a registerable struct as IpAccessRecord from the current geolocation and the request information
func (impl *SupermanDetectorImpl) GenerateIpAccessRecord(request *supermandetector.IpAccessRequest, currentGeo *supermandetector.CurrentGeo) *supermandetector.IpAccessRecord {
	return supermandetector.NewIpAccessRecord(&supermandetector.IpAccessRecord{
		Username:       request.Username,
		Unix_timestamp: request.Unix_timestamp,
		Event_uuid:     request.Event_uuid,
		Ip_address:     request.Ip_address,
		Lat:            currentGeo.Lat,
		Lon:            currentGeo.Lon,
		Radius:         currentGeo.Radius,
	})
}

// RegisterIpAccessRecord is an implementation to register ip access to database as a record
func (impl *SupermanDetectorImpl) RegisterIpAccessRecord(ipRecord *supermandetector.IpAccessRecord) error {
	tx, err := impl.ipaccessdb.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into ipaccess(username, unix_timestamp, event_uuid, ip_address, lat, lon, radius) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(ipRecord.Username, ipRecord.Unix_timestamp, ipRecord.Event_uuid, ipRecord.Ip_address, ipRecord.Lat, ipRecord.Lon, ipRecord.Radius)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

// GetSubsequentIpAccess is an implementation to get a nearest preceding ip access from current ip access
func (impl *SupermanDetectorImpl) GetPrecedingIpAccess(ipRecord *supermandetector.IpAccessRecord) (*supermandetector.IpAccess, error) {
	stmt, err := impl.ipaccessdb.Prepare("select ip_address, lat, lon, radius, unix_timestamp from ipaccess where unix_timestamp < ? order by unix_timestamp limit 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var ip_address, lat, lon, radius, unix_timestamp string
	err = stmt.QueryRow(ipRecord.Unix_timestamp).Scan(&ip_address, &lat, &lon, &radius, &unix_timestamp)
	if err == sql.ErrNoRows {
		log.Printf("No PrecedingIpAccess\n")
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	destination := haversine.Coord{Lat: float64(ipRecord.Lat), Lon: float64(ipRecord.Lon)}
	originLat, _ := strconv.ParseFloat(lat, 64)
	originLon, _ := strconv.ParseFloat(lon, 64)
	originRadius, _ := strconv.Atoi(radius)
	originUnixTimestamp, _ := strconv.Atoi(unix_timestamp)
	origin := haversine.Coord{Lat: originLat, Lon: originLon}
	originTime, _ := strconv.Atoi(unix_timestamp)
	time := (int(ipRecord.Unix_timestamp) - originTime) / 3600
	speed := impl.CalculateSpeed(origin, destination, time)

	return supermandetector.NewIpAccess(&supermandetector.IpAccess{
		Ip:        supermandetector.IPAddress(ip_address),
		Speed:     int32(speed),
		Lat:       originLat,
		Lon:       originLon,
		Radius:    int32(originRadius),
		Timestamp: int32(originUnixTimestamp),
	}), nil
}

// GetSubsequentIpAccess is an implementation to get a nearest subsequent ip access from current ip access
func (impl *SupermanDetectorImpl) GetSubsequentIpAccess(ipRecord *supermandetector.IpAccessRecord) (*supermandetector.IpAccess, error) {
	stmt, err := impl.ipaccessdb.Prepare("select ip_address, lat, lon, radius, unix_timestamp from ipaccess where unix_timestamp > ? order by unix_timestamp limit 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var ip_address, lat, lon, radius, unix_timestamp string
	err = stmt.QueryRow(ipRecord.Unix_timestamp).Scan(&ip_address, &lat, &lon, &radius, &unix_timestamp)
	if err == sql.ErrNoRows {
		log.Printf("No SubsequentIpAccess\n")
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	origin := haversine.Coord{Lat: float64(ipRecord.Lat), Lon: float64(ipRecord.Lon)}
	destinationLat, _ := strconv.ParseFloat(lat, 64)
	destinationLon, _ := strconv.ParseFloat(lon, 64)
	destinationRadius, _ := strconv.Atoi(radius)
	destinationUnixTimestamp, _ := strconv.Atoi(unix_timestamp)
	destination := haversine.Coord{Lat: destinationLat, Lon: destinationLon}
	destinationTime, _ := strconv.Atoi(unix_timestamp)
	time := (destinationTime - int(ipRecord.Unix_timestamp)) / 3600
	speed := impl.CalculateSpeed(origin, destination, time)

	return supermandetector.NewIpAccess(&supermandetector.IpAccess{
		Ip:        supermandetector.IPAddress(ip_address),
		Speed:     int32(speed),
		Lat:       destinationLat,
		Lon:       destinationLon,
		Radius:    int32(destinationRadius),
		Timestamp: int32(destinationUnixTimestamp),
	}), nil
}

// CalculateSpeed is an implementation to calculate speed from the latitude and longitude of origin and destination with the time
func (impl *SupermanDetectorImpl) CalculateSpeed(origin haversine.Coord, destination haversine.Coord, time int) int {
	mi, _ := haversine.Distance(origin, destination)
	return int(mi / float64(time))
}

// PostIpAccessRequest is an implementation for the api logic
func (impl *SupermanDetectorImpl) PostIpAccessRequest(context *rdl.ResourceContext, request *supermandetector.IpAccessRequest) (*supermandetector.IpAccessResponse, error) {

	response := supermandetector.NewIpAccessResponse()

	currentGeo, err := impl.IpAccessRequest2CurrentGeo(request)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get city from ip, Error:%v", err)
		log.Printf(string(errMsg))
		return nil, &rdl.ResourceError{Code: 200, Message: errMsg}
	}
	response.CurrentGeo = currentGeo

	record := impl.GenerateIpAccessRecord(request, response.CurrentGeo)
	err = impl.RegisterIpAccessRecord(record)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to register IpAccessRecord, Error:%v", err)
		log.Printf(string(errMsg))
		return nil, &rdl.ResourceError{Code: 200, Message: errMsg}
	}

	response.PrecedingIpAccess, err = impl.GetPrecedingIpAccess(record)
	if err != nil {
		errMsg := fmt.Sprintf("Failed get PrecedingIpAccess, Error:%v", err)
		return nil, &rdl.ResourceError{Code: 200, Message: errMsg}
	}
	if response.PrecedingIpAccess != nil {
		response.TravelToCurrentGeoSuspicious = new(bool)
		*response.TravelToCurrentGeoSuspicious = (response.PrecedingIpAccess.Speed > impl.speedThreshold)
		log.Printf("PrecedingIpAccess: %v\n", *response.PrecedingIpAccess)
	}

	response.SubsequentIpAccess, err = impl.GetSubsequentIpAccess(record)
	if err != nil {
		errMsg := fmt.Sprintf("Failed get SubsequentIpAccess, Error:%v", err)
		return nil, &rdl.ResourceError{Code: 200, Message: errMsg}
	}
	if response.SubsequentIpAccess != nil {
		response.TravelFromCurrentGeoSuspicious = new(bool)
		*response.TravelFromCurrentGeoSuspicious = (response.SubsequentIpAccess.Speed > impl.speedThreshold)
		log.Printf("SubsequentIpAccess: %v\n", *response.SubsequentIpAccess)
	}

	return response, nil
}

//
// the following is to support TLS-based authentication, and self-authorization that just logs what if *could* enforce.
//

func (impl *SupermanDetectorImpl) Authorize(action string, resource string, principal rdl.Principal) (bool, error) {
	return true, nil
}

func (impl *SupermanDetectorImpl) Authenticate(context *rdl.ResourceContext) bool {
	return false
}
