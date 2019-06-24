package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/ardielle/ardielle-go/rdl"
	"gitlab.com/cty3000/superman-detector/supermandetector"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
)

func TestNewSupermanDetectorImpl(t *testing.T) {
	type args struct {
		baseUrl        string
		ipaccessdb     *sql.DB
		geodb          *geoip2.Reader
		speedThreshold int32
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(*SupermanDetectorImpl, *SupermanDetectorImpl) error
		afterFunc  func()
		want       *SupermanDetectorImpl
		wantErr    error
	}
	tests := []test{
		func() test {
			ipaccessdb, _ := sql.Open("sqlite3", "./ipaccess.db")
			geodb, _ := geoip2.Open("GeoLite2-City.mmdb")
			args := args{
				baseUrl:        "http://0.0.0.0:80/",
				ipaccessdb:     ipaccessdb,
				geodb:          geodb,
				speedThreshold: 500,
			}
			return test{
				name: "Check success",
				args: args,
				checkFunc: func(gotS, wantS *SupermanDetectorImpl) error {
					if !reflect.DeepEqual(gotS.baseUrl, wantS.baseUrl) ||
						!reflect.DeepEqual(gotS.speedThreshold, wantS.speedThreshold) ||
						reflect.TypeOf(gotS.ipaccessdb) != reflect.TypeOf(wantS.ipaccessdb) ||
						reflect.TypeOf(gotS.geodb) != reflect.TypeOf(wantS.geodb) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: &SupermanDetectorImpl{
					baseUrl:        args.baseUrl,
					ipaccessdb:     args.ipaccessdb,
					geodb:          args.geodb,
					speedThreshold: args.speedThreshold,
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			got, err := NewSupermanDetectorImpl(tt.args.baseUrl)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, error: %v", err)
				return
			} else if tt.wantErr != nil {
				if err == nil {
					t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, err)
				}
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func TestIpAccessRequest2CurrentGeo(t *testing.T) {
	type args struct {
		baseUrl string
		request supermandetector.IpAccessRequest
		ipaccessdb     *sql.DB
		geodb          *geoip2.Reader
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(*supermandetector.CurrentGeo, *supermandetector.CurrentGeo) error
		afterFunc  func()
		want       *supermandetector.CurrentGeo
		wantErr    error
	}
	tests := []test{
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514761200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e42",
					Ip_address:     "91.207.175.104",
				},
			}
			return test{
				name: "Check success",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.CurrentGeo) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: &supermandetector.CurrentGeo{
					Lat:    34.0549,
					Lon:    -118.2578,
					Radius: 200,
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			impl, _ := NewSupermanDetectorImpl(tt.args.baseUrl)

			if tt.args.ipaccessdb != nil {
				impl.ipaccessdb = tt.args.ipaccessdb
			}
			if tt.args.geodb != nil {
				impl.geodb = tt.args.geodb
			}

			got, err := impl.IpAccessRequest2CurrentGeo(&tt.args.request)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, error: %v", err)
				return
			} else if tt.wantErr != nil {
				if err == nil {
					t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, err)
				}
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func TestGenerateIpAccessRecord(t *testing.T) {
	type args struct {
		baseUrl    string
		request    supermandetector.IpAccessRequest
		currentGeo supermandetector.CurrentGeo
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(*supermandetector.IpAccessRecord, *supermandetector.IpAccessRecord) error
		afterFunc  func()
		want       *supermandetector.IpAccessRecord
		wantErr    error
	}
	tests := []test{
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514761200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e42",
					Ip_address:     "91.207.175.104",
				},
				currentGeo: supermandetector.CurrentGeo{
					Lat:    34.0549,
					Lon:    -118.2578,
					Radius: 200,
				},
			}
			return test{
				name: "Check success",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccessRecord) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514761200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e42",
					Ip_address:     "91.207.175.104",
					Lat:            34.0549,
					Lon:            -118.2578,
					Radius:         200,
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			impl, _ := NewSupermanDetectorImpl(tt.args.baseUrl)
			got := impl.GenerateIpAccessRecord(&tt.args.request, &tt.args.currentGeo)

			if tt.checkFunc != nil {
				err := tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func TestGetPrecedingIpAccess(t *testing.T) {
	type args struct {
		baseUrl         string
		currentRecord   *supermandetector.IpAccessRecord
		precedingRecord *supermandetector.IpAccessRecord
		ipaccessdb     *sql.DB
		geodb          *geoip2.Reader
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(*supermandetector.IpAccess, *supermandetector.IpAccess) error
		afterFunc  func()
		want       *supermandetector.IpAccess
		wantErr    error
	}
	tests := []test{
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				currentRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
					Lat:            39.2293,
					Lon:            -76.6907,
					Radius:         10,
				},
				precedingRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514761200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e42",
					Ip_address:     "91.207.175.104",
					Lat:            34.0549,
					Lon:            -118.2578,
					Radius:         200,
				},
			}
			return test{
				name: "Check success",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccess) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: &supermandetector.IpAccess{
					Ip:        "91.207.175.104",
					Speed:     2311,
					Lat:       34.0549,
					Lon:       -118.2578,
					Radius:    200,
					Timestamp: 1514761200,
				},
			}
		}(),
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				currentRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
					Lat:            39.2293,
					Lon:            -76.6907,
					Radius:         10,
				},
			}
			return test{
				name: "Check Failure",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccess) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			impl, _ := NewSupermanDetectorImpl(tt.args.baseUrl)

			if tt.args.ipaccessdb != nil {
				impl.ipaccessdb = tt.args.ipaccessdb
			}
			if tt.args.geodb != nil {
				impl.geodb = tt.args.geodb
			}

			e := impl.RegisterIpAccessRecord(tt.args.currentRecord)
			if e == nil && tt.args.precedingRecord != nil {
				e = impl.RegisterIpAccessRecord(tt.args.precedingRecord)
			}

			if tt.wantErr == nil && e != nil {
				t.Errorf("failed to instantiate, error: %v", e)
				return
			} else if tt.wantErr != nil {
				if e == nil {
					t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, e)
					return
				}
				if tt.wantErr.Error() != e.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, e)
					return
				}

			} else {
				got, err := impl.GetPrecedingIpAccess(tt.args.currentRecord)
				if tt.wantErr == nil && err != nil {
					t.Errorf("failed to instantiate, error: %v", err)
					return
				} else if tt.wantErr != nil {
					if err == nil {
						t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, err)
					}
					if tt.wantErr.Error() != err.Error() {
						t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
					}
				}

				if tt.checkFunc != nil {
					err = tt.checkFunc(got, tt.want)
					if tt.wantErr == nil && err != nil {
						t.Errorf("compare check failed, err: %v", err)
						return
					}
				}
			}
		})
	}
}

func TestGetSubsequentIpAccess(t *testing.T) {
	type args struct {
		baseUrl          string
		currentRecord    *supermandetector.IpAccessRecord
		subsequentRecord *supermandetector.IpAccessRecord
		ipaccessdb     *sql.DB
		geodb          *geoip2.Reader
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(*supermandetector.IpAccess, *supermandetector.IpAccess) error
		afterFunc  func()
		want       *supermandetector.IpAccess
		wantErr    error
	}
	tests := []test{
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				currentRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
					Lat:            39.2293,
					Lon:            -76.6907,
					Radius:         10,
				},
				subsequentRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514851200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e40",
					Ip_address:     "24.242.71.20",
					Lat:            30.3773,
					Lon:            -97.71,
					Radius:         5,
				},
			}
			return test{
				name: "Check success",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccess) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: &supermandetector.IpAccess{
					Ip:        "24.242.71.20",
					Speed:     55,
					Lat:       30.3773,
					Lon:       -97.71,
					Radius:    5,
					Timestamp: 1514851200,
				},
			}
		}(),
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				currentRecord: &supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
					Lat:            39.2293,
					Lon:            -76.6907,
					Radius:         10,
				},
			}
			return test{
				name: "Check Failure",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccess) error {
					if !reflect.DeepEqual(gotS, wantS) {

						return fmt.Errorf("got: %+v, want: %+v", gotS, wantS)
					}
					return nil
				},
				want: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			impl, _ := NewSupermanDetectorImpl(tt.args.baseUrl)

			if tt.args.ipaccessdb != nil {
				impl.ipaccessdb = tt.args.ipaccessdb
			}
			if tt.args.geodb != nil {
				impl.geodb = tt.args.geodb
			}

			e := impl.RegisterIpAccessRecord(tt.args.currentRecord)
			if e == nil && tt.args.subsequentRecord != nil {
				e = impl.RegisterIpAccessRecord(tt.args.subsequentRecord)
			}

			if tt.wantErr == nil && e != nil {
				t.Errorf("failed to instantiate, error: %v", e)
				return
			} else if tt.wantErr != nil {
				if e == nil {
					t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, e)
					return
				}
				if tt.wantErr.Error() != e.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, e)
					return
				}

			} else {
				got, err := impl.GetSubsequentIpAccess(tt.args.currentRecord)
				if tt.wantErr == nil && err != nil {
					t.Errorf("failed to instantiate, error: %v", err)
					return
				} else if tt.wantErr != nil {
					if err == nil {
						t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, err)
					}
					if tt.wantErr.Error() != err.Error() {
						t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
					}
				}

				if tt.checkFunc != nil {
					err = tt.checkFunc(got, tt.want)
					if tt.wantErr == nil && err != nil {
						t.Errorf("compare check failed, err: %v", err)
						return
					}
				}
			}
		})
	}
}

func TestPostIpAccessRequest(t *testing.T) {
	type args struct {
		baseUrl          string
		request          supermandetector.IpAccessRequest
		precedingRecord  supermandetector.IpAccessRecord
		subsequentRecord supermandetector.IpAccessRecord
		ipaccessdbClose  bool
		geodbClose       bool
	}
	type test struct {
		name       string
		args       args
		beforeFunc func(test)
		checkFunc  func(*supermandetector.IpAccessResponse, *supermandetector.IpAccessResponse) error
		afterFunc  func()
		want       *supermandetector.IpAccessResponse
		wantErr    error
	}
	tests := []test{
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
				},
				precedingRecord: supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514761200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e42",
					Ip_address:     "91.207.175.104",
					Lat:            34.0549,
					Lon:            -118.2578,
					Radius:         200,
				},
				subsequentRecord: supermandetector.IpAccessRecord{
					Username:       "bob",
					Unix_timestamp: 1514851200,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e40",
					Ip_address:     "24.242.71.20",
					Lat:            30.3773,
					Lon:            -97.71,
					Radius:         5,
				},
			}
			return test{
				name: "Check success",
				args: args,
				beforeFunc: func(tt test) {
					*tt.want.TravelToCurrentGeoSuspicious = true
					*tt.want.TravelFromCurrentGeoSuspicious = false
				},
				checkFunc: func(gotS, wantS *supermandetector.IpAccessResponse) error {
					if !reflect.DeepEqual(gotS.CurrentGeo, wantS.CurrentGeo) {
						return fmt.Errorf("CurrentGeo got: %+v, want: %+v", gotS.CurrentGeo, wantS.CurrentGeo)
					}
					if !reflect.DeepEqual(*gotS.TravelToCurrentGeoSuspicious, *wantS.TravelToCurrentGeoSuspicious) {
						return fmt.Errorf("TravelToCurrentGeoSuspicious got: %+v, want: %+v", *gotS.TravelToCurrentGeoSuspicious, *wantS.TravelToCurrentGeoSuspicious)
					}
					if !reflect.DeepEqual(*gotS.TravelFromCurrentGeoSuspicious, *wantS.TravelFromCurrentGeoSuspicious) {
						return fmt.Errorf("TravelFromCurrentGeoSuspicious got: %+v, want: %+v", *gotS.TravelFromCurrentGeoSuspicious, *wantS.TravelFromCurrentGeoSuspicious)
					}
					if !reflect.DeepEqual(gotS.PrecedingIpAccess, wantS.PrecedingIpAccess) {
						return fmt.Errorf("PrecedingIpAccess got: %+v, want: %+v", gotS.PrecedingIpAccess, wantS.PrecedingIpAccess)
					}
					if !reflect.DeepEqual(gotS.SubsequentIpAccess, wantS.SubsequentIpAccess) {
						return fmt.Errorf("SubsequentIpAccess got: %+v, want: %+v", gotS.SubsequentIpAccess, wantS.SubsequentIpAccess)
					}
					return nil
				},
				want: &supermandetector.IpAccessResponse{
					CurrentGeo: &supermandetector.CurrentGeo{
						Lat:    39.2293,
						Lon:    -76.6907,
						Radius: 10,
					},
					TravelToCurrentGeoSuspicious: new(bool),
					TravelFromCurrentGeoSuspicious: new(bool),
					PrecedingIpAccess: &supermandetector.IpAccess{
						Ip:        "91.207.175.104",
						Speed:     2311,
						Lat:       34.0549,
						Lon:       -118.2578,
						Radius:    200,
						Timestamp: 1514761200,
					},
					SubsequentIpAccess: &supermandetector.IpAccess{
						Ip:        "24.242.71.20",
						Speed:     55,
						Lat:       30.3773,
						Lon:       -97.71,
						Radius:    5,
						Timestamp: 1514851200,
					},
				},
			}
		}(),
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "",
				},
			}
			return test{
				name: "Check error to get city from ip",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccessResponse) error {
					if !reflect.DeepEqual(gotS, wantS) {
						return fmt.Errorf("got: %+v, want: %+v", gotS.CurrentGeo, wantS.CurrentGeo)
					}
					return nil
				},
				want: nil,
				wantErr: &rdl.ResourceError{
					Code: 200,
					Message: fmt.Sprintf("Failed to get city from ip, Error:%s", "ipAddress passed to Lookup cannot be nil"),
				},
			}
		}(),
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
				},
				ipaccessdbClose: true,
			}
			return test{
				name: "Check error to ",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccessResponse) error {
					if !reflect.DeepEqual(gotS, wantS) {
						return fmt.Errorf("got: %+v, want: %+v", gotS.CurrentGeo, wantS.CurrentGeo)
					}
					return nil
				},
				want: nil,
				wantErr: &rdl.ResourceError{
					Code: 200,
					Message: fmt.Sprintf("Failed to register IpAccessRecord, Error:%s", "sql: database is closed"),
				},
			}
		}(),
		func() test {
			args := args{
				baseUrl: "http://0.0.0.0:80/",
				request: supermandetector.IpAccessRequest{
					Username:       "bob",
					Unix_timestamp: 1514764800,
					Event_uuid:     "85ad929a-db03-4bf4-9541-8f728fa12e41",
					Ip_address:     "206.81.252.7",
				},
				geodbClose: true,
			}
			return test{
				name: "Check error to ",
				args: args,
				checkFunc: func(gotS, wantS *supermandetector.IpAccessResponse) error {
					if !reflect.DeepEqual(gotS, wantS) {
						return fmt.Errorf("got: %+v, want: %+v", gotS.CurrentGeo, wantS.CurrentGeo)
					}
					return nil
				},
				want: nil,
				wantErr: &rdl.ResourceError{
					Code: 200,
					Message: fmt.Sprintf("Failed to get city from ip, Error:%s", "cannot call Lookup on a closed database"),
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt)
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			impl, _ := NewSupermanDetectorImpl(tt.args.baseUrl)

			if tt.args.ipaccessdbClose {
				impl.ipaccessdb.Close()
			}
			if tt.args.geodbClose {
				impl.geodb.Close()
			}

			impl.RegisterIpAccessRecord(&tt.args.subsequentRecord)
			impl.RegisterIpAccessRecord(&tt.args.precedingRecord)

			got, err := impl.PostIpAccessRequest(nil, &tt.args.request)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, error: %v", err)
				return
			} else if tt.wantErr != nil {
				if err == nil {
					t.Errorf("got nil error, want: %v, got: %v", tt.wantErr, err)
				}
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}
