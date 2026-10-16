package types

import (
	"slices"
	"testing"

	"github.com/alecthomas/kingpin/v2"
)

func TestCommaSeparated(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want []string
	}{
		{args: []string{}, want: []string{"cluster", "namespace"}},
		{args: []string{"--labels=namespace"}, want: []string{"namespace"}},
		{args: []string{"--labels= cluster , ,kind "}, want: []string{"cluster", "kind"}},
		{args: []string{"--labels="}, want: []string{}},
	} {
		labels := []string{}
		app := kingpin.New("test", "")
		app.Flag("labels", "").Default("cluster,namespace").SetValue(CommaSeparated(&labels))

		if _, err := app.Parse(tc.args); err != nil {
			t.Fatalf("%v: parse: %v", tc.args, err)
		}
		if !slices.Equal(labels, tc.want) {
			t.Errorf("%v: labels = %q, want %q", tc.args, labels, tc.want)
		}
	}
}
