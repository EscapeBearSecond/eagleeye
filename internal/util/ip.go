package util

import (
	"encoding/binary"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/projectdiscovery/nuclei/v3/pkg/utils/expand"
)

// IsCIDR 判断输入是否为CIDR
func IsCIDR(str string) bool {
	_, _, err := net.ParseCIDR(str)
	return err == nil
}

// IsIP 判断输入是否为IP
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsIPv4 判断输入是否为IPv4
func IsIPv4(str string) bool {
	parsedIP := net.ParseIP(str)
	return parsedIP != nil && parsedIP.To4() != nil && strings.Contains(str, ".")
}

// IsPort 判断输入是否为Port
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsHostPort 判断输入是否为HostPort
func IsHostPort(str string) bool {
	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return false
	}

	return IsIP(host) && IsPort(port)
}

// IsIPRange 判断输入是否为IPRange
func IsIPRange(input string) bool {
	ipRange := strings.Split(input, "-")
	if len(ipRange) != 2 {
		return false
	}

	startIP := net.ParseIP(ipRange[0])
	endIP := net.ParseIP(ipRange[1])

	if startIP == nil || endIP == nil {
		return false
	}

	startIPUint := binary.BigEndian.Uint32(startIP.To4())
	endIPUint := binary.BigEndian.Uint32(endIP.To4())

	return endIPUint > startIPUint
}

func OffsetIP(ip string, size uint32) string {
	ipUint := binary.BigEndian.Uint32(net.ParseIP(ip).To4())
	ipUint += size

	return net.IPv4(
		byte(ipUint>>24),
		byte(ipUint>>16),
		byte(ipUint>>8),
		byte(ipUint),
	).String()
}

func IPRangeSize(ipRange string) (string, uint32) {
	ipRangeSplit := strings.Split(ipRange, "-")

	startIP := net.ParseIP(ipRangeSplit[0])
	endIP := net.ParseIP(ipRangeSplit[1])

	startIPUint := binary.BigEndian.Uint32(startIP.To4())
	endIPUint := binary.BigEndian.Uint32(endIP.To4())

	return ipRangeSplit[0], endIPUint - startIPUint + 1
}

func CIDRSize(cidr string) (string, uint32) {
	_, ipnet, _ := net.ParseCIDR(cidr)
	ones, bits := ipnet.Mask.Size()
	return ipnet.IP.String(), uint32(math.Pow(2, float64(bits-ones)))
}

// ExpandIPRange 扩展IPRange为IP列表
func ExpandIPRange(input string) []string {
	ipRange := strings.Split(input, "-")

	startIP := net.ParseIP(ipRange[0])
	endIP := net.ParseIP(ipRange[1])

	startIPUint := binary.BigEndian.Uint32(startIP.To4())
	endIPUint := binary.BigEndian.Uint32(endIP.To4())

	var ret []string
	for i := startIPUint; i <= endIPUint; i++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, i)
		ret = append(ret, net.IPv4(ip[0], ip[1], ip[2], ip[3]).String())
	}
	return ret
}

// ExpandCIDR 扩展CIDR为IP列表
var ExpandCIDR = expand.CIDR

// ToBytes 将ip或者ip:port转换为[]byte
func ToBytesAddr(input string) []byte {
	if IsIP(input) {
		return net.ParseIP(input).To4()
	} else if IsHostPort(input) {
		ip := net.ParseIP(strings.Split(input, ":")[0]).To4()

		port, _ := strconv.Atoi(strings.Split(input, ":")[1])
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(port))

		return append(ip, portBytes...)
	}
	return nil
}

// ToStringAddr 将[]byte转换为ip或者ip:port
func ToStringAddr(input []byte) string {
	if len(input) == 4 {
		return net.IP(input).String()
	} else if len(input) == 6 {
		return net.IP(input[:4]).String() + ":" +
			strconv.Itoa(int(binary.BigEndian.Uint16(input[4:])))
	}
	return ""
}
