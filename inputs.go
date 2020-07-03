package pulseaudio

import "fmt"

// Input represents PulseAudio input.
type Input struct {
	client    *Client
	CardID    string
	PortID    string
	CardName  string
	PortName  string
	Available bool
	Direction int
}

// Activate sets this input as the main one.
func (o Input) Activate() error {
	c := o.client
	cards, err := c.Cards()
	if err != nil {
		return err
	}

	if o.CardID == "all" && o.PortID == "none" {
		for _, otherCard := range cards {
			err = c.SetCardProfile(otherCard.Index, "off")
			if err != nil {
				return err
			}
		}
		return nil
	}

	var found bool
	var card Card
	for _, card = range cards {
		if card.Name == o.CardID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("PulseAudio error: card %s is no longer available", o.CardID)
	}

	found = false
	var port port
	for _, port = range card.Ports {
		if port.Name == o.PortID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("PulseAudio error: port %s is no longer available", o.PortID)
	}

	for _, otherCard := range cards {
		if otherCard.Index == card.Index {
			continue
		}
		err = c.SetCardProfile(otherCard.Index, "off")
		if err != nil {
			return err
		}
	}
	bestProfile := port.Profiles[0]
	for _, profile := range port.Profiles {
		if profile.Priority > bestProfile.Priority {
			bestProfile = profile
		}
	}
	err = c.SetCardProfile(card.Index, bestProfile.Name)
	if err != nil {
		return err
	}
	sinks, err := c.sinks()
	if err != nil {
		return err
	}
	s, err := c.ServerInfo()
	if err != nil {
		return err
	}
	for _, sink := range sinks {
		if sink.CardIndex != card.Index {
			continue
		}
		if s.DefaultSink == sink.Name {
			continue
		}
		return c.setDefaultSink(sink.Name)
	}
	return nil
}

// Inputs returns a list of all audio inputs and an index of the active audio input.
//
// The last audio input is always called "None" and indicates that audio is disabled.
func (c *Client) Inputs() (outputs []Input, activeIndex int, err error) {
	s, err := c.ServerInfo()
	if err != nil {
		return nil, 0, err
	}
	sources, err := c.Sources()
	if err != nil {
		return nil, 0, err
	}
	cards, err := c.Cards()
	if err != nil {
		return nil, 0, err
	}

	activeIndex = -1
	for _, card := range cards {
		for _, port := range card.Ports {
			if port.Direction != 2 {
				continue
			}
			for _, source := range sources {
				if source.Name != s.DefaultSource {
					continue
				}
				if source.CardIndex != card.Index {
					continue
				}
				if source.ActivePortName != port.Name {
					continue
				}
				activeIndex = len(outputs)
			}
			outputs = append(outputs, Input{
				client:    c,
				CardID:    card.Name,
				CardName:  card.PropList["device.description"],
				PortID:    port.Name,
				PortName:  port.Description,
				Available: port.Available != 1,
				Direction: int(port.Direction),
			})
		}
	}
	if activeIndex == -1 {
		activeIndex = len(outputs)
	}
	outputs = append(outputs, Input{
		client:    c,
		CardID:    "all",
		CardName:  "All",
		PortID:    "none",
		PortName:  "None",
		Available: false,
		Direction: -1,
	})
	return
}
