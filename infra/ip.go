package infra

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetIp() string {
	resp, err := http.Get("http://ip.dhcp.cn/?ip")
	if err != nil {
		Logger.Errorln(err)
		return ""
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger.Errorln(err)
		return ""
	}

	return fmt.Sprintf("%s", ip)
}
