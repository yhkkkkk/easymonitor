package xtime

import (
	"time"
)

var Zone *time.Location

const (
	ISO8601 = "2006-01-02T15:04:05.000Z07:00"
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type TimeLimit struct {
	Seconds uint `toml:"Seconds"`
	Minutes uint `toml:"Minutes"`
	Days    uint `toml:"Days"`
}

func (t TimeLimit) GetTimeDuration() time.Duration {
	return time.Duration(t.Seconds)*time.Second +
		time.Duration(t.Minutes)*time.Minute +
		time.Duration(t.Days)*24*time.Hour
}

func (t TimeLimit) GetSeconds() int {
	s := t.Seconds
	s += t.Minutes * 60
	s += t.Days * 3600 * 24
	return int(s)
}

func FixedZone(zoneStr string) {
	z, err := time.LoadLocation(zoneStr)
	if err != nil {
		panic(err)
	} else {
		Zone = z
	}
}

func Now() time.Time {
	t := time.Now()
	t.In(Zone)
	return t
}

func TimeFormatISO8601(t time.Time) string {
	return t.In(Zone).Format(ISO8601)
}

func Parse(s string) time.Time {
	t, _ := time.ParseInLocation(time.RFC3339, s, Zone)
	return t
}

func TimeParse(s string) time.Time {
	t, _ := time.Parse(RFC3339, s)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc)
}

func DtimeParseStr(input string) string {
	t, _ := time.Parse(time.RFC3339, input)
	// 将时间转换为UTC，并格式化为目标格式
	output := t.UTC().Format("2006-01-02T15:04:05.000Z")
	return output
}

func DtimeParse(t *time.Time) *time.Time {
	if t != nil {
		*t = t.UTC()
		output, _ := time.Parse("2006-01-02T15:04:05.000Z", t.Format("2006-01-02T15:04:05.000Z"))
		return &output
	}
	return nil
}
