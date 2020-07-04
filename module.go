package pulseaudio

import "io"

// Module contains information about a pulseaudio module
type Module struct {
	Index    uint32
	Name     string
	Argument string
	NUsed    uint32
	PropList map[string]string
}

// ReadFrom deserializes a PA module packet
func (s *Module) ReadFrom(r io.Reader) (int64, error) {
	err := bread(r,
		uint32Tag, &s.Index,
		stringTag, &s.Name,
		stringTag, &s.Argument,
		uint32Tag, &s.NUsed,
		&s.PropList)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

// ModuleList queries pulseaudio for a list of loaded modules and returns an array
func (c *Client) ModuleList() ([]Module, error) {
	b, err := c.request(commandGetModuleInfoList)
	if err != nil {
		return nil, err
	}
	var modules []Module
	for b.Len() > 0 {
		var module Module
		err = bread(b, &module)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}
