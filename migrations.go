package migrations

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Order int
	Up    string
	Down  string
}
type Migrations []*Migration

func (m Migrations) Len() int { return len(m) }
func (m Migrations) Less(i, j int) bool {
	return m[i].Order < m[j].Order
}
func (m Migrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func FromDir(directory string) []*Migration {
	out := make([]*Migration, 0)
	if err := filepath.Walk(directory, BuildMigrations(&out)); err != nil {
		panic(err)
	}
	sort.Sort(Migrations(out))
	return out
}

func BuildMigrations(m *[]*Migration) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ignore nested dirs, why are these here?
		if info.IsDir() {
			return nil
		}
		up, down := getUpAndDownFromFile(path)
		migration := &Migration{
			Order: getOrderFromFilename(info.Name()),
			Up:    up,
			Down:  down,
		}
		*m = append(*m, migration)
		return nil
	}
}

func getOrderFromFilename(name string) int {
	i, err := strconv.Atoi(strings.TrimSuffix(name, ".sql"))
	if err != nil {
		panic(err)
	}
	return i
}

const SplitMarker = "-- SPLIT --"

func getUpAndDownFromFile(name string) (string, string) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(string(b), SplitMarker)
	return parts[0], parts[1]
}
