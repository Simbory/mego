package mego

import (
	"strings"
	"strconv"
	"fmt"
)

type RouteConstraint interface {
	Init(setting string)
	FindMatch(urlPath string) (bool, string)
}

type IntRouteConstraint struct {
	Min int64
	Max int64
}

func (irc *IntRouteConstraint) Init(setting string) error {
	if len(setting) == 0 {
		return
	}
	if strings.Contains(setting, "~") {
		sp := strings.SplitN(setting, "~", 2)
		min,err := strconv.ParseInt(sp[0], 0, 64)
		if err != nil {
			return nil
		}
		max,err := strconv.ParseInt(sp[1], 0, 64)
		if err != nil {
			return nil
		}
		if max < min {
			return fmt.Errorf("Invalid setting string, the min value must be smaller than the max value: %s", setting)
		}
		irc.Min = min
		irc.Max = max
	} else {
		if v,err := strconv.ParseInt(setting, 0, 64); err != nil {
			return err
		} else {
			irc.Max = v
			irc.Min = v
		}
	}
	return nil
}