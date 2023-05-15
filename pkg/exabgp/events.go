package exabgp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gizmoguy/exabgp_exporter/pkg/exabgp/messages"
)

var lastStatus *status

type status struct {
	state  string
	reason string
}

// GetStatus returns the last known status of the exabgp instance
func GetStatus() string {
	if lastStatus == nil {
		lastStatus = &status{
			state:  "unknown",
			reason: "no known last status",
		}
	}
	return lastStatus.state
}

// GetStatusReason returns any reason we may have for the current status
func GetStatusReason() string {
	return lastStatus.reason
}

// Event represents a fully parsed event
type Event struct {
	messages.BaseEvent
	Peer          Peer
	Self          Self
	Direction     string
	announcements *Announcements
	withdrawals   *Withdrawals
	sync.RWMutex
}

// GetVersion returns the exabgp version in an event
func (e *Event) GetVersion() string {
	return e.BaseEvent.Version
}

// GetWithdrawals gets all the withdraw messages in the event
func (e *Event) GetWithdrawals() *Withdrawals {
	e.RLock()
	defer e.RUnlock()
	return e.withdrawals
}

func (e *Event) setWithdrawals(rw *Withdrawals) {
	e.Lock()
	e.withdrawals = rw
	e.Unlock()
}

// GetAnnouncements gets all the announce messages in the event
func (e *Event) GetAnnouncements() *Announcements {
	e.RLock()
	defer e.RUnlock()
	return e.announcements
}

func (e *Event) setAnnouncements(ra *Announcements) {
	e.Lock()
	e.announcements = ra
	e.Unlock()
}

// Peer represents a neighbor and its state
type Peer struct {
	IP     string
	ASN    int
	State  string
	Reason string
}

// Self represents the local bgp instance
type Self struct {
	IP  string
	ASN int
}

// Announcements represents all the bgp `announce` messages
type Announcements struct {
	IPV4Unicast map[string]*IPv4UnicastAnnouncement
	IPV4Flow    map[string]*IPv4FlowAnnouncement
	IPV6Unicast map[string]*IPv6UnicastAnnouncement
	IPV6Flow    map[string]*IPv6FlowAnnouncement
}

// Withdrawals represents all the bgp `withdraw` messages
type Withdrawals struct {
	IPv4Unicast []IPv4UnicastWithdrawal
	IPv4Flow    []IPv4Flow
	IPv6Unicast []IPv6UnicastWithdrawal
	IPv6Flow    []IPv6Flow
}

// IPv4UnicastAnnouncement represents an `ipv4 unicast` family announce
type IPv4UnicastAnnouncement struct {
	Attributes messages.Attribute
	NLRI       []string
}

// IPv4UnicastWithdrawal represents an 'ipv4 unicast' family route withdraw
type IPv4UnicastWithdrawal struct {
	Attributes messages.Attribute
	NLRI       []string
}

// IPv4FlowAnnouncement represents an 'ipv4 flow' flow announce
type IPv4FlowAnnouncement struct {
	Attributes messages.Attribute
	Flows      []IPv4Flow
}

// IPv4Flow represents an 'ipv4 flow'
type IPv4Flow struct {
	Attributes  messages.Attribute
	Destination []string
	Source      []string
	String      string
}

// IPv6UnicastAnnouncement represents an `ipv6 unicast` family announce
type IPv6UnicastAnnouncement struct {
	Attributes messages.Attribute
	NLRI       []string
}

// IPv6UnicastWithdrawal represents an 'ipv6 unicast' family route withdraw
type IPv6UnicastWithdrawal struct {
	Attributes messages.Attribute
	NLRI       []string
}

// IPv6FlowAnnouncement represents an 'ipv6 flow' flow announce
type IPv6FlowAnnouncement struct {
	Attributes messages.Attribute
	Flows      []IPv6Flow
}

// IPv6Flow represents an 'ipv6 flow'
type IPv6Flow struct {
	Attributes  messages.Attribute
	Destination []string
	Source      []string
	String      string
}

// This tries to fix any non-utf8 json generated by exabgp
func safeUnmarshal(data []byte, t interface{}) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		// second pass and give up
		for _, seq := range []string{"\x01", "\x00", "\x04"} {
			data = bytes.Replace(data, []byte(seq), []byte{}, -2) // we don't want to remove the final one since that would merge lines
		}
		err = json.Unmarshal(data, t)
	}
	return err
}

// ParseEvent parses an exabgp json message
func ParseEvent(data []byte) (*Event, error) {
	jsonEvent := &messages.JSONEvent{}

	err := safeUnmarshal(data, jsonEvent)
	if err != nil {
		return nil, err
	}
	event := &Event{
		BaseEvent: jsonEvent.BaseEvent,
		Peer: Peer{
			IP:  jsonEvent.Neighbor.Address.Peer,
			ASN: jsonEvent.Neighbor.ASN.Peer,
		},
		Self:      Self{IP: jsonEvent.Neighbor.Address.Local, ASN: jsonEvent.Neighbor.ASN.Local},
		Direction: jsonEvent.Neighbor.Direction,
	}
	switch jsonEvent.Type {
	case "update":
		ra, rw, err := parseUpdateMessage(jsonEvent.Neighbor.Message.Update)
		if err != nil {
			return event, err
		}
		event.setAnnouncements(ra)
		event.setWithdrawals(rw)
	case "state":
		event.Peer.Reason = jsonEvent.Neighbor.Reason
		event.Peer.State = jsonEvent.Neighbor.State
		lastStatus = &status{
			state:  event.Peer.State,
			reason: event.Peer.Reason,
		}
	case "notification":
		// nothing yet
	case "open":
		// we don't do these right now
	case "keepalive":
		// nothing to do yet
		// increment counter in future?
	case "signal":
		// nothing yet
	default:
		return nil, fmt.Errorf("Cannot handle event type: %s [data: %s]", jsonEvent.Type, data)
	}
	return event, nil
}

func makeAnnouncements() *Announcements {
	return &Announcements{
		IPV4Unicast: make(map[string]*IPv4UnicastAnnouncement),
		IPV4Flow:    make(map[string]*IPv4FlowAnnouncement),
		IPV6Unicast: make(map[string]*IPv6UnicastAnnouncement),
		IPV6Flow:    make(map[string]*IPv6FlowAnnouncement),
	}
}

func parseUpdateMessage(u messages.UpdateMessageFull) (*Announcements, *Withdrawals, error) {
	ra := makeAnnouncements()
	rw := &Withdrawals{}

	// ipv4-unicast announce
	for nexthop, routes := range u.Announce.IPv4Unicast {
		if _, ok := ra.IPV4Unicast[nexthop]; !ok {
			ra.IPV4Unicast[nexthop] = &IPv4UnicastAnnouncement{}
		}
		for _, rs := range routes {
			switch r := rs.(type) {
			case string:
				ra.IPV4Unicast[nexthop].NLRI = append(ra.IPV4Unicast[nexthop].NLRI, r)
			case map[string]interface{}:
				for k, v := range r {
					if _, ok := v.(string); !ok {
						return nil, nil, fmt.Errorf("got a non-string value for %s: %s", k, v)
					}
					ra.IPV4Unicast[nexthop].NLRI = append(ra.IPV4Unicast[nexthop].NLRI, v.(string))
				}
			default:
				return nil, nil, fmt.Errorf("unable to parse route: %+v", rs)
			}
		}
		ra.IPV4Unicast[nexthop].Attributes = u.Attribute
	}

	// ipv4-unicast withdraws
	ws4 := []string{}
	for _, r := range u.Withdraw.IPv4Unicast {
		switch r := r.(type) {
		case string:
			ws4 = append(ws4, r)
		case map[string]interface{}:
			for k, v := range r {
				if _, ok := v.(string); !ok {
					return nil, nil, fmt.Errorf("got a non-string value for %s: %s", k, v)
				}
				ws4 = append(ws4, v.(string))
			}

		default:
			return nil, nil, fmt.Errorf("unable to parse route: %+v", r)
		}
	}
	rw.IPv4Unicast = append(rw.IPv4Unicast, IPv4UnicastWithdrawal{Attributes: u.Attribute, NLRI: ws4})

	// ipv4-flow announce
	for nexthop, flows := range u.Announce.IPv4Flow {
		if _, ok := ra.IPV4Flow[nexthop]; !ok {
			ra.IPV4Flow[nexthop] = &IPv4FlowAnnouncement{}
		}
		for _, flow := range flows {
			f := IPv4Flow{
				Destination: flow.DestinationIPv4,
				Source:      flow.SourceIPv4,
				String:      flow.String,
			}
			ra.IPV4Flow[nexthop].Flows = append(ra.IPV4Flow[nexthop].Flows, f)
		}
		ra.IPV4Flow[nexthop].Attributes = u.Attribute
	}

	for _, flow := range u.Withdraw.IPv4Flow {
		rw.IPv4Flow = append(rw.IPv4Flow, IPv4Flow{Attributes: u.Attribute, Destination: flow.DestinationIPv4, Source: flow.SourceIPv4, String: flow.String})
	}

	// ipv6-unicast announce
	for nexthop, routes := range u.Announce.IPv6Unicast {
		if _, ok := ra.IPV6Unicast[nexthop]; !ok {
			ra.IPV6Unicast[nexthop] = &IPv6UnicastAnnouncement{}
		}
		for _, rs := range routes {
			switch r := rs.(type) {
			case string:
				ra.IPV6Unicast[nexthop].NLRI = append(ra.IPV6Unicast[nexthop].NLRI, r)
			case map[string]interface{}:
				for k, v := range r {
					if _, ok := v.(string); !ok {
						return nil, nil, fmt.Errorf("got a non-string value for %s: %s", k, v)
					}
					ra.IPV6Unicast[nexthop].NLRI = append(ra.IPV6Unicast[nexthop].NLRI, v.(string))
				}
			default:
				return nil, nil, fmt.Errorf("unable to parse route: %+v", rs)
			}
		}
		ra.IPV6Unicast[nexthop].Attributes = u.Attribute
	}

	// ipv6-unicast withdraws
	ws6 := []string{}
	for _, r := range u.Withdraw.IPv6Unicast {
		switch r := r.(type) {
		case string:
			ws6 = append(ws6, r)
		case map[string]interface{}:
			for k, v := range r {
				if _, ok := v.(string); !ok {
					return nil, nil, fmt.Errorf("got a non-string value for %s: %s", k, v)
				}
				ws6 = append(ws6, v.(string))
			}

		default:
			return nil, nil, fmt.Errorf("unable to parse route: %+v", r)
		}
	}
	rw.IPv6Unicast = append(rw.IPv6Unicast, IPv6UnicastWithdrawal{Attributes: u.Attribute, NLRI: ws6})

	// ipv6-flow announce
	for nexthop, flows := range u.Announce.IPv6Flow {
		if _, ok := ra.IPV6Flow[nexthop]; !ok {
			ra.IPV6Flow[nexthop] = &IPv6FlowAnnouncement{}
		}
		for _, flow := range flows {
			f := IPv6Flow{
				Destination: flow.DestinationIPv6,
				Source:      flow.SourceIPv6,
				String:      flow.String,
			}
			ra.IPV6Flow[nexthop].Flows = append(ra.IPV6Flow[nexthop].Flows, f)
		}
		ra.IPV6Flow[nexthop].Attributes = u.Attribute
	}

	for _, flow := range u.Withdraw.IPv6Flow {
		rw.IPv6Flow = append(rw.IPv6Flow, IPv6Flow{Attributes: u.Attribute, Destination: flow.DestinationIPv6, Source: flow.SourceIPv6, String: flow.String})
	}

	return ra, rw, nil
}
