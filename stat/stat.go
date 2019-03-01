package stat

import (
  "fmt"
  "time"
  "net"
  
  "github.com/webability-go/xamboo/config"
  "github.com/webability-go/xamboo/engine/context"
)

/*
This code keeps tracks and stats of the whole webserver and served pages and requests
*/

type RequestStat struct {
  Id        uint64
  StartTime time.Time
  Time      time.Time
  Request   string
  Protocol  string
  Method    string
  Code      int
  Length    int
  Duration  time.Duration
  IP        string
  Port      string
  Alive     bool
  Context  *context.Context
}

type SiteStat struct {
  RequestsTotal  int                     // num requests total, anything included
  RequestsServed map[int]int             // by response code
  LengthServed   int                     // length total, anything included
  Requests       []*RequestStat          // the last minute requests
}

type Stat struct {
  Start          time.Time
  RequestsTotal  int                     // num requests total, anything included
  LengthServed   int                     // length total, anything included
  RequestsServed map[int]int             // by response code
  Requests       []*RequestStat          // by microtime. keep last minute requests
  
  SitesStat      map[string]*SiteStat    // Every site stat. referenced by ID (from config)
}

var SystemStat = CreateStat()
var RequestCounter uint64

func CreateStat() *Stat {
  s := &Stat{
    Start: time.Now()
    RequestsTotal: 0,
    RequestsServed: make(map[int]int),
    LengthServed: 0,
    SitesStat: make(map[string]*SiteStat),
  }
  for _, host := range config.Config.Hosts {
    s.SitesStat[host.Name] = &SiteStat{
      RequestsServed: make(map[int]int),
    }
  }
  
  // launch cleaning thread, while the xamboo go system works
  go s.Clean()
  
  return s
}

func (s* Stat)Clean() {
  // 1. clean Requests from stat
  fmt.Println("Stats cleaner launched. Clean every minute.")
  for {
    n := time.Now()
    // we keep 2 minutes
    delta := time.Minute * 2
    last := 0
    
    // if it's alive: no delete
    for i, r := range s.Requests {
      if r.Time.Add(delta).Before(n) {
        last = i
      } else {
        break
      }
    }
    s.Requests = s.Requests[last:]
    // we clean every 60 seconds
    time.Sleep(time.Minute)
  }
}

func CreateRequestStat(request string, method string, protocol string, code int, length int, duration time.Duration, remoteaddr string) *RequestStat {
  
  SystemStat.RequestsTotal ++
  SystemStat.LengthServed += length

  ip,port,_ := net.SplitHostPort(remoteaddr)

  r := &RequestStat{
    Id: RequestCounter,
    StartTime: time.Now(),
    Time: time.Now(),
    Request: request,
    Method: method,
    Protocol: protocol,
    Code: code,
    Length: length,
    Duration: duration,
    IP: ip,
    Port: port,
    Alive: true,
  }
  RequestCounter++
  SystemStat.Requests = append(SystemStat.Requests, r)
  
  // Adding stat to the site:
  return r
}

func (r *RequestStat)UpdateStat(code int, length int) {
  r.Time = time.Now()
  if code != 0 { r.Code = code }
  r.Length += length
  SystemStat.LengthServed += length
  r.Duration = r.Time.Sub(r.StartTime)
}

func (r *RequestStat)UpdateProtocol(protocol string) {
  r.Protocol = protocol
}

func (r *RequestStat)End() {
  
  // Call stats ? (code entry)
  
  // closed case
  r.Alive = false
}

