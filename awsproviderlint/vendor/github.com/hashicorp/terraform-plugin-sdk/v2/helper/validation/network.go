package validation

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IsIPAddress is a SchemaValidateFunc which tests if the provided value is of type string and is a single IP (v4 or v6)
func IsIPAddress(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	ip := net.ParseIP(v)
	if ip == nil {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IP, got: %s", k, v))
	}

	return warnings, errors
}

// IsIPv6Address is a SchemaValidateFunc which tests if the provided value is of type string and a valid IPv6 address
func IsIPv6Address(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	ip := net.ParseIP(v)
	if six := ip.To16(); six == nil {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IPv6 address, got: %s", k, v))
	}

	return warnings, errors
}

// IsIPv4Address is a SchemaValidateFunc which tests if the provided value is of type string and a valid IPv4 address
func IsIPv4Address(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	ip := net.ParseIP(v)
	if four := ip.To4(); four == nil {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IPv4 address, got: %s", k, v))
	}

	return warnings, errors
}

// IsIPv4Range is a SchemaValidateFunc which tests if the provided value is of type string, and in valid IP range
func IsIPv4Range(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	ips := strings.Split(v, "-")
	if len(ips) != 2 {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IP range, got: %s", k, v))
		return warnings, errors
	}

	ip1 := net.ParseIP(ips[0])
	ip2 := net.ParseIP(ips[1])
	if ip1 == nil || ip2 == nil || bytes.Compare(ip1, ip2) > 0 {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IP range, got: %s", k, v))
	}

	return warnings, errors
}

// IsCIDR is a SchemaValidateFunc which tests if the provided value is of type string and a valid CIDR
func IsCIDR(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if _, _, err := net.ParseCIDR(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid IPv4 Value, got %v: %v", k, i, err))
	}

	return warnings, errors
}

// IsCIDRNetwork returns a SchemaValidateFunc which tests if the provided value
// is of type string, is in valid Value network notation, and has significant bits between min and max (inclusive)
func IsCIDRNetwork(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		_, ipnet, err := net.ParseCIDR(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("expected %s to contain a valid Value, got: %s with err: %s", k, v, err))
			return warnings, errors
		}

		if ipnet == nil || v != ipnet.String() {
			errors = append(errors, fmt.Errorf("expected %s to contain a valid network Value, expected %s, got %s",
				k, ipnet, v))
		}

		sigbits, _ := ipnet.Mask.Size()
		if sigbits < min || sigbits > max {
			errors = append(errors, fmt.Errorf("expected %q to contain a network Value with between %d and %d significant bits, got: %d", k, min, max, sigbits))
		}

		return warnings, errors
	}
}

// IsMACAddress is a SchemaValidateFunc which tests if the provided value is of type string and a valid MAC address
func IsMACAddress(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := net.ParseMAC(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid MAC address, got %v: %v", k, i, err))
	}

	return warnings, errors
}

// IsPortNumber is a SchemaValidateFunc which tests if the provided value is of type string and a valid TCP Port Number
func IsPortNumber(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be integer", k))
		return warnings, errors
	}

	if 1 > v || v > 65535 {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number, got: %v", k, v))
	}

	return warnings, errors
}

// IsPortNumberOrZero is a SchemaValidateFunc which tests if the provided value is of type string and a valid TCP Port Number or zero
func IsPortNumberOrZero(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be integer", k))
		return warnings, errors
	}

	if 0 > v || v > 65535 {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number or 0, got: %v", k, v))
	}

	return warnings, errors
}
