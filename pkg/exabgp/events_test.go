package exabgp

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// test messages are take from here (https://github.com/Exa-Networks/exabgp/wiki/Controlling-ExaBGP-:-API-for-received-messages)
// as well as messages seen in the wild that did not parse
var testDataFile = filepath.Join("testdata", "exabgp.log")

var testInvalidDataFile = filepath.Join("testdata", "exabgp.log.1")

func testGetTotalLinesInFile(t *testing.T, f string) int {
	file, err := os.Open(f)
	defer file.Close() // nolint: errcheck

	require.NoError(t, err)

	s := bufio.NewScanner(file)
	totalLines := 0
	for s.Scan() {
		totalLines++
	}
	return totalLines
}

func TestParseTestData(t *testing.T) {
	file, err := os.Open(testDataFile)
	defer file.Close() // nolint: errcheck
	require.NoError(t, err)

	totalLines := testGetTotalLinesInFile(t, testDataFile)

	parsedEvents := 0
	reader := bufio.NewReader(file)
	for {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		_, jerr := ParseEvent(l)
		require.NoError(t, jerr, string(l))
		parsedEvents++
	}
	require.Equal(t, totalLines, parsedEvents)

}

func TestParseInvalidData(t *testing.T) {
	file, err := os.Open(testInvalidDataFile)
	defer file.Close() // nolint: errcheck
	require.NoError(t, err)

	totalLines := testGetTotalLinesInFile(t, testInvalidDataFile)
	parsedEvents := 0
	reader := bufio.NewReader(file)
	for {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		_, jerr := ParseEvent(l)
		require.NoError(t, jerr, string(l))
		parsedEvents++
	}
	require.Equal(t, totalLines, parsedEvents)
}

func TestIPv4Announce(t *testing.T) {
	var testString = `{ "exabgp": "4.0.1", "time": 1554843223.5592246, "host" : "node1", "pid" : 31372, "ppid" : 1, "counter": 11, "type": "update", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "direction": "send", "message": { "update": { "attribute": { "origin": "igp", "med": 100, "local-preference": 100 }, "announce": { "ipv4 unicast": { "192.168.1.184": [ "192.168.88.2/32" ] } } } } } }`
	evt, err := ParseEvent([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "4.0.1", evt.GetVersion())
	announcements := evt.GetAnnouncements()
	require.NotEmpty(t, announcements.IPV4Unicast["192.168.1.184"].Attributes.Med)
	require.NotEmpty(t, announcements.IPV4Unicast["192.168.1.184"].Attributes.Origin)
	require.NotEmpty(t, announcements.IPV4Unicast["192.168.1.184"].Attributes.LocalPreference)
	require.Len(t, announcements.IPV4Unicast, 1)
	require.Len(t, announcements.IPV4Unicast["192.168.1.184"].Routes, 1)
	require.Contains(t, announcements.IPV4Unicast["192.168.1.184"].Routes, "192.168.88.2/32")
}

func TestIPv4AnnounceMulti(t *testing.T) {
	var testString = `{ "exabgp": "4.0.1", "time": 1554987198.3642054, "host" : "node1", "pid" : 14339, "ppid" : 1, "counter": 13, "type": "update", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.158" }, "asn": { "local": 64496, "peer": 64496 } , "direction": "send", "message": { "update": { "attribute": { "origin": "igp", "local-preference": 100 }, "announce": { "ipv4 unicast": { "192.168.1.184": [ "0.0.0.0/0", "0.0.0.0/0", "0.0.0.0/0", "0.0.0.0/0", "192.168.88.0/24" ] } } } } } }`
	evt, err := ParseEvent([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "4.0.1", evt.GetVersion())
	announcements := evt.GetAnnouncements()
	require.NotEmpty(t, announcements.IPV4Unicast["192.168.1.184"].Attributes.Origin)
	require.NotEmpty(t, announcements.IPV4Unicast["192.168.1.184"].Attributes.LocalPreference)
	require.Len(t, announcements.IPV4Unicast, 1)
	require.Len(t, announcements.IPV4Unicast["192.168.1.184"].Routes, 5)
	require.Contains(t, announcements.IPV4Unicast["192.168.1.184"].Routes, "192.168.88.0/24")
}

func TestIPv4AnnounceFlow(t *testing.T) {
	var testString = `{ "exabgp": "4.0.1", "time": 1554987723.0377939, "host" : "node1", "pid" : 14339, "ppid" : 1, "counter": 15, "type": "update", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.158" }, "asn": { "local": 64496, "peer": 64496 } , "direction": "send", "message": { "update": { "attribute": { "origin": "igp", "local-preference": 100, "extended-community": [ { "value": 9225060887780392960, "string": "rate-limit:1" } ] }, "announce": { "ipv4 flow": { "no-nexthop": [ { "destination-ipv4": [ "170.170.170.170/32" ], "source-ipv4": [ "170.170.170.170/32" ], "string": "flow destination-ipv4 170.170.170.170/32 source-ipv4 170.170.170.170/32" } ] } } } } } }`
	evt, err := ParseEvent([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "4.0.1", evt.GetVersion())
	announcements := evt.GetAnnouncements()
	require.NotEmpty(t, announcements.IPV4Flow["no-nexthop"].Attributes.Origin)
	require.NotEmpty(t, announcements.IPV4Flow["no-nexthop"].Attributes.LocalPreference)
	/*
		{
			"ipv4 flow": {
				"no-nexthop": [{
					"destination-ipv4": ["170.170.170.170/32"],
					"source-ipv4": ["170.170.170.170/32"],
					"string": "flow destination-ipv4 170.170.170.170/32 source-ipv4 170.170.170.170/32"
				}]
			}
		}
	*/
	require.NotNil(t, announcements.IPV4Flow["no-nexthop"])
	require.Len(t, announcements.IPV4Flow["no-nexthop"].Flows, 1)
	require.Contains(t, announcements.IPV4Flow["no-nexthop"].Flows[0].Destination, "170.170.170.170/32")
	require.Contains(t, announcements.IPV4Flow["no-nexthop"].Flows[0].Source, "170.170.170.170/32")
	require.Equal(t, "flow destination-ipv4 170.170.170.170/32 source-ipv4 170.170.170.170/32", announcements.IPV4Flow["no-nexthop"].Flows[0].String)
}
func TestIPv4Withdraw(t *testing.T) {
	var testString = `{ "exabgp": "4.0.1", "time": 1554850881.0072424, "host" : "node1", "pid" : 1026, "ppid" : 1, "counter": 6, "type": "update", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "direction": "send", "message": { "update": { "withdraw": { "ipv4 unicast": [ "192.168.88.2/32" ] } } } } }`
	evt, err := ParseEvent([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "4.0.1", evt.GetVersion())
	w := evt.GetWithdrawals()
	require.NotNil(t, w)
	require.Len(t, w.IPv4Unicast, 1)
	require.Contains(t, w.IPv4Unicast[0].Routes, "192.168.88.2/32")
}

func TestIPv4WithdrawMulti(t *testing.T) {
	var testString = `{ "exabgp": "4.0.1", "time": 1554987394.5413187, "host" : "node1", "pid" : 14339, "ppid" : 1, "counter": 14, "type": "update", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.158" }, "asn": { "local": 64496, "peer": 64496 } , "direction": "send", "message": { "update": { "withdraw": { "ipv4 unicast": [ "0.0.0.0/0", "0.0.0.0/0", "0.0.0.0/0", "0.0.0.0/0", "192.168.88.0/24" ] } } } } }`
	evt, err := ParseEvent([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "4.0.1", evt.GetVersion())
	w := evt.GetWithdrawals()
	require.NotNil(t, w)
	require.Len(t, w.IPv4Unicast, 1)
	require.Len(t, w.IPv4Unicast[0].Routes, 5)
	require.Contains(t, w.IPv4Unicast[0].Routes[4], "192.168.88.0/24")
}
func TestPeerState(t *testing.T) {
	tc := map[string]string{
		"up":        `{ "exabgp": "4.0.1", "time": 1554851049.928668, "host" : "node1", "pid" : 8059, "ppid" : 1, "counter": 25, "type": "state", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "state": "up" } }`,
		"down":      `{ "exabgp": "4.0.1", "time": 1554851049.9405053, "host" : "node1", "pid" : 8059, "ppid" : 1, "counter": 26, "type": "state", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "state": "down", "reason": "peer reset, message () error()" } }`,
		"connected": `{ "exabgp": "4.0.1", "time": 1554851063.9655015, "host" : "node1", "pid" : 8059, "ppid" : 1, "counter": 27, "type": "state", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "state": "connected" } }`,
	}
	for name, test := range tc {
		t.Run(name, func(t *testing.T) {
			evt, err := ParseEvent([]byte(test))
			require.NoError(t, err)
			require.Equal(t, evt.Peer.State, GetStatus())
			if evt.Peer.State == "down" {
				require.NotEmpty(t, GetStatusReason())
			}
		})
	}
}

func TestPeerError(t *testing.T) {
	tc := map[string]string{
		"Unsupported Capability":    `{ "exabgp": "4.0.1", "time": 1555087685.177178, "host" : "node1", "pid" : 21053, "ppid" : 1, "counter": 3, "type": "state", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "state": "down", "reason": "peer reset, message (notification received (2,7)) error(OPEN message error / Unsupported Capability / )" } }`,
		"TCP connection was closed": `{ "exabgp": "4.0.1", "time": 1555087685.2088819, "host" : "node1", "pid" : 21053, "ppid" : 1, "counter": 5, "type": "state", "neighbor": { "address": { "local": "192.168.1.184", "peer": "192.168.1.2" }, "asn": { "local": 64496, "peer": 64496 } , "state": "down", "reason": "peer reset, message (closing connection) error(the TCP connection was closed by the remote end)" } }`,
	}
	for name, test := range tc {
		t.Run(name, func(t *testing.T) {
			_, err := ParseEvent([]byte(test))
			require.NoError(t, err)
			require.Equal(t, "down", GetStatus())
			require.Contains(t, GetStatusReason(), name)
		})
	}
}
