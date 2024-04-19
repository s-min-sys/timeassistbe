package timeassist

type ShowItem struct {
	ID       string `yaml:"ID"`
	StartUTC int64  `yaml:"StartUTC"`
	EndUTC   int64  `yaml:"EndUTC"`
}
