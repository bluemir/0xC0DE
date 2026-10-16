package types

import (
	"strings"

	"github.com/alecthomas/kingpin/v2"
)

// CommaSeparated 는 쉼표로 구분한 flag 값을 target 에 []string 으로 받는 kingpin Value 를 만든다.
//
//	app.Flag("labels", "").Default("a,b").SetValue(types.CommaSeparated(&conf.Labels))
func CommaSeparated(target *[]string) kingpin.Value {
	return (*commaSeparated)(target)
}

type commaSeparated []string

func (c *commaSeparated) Set(value string) error {
	items := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	*c = items
	return nil
}

func (c *commaSeparated) String() string {
	return strings.Join(*c, ",")
}
